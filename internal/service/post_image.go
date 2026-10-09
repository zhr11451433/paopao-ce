package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"paopao/internal/db"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// 图片上传  POST /upload
// ----------------------------------------------------------------------------
// 设计：文件存磁盘，数据库只存"路径字符串"（不是图片二进制！）
//   ① POST /upload  (multipart，真正传文件的字节)
//        → {"images":[{"id":1,"url":"/uploads/2026/02/18/xxx.jpg"}]}
//   ② 前端拿到 url，立刻能显示缩略图预览
//   ③ POST /posts  (JSON，带上图片 id)
//        {"title":"...","content":"...","images":[1,2]}
//      服务端在【事务】里：创建帖子 + 把这些图片的 post_id 绑定上去
//
// 【两条铁律，务必记住】
//   1. 文件内容必须由客户端"上传字节"到服务端，由服务端落盘。
//   2. url 必须【服务端自己生成】，绝不能采信客户端传来的 url。
//      否则客户端能填任意地址：别人的图、外站地址、恶意链接。
// ============================================================================

type PostImageHandler struct {
	db *gorm.DB
}

func NewPostImageHandler(db *gorm.DB) *PostImageHandler {
	return &PostImageHandler{db: db}
}

// 扩展名白名单：挡掉 .php / .jsp / .exe 这类
var allowedExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

const maxSize int64 = 5 * 1024 * 1024 // 单张最大 5MB（1024*1024 = 1MB）

// ✅ 你已修正：原来是 32 << 20，现在 50 << 20
// 【为什么必须改】单张 5MB × 最多 9 张 = 45MB。如果总上限只有 32MB，
//
//	"合法地传 9 张 5MB 的图"会在解析表单阶段就被截断，报的还是
//	500 "解析表单失败" —— 错误信息完全误导人。现在 50MB > 45MB，两个限制自洽了。
//	`<< 20` = 左移 20 位 = 乘以 2^20 = 乘以 1048576，就是 MB。
const reqMaxSize int64 = 50 << 20 // 整个请求最大 50MB（9 × 5MB = 45MB < 50MB ✅）

// 【新增】上传根目录集中定义一处。
// 【为什么要提这个】原代码里写死了相对路径 "uploads/..."，它是相对于
//
//	【进程的工作目录】的（router.go 里的 r.Static 同理）。
//	这和你最早踩的 config.yaml 找不 到 是 同一类问题：换个目录启动 go run，
//	文件就存到别处、静态服务也找不到。
//	集中成一个常量，以后想改成从 .env / config 读，只改这一行。
const uploadBaseDir = "uploads"

// SafeName 只保留白名单扩展名，主名用 UUID
func SafeName(original string) (string, error) {
	ext := strings.ToLower(filepath.Ext(original)) // ".JPG" → ".jpg"，防大小写绕过白名单
	if !allowedExt[ext] {
		return "", errors.New("不支持的图片类型: " + ext)
	}
	// ① UUID 全局唯一 → 不会同名覆盖
	// ② 原始文件名的主体部分全部丢弃 → 防路径穿越（如 "../../etc/passwd"）
	// ③ 只保留扩展名，而它来自白名单，所以是安全的
	return uuid.NewString() + ext, nil
}

