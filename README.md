# 泡泡 paopao —— 一个模仿 [rocboss/paopao-ce](https://github.com/rocboss/paopao-ce) 的开源学习项目

> **本项目是模仿 [https://github.com/rocboss/paopao-ce](https://github.com/rocboss/paopao-ce) 写的个人练习项目，不是原创作品。**
>
> 界面布局、页面结构、交互流程、路由与组件划分都参考了 **paopao-ce**（原项目地址：<https://github.com/rocboss/paopao-ce>）；
> 后端接口是我按自己的理解用 Go + Gin 重新实现的，并非原项目的代码。
> 项目里大量注释写的是「作者原版是怎么写的 / 我为什么这样改」，就是为了在学习时对照。
> 如果这个项目对你有帮助，请给原项目点个 Star：**<https://github.com/rocboss/paopao-ce>**

一个「清新文艺的微社区」的最小可用实现：用户注册登录、发帖配图、评论、点赞、关注、关注流、个人主页。

后端 **Go + Gin + GORM + MySQL**，前端 **Vue 3 + Vite + TypeScript + Element Plus + Pinia**。

---

## 目录

- [一、技术栈](#一技术栈)
- [二、目录结构](#二目录结构)
- [三、已实现的业务功能](#三已实现的业务功能)
- [四、API 接口清单](#四api-接口清单)
- [五、数据模型](#五数据模型)
- [六、本地运行](#六本地运行)
- [七、未完成的功能（未来补做）](#七未完成的功能未来补做)
- [八、已知问题与说明](#八已知问题与说明)

---

## 一、技术栈

### 后端

| 用途 | 选型 |
| --- | --- |
| 语言 | Go |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL |
| 鉴权 | JWT（`internal/auth/jwt.go`） |
| 密码 | bcrypt 加密存储 |
| 图片存储 | 本地磁盘（`uploads/YYYY/MM/DD/`），由 `r.Static("/uploads", ...)` 提供静态访问 |

### 前端

| 用途 | 选型 |
| --- | --- |
| 框架 | Vue 3（`<script setup>` 组合式 API） |
| 构建 | Vite |
| 语言 | TypeScript |
| UI 组件库 | Element Plus + `@element-plus/icons-vue` / `@vicons/*` |
| 状态管理 | Pinia |
| 路由 | Vue Router（**hash 模式**） |
| HTTP | axios（统一封装在 `web/src/utils/request.ts`） |
| 样式 | SCSS / CSS 变量 |

> 路由用 hash 模式（URL 形如 `/#/post/1`）而不是 history 模式：history 模式下刷新 `/post/1` 时浏览器会真的向服务器请求这个路径，而 Go 后端没有配 SPA 兜底路由，会直接 404；hash 模式刷新永远只请求 `/`，不需要后端配合。

---

## 二、目录结构

```
paopao/
├── main.go                     # 程序入口：加载配置 → 连库 → 自动建表 → 启动路由
├── go.mod / go.sum
├── .env                        # 数据库与 JWT 配置（不要提交到仓库！）
├── uploads/                    # 用户上传的图片（按年月日分目录）
├── internal/
│   ├── auth/jwt.go             # JWT 签发与解析
│   ├── config/config.go        # 读取 .env
│   ├── db/
│   │   ├── database.go         # 建立 MySQL 连接
│   │   └── model.go            # 全部数据模型 + AutoMigrate
│   ├── middleware/middleware.go# 鉴权中间件（解析 Bearer Token，把 user_id 放进上下文）
│   └── service/
│       ├── router.go           # 全部路由注册
│       ├── service.go          # 注册 / 登录 / 当前用户 + 公共工具函数
│       ├── post.go             # 帖子的增删改查
│       ├── post_image.go       # 图片上传 + 发帖绑图
│       ├── comment.go          # 评论
│       ├── like.go             # 点赞
│       ├── follow.go           # 关注 / 粉丝
│       └── feed.go             # 关注流
└── web/                        # 前端
    ├── vite.config.ts          # 开发代理：/api 与 /uploads 转发到 :8080
    ├── package.json
    └── src/
        ├── api/                # 按业务拆分的接口层
        │   ├── auth.ts         # 注册 / 登录 / 当前用户
        │   ├── post.ts         # 帖子 / 评论 / 点赞
        │   ├── upload.ts       # 图片上传
        │   └── user.ts         # 用户 / 关注
        ├── components/
        │   ├── AppSidebar.vue  # 左侧导航栏（桌面端左栏 + 移动端抽屉复用）
        │   ├── AppRightbar.vue # 右侧栏（搜索 + 版权）
        │   ├── MainNav.vue     # 顶部标题栏
        │   ├── MobileNav.vue   # 移动端底部导航
        │   ├── AuthDialog.vue  # 登录 / 注册弹窗
        │   ├── HomeCompose.vue # 发帖框
        │   ├── PostItem.vue    # 帖子卡片（信息流里的那一条）
        │   ├── PostDetail.vue  # 帖子详情主体
        │   ├── PostImage.vue   # 九宫格图片 + 大图预览
        │   ├── PostLink.vue    # 外链卡片
        │   ├── PostSkeleton.vue# 骨架屏
        │   ├── InfiniteLoadMore.vue # 无限滚动加载
        │   ├── CommentItem.vue # 单条评论
        │   ├── ComposeComment.vue # 评论输入框
        │   └── UserCard.vue    # 用户卡片（含关注按钮）
        ├── views/
        │   ├── HomeView.vue       # 广场（信息流）
        │   ├── PostDetailView.vue # 帖子详情
        │   ├── ProfileView.vue    # 我的主页
        │   ├── UserView.vue       # 他人主页
        │   ├── FollowingView.vue  # 关注 / 粉丝列表
        │   ├── PlaceholderView.vue# 占位页（话题 / 消息 / 收藏 / 好友 / 设置）
        │   └── NotFoundView.vue   # 404
        ├── store/
        │   ├── main.ts         # 主题、布局、登录弹窗等全局 UI 状态
        │   └── user.ts         # 登录用户与 Token
        ├── utils/
        │   ├── request.ts      # axios 实例 + 拦截器 + 统一错误类型
        │   ├── token.ts        # Token / 用户信息的本地持久化
        │   ├── formatTime.ts   # 时间格式化（刚刚 / N 分钟前 / 昨天…）
        │   └── count.ts        # 数字格式化（1.2万）
        ├── types/api.ts        # 全部接口的请求 / 响应类型
        ├── router/index.ts     # 路由表 + 登录守卫
        ├── App.vue             # 三栏布局外壳
        └── main.ts             # 应用入口
```

---

## 三、已实现的业务功能

### 1. 用户系统

| 功能 | 状态 | 说明 |
| --- | --- | --- |
| 注册 | ✅ | 用户名 + 邮箱 + 密码，邮箱唯一；密码 bcrypt 加密 |
| 登录 | ✅ | 邮箱 + 密码，成功返回 JWT |
| 登录态保持 | ✅ | Token 存 localStorage，请求自动带 `Authorization: Bearer <token>` |
| 退出登录 | ✅ | 清 Token 并回广场 |
| 路由登录守卫 | ✅ | `/profile`、`/messages` 等需要登录；未登录时弹登录框，冷启动则退回广场 |
| 401 统一处理 | ✅ | Token 过期时自动清除并弹出登录框（不会把「密码错误」的 401 也当成过期） |

### 2. 帖子

| 功能 | 状态 | 说明 |
| --- | --- | --- |
| 发帖 | ✅ | 正文 + 最多 9 张图；正文第一行自动作为标题（后端 title 必填且限 100 字） |
| 图片上传 | ✅ | multipart 上传（字段名 `image`），校验类型与单张 5MB 上限，按日期分目录落盘 |
| 广场信息流 | ✅ | 分页 + 无限滚动 + 骨架屏 + 空状态 |
| 帖子详情 | ✅ | 正文、图片九宫格、大图预览、评论区 |
| 编辑帖子 | ⚠️ 后端已就绪 | `PUT /posts/:id` 已实现，前端暂未接入入口 |
| 删除帖子 | ✅ | 仅作者本人可删（后端校验归属） |
| 帖子配图展示 | ✅ | 九宫格布局 + 点击放大预览 |

### 3. 互动

| 功能 | 状态 | 说明 |
| --- | --- | --- |
| 评论 | ✅ | 评论列表（分页）、发表评论、删除自己的评论 |
| 点赞 / 取消点赞 | ✅ | 联合唯一索引保证同一人同一帖只能赞一次；前端乐观更新计数 |
| 关注 / 取消关注 | ✅ | 不能关注自己；重复关注返回「您已关注该用户」 |
| 关注流 | ✅ | `GET /feed`：我关注的人 + 我自己的帖子，按时间倒序 |
| 关注 / 粉丝列表 | ✅ | 分页展示，可互相跳转 |

### 4. 个人主页

| 功能 | 状态 | 说明 |
| --- | --- | --- |
| 我的主页 | ✅ | 我的帖子 / 我赞过的 / 我的关注 / 我的粉丝 四个标签页 |
| 他人主页 | ✅ | 展示对方的帖子、关注数、粉丝数，可直接关注 |
| 用户卡片 | ✅ | 列表中悬浮展示用户信息 + 关注按钮 |

### 5. 界面与体验

| 功能 | 状态 | 说明 |
| --- | --- | --- |
| 三栏布局 | ✅ | 左导航 + 中间内容流（最大 620px）+ 右栏 |
| 响应式 | ✅ | 窄屏时左栏收起为抽屉，底部出现移动端导航 |
| 暗色主题 | ✅ | 跟随系统 + 手动切换，选择存 localStorage |
| 骨架屏 / 空状态 | ✅ | 加载中显示骨架，无数据显示「暂无数据」 |
| 无限滚动 | ✅ | 触底自动加载下一页，到底显示「没有更多了」 |

---

## 四、API 接口清单

- 统一前缀：无（开发时通过 Vite 代理 `/api` → `http://127.0.0.1:8080`，代理会去掉 `/api` 前缀）
- 鉴权方式：请求头 `Authorization: Bearer <token>`
- 请求体：除图片上传外均为 `application/json`
- 失败响应统一为 `{"error": "错误信息"}`

### 4.1 认证

| 方法 | 路径 | 鉴权 | 说明 | 成功响应 |
| --- | --- | --- | --- | --- |
| `POST` | `/register` | 否 | 注册 | `200 {"status":"注册成功","id":1,"name":"xx","email":"xx"}` |
| `POST` | `/login` | 否 | 登录 | `200 {"token":"...","user":{"id":1,"name":"xx","email":"xx"}}` |
| `GET` | `/me` | 是 | 当前登录用户 | `200 {"user_id":1,"role":"customer"}` |

请求体：

```jsonc
// POST /register
{ "name": "泡泡", "email": "a@b.com", "password": "123456" }

// POST /login
{ "email": "a@b.com", "password": "123456" }
```

失败情况：

| 场景 | 状态码 | 响应 |
| --- | --- | --- |
| 邮箱已被注册 | 400 | `{"status":"用户已存在"}` |
| 邮箱未注册 | 400 | `{"error":"用户不存在，请先注册"}` |
| 密码错误 | 401 | `{"error":"邮箱或密码错误"}` |
| 请求体格式错误 | 400 | `{"error":"解析失败"}` |

### 4.2 帖子

| 方法 | 路径 | 鉴权 | 说明 | 成功响应 |
| --- | --- | --- | --- | --- |
| `GET` | `/posts` | 否 | 帖子列表（分页） | `200 {"total":10,"page":1,"page_size":10,"list":[...]}` |
| `GET` | `/posts/me` | 是 | 我发布的帖子 | `200 {"total":3,"page":1,"page_size":10,"list":[...]}` |
| `GET` | `/posts/:id` | 否 | 帖子详情（含评论） | `200 { ...帖子详情 }` |
| `POST` | `/posts` | 是 | 发帖 | `201 {"post_id":1}` |
| `PUT` | `/posts/:id` | 是 | 编辑帖子 | `200 {"status":"更新成功"}` |
| `DELETE` | `/posts/:id` | 是 | 删除帖子 | `200 {"status":"删除成功"}` |
| `GET` | `/feed` | 是 | 关注流 | `200 {"page":1,"page_size":10,"total":5,"posts":[...]}` |

分页参数：`?page=1&page_size=10`（默认 `page=1`、`page_size=10`，`page_size` 上限 100）。

列表项结构：

```jsonc
// GET /posts      的 list 元素
{ "id": 1, "title": "标题", "created_at": "...", "author": "泡泡", "comment_count": 2,
  "images": [{ "id": 1, "url": "/uploads/2026/10/13/xxx.png" }] }

// GET /posts/me   的 list 元素（没有 author，作者就是自己）
{ "id": 1, "title": "标题", "created_at": "...", "comment_count": 2, "images": [...] }
```

详情结构：

```jsonc
// GET /posts/:id
{
  "id": 1, "title": "标题", "content": "正文", "created_at": "...", "author": "泡泡",
  "images": [{ "id": 1, "url": "/uploads/..." }],
  "comments": [{ "id": 1, "content": "评论", "author": "某某", "created_at": "..." }]
}
```

发帖请求体：

```jsonc
// POST /posts
{ "title": "标题（必填，最长 100 字）", "content": "正文（必填）", "images": [1, 2, 3] }
// images 是上传接口返回的图片 id，先上传拿到 id，发帖时再绑定
```

失败情况：

| 场景 | 状态码 | 响应 |
| --- | --- | --- |
| 帖子不存在 | 404 | `{"error":"帖子不存在"}` |
| 不是自己的帖子 | 403 | `{"error":"无权操作"}` |
| 标题或正文缺失 | 400 | `{"error":"参数错误：title / content 必填，title 最长 100 字"}` |
| 图片超过 9 张 | 400 | `{"error":"一次最多上传9张图片"}` |

### 4.3 图片

| 方法 | 路径 | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/upload` | 是 | 上传图片，`multipart/form-data`，**文件字段名必须是 `image`**（可重复以一次传多张） |
| `GET` | `/uploads/*` | 否 | 静态图片访问 |

响应：

```jsonc
// POST /upload
{ "images": [{ "id": 1, "url": "/uploads/2026/10/13/9f3c....png" }] }
```

限制与校验：

| 项目 | 规则 |
| --- | --- |
| 文件字段名 | `image`（不是 `file`，字段名错了会返回「没有上传图片」） |
| 单次数量 | 最多 9 张 |
| 单张大小 | 不超过 5 MB |
| 整个请求 | 不超过 50 MB |
| 允许类型 | `.jpg` `.jpeg` `.png` `.gif` `.webp`，并用文件头魔数二次校验 |
| 存储路径 | `uploads/YYYY/MM/DD/<uuid>.<ext>` |

失败情况：`{"error":"解析表单失败（请用 multipart/form-data，文件字段名为 image）"}`、`{"error":"没有上传图片"}`、`{"error":"一次最多上传9张图片"}`、`{"error":"图片[1]超过5MB"}`、`{"error":"不支持的图片类型: .txt"}`。

### 4.4 评论

| 方法 | 路径 | 鉴权 | 说明 | 成功响应 |
| --- | --- | --- | --- | --- |
| `GET` | `/posts/:id/comments` | 否 | 评论列表（分页） | `200 {"total":2,"page":1,"page_size":10,"comments":[...]}` |
| `POST` | `/posts/:id/comments` | 是 | 发表评论 | `201 {"status":"评论成功"}` |
| `DELETE` | `/comments/:id` | 是 | 删除评论 | `200 {"status":"删除成功"}` |

```jsonc
// POST /posts/1/comments
{ "content": "评论内容（必填，最长 512 字）" }

// GET /posts/1/comments 的 comments 元素
{ "id": 1, "content": "评论内容", "author": "某某", "created_at": "..." }
```

失败情况：评论不存在 404 `{"error":"评论不存在"}`；不是自己的评论 403 `{"error":"无权操作"}`。

### 4.5 点赞

| 方法 | 路径 | 鉴权 | 说明 | 成功响应 |
| --- | --- | --- | --- | --- |
| `POST` | `/posts/:id/likes` | 是 | 点赞 | `201 {"status":"点赞成功"}`；已赞过则 `200 {"status":"已点赞"}` |
| `DELETE` | `/posts/:id/likes` | 是 | 取消点赞 | `200 {"status":"已取消点赞"}` |
| `GET` | `/posts/:id/likes` | 否 | 某帖的点赞数 | `200 {"id":1,"title":"标题","like_count":5}` |
| `GET` | `/posts/me/likes` | 是 | 我赞过的帖子 | `200 [{"author":"泡泡","title":"标题","id":1}]` |

### 4.6 关注

| 方法 | 路径 | 鉴权 | 说明 | 成功响应 |
| --- | --- | --- | --- | --- |
| `POST` | `/users/:id/follow` | 是 | 关注某人 | `201 {"status":"关注成功"}`；已关注则 `200 {"status":"您已关注该用户"}` |
| `DELETE` | `/users/:id/follow` | 是 | 取消关注 | `200 {"status":"成功取消关注"}` |
| `GET` | `/users/:id/following` | 否 | 某人关注了谁 | `200 {"user":"泡泡","total":3,"page":1,"page_size":10,"follows":[...]}` |
| `GET` | `/users/:id/followers` | 否 | 某人的粉丝 | `200 {"user":"泡泡","total":3,"page":1,"page_size":10,"fans":[...]}` |

失败情况：关注自己 400 `{"error":"您无法关注自己"}`；用户不存在 404 `{"error":"用户不存在"}`。

---

## 五、数据模型

| 表 | 说明 | 主要字段 |
| --- | --- | --- |
| `users` | 用户 | `id`、`name`、`email`（唯一）、`password`（bcrypt，JSON 不输出）、`role` |
| `posts` | 帖子 | `id`、`user_id`、`title`(varchar 100)、`content`(text) |
| `comments` | 评论 | `id`、`user_id`、`post_id`、`content`(varchar 512) |
| `post_likes` | 点赞 | `id`、`post_id` + `user_id`（**联合唯一索引**，防止重复点赞） |
| `follows` | 关注关系 | `id`、`user_id`（谁）、`follow_id`（关注谁），联合唯一索引 |
| `post_images` | 帖子图片 | `id`、`post_id`（可空）、`user_id`、`url`、`file_name`、`sort`、`status`（0=临时未绑定，1=已绑定帖子） |

表结构由 GORM 的 `AutoMigrate` 在启动时自动创建，**不需要手动执行建表 SQL**。

---

## 六、本地运行

### 1. 准备数据库

```sql
CREATE DATABASE paopao DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
```

### 2. 配置后端

在项目根目录创建 `.env`（**不要提交到 Git**）：

```dotenv
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=你的数据库密码
DB_NAME=paopao
JWT_SECRET=换成你自己的随机密钥
```

### 3. 启动后端

```bash
cd paopao
go mod download
go run main.go
# 监听 http://127.0.0.1:8080
```

### 4. 启动前端

```bash
cd web
npm install
npm run dev
# 打开 http://localhost:5173
```

前端在开发环境下通过 Vite 代理访问后端（`web/vite.config.ts`）：

- `/api/*` → `http://127.0.0.1:8080/*`（去掉 `/api` 前缀）
- `/uploads/*` → `http://127.0.0.1:8080/uploads/*`

这样浏览器始终同源，不需要后端配置 CORS。

### 5. 其他可用命令

```bash
npm run type-check   # vue-tsc 类型检查
npm run build        # 类型检查 + 生产构建
npm run preview      # 预览构建产物
```

---

## 七、未完成的功能（未来补做）

> 下面这些功能**目前没有实现**，列出来作为后续计划。

### 内容与发现

- [ ] **话题 / 标签体系**（`#话题`）：目前左侧「话题」是占位页，帖子里也不解析 `#tag`
- [ ] **搜索**：按关键词搜帖子、搜用户、搜话题（右栏搜索框目前只是 UI）
- [ ] **热门推荐排序**：广场的「热门推荐」筛选现在与「全部」结果相同，因为后端只支持按 id 倒序；需要按点赞 / 评论 / 时间加权算热度
- [ ] **帖子置顶 / 加精**
- [ ] **外链卡片解析**：`PostLink.vue` 已有组件，后端没有链接元信息抓取

### 社交

- [ ] **消息通知系统**：评论 / 点赞 / 关注的站内通知、未读数红点（左侧「消息」是占位页）
- [ ] **好友关系**：互相关注成为好友、好友可见的帖子（左侧「好友」是占位页）
- [ ] **私信聊天**
- [ ] **收藏**：收藏帖子、收藏夹（左侧「收藏」是占位页）
- [ ] **评论点赞**
- [ ] **评论回复（楼中楼）**、评论 @提及
- [ ] **举报 / 拉黑**

### 用户

- [ ] **用户头像**：后端 `User` 表还没有 `avatar` 字段，目前用昵称首字母占位
- [ ] **个人设置页**：改昵称、改头像、改密码、绑定手机（左侧「设置」是占位页）
- [ ] **钱包 / 积分 / 付费内容**
- [ ] **用户主页的关注按钮状态**：`is_following` 需要后端在用户信息里一并返回

### 媒体与内容形态

- [ ] **视频**：发视频、视频播放器
- [ ] **附件**：上传并下载附件、附件定价
- [ ] **音频**

### 工程与管理

- [ ] **管理后台**：站点配置、用户封禁、内容审核
- [ ] **软删除 + 回收站**：目前是物理删除
- [ ] **点赞数 / 评论数冗余字段**：现在每次都 `COUNT(*)`，数据量上来后需要冗余计数 + 缓存
- [ ] **统一分页响应结构**：现在列表接口有的返回 `list`、有的返回 `posts`、有的返回 `comments`、有的直接返回数组，应该统一
- [ ] **IP 归属地**
- [ ] **关注流缓存**：关注人数多时 `IN (关注列表)` 会变慢，需要推模式或缓存时间线
- [ ] **单元测试 / 接口测试**
- [ ] **Docker Compose 一键启动**
- [ ] **CI（构建 + 类型检查）**

---

## 八、已知问题与说明

1. **5 个占位页**：左侧导航里的「话题」「消息」「收藏」「好友」「设置」目前都指向同一个 `PlaceholderView.vue`，点进去只有一句提示。它们保留在导航里是为了和原项目的布局一致。

2. **列表接口不返回作者 id**：`GET /posts` 的列表项只有作者昵称（`author`），没有 `user_id`。这导致信息流里无法判断「这条是不是我发的」（所以不显示删除按钮），也无法在卡片上直接关注作者。补上 `user_id` 即可解决。

3. **点赞状态靠前端记录**：列表接口不返回「我是否赞过这条」，前端用 localStorage 记录已赞的帖子 id。换设备或清缓存后，已赞过的帖子会显示为未赞状态（但再点一次后端会返回「已点赞」，不会重复计数）。

4. **发帖标题是自动生成的**：原项目发布框没有标题输入，而本项目的后端 `title` 是必填字段，所以取正文第一行截断到 100 字作为标题。

5. **「热门推荐」暂无效果**：见「未完成」一节的说明。

6. **图片存在本地磁盘**：`uploads/` 目录目前是**已经提交进仓库**的（含开发时留下的测试图片）。生产环境应换成对象存储；如果不想让测试图片进仓库，可以执行 `git rm -r --cached uploads`，再把 `uploads/` 加进 `.gitignore`（目录本身保留，只是不再跟踪）。

7. **`.env` 里是真实配置**：如果你的仓库里已经提交过 `.env`，请务必执行 `git rm --cached .env` 把它从版本库里移除 —— `.gitignore` 只对**尚未跟踪**的文件生效，对已提交过的文件没用。可以参考根目录的 `.env.example` 新建自己的配置。

---

## 致谢

- 原项目：**[rocboss/paopao-ce](https://github.com/rocboss/paopao-ce)** —— 本项目的外观与交互设计全部参考自它
- 前端组件库：**[Element Plus](https://element-plus.org/)**
- 图标：**[@vicons](https://github.com/07akioni/xicons)**

本项目仅用于个人学习，如涉及版权问题请联系删除。