// checkImageContent 读文件开头 512 字节，用 magic bytes 判断"真实类型"。
// 【为什么需要】后缀名能随便改，把木马命名成 a.jpg 一样传得上来；
//
//	但文件开头的字节特征改不了：JPEG 开头是 FF D8 FF，PNG 是 89 50 4E 47……
//	这就是"防改后缀攻击"。
//
// 【为什么抽成独立函数】见下面「 关于 defer」那段说明 —— 核心原因是：
//
//	函数级的 defer 是正确用法，循环体里的 defer 才是坏习惯。
func checkImageContent(fh *multipart.FileHeader) error {
	// fh 只是"文件元信息"（Filename / Size / Header），不含内容。
	// 想知道里面到底是什么，必须 Open() 拿到一个读句柄。
	file, err := fh.Open()
	if err != nil {
		return errors.New("打开上传文件失败")
	}
	defer file.Close() // ✅ 这里 defer 是函数级，本次调用结束就关，完全正确

	buf := make([]byte, 512)
	// 为什么是 512？因为 http.DetectContentType 只需要前 512 字节就能识别类型。
	n, err := file.Read(buf)
	//n 本次 Read 调用实际读到的字节数量
	// ✅ 你已修正：原来是 io.ReadFull(file, buf) 且只放行 io.EOF
	//   原代码： _, err = io.ReadFull(file, buf)
	//           if err != nil && err != io.EOF { 返回 500 "读取图片失败" }
	// 【原代码的问题】io.ReadFull 的语义是"必须读满 512 字节"，它的错误分两种：
	//     · 一个字节都没读到        → io.EOF
	//     · 读到 100 字节就到文件尾 → io.ErrUnexpectedEOF（注意：不是 io.EOF！）
	//   原代码只放行了 io.EOF，于是【所有小于 512 字节的图片】都会走进报错分支，
	//   返回 500。小图标、小 GIF、1x1 像素图都会中招。
	// ✅ 你现在改成 file.Read：单次读，读多少算多少，配合下面的 buf[:n] 用，逻辑对了。
	// 【一个更严格的小提示】严格来说 Read 不保证"一次读满"，它可能只返回一部分。
	//   对本地文件和内存 Reader 来说基本不会发生，所以你现在这样写没问题；
	//   想 100% 稳的话，用 ReadFull 并同时放行 io.EOF 和 io.ErrUnexpectedEOF。
	if err != nil && err != io.EOF {
		return errors.New("读取图片失败")
	}

	contentType := http.DetectContentType(buf[:n])
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("不是有效图片（实际检测到 %s）", contentType)
	}
	return nil
}

func (h *PostImageHandler) Upload(c *gin.Context) {
	// ---------------- 0. 取当前用户 ----------------
	// user_id 是 JWT 中间件验签成功后塞进 context 的（c.Set("user_id", ...)）。
	// 图片要记归属（UserID），将来发帖绑定时才能校验"这是不是我的图"。
	userID, ok := GetUserID(c)
	if !ok {
		return // GetUserID 内部已经写过响应了，直接返回
	}

	// ---------------- 1. 给整个请求体设上限（第一道闸门）----------------
	// MaxBytesReader 把 Body 包一层：超过 reqMaxSize 就中断读取并报错。
	// 【为什么要在解析之前做】防止有人塞一个 10GB 的请求把内存/磁盘打爆。
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, reqMaxSize)

	// ---------------- 2. 解析 multipart 表单 ----------------
	// multipart/form-data 是 HTML 表单上传文件的编码格式：
	//     普通字段 → form.Value["key"]
	//     文件字段 → form.File["key"]（类型 []*multipart.FileHeader，一个字段可带多个文件）
	// gin 会先把内容缓存在内存（≤ MaxMultipartMemory = 8MB），超出就落到系统临时文件。
	form, err := c.MultipartForm()
	if err != nil {
		// ❌ 原代码：c.JSON(http.StatusInternalServerError, gin.H{"error": "解析表单失败"})
		// 【问题】表单格式不对是【客户端】的错。返回 500 会让前端以为服务端挂了，
		//   也不利于排查问题。记住这条分界线：4xx = "你传错了"，5xx = "我这边挂了"。
		// ✅ 改成 400，并把用法提示写进错误信息，方便调试。
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "解析表单失败（请用 multipart/form-data，文件字段名为 image）",
		})
		return
	}
	fhs := form.File["image"]
	// 取名为 "image" 的文件字段。前端要多传几张，就重复使用同一个字段名。
	if len(fhs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有上传图片"})
		return
	}
	if len(fhs) > 9 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "一次最多上传9张图片"})
		return
	}

	// ---------------- 3. "失败就清理"的兜底（新增）----------------
	// 【解决什么问题】原代码是"边校验边保存"：假如第 3 张校验失败直接 return，
	//   前 2 张其实已经写进磁盘了，但后面的批量入库永远不会执行
	//   → 磁盘上留下"孤儿文件"（磁盘有文件、数据库没记录，谁也不知道它属于谁）。
	// 【怎么解决】把"已经写成功的文件路径"记下来；只要函数没走到最后一步，
	//   就靠 defer 把这些文件统统删掉。
	//   这是很常用的"补偿 / 回滚"思路：没有数据库事务保护的磁盘操作，得自己收尾。
	written := make([]string, 0, len(fhs))
	success := false
	defer func() {
		if !success {
			for _, w := range written {
				_ = os.Remove(w) // 尽力删除；删不掉也不影响已经发出去的错误响应
			}
		}
	}()

	images := make([]db.PostImage, 0, len(fhs))
	// ✅ 你已修正：原来是 make([]db.PostImage, len(fhs))，这个对照务必记牢 ——
	//     make([]T, n)     → 长度就是 n，里面装着 n 个【零值元素】
	//     make([]T, 0, n)  → 长度 0（空切片），只是【预留】了 n 个位置（容量）
	//   用第一种再配合 append，最终长度会变成 2n：前 n 个是零值
	//   （Url=""、FileName=""、UserID=0），后 n 个才是真数据，
	//   然后 db.Create 会把这 2n 条【全部插进数据库】= 一条条垃圾行。
	//   （你文件末尾的 resp 就是写对的 make([]gin.H, 0, len(images))，正好对照着记。）

	for sortIdx, fh := range fhs {
		if fh.Size > maxSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("图片[%d]超过5MB", sortIdx+1)})
			return
		}
		// fh.Size 是 gin 解析表单时就算好的大小，不用读文件就能判断，很便宜。

		// ❌ 原代码（整块被下面这行调用替代）：
		//
		//   file, err := fh.Open()
		//   if err != nil { 返回 500 "打开上传文件失败" }
		//   defer file.Close()                    // ← 问题 1：循环里 defer
		//   buf := make([]byte, 512)
		//   n, err := file.Read(buf)
		//   if err != nil && err != io.EOF { 返回 500 "读取图片失败" }
		//   contentType := http.DetectContentType(buf[:n])
		//   if !strings.HasPrefix(contentType, "image/") { 返回 400 "不是有效图片" }
		//   _, err = file.Seek(0, io.SeekStart)   // ← 问题 2：这句是多余的（见文末）
		//   if err != nil { 返回 500 "文件指针重置失败" }
		//
		// 【问题 1 详解：循环里不要写 defer】
		//   defer 的执行时机是"【函数】返回时"，不是"循环这一轮结束时"。
		//   所以上面那句 defer file.Close() 会把 9 个文件句柄一直拖到整个 Upload
		//   函数结束（包括后面入库的时间）才关。这叫 defer-in-loop，是很常见的坏习惯。
		// ✅ 修法：把这一整块抽成 checkImageContent 函数。这样一来 ——
		//   函数级的 defer 是【正确】用法（一次调用只开 1 个句柄，用完就关），
		//   而循环体里只剩一行调用，既干净也没有句柄泄漏风险。
		//   【记一句话】函数级 defer = 好习惯；循环里 defer = 坏习惯。
		if err := checkImageContent(fh); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("图片[%d]%s", sortIdx+1, err.Error())})
			return
		}

		// 生成 uuid 文件名 + 校验后缀
		newName, err := SafeName(fh.Filename)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 按日期生成目录：uploads/2026/02/18/
		now := time.Now()
		dateDir := fmt.Sprintf("%s/%04d/%02d/%02d", uploadBaseDir, now.Year(), now.Month(), now.Day())
		// %04d = 至少 4 位、不足补 0 → 2026；%02d = 至少 2 位 → 02 / 18
		// 【为什么按日期分目录】单个目录塞几十万文件会让文件系统变慢；
		//   按年 / 月 / 日分层，每个目录的文件数就控制在合理范围。
		// （这里用 uploadBaseDir 常量替代原来写死的 "uploads" 字符串。）

		// MkdirAll = "父目录不存在也一起建"（相当于 mkdir -p），已存在不报错。
		// 0755 是权限位（Linux 概念：属主可读写执行、其他人可读可执行）；
		//   Windows 上基本被忽略，写上是为了跨平台。
		if err := os.MkdirAll(dateDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建存储目录失败"})
			return
		}

		// 磁盘路径用 filepath.Join：它会按【操作系统】选分隔符（Windows 用 \），
		// 并清理多余的 / 和重复分隔符。写磁盘就该用它。
		savePath := filepath.Join(dateDir, newName)

		if err := c.SaveUploadedFile(fh, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存图片失败"})
			return
		}
		// 【c.SaveUploadedFile 内部做了什么】fh.Open() → os.Create(目标路径) → io.Copy。
		//   也就是"把上传的临时文件内容写到磁盘指定位置"，这是 gin 提供的便捷方法。
		// 【关键】它接收的是 fh（FileHeader），会【自己重新 Open 一个新句柄】 ——
		//   这正是原代码里那句 file.Seek(0, ...) 变成无用代码的原因（见文末说明）。
		written = append(written, savePath) // 记下来，万一后面失败好清理

		// ⚠️【原代码在 Windows 上有个坑，这里修掉了】
		//   原代码：imgUrl := "/" + filepath.Join(dateDir, newName)
		// 【问题】filepath.Join 在 Windows 上返回的是【反斜杠】：
		//       uploads\2026\02\18\xxx.jpg
		//   拼出来的 url 就变成 /uploads\2026\02\18\xxx.jpg ——
		//   url 里出现反斜杠是非法的，浏览器访问会出问题。
		//   你正好在 Windows 上开发，这个坑一定会踩到。
		// ✅ 修法：url 是"网络路径"，永远用正斜杠，所以要用 path.Join
		//   （注意是 path 包，不是 filepath 包）。
		// 【记住这对区别】filepath = 本地文件系统路径；path = url / 网络路径。
		//   上面 savePath 用 filepath.Join、这里 imgUrl 用 path.Join ——
		//   一个 filepath 一个 path 不是笔误，而是各自场景不同。
		imgUrl := "/" + path.Join(dateDir, newName)

		postImg := db.PostImage{
			PostID:   nil,     // 还没绑定帖子（*uint 类型，用 nil 表达"空"）
			UserID:   userID,  // 谁传的 —— 发帖绑定时用来校验归属
			Url:      imgUrl,  // 服务端生成的 url，不是客户端给的
			FileName: newName, // 磁盘上的真实文件名
			Sort:     sortIdx, // 第几张，保留上传顺序
			Status:   0,       // 0 = 临时未绑定（注：和 PostID==nil 冗余，见文末）
		}
		images = append(images, postImg)
	}

	// ---------------- 4. 批量入库 ----------------
	if err := h.db.Create(&images).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入图片数据库失败"})
		return
		// 注意：这里 return 之后，上面 defer 的清理会把已经落盘的图片删掉，
		//   所以不会留下"磁盘有图、数据库没记录"的孤儿文件。
	}
	// GORM 传切片 = 一条 SQL 批量 INSERT，比循环里单条插入快得多。
	// 【重要副作用】GORM 会把数据库生成的自增 ID 【回填】到切片的每个元素上，
	//   所以下面可以直接读到 img.ID —— 这是 GORM 很方便的一点。

	// ---------------- 5. 回填 ID 并返回 ----------------
	resp := make([]gin.H, 0, len(images))
	// ✅ 这里就是正确写法 make([]T, 0, n)，和上面那个 ❌ 正好是一对，对照着记。
	for _, img := range images {
		resp = append(resp, gin.H{
			"id":  img.ID,  // 发帖时用它绑定（比用 url 安全：可以校验归属）
			"url": img.Url, // 前端立刻拿去预览
		})
	}

	success = true // 走到这里才算全部成功，defer 里的清理才不会执行
	c.JSON(http.StatusOK, gin.H{"images": resp})
}

// ============================================================================
// 【文末说明 1】原代码里的 file.Seek(0, io.SeekStart) 为什么删掉了
//
//   原代码流程：fh.Open() → 读 512 字节做 magic 校验 → Seek(0) 回到开头
//              → c.SaveUploadedFile(fh, savePath) 保存
//
//   【问题】c.SaveUploadedFile 内部会【自己重新 Open 一次】文件（用一个全新的句柄），
//     跟你前面那个句柄毫无关系。所以你对旧句柄做的 Seek(0)，对保存过程
//     【没有任何影响】—— 它是一句无用代码（而且 file 句柄已经交给 checkImageContent
//     在函数内关掉了，那里也根本不需要 Seek）。
//
//   【三种写法，选一种，别混】
//     写法 A（本次采用）：读校验和落盘各自打开自己的句柄，各关各的。
//            最不容易出错，代价是多开一次文件。
//     写法 B：复用同一个句柄读校验 + 落盘 —— 这时 Seek(0) 是【必须的】！
//            否则 io.Copy 会把"已经读过 512 字节之后的剩余内容"写出去，
//            保存下来的文件会少 512 字节，图片直接损坏。
//     写法 C：干脆不做 magic 校验，直接用 c.SaveUploadedFile。
//            最省事，但会失去"防改后缀攻击"的能力，不推荐。
//
//   【一句话总结】Seek 只在"同一个句柄既读又写"的时候才需要。
//

// ----------------------------------------------------------------------------
// 【文末说明 2】Status 字段和 PostID 是冗余的
//
//   PostImage.Status（0=临时 / 1=已绑定）所表达的意思，其实 PostID == nil
//   已经完全表达了：没绑定就是 nil，绑定了就有值。
//   同一件事用两个字段表示，早晚会不一致（比如绑定了 PostID 却忘了改 Status，
//   或者反过来）。→ 建议二选一：删掉 Status，或者保留它但别再依赖 PostID==nil 判断。
//

// ----------------------------------------------------------------------------
// 【文末说明 3】下一步：把图片绑到帖子上
//
//   POST /posts 的请求体加 images 字段（建议用【图片 id】而不是 url，因为 id 能校验归属）：
//       {"title":"...","content":"...","images":[1,2]}
//   然后在一个【事务】里做两件事：
//     ① 查出这些 PostImage，校验 UserID == 当前用户（不能绑别人的图）
//        且 PostID == nil（不能重复绑定）
//     ② 创建 Post，再把这些图片的 PostID 更新成新帖 id
//   【为什么必须用事务】帖子建成功但图片绑定失败 → 帖子没图；
//     反过来 → 图片挂在一个不存在的帖子上。要么全成，要么全不成。

type ReqPost struct {
	Title   string `json:"title" binding:"required,max=100"`
	Content string `json:"content" binding:"required"`
	Images  []uint `json:"images"`
}

func (h *PostImageHandler) Posts(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	var reqPost ReqPost
	if err := c.ShouldBindJSON(&reqPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：title / content 必填，title 最长 100 字"})
		return
	}
	if len(reqPost.Images) > 9 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "一次最多上传9张图片"})
		return
	}
	var postID uint
	err := h.db.Transaction(func(tx *gorm.DB) error {
		post := db.Post{
			UserID:  userID,
			Title:   reqPost.Title,
			Content: reqPost.Content,
		}
		if err := tx.Create(&post).Error; err != nil {
			return err
		}
		postID = post.ID
		// 没有图片直接返回成功
		if len(reqPost.Images) == 0 {
			return nil
		}
		// 2.批量更新：条件：图片ID在列表、属于当前用户、post_id尚未绑定(null)
		result := tx.Model(&db.PostImage{}).
			Where("id IN ? AND user_id = ? AND post_id IS NULL", reqPost.Images, userID).
			Updates(map[string]any{
				"post_id": postID,
				"status":  1,
			})
		if result.Error != nil {
			return result.Error
		}
		// 校验：实际修改行数必须等于传入图片数量，否则说明有图片：不存在/不是你的/已经绑定别的帖子
		if result.RowsAffected != int64(len(reqPost.Images)) {
			return errors.New("部分图片无效：图片不存在、不属于你或已经绑定其他帖子")
		}
		return nil
	})
	if err != nil {
		// 区分业务错误(400)和数据库底层错误(500)
		// 判断是不是业务逻辑报错
		if err.Error() == "部分图片无效：图片不存在、不属于你或已经绑定其他帖子" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建帖子失败"})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"post_id": postID,
	})
}
