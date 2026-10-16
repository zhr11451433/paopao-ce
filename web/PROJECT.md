# 泡泡社区 · 前端复刻项目说明书

> 本项目是 [paopao-ce](https://github.com/rocboss/paopao-ce) 的**前端复刻**：`F:\paopao\web`（Vue 3 + TS + Element Plus）。
> 这是给 AI 助手的**唯一作业依据**——铁律、后端契约、任务队列、当前状态全在这里。
>
> **【AI 请注意】每次开工前完整读完本文件再动手，不要凭记忆、不要凭上一轮印象。**
> **【我（人类）请注意】AI 每轮自己改 §11 的状态；我只要回一个字。**

> **【当前状态 · 第 40 轮】**
> T1–T18 已验收。S 档停用。C1–C12 已验收。**C13 已交付、等验收**（`GET /posts?style=hots` 按点赞数排序）。下一件是 C14，要等 C13 标 ✅ 才能做。C14 只改 `HomeView.vue`。
>
> **§11 的节号沿用旧版**：11.2 / 11.3 / 11.10 已大幅压缩，节号保留是为了不破坏 §11.4 / §12 这些交叉引用。

---

## 0. 铁律（最高优先级，任何情况不得违反）

### 0.1 功能照作者，外观用我现在这套

作者源码只用来回答两件事：**这个功能做什么、点下去发生什么**。
外观**不要再照抄**作者的 Naive UI / less。那套我看着不舒服。

新页面、新组件向本项目**已经验收的页面**看齐：

- 布局和主题：`src\assets\css\main.scss`（三栏尺寸、主色、暗色）。三栏的 `position: fixed` **不要改成 flex**，这是已经定下来的壳，不是作者审美。
- 卡片和列表怎么排：看 `PostItem.vue`、`MainNav.vue`、`HomeView.vue`、`ProfileView.vue`。
- 组件库用 Element Plus。第 6 节只负责「作者的组件换成哪个 EP 组件」，**不负责**把间距、阴影、密度量成作者那样。

已经验收的页面不要为了「统一成新风格」再改一遍。

作者有、我没有的交互要补上。作者没有的交互不要自己加。

### 0.2 不要擅自改动我没有明确要求的文件

尤其**不要「顺手优化」已经跑通的代码**。
每轮开始前我会写清「本轮只做」和「本轮不做」，请严格守住边界。

### 0.3 接口形状以「我的后端」为准，不抄作者的 `/v1`

**作者的后端和我的后端不是同一个。** 字段名、路径、错误格式、返回结构都按我现有的 gin 写法来，不抄作者。

从第 26 轮起，**缺接口由 AI 直接改 `F:\paopao` 的 Go**（见 §0.9），不再只出配方、不动文件。
新接口必须长得像我已经有的那些：没有 `{code,data,msg}` 信封，列表用顶层 `total`，路径用 `/posts`、`/users/:id` 这种，不要写成作者的 `/v1/post`。

- 作者的 `Tweet` 类型 → 我的 `PostListItem`
- 作者的 `{code, data, msg}` 信封 → 我的**没有信封**
- 作者的 `rsp.pager.total_rows` → 我的**顶层 `total`**

不确定时查第 7 节；第 7 节也没写的，**直接问我**，不要猜。

### 0.4 每次只做一件事

**一次只做 C 档里的一行。** 一行是一个后端改动，或一个前端组件 / 一个页面。
禁止一轮里交付多个组件，也禁止同一轮里后端和前端一起交。

### 0.9 ★ 先写后端，再写前端（第 26 轮起）

从 §11.4 **C 档**取第一个不是 ✅ 的任务。

| 这一行是 | 这一轮做什么 | 这一轮不做什么 |
|---|---|---|
| **后端** | 只改 `F:\paopao` 里这一件事要用的 Go；在 `src\types\api.ts` 和 `src\api\` 补上类型和函数 | 不写 `.vue`，不顺手改下一个接口 |
| **前端** | 只改这一行点名的那一个组件或页面，接上一行已经写好的接口 | 不改 Go。上一行后端还不是 ✅ 就不能开始 |

后端写法照我现有的 handler（先看 `internal\service\router.go`、`internal\service\post.go`、`internal\db\model.go`）：`AutoMigrate` 能建表就不要手写 SQL。**一个接口要动哪几个文件、每步写什么，见 §11.11.3；交付前怎么自己验，见 §10.7。**

AI 写的后端代码，风格要跟我已经写过的大致相似：命名、错误返回、handler 拆法都向现有文件看齐，不要换成另一套写法。

每写一个前端组件之前，不用我再开口，AI 自己先按这个格式检查作者是怎么做的，并且真的去读作者源码再动手：`用 REA 检查 <作者文件或功能>。<这个组件作者是怎么实现的？>`。例如：`用 REA 检查 components\user-card.vue。作者的用户资料卡是怎么取出资料并渲染的？`

### 0.5 【已作废】不要直接改我的文件，把代码发给我

> **⚠️ 2026-10-15 起本条作废**：AI **直接改我的项目文件**，不再让我粘贴代码。
> **⚠️ 2026-10-15 第 26 轮起**：后端 Go 也一样，AI 直接改 `F:\paopao`，但只限当前这一行 C 档（§0.9）。
> 本条只在 §10.1 第 6 条被撤销时才重新生效。

### 0.6 讲解必须从作者源码出发

我需要知道「原版这个功能是怎么做的」，而不只是「你写了什么」。
请带**文件名 + 行号**引用原版代码，说明行为。外观为什么没照抄，用 §0.1 那一句带过即可。

### 0.7 不清楚就问，不要编

宁可少做一步，不要猜着做错一步。特别是：不确定某个接口的路径或字段时，**问**。

### 0.8 我是第一次写全栈项目的前端

- 不要假设我懂某个 API 或某个「显然」的写法
- 关键处请写明**为什么**这么写，而不只是写什么
- 遇到容易踩的坑，主动提醒

---

## 1. 三方路径

| 角色 | 路径 | 说明 |
|---|---|---|
| **本项目** | `F:\paopao\web` | Vue 3 + TS + Element Plus |
| **参考源码** | `F:\github_paopao\paopao-ce\web` | 作者前端（Naive UI + less） |
| **我的后端** | `F:\paopao` | Go + gin，端口 **8080**。第 26 轮起 AI **可以改**这里的 Go，但只限当前这一行 C 档任务（见 §0.9） |
| 线上演示 | https://www.slowsay.com/ | 作者官方演示站 |
| API 文档 | Notion | 链接见会话 |

**启动命令**

```bash
# 后端
cd F:\paopao && go run main.go

# 前端
cd F:\paopao\web && npm run dev
```

**作者源码关键文件索引**（引用时从这里找）

```
web\src\views\Home.vue                广场页（599 行）
web\src\views\Post.vue                帖子详情
web\src\views\Profile.vue             我的主页
web\src\views\User.vue                用户主页
web\src\components\sidebar.vue        左栏（431 行）
web\src\components\rightbar.vue       右栏（284 行）
web\src\components\main-nav.vue       顶部卡片（162 行）
web\src\components\auth.vue           登录注册弹窗
web\src\components\post-item.vue      帖子卡片（广场页的心脏）
web\src\components\compose.vue        发帖框
web\src\assets\css\main.less          全局布局地基
web\src\utils\request.ts              axios 封装（96 行）
web\src\store\user.ts                 token / 用户信息
web\src\composables\usePagination.ts  分页组合式函数
```

---

## 2. 技术栈差异（**必须做转换，不能照抄**）

| | 作者 | 本项目 |
|---|---|---|
| UI 库 | **Naive UI** | **Element Plus** |
| 样式语言 | less | scss（已装 `sass`） |
| 状态管理 | pinia + vuex | **pinia** |
| 路由 | vue-router | vue-router 5，**hash 模式** |
| 图标 | `@vicons/*` + Naive 内置 | `@vicons/ionicons5`、`@vicons/tabler`、`@element-plus/icons-vue` |
| 无限滚动 | `v3-infinite-loading` | **需自己实现**（IntersectionObserver） |
| 日期库 | moment | 不用第三方，自己写工具函数 |

**结论：作者的 `<n-xxx>` 组件、`useMessage()`、`h(Icon)`、`n-menu` 的 `:options` 写法，全部不能照抄。**
转换对照表见第 6 节。

---

## 3. 目录职责

```
F:\paopao\web\
├─ index.html              入口 HTML（favicon / title / lang）
├─ vite.config.ts          别名 @ + 跨域 proxy
├─ public\                 原样拷到站点根目录的静态文件（favicon 放这里）
└─ src\
   ├─ main.ts              应用入口：pinia → router → ElementPlus
   ├─ App.vue              三栏外壳（左栏 / 中栏 / 右栏）
   ├─ assets\
   │  ├─ reset.css        浏览器默认样式重置
   │  └─ css\main.scss    全局主题变量 + 布局地基
   ├─ router\index.ts      路由表 + 守卫
   ├─ store\               pinia：main.ts（UI 状态）、user.ts（登录态）
   ├─ types\api.ts         后端接口的出入参类型（**接口契约的唯一出处**）
   ├─ utils\
   │  ├─ request.ts       axios 封装（token 注入 / 401 处理）
   │  ├─ token.ts         token 读写
   │  └─ formatTime.ts    日期格式化（替代作者的 moment）
   ├─ api\                 一接口一函数：auth / post / user / upload
   ├─ components\          可复用组件（**不注册进路由**）
   └─ views\               页面（**每个都在 router 里注册**）
```

**`components` 还是 `views`？**

判断标准只有一条：**这个文件要出现在 `router/index.ts` 里吗？** 是 → `views`；不是 → `components`。

**引用规则**

- `@` = `F:\paopao\web\src`
- `.ts` 文件不用写后缀，**`.vue` 必须写后缀**

---

## 4. 布局尺寸常量（**抄死的，不要重新设计**）

来自作者 `web\src\assets\css\main.less` 与 `web\src\components\{sidebar,rightbar}.vue`。

```
--content-main        620px      中栏最大宽度
左栏 AppSidebar       200px      position: fixed;
                                 right: calc(50% + var(--content-main) / 2 + 10px);
右栏 AppRightbar      240px      position: fixed;
                                 left:  calc(50% + var(--content-main) / 2 + 10px);
栏间距                10px
移动端断点            821px
```

**左栏菜单项**

```
height: 42px           border-radius: 21px
padding: 0 18px 0 32px 图标 24px
```

**顶栏卡片（MainNav）**

```css
position: sticky;
top: 0;
border-radius: 0;
border-bottom: 0;
background: rgba(255, 255, 255, 0.75);   /* 暗色: rgba(16, 16, 20, 0.75) */
backdrop-filter: blur(12px);
```

**主色**

```
亮色 #18a058        暗色 #63e2b7
```

> **⚠️ 原理（不要改）**：只有**中栏在文档流**里，左右两栏用 `position: fixed` 挂在它两侧。
> 不要改成 flex 三栏布局——宽度会一样，但**滚动行为不一样**，视觉就不像了。

---

## 5. 主题与暗色模式

- 主题状态：`store/main.ts`，`localStorage` 键 `PAOPAO_THEME`
- 切换方式：给 `<html>` 加 `.dark` class
- Element Plus 暗色变量：`main.ts` 里已 `import 'element-plus/theme-chalk/dark/css-vars.css'`
- **CSS 引入顺序不能反**（`main.ts` 里有注释）：

  ```
  element-plus/dist/index.css  →  dark/css-vars.css  →  reset.css  →  main.scss
  ```

  同权重时后来者胜，`main.scss` 必须最后，否则主色还是 Element Plus 默认的蓝。

---

## 6. Naive UI → Element Plus 对照表

**遇到新组件就往这张表加一行，不要每次重新推导。**

| 作者（Naive UI） | 本项目（Element Plus） | 注意 |
|---|---|---|
| `n-card` | `el-card` | `size="small"` 可对齐 |
| `n-list` / `n-list-item` | **无对应** | 用 `<div>` 自己写分割线 |
| `n-avatar` | `el-avatar` | `round` → `shape="circle"` |
| `n-button` | `el-button` | `secondary` → `plain`；`quaternary` → `text` |
| `n-icon` | `el-icon` | `:size="24"` 照抄 |
| `n-menu` | `el-menu` | **差异最大**：作者用 `h(Icon)` + `:options`，EP 用模板 + 插槽 |
| `n-image` | `el-image` | 不传 `preview-src-list` 即关闭预览 |
| `n-image-group` | **无对应** | 用 `el-image` 的 `:preview-src-list`（传整组 URL）+ `:initial-index`，**必须加 `:preview-teleported="true"`**，否则预览层会被外层的 `overflow:hidden` 裁掉 |
| `n-grid` / `n-gi` | **无对应** | 用 CSS grid：`display: grid; grid-template-columns: repeat(N, 1fr); gap: 4px` |
| `n-drawer` | `el-drawer` | EP 默认**不** teleport，需要时显式写 `append-to-body` |
| `n-empty` | `el-empty` | |
| `n-switch` | `el-switch` | |
| `n-input` | `el-input` | |
| `n-dropdown` | `el-dropdown` | |
| `n-tabs` | `el-tabs` | 作者用 `#prefix` 插槽，EP 没有，改用 flex 排 |
| `n-badge` | `el-badge` | |
| `n-modal` / `useDialog()` | `el-dialog` / `ElMessageBox` | |
| `n-popover` | `el-popover` | |
| `n-tooltip` | `el-tooltip` | |
| `n-spin` | `v-loading` / `el-skeleton` | 页面级用 `v-loading`；**行内小转圈**（作者写 `n-spin :size="14"`）用 `<el-icon class="is-loading" :size="14"><Loading /></el-icon>`，`Loading` 来自 `@element-plus/icons-vue`。`is-loading` 这个类名就是 EP 自己的旋转动画 |
| `n-space` | **无对应** | 它就是个 flex 容器，照抄它的两个参数即可：`justify="center"` → `display:flex; justify-content:center`；`justify="space-between"` → `display:flex; justify-content:space-between` |
| `n-form` / `n-form-item` | `el-form` / `el-form-item` | |
| `n-tag` | `el-tag` | |
| `n-pagination` | `el-pagination` | |
| `n-skeleton text :repeat="3"` | `el-skeleton-item variant="text"` | 不要用 `<el-skeleton :rows="3">`，EP 会把最后一行擅自缩成 60% 宽 |
| `n-result status="404"` | `el-result` | EP 只有 success / warning / info / error 四种内置图标，没有 404 插画，要用 `#icon` 插槽自己画 |
| `useMessage()` | `ElMessage` | **函数式，直接 import，不需要 Provider** |
| `v3-infinite-loading` | 自己做 `IntersectionObserver` | |
| `n-config-provider :theme` | **无对应** | 暗色靠 `.dark` class + CSS 变量 |

---

## 7. 后端 API 约定 ★★★（**这是我的后端，不是作者的**）

### 7.1 基础

| 项 | 值 |
|---|---|
| baseURL | `/api`（由 `vite.config.ts` 的 proxy 去掉 `/api` 转发到 `http://127.0.0.1:8080`） |
| 鉴权 | `Authorization: Bearer <token>`（`utils/request.ts` 自动注入） |
| token 存储 | `localStorage`，由 `utils/token.ts` 管理 |
| 超时 | 15000 ms |
| 分页参数 | `page`（默认 1）、`page_size`（默认 10，**上限 100**） |

### 7.2 【最重要】没有 `{code, data, msg}` 信封

**HTTP 200 时，响应 body 本身就是数据。**

```
我的后端： { "total": 10, "list": [...] }
作者后端： { "code": 0, "data": { ... }, "msg": "ok" }
```

> **所以作者的 `utils/request.ts` 响应拦截器逻辑（`const { data, code } = response.data`）绝对不能抄。**
> 抄了会导致**每个请求都拿不到数据**，而且不报错，非常隐蔽。
> 本项目已在 `utils/request.ts` 正确处理：`(response) => response.data`。

### 7.3 错误格式

| 情况 | 表现 |
|---|---|
| 4xx / 5xx | body 形如 `{"error": "..."}`，少数用 `{"status": "..."}` |
| **401** | 未登录 / token 过期 → `request.ts` 会清 token 并触发 `setUnauthorizedHandler` 注册的回调（在 `App.vue` 里接的 `userStore.logout()`） |
| 成功但有话要说 | 一批写操作返回 `{"status": "中文提示"}`，**形状还不统一**（少数用 `{"error":"..."}` 却仍是 200） |

统一用 `StatusResult` 类型接。

### 7.4 【极易出错】六个列表接口的 key 各不相同

| 接口 | 列表字段 |
|---|---|
| `GET /posts` | `list` |
| `GET /posts/me` | `list` |
| `GET /feed` | **`posts`**（不是 list） |
| `GET /posts/:id/comments` | `comments` |
| `GET /users/:id/following` | `follows` |
| `GET /users/:id/followers` | `fans` |
| `GET /posts/me/likes` | **裸数组，没有任何包装** |

> **禁止**照抄作者的 `rsp.pager.total_rows` —— 我的后端是**顶层 `total`**。

### 7.5 字段易错点

1. **`images` 可能是 `null` 而不是 `[]`**（Go 的 nil 切片序列化成 `null`）。
   模板里必须写 `v-if="post.images?.length"`，不能直接 `v-for`。
2. **`created_at` 是 RFC3339 字符串**（如 `2026-02-18T10:30:00+08:00`），前端自己格式化。
   作者用 `moment`，本项目不用第三方，自己写工具函数。
3. **`GET /me` 只返回 `{ user_id, role }`**，**没有昵称和邮箱**。
   需要用户资料要另想办法。
4. **登录响应是嵌套的**：`{ token, user: { id, name, email } }`，不是平的。
5. **`GET /posts/:id/likes` 返回的是「点赞数」**，不是点赞的人名单。
6. **`GET /posts/me` 的列表项没有 `author` 字段**（自己的帖子不用显示作者）。

### 7.6 完整契约在哪

**全部类型定义在 `src\types\api.ts`，以那个文件为准。**
新增接口 → **先往 `types/api.ts` 加类型，再写 `api/` 里的函数**。
路径细节以 `src\api\` 下的实际实现 + Notion 文档为准。

---

## 8. 命名与写法约定

| 项 | 约定 | 例 |
|---|---|---|
| 页面文件 | PascalCase + `View` 结尾 | `HomeView.vue` |
| 组件文件 | PascalCase | `AppSidebar.vue` |
| 变量 / 函数 | camelCase | `loadPosts()` |
| 常量 | UPPER_SNAKE_CASE | `MOBILE_BREAKPOINT` |
| 布尔量 | `is` / `has` / `show` 开头 | `isMobile` |
| 路由 name | kebab 或单词 | `home`、`post` |

**SFC 三段固定顺序**

```vue
<script setup lang="ts">
// 1. import
// 2. 常量 / 类型
// 3. 响应式状态
// 4. 计算属性
// 5. 生命周期
// 6. 方法
</script>

<template>
  <!-- ... -->
</template>

<style scoped lang="scss">
  /* 尽量用 Element Plus 的 CSS 变量，别写死颜色 */
</style>
```

**编码要求**

- 用 TypeScript，**不用 `any`，不用 `as any`**（`as unknown as` 是例外，`request.ts` 里有必要用法）
- 注释用**中文**，解释「**为什么**」而不是「是什么」
- 每个 `api/` 函数返回 `Promise<T>`，`T` 来自 `types/api.ts`
- 不要引入新的第三方依赖，**需要时先问**

---

## 9. 先后端、再前端（**第 26 轮起的工作方式**）

> 旧的「阶段 A 假数据 → 阶段 B 接接口」**停用**。不再先做一套不能用的样子等后端。
> 每一件功能拆成两行，写在 §11.4 C 档里：上一行后端 ✅ 之后，下一行才写那一个前端组件。

| | **后端那一行** | **前端那一行** |
|---|---|---|
| 改哪里 | `F:\paopao` 的 Go + `web\src\types\api.ts` + `web\src\api\` | 点名的那一个 `.vue` |
| 外观 | 不涉及 | 跟 `HomeView` / `PostItem` / `ProfileView` 同一套，不照抄作者 |
| 怎么验收 | PowerShell 或浏览器能打到新接口，字段跟 `types\api.ts` 一致 | 页面上点得通，数据是真的 |
| 假数据 | 不用 | 不用。接口还没有就不要先画按钮 |

下面这段旧规矩仍然有效：**字段名用我的后端，不用作者的。**

> **假数据的「字段名」必须用本项目后端的，不能用作者的。**

```ts
// ❌ 错：按作者的 Tweet 字段名写
const mockPosts = [{ content: '...', attachment: '...', user: {...} }]

// ✅ 对：值随便编，但形状 = 我的 PostListItem
const mockPosts: PostListItem[] = [
  { id: 1, title: '今天天气真好', created_at: '2026-02-18T10:30:00+08:00',
    author: 'paopao', comment_count: 3, images: null },
]
```

**为什么**：类型先按我的后端定死，模板才不用返工。C 档不再用假数据顶页面。

### 作者有、但我先不做的

这些依赖外部服务，**不进 C 档**。不要自己接支付宝、短信网关、搜索引擎：

- 手机号绑定、短信验证码、图形验证码、激活码（作者 `Setting.vue` 里那几块，`types\api\Captcha.d.ts`）
- 支付宝充值（`Wallet.vue` 的支付）。钱包若要做，只做本地余额和账单，排在 C 档后部
- 付费附件（`Attachment.d.ts` 的 precheck / signed_url）
- 多云对象存储、Meilisearch / Zinc、Redis 缓存
- 公告页：作者 `views\Anouncement.vue` 自己也是空的（`onMounted` 里是 TODO），不单列
- 管理后台的通用配置 schema（`AdminSettings.vue`）。C 档只留「管理员禁言」那一件

---

## 10. 开工与收尾的固定格式

### 10.1 每个新会话开头发一段（**设定自驱动契约**）

> **用 Cursor 时不用每次发** —— 同样的内容已经写进 `F:\paopao\web\.cursor\rules\paopao.mdc`，开箱常驻。
> 换别的工具（网页版 AI、ChatGPT）时才需要贴。要点是这 7 条：

1. **【项目约定】** 全部在 `F:\paopao\web\PROJECT.md`。每轮开工必须**重新完整读一遍**，不要凭记忆、不要凭上一轮印象。
2. **【取活】** 从 §11.4 **C 档**取第一个状态不是 ✅ 的任务。一次只做一行，不跳号、不合并。S 档已停用，不要再取。
   第一个不是 ✅ 的是 🚧（上一轮已交付、等我验收），说明我回了「继续」→ 先把它改成 ✅，再往下取。
3. **【每轮固定动作】** 读 PROJECT.md → 完整读作者对应源码（只学行为）→ 读本项目同类文件（学外观）→ **直接改文件** → **自己更新 PROJECT.md**（标 🚧，并执行 §11.9）→ 末尾给 §10.2 的三段 → 停住等我说话。
4. **【顺序】** 按第 9 节和第 0.9 节：后端行只写 Go + 类型 + api 函数，风格跟我现有 Go 大致相似；前端行只写那一个组件。写组件前自己先问并读作者源码：`用 REA 检查 <作者文件或功能>。<这个组件作者是怎么实现的？>`。外观用本项目现有页面，不照抄作者。
5. **【绝对不许】** 改 §12 禁区 / 「顺手优化」已完成的页面 / 抄作者的 `/v1` 和信封 / 一轮做两个组件 / 假装功能能用了。
6. **【改文件】** 前端和后端都可以直接改，不要让我自己复制代码。后端只限当前这一行。
7. **【目标】** 把 §11.4 C 档里作者有、我还没有的功能，一个一个做成真的。

### 10.2 每轮末尾 AI 必须给的三段（**我只看这个就行**）

```
【本轮验收】T4 帖子卡片 —— 做完了。
  你要看的：打开 http://localhost:5173/ ，往下滚，每条帖子应该……
  你可以点：点作者名旁边的"…"，应该……
  已知不行的：点赞数字不显示（11.5 坑 2）——这是后端缺字段，不是 bug

  【自检】go build ./... 过；GET /users/1 返回 200，字段有 id/name/role；页面 /user/1 打开，Console 无红色报错。

【下一轮预告】T5 帖子图片 —— 参考 components\post-image.vue
  新建 src\components\PostImage.vue，改 src\components\PostItem.vue。共 2 个文件。

【等你一个字】回「继续」= 验收通过 + 放行 T5。
  哪里不对就直接说现象（"头像偏左"），别猜原因。
```

**关键点：把「验收这一轮」和「放行下一轮」合成同一个动作。**
我回一句「继续」，同时完成了两件事，所以每轮我只需要说一个字，不用来回两趟。

### 10.3 我每轮只需要回一个词

| 我回什么 | AI 要做什么 |
|---|---|
| `继续` | 把上一行 🚧 改成 ✅，开始下一轮 |
| `T4 不行：<现象>` | 状态改回 🚧，重做 T4，**不要往下走** |
| `停` | 停下来，不要开始下一轮 |
| `先做 T12` | 覆盖队列，跳到 T12 |
| `T4 没问题，但顺手把 xxx 也改了` | 只加这一件，加完仍然停住等验收 |

### 10.4 什么时候还要单独出「学习报告」

默认**不需要**——「下一轮预告」已经把方向说清楚了，我在回「继续」时就已经确认过方向。

**例外**：下面三种情况，AI 必须先单独出报告，等我回「可以」再写代码——

1. 这个任务要动 **3 个以上文件**，或要改已有文件的核心逻辑
2. 要**新增依赖**
3. ~~任务落在 B 档~~ **作废**。S1–S12 已停用，C 档按行直接做，不用先等「可以」

> 理由：这三种情况猜错的代价太大。其余情况靠「下一轮预告」拦住就够了。

### 10.5 验收时怎么描述问题（**给我自己看的**）

**说现象，不要说猜测。**

- ✅「头像偏左了，应该跟作者名左对齐」→ AI 能直接定位
- ❌「可能是 margin 算错了」→ 我的猜测通常是错的，反而会把它带偏

**手边有原版截图就贴图** —— 我可以直接从图上量出尺寸，比用文字描述省事得多。

### 10.6 原版的三条交付硬要求（**仍然有效**）

1. **每个文件给完整路径**，改动段落给完整内容，不给片段、不给省略号。
   > **⚠️ 2026-10-15 起**：文件已由 AI 直接写入（10.1 第 6 条），我不用再粘贴；但回复里**仍然要给完整路径**，方便我回看。
2. **末尾必须附验收清单**，写清楚「打开哪个地址、看到什么、点什么、会发生什么」。
3. **不确定的地方当场问**，不要猜一个看起来合理的答案糊过去。

### 10.7 交付前自检（**AI 自己跑完再说话，不用我提醒**）

改完文件、还没写汇报**之前**，自己按顺序跑一遍。哪一步红了就先修；修不好就如实说，
**不准把一个红的交付当绿的报给我**。

**① 先跑命令**

```powershell
# 后端行（每改一次 Go 都要）
cd F:\paopao; go build ./...     # 没输出 = 编译过
cd F:\paopao; go vet ./...       # 没输出 = 没低级错误

# 前端行
cd F:\paopao\web; npm run type-check   # 无 TS 报错
cd F:\paopao\web; npm run build-only   # 能出 dist = 模板和 scss 没写坏
```

**② 再真的打开页面看（Cursor 里有浏览器就一定要用）**

- **后端行**：先重启服务 —— `cd F:\paopao; go run main.go`
  （**gin 没有热重载，不重启改动不生效**，这是最容易骗过自己的一步）。
  然后打开要验的那个地址，例如 `http://127.0.0.1:8080/users/1`。
  看到 JSON、且字段名跟 `src\types\api.ts` 那个 interface 对得上，才算过；
  看到 `{"error":...}`、404、或字段名对不上，就是没过。
- **前端行**：`cd F:\paopao\web; npm run dev`（地址 **http://localhost:5173/**，hash 路由）。
  打开这一行对应的那一页，**看两处**：① 页面上这个功能真的点得通、数据是真的；
  ② 开发者工具 **Console 有没有红色报错**、**Network 里那个请求是不是 200**。
- 页面白屏 / `Failed to fetch` / 502 —— 先确认后端起着，再怀疑自己的代码。

**③ 库里没数据就先造（不要把假返回写进代码来「验」）**

验新接口时表常常是空的。用 PowerShell 打真接口造数据：

```powershell
# 注册一个测试用户（邮箱重复会返回「用户已存在」，换个邮箱即可）
Invoke-RestMethod -Uri http://127.0.0.1:8080/register -Method Post `
  -ContentType 'application/json' -Body '{"name":"测试甲","email":"t1@test.com","password":"123456"}'

# 登录拿 token（验需要登录的接口时带上 Authorization: Bearer <token>）
$r = Invoke-RestMethod -Uri http://127.0.0.1:8080/login -Method Post `
  -ContentType 'application/json' -Body '{"email":"t1@test.com","password":"123456"}'
$r.token
```

数据库是 **MySQL**，连接串读 `F:\paopao\.env`（`internal\config\config.go`；**缺 `.env` 会在启动时就报「配置加载失败」**）。
造进去的测试数据留着不影响功能，**不要 DROP 表、不要删别人的数据**。

**④ 自检结果要写进汇报**

在 §10.2 的【本轮验收】里加一行：

```
【自检】go build ./... 过；GET /users/1 返回 200，有 id/name/role；页面 /user/1 打开，Console 无红色报错。
```

**没跑的就写「没跑，因为……」，不要空着，更不要编。**

---

## 11. 进度与任务队列

> **【AI 取活的地方】每轮开场把这一节读完就够，不用通读全文。**

### 11.1 总账（第 26 轮修订）

已经能用的：登录注册、广场「全部」、关注流、发帖（文字 + 图片）、点赞、一层评论、删评论、删帖、复制链接、我的主页帖子列表、我喜欢（只有作者和标题）、关注 / 粉丝列表、别人主页上的关注按钮、404、暗色。

还没有的：全部在下面 **C 档**，按行做。S 档停用。

作者有、明确先不做的：见第 9 节末尾（短信、支付宝、付费附件、公告空页、管理配置台）。

### 11.2 基础设施（**全部 ✅，别再动**）

`src\main.ts`、`src\App.vue`、`src\router\index.ts`、`src\assets\css\main.scss`、`src\utils\{request,token,formatTime}.ts`、
`src\types\api.ts`、`src\api\{auth,post,user,upload}.ts`、`src\store\{main,user}.ts`

C 档某一行点名要改其中的类型、api 函数或**新增**路由时可以改。不要改已有接口的形状，不要改三栏结构和 import 顺序。

### 11.3 已完成（**别再重做**）

- **页面**：`HomeView`（广场 `/`）、`PostDetailView`（`/post/:id`）、`FollowingView`（`/following?id=&n=&t=`）、`ProfileView`（`/profile`）、`UserView`（`/user/:id`，**阶段 A**）、`NotFoundView`（`/404`），外加 `PlaceholderView` 顶着 `/topic` `/messages` `/collection` `/contacts` `/setting` 五个占位路由。
- **组件**（14 个，`src\components\`）：`AppSidebar`、`AppRightbar`、`MainNav`、`MobileNav`、`HomeCompose`、`PostItem`、`PostImage`、`PostSkeleton`、`InfiniteLoadMore`、`PostLink`（阶段 A）、`CommentItem`、`ComposeComment`、`UserCard`（C3 起名字和计数来自 `GET /users/:id`）、`PostDetail`。
- **工具**：`src\utils\formatTime.ts`、`src\utils\count.ts`。

> 这份清单**只到这一行为止**。要确认某个文件在不在，**直接列 `src\components\` `src\views\` 目录**，不要依赖本文件。
> 「阶段 A」的含义见第 9 节。

---

### 11.4 ★ 任务队列（**AI 从这里取活，一次一个**）

#### A 档 —— 后端接口齐全，能做完整功能（按第 9 节走 A → B 两阶段）

| # | 任务 | 要创建 / 修改的文件 | 参考源码 | 需要哪些后端接口 | 依赖 | 状态 |
|---|---|---|---|---|---|---|
| **T17** | 404 页 | `src\views\NotFoundView.vue`（已改写） | `views\404.vue` | 无 | 无 | ✅ 第 23 轮验收通过 |
| **T18** | ★ **广场页阶段 B** | `src\views\HomeView.vue`、`src\components\PostItem.vue`、`src\components\PostDetail.vue`、`src\types\api.ts` | `views\Home.vue` 的 `loadPosts`（`:345-375`）、`components\post-item.vue` | `GET /posts`、`GET /feed`、`POST` / `DELETE /posts/:id/likes` | B0 后端已落地 | ✅ 第 25 轮验收通过 |

> T18 已按这两条做完：外链选了 ②（`PostLink.vue` 还在，没挂到真帖子上）；「热门推荐」仍是死控件。

#### B 档 —— **停用**

S1–S12 不再取活。那些功能改到下面的 C 档，拆成「先后端、再一个组件」。

#### C 档 —— 作者有、我还没有（**从这里取活**）

对照来源：作者 `web\src\types\api\*.d.ts`、`api\post.ts`、各 `views\`。行为照作者，接口形状照我现有的 gin，外观照 `HomeView` / `PostItem`。

状态空着 = 还没做。一轮只做一行。

**取活规则（怕一次做太多，就靠这四条）**

1. **一行里明显塞了多个独立功能的，只做第一件。** 例如 C46「改昵称 / 改密码 / 换头像」是三件、C27 话题是四件 —— 只做第一件，然后在【下一轮预告】里写「建议把 C46 拆成 C46a / C46b / C46c」，等我说了再做下一件。
2. **宁可拆细，不许合并。** 一行做完还剩另一半，**这一行不算 ✅**：标 🚧 并注明「只做了 X，还剩 Y」。**✅ 只代表这一行整行做完、且我验收通过。**
3. **§10.4 的「3 个以上文件」按 ≥3 算。** 正好 3 个文件也算（例如 C58：新建页面 + 改 `AppSidebar.vue` + 改路由），先出学习报告等我回「可以」。
4. **开工前先看这一行有多重。** 作者对应的那个文件越大越可能要拆：`Setting.vue` 779 行（C46）、`User.vue` 652（C2 / C5 / C50 / C56 / C60）、`Home.vue` 599、`Messages.vue` 380（C34 / C35 / C36）、`Wallet.vue` 332（C58）、`Following.vue` 201（C9 / C10）。**超过 400 行的按第 1 条拆。**

##### 用户资料和别人的帖子

| # | 侧 | 做什么 | 作者依据 | 只改这些 | 状态 |
|---|---|---|---|---|---|
| **C1** | 后端 | `GET /users/:id`：`id`、`name`、`role`，以及关注数、粉丝数、帖子数 | `User.d.ts` 的 `profile` | Go + `types\api.ts` + `api\user.ts` | ✅ 第 29 轮 |
| **C2** | 前端 | `UserView.vue` 头部改用 C1，不再从关注列表借名字 | `views\User.vue` | 只改这一个页面 | ✅ 第 30 轮 |
| **C3** | 前端 | `UserCard.vue` 改用 C1 的真资料 | `components\user-card.vue` | 只改这一个组件 | ✅ 第 31 轮 |
| **C4** | 后端 | `GET /users/:id/posts`，分页，每一项用现有 `PostListItem` | `User.d.ts` 的 `posts` | Go + 类型 + `api\user.ts` | ✅ 第 32 轮 |
| **C5** | 前端 | `UserView.vue` 的帖子列表接 C4，去掉永远空态 | `views\User.vue` 的「泡泡」tab | 只改这一个页面 | ✅ 第 33 轮。只接了「泡泡」；另外四个 tab 仍空，那是 C56 |
| **C6** | 后端 | `GET /posts/me` 补上 `user_id`、`like_count`、`is_liked`（广场已经有） | 我的主页空枪，见 §11.8 | Go + `types\api.ts` | ✅ 第 34 轮 |
| **C7** | 前端 | `ProfileView.vue` 的帖子卡片能点赞、作者名能点 | `views\Profile.vue` | 只改这一个页面 | ✅ 第 35 轮 |

##### 关注、搜索、热门、评论

| # | 侧 | 做什么 | 作者依据 | 只改这些 | 状态 |
|---|---|---|---|---|---|
| **C8** | 后端 | 帖子列表和详情加 `is_following` | `post-item.vue:218-224` | Go + `PostListItem` / `PostDetail` | ✅ 第 36 轮 |
| **C9** | 前端 | `PostItem.vue`「更多」里加上关注 / 取关 | 同上 | 只改这一个组件 | ✅ 第 37 轮 |
| **C10** | 前端 | `PostDetail.vue`「更多」里加上关注 / 取关 | `post-detail.vue:336-348` | 只改这一个组件 | ✅ 第 38 轮 |
| **C11** | 后端 | `GET /posts` 增加 `q`，按标题和正文包含匹配。不要接 Meilisearch | `Home.vue:345-350` 的 search | Go + `PageParams` 或单独参数 | ✅ 第 39 轮。用的是单独的 `PostListParams.q`，没放进共用的 `PageParams` |
| **C12** | 前端 | `AppRightbar.vue` 加上搜索框，提交后打开 `/?q=` | `rightbar.vue` 的 search-wrap | 只改这一个组件 | ✅ 第 40 轮 |
| **C13** | 后端 | `GET /posts?style=hots` 按点赞数排序 | `Home.vue` 的 `hots` | Go | 🚧 第 40 轮，**等你验收**。查询参数写进了 `PostListParams.style` |
| **C14** | 前端 | `HomeView.vue` 的「热门推荐」真的换列表 | `Home.vue` 的 loadPosts | 只改这一个页面 | |
| **C15** | 后端 | `GET /posts/:id/comments` 增加 `sort=default\|hots\|newest` | `Post.vue:21-23` | Go | |
| **C16** | 前端 | 详情页三个排序 Tab 按 C15 重新请求 | `views\Post.vue` | 只改 `PostDetailView.vue` | |
| **C17** | 后端 | 评论可以带图片 id（表和 `POST /posts/:id/comments`） | `compose-comment.vue` 的图片 | Go + 类型 | |
| **C18** | 前端 | `ComposeComment.vue` 发布时把图片发出去 | 同上 | 只改这一个组件 | |

##### 收藏、回复、评论赞

| # | 侧 | 做什么 | 作者依据 | 只改这些 | 状态 |
|---|---|---|---|---|---|
| **C19** | 后端 | 收藏表；收藏 / 取消；`GET /posts/me/collections`；帖子上带 `collection_count`、`is_collected` | `api\post.ts` 的 `/v1/post/collection`，`User.d.ts` 的 collections | Go + 类型 + api | |
| **C20** | 前端 | `PostItem.vue` 的收藏图标接 C19，并显示数字 | `post-item.vue` 收藏按钮 | 只改这一个组件 | |
| **C21** | 前端 | 用真页面换掉 `/collection` 占位 | `views\Collection.vue` | 新建 `CollectionView.vue`，路由只把占位换成它 | |
| **C22** | 后端 | 评论加 `parent_id`；`POST /comments/:id/replies`；列表能看出回复谁 | `compose-reply.vue`、`reply-item.vue` | Go + `CommentItem` | |
| **C23** | 前端 | 新建 `ComposeReply.vue` | `components\compose-reply.vue` | 只新建这一个组件 | |
| **C24** | 前端 | 新建 `ReplyItem.vue`，并在 `CommentItem.vue` 里挂上（挂载算这一行，因为回复项离开评论项无法看见） | `components\reply-item.vue` | 这两个文件 | |
| **C25** | 后端 | 评论点赞 / 取消，评论上带 `like_count`、`is_liked` | `api\post.ts` 的 thumbsup / thumbsdown。点踩先不做，只做点赞 | Go | |
| **C26** | 前端 | `CommentItem.vue` 加上点赞 | 同上 | 只改这一个组件 | |

##### 话题

| # | 侧 | 做什么 | 作者依据 | 只改这些 | 状态 |
|---|---|---|---|---|---|
| **C27** | 后端 | 发帖时从正文提取 `#话题`；`GET /tags`；关注 / 取关话题；`GET /posts?tag=` | `api\post.ts` 的 tags、topic/follow；`Topic.vue` | Go + 类型 + api | |
| **C28** | 前端 | `PostItem.vue` 正文里的 `#话题` 能点到 `/?tag=` | `utils\content.ts` 的 tagExp | 只改这一个组件 | |
| **C29** | 前端 | 话题页换掉 `/topic` 占位。Tab 先做「热门 / 最新」两个，关注和钉住等 C27 的关注接口有了再用 | `views\Topic.vue` | 新建 `TopicView.vue` | |
| **C30** | 前端 | 新建 `TagItem.vue`，只被话题页使用 | `components\tag-item.vue` | 只新建这一个组件，并在 `TopicView.vue` 里换成它 | |
| **C31** | 前端 | `AppRightbar.vue` 加上热门话题，数据用 `GET /tags` | `rightbar.vue` 话题块 | 只改这一个组件 | |

##### 消息、好友、私信

| # | 侧 | 做什么 | 作者依据 | 只改这些 | 状态 |
|---|---|---|---|---|---|
| **C32** | 后端 | 消息表。有人评论我、关注我时写一条。`GET /messages`、未读数、单条已读、全部已读 | `User.d.ts` 的 messages / msgcount / message.read | Go + 类型 + api | |
| **C33** | 前端 | `AppSidebar.vue` 消息项上的未读数 | `sidebar.vue` 消息角标 | 只改这一个组件 | |
| **C34** | 前端 | 消息页换掉 `/messages` 占位 | `views\Messages.vue` | 新建 `MessagesView.vue` | |
| **C35** | 前端 | 新建 `MessageItem.vue` | `components\message-item.vue` | 这一个组件 + 在消息页里用它 | |
| **C36** | 前端 | 新建 `MessageSkeleton.vue` | `components\message-skeleton.vue` | 这一个组件 + 在消息页里用它 | |
| **C37** | 后端 | 好友申请、同意、拒绝、删除、联系人列表。帖子上加 `is_friend` | `Friend.d.ts`；`post-item.vue:233-239` | Go + 类型 + api | |
| **C38** | 前端 | 好友页换掉 `/contacts` 占位 | `views\Contacts.vue` | 新建 `ContactsView.vue` | |
| **C39** | 前端 | 新建 `WhisperAddFriend.vue` | `components\whisper-add-friend.vue` | 只这一个组件 | |
| **C40** | 后端 | 私信：发给某人一条，收进 C32 的消息表，`style` 能滤出悄悄话 | `User.d.ts` 的 `whisper` | Go | |
| **C41** | 前端 | 新建 `Whisper.vue` | `components\whisper.vue` | 只这一个组件 | |
| **C42** | 前端 | `PostItem.vue`「更多」里加上添加 / 删除好友 | `post-item.vue:233-239` | 只改这一个组件 | |

##### 设置、头像、帖子管理

| # | 侧 | 做什么 | 作者依据 | 只改这些 | 状态 |
|---|---|---|---|---|---|
| **C43** | 后端 | `PUT /me/name` 改昵称 | `User.d.ts` 的 nickname | Go + api | |
| **C44** | 后端 | `PUT /me/password` 改密码，要旧密码 | `User.d.ts` 的 password | Go + api | |
| **C45** | 后端 | 用户表加头像；`POST /me/avatar` 上传后写入 | `User.d.ts` 的 avatar | Go + api | |
| **C46** | 前端 | 设置页换掉 `/setting` 占位：改昵称、改密码、换头像。不要做手机号和激活码 | `views\Setting.vue` | 新建 `SettingView.vue` | |
| **C47** | 前端 | 新建 `UserAvatar.vue`，按 url 显示头像，没有 url 就用现在的首字母 | 作者各处的 `user.avatar` | 只新建这一个组件，本轮不替换调用方 | |
| **C48** | 前端 | `PostItem.vue` 改用 `UserAvatar` | `post-item.vue:5` | 只改这一个组件 | |
| **C49** | 前端 | `ProfileView.vue` 改用 `UserAvatar` | `views\Profile.vue` | 只改这一个页面 | |
| **C50** | 前端 | `UserView.vue` 改用 `UserAvatar` | `views\User.vue` | 只改这一个页面 | |
| **C51** | 后端 | 帖子增加可见范围（公开 / 粉丝 / 好友 / 私密）、锁定、置顶、亮点，以及对应的修改接口。广场默认只出公开且未锁定的 | `api\post.ts` 的 lock / stick / highlight / visibility | Go + 类型 | |
| **C52** | 前端 | `PostDetail.vue`「更多」里加上这些操作（仅作者本人或管理员看得到该看的项） | `post-detail.vue:356-446` | 只改这一个组件 | |
| **C53** | 后端 | `GET /posts/me?style=` 支持 comment / highlight / media；`GET /posts/me/likes` 改成和帖子列表一样的完整结构，不要再只回作者和标题 | `Profile.vue:82-86` | Go | |
| **C54** | 前端 | `ProfileView.vue` 补上「评论 / 亮点 / 图文」三个 tab，「喜欢」改成完整卡片 | 同上 | 只改这一个页面 | |
| **C55** | 后端 | `GET /users/:id/posts?style=` 支持同样四个过滤 | `User.vue:93-97` | Go | |
| **C56** | 前端 | `UserView.vue` 补上同样的 tab | 同上 | 只改这一个页面 | |

##### 钱包、禁言、链接、视频和附件

| # | 侧 | 做什么 | 作者依据 | 只改这些 | 状态 |
|---|---|---|---|---|---|
| **C57** | 后端 | 本地余额和账单。充值只是给自己加一笔，**不接支付宝** | `User.d.ts` 的 wallet.bills。支付接口不做 | Go + 类型 + api | |
| **C58** | 前端 | 新建 `WalletView.vue`，左栏加上入口 | `views\Wallet.vue` | 这一个页面 + `AppSidebar.vue` 加一项 + 路由加一条 | |
| **C59** | 后端 | 打开管理员中间件；`POST /users/:id/status` 禁言 / 解禁 | `Admin.d.ts` 的 user.status；`router.go:29` 现在是注释 | Go | |
| **C60** | 前端 | `UserView.vue` 管理员能看到禁言 / 解禁 | `User.vue:408-414` | 只改这一个页面 | |
| **C61** | 后端 | 帖子增加 `links`（url + 标题）。发帖体可以带链接 | `components\post-link.vue`，内容类型 6 | Go + 类型 | |
| **C62** | 前端 | `PostItem.vue` 有链接就渲染已有的 `PostLink.vue` | 同上 | 只改 `PostItem.vue` | |
| **C63** | 前端 | `HomeCompose.vue` 的「链接」按钮能填一条链接并随帖子提交 | `compose.vue` 的链接 | 只改这一个组件 | |
| **C64** | 后端 | 上传允许视频和普通附件，帖子能带上它们 | `post-video.vue`、`post-attachment.vue` | Go + 上传接口 | |
| **C65** | 前端 | 新建 `PostVideo.vue` | `components\post-video.vue` | 只这一个组件 | |
| **C66** | 前端 | 新建 `PostAttachment.vue` | `components\post-attachment.vue` | 只这一个组件 | |
| **C67** | 前端 | `PostItem.vue` 渲染视频和附件 | 作者 `usePostContent.ts` | 只改这一个组件 | |
| **C68** | 前端 | `HomeCompose.vue` 的「视频」「附件」按钮接上传 | `compose.vue` | 只改这一个组件 | |

> 左栏作者还有「公告」「系统配置」。公告页作者自己是空的，系统配置整页不做，所以这两项不加。
> 「可见范围」按钮在发帖框里（作者 compose 的 eye）。它依赖 C51。做 C63 时不要顺手做它；若要做，等 C51 ✅ 之后另说「先做发帖可见范围」，不要插进当前行。

---

### 11.5 ⚠️ 三个「以为能做、其实做不全」的坑（**做 T18 / 配方 B0 之前必读**）

**坑 1 —— 帖子卡片上的作者名，点不进他的主页**

```go
// F:\paopao\internal\service\post.go:60-67
type PostListItem struct {
	ID           uint        `json:"id"`
	Title        string      `json:"title"`
	CreatedAt    time.Time   `json:"created_at"`
	Author       string      `json:"author"`     // ← 只有名字，没有 id
	CommentCount int         `json:"comment_count"`
	Images       []ImageItem `json:"images"`
}
```

`author` 只是一个字符串，**没有 `user_id` 可跳**；也没有 `avatar`，头像只能用默认图。
后果：`UserView`（`/user/:id`）从帖子列表进不去，只能从关注列表进。

> ✅ **有解了：配方 B0（见 11.11.5）。** 后端给 `PostListItem` / `PostDetail` 各加一个 `user_id` 就行，改动极小。
> **正因为这个字段决定模板结构，B0 必须排在 T18 前面做。**

**坑 2 —— 点赞按钮显示不出「我赞过没」和「几个人赞过」**

`PostListItem` 里**既没有 `like_count`，也没有 `is_liked`**。
作者的 `post-item.vue` 直接读 `tweet.likes_count` / `tweet.is_liked`，**照抄拿到的是 `undefined`**
（不报错，按钮永远不高亮、数字永远空）。唯一办法是每条帖子单独打一次 `GET /posts/:id/likes`——
一页 10 条就是 10 个额外请求。

> ✅ **有解了：配方 B0（见 11.11.5）。** **关键是别在 for 循环里逐条查** ——
> B0 给的是 `commentCountMap` 那套「一次查完一整页、按 `post_id` 分组」的批量写法，照抄即可。

> ⚠️ **阶段 A 会骗过你**：假数据里写了 `like_count: 3`，页面看着很正常；阶段 B 换真数据才炸。

**坑 3 —— 评论区做不了二级回复**

```go
// F:\paopao\internal\service\post.go:36-41
type CommentItem struct {
	ID        uint      `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}
```

没有 `parent_id`、没有 `reply_to`、没有 `like_count`。评论是**平铺一层**的。
作者的 `compose-reply.vue` + `reply-item.vue` 整套只能做阶段 A，做完也接不上数据（= S9）。

---

### 11.6 还活着的接口契约缺口

> 这里**只留还没兑现的**；已经处理掉的偏差记录在代码注释里。

| 位置 | 情况 | 谁来做 |
|---|---|---|
| `src\types\api.ts` 的 `FollowingResult` / `FollowerResult` | 后端 `F:\paopao\internal\service\follow.go:136-142`（及 `:152-201`）**已经返回**顶层 `"user": <昵称>`，契约里漏了这两个字段。`UserView.vue` 现在用一个就地收窄的类型别名顶着 | 前端自己的活，确认后补两行、删别名 |
| `F:\paopao\internal\service\follow.go` 后半段 | `GET /users/:id/followers` 大概率也返回顶层 `user`，**未确认** | 顺手核一下 |

---

### 11.7 建议后端补的接口（**这是「后端工作清单」，不是前端任务**）

> **第 26 轮起本表不再派活。** 缺的功能在 §11.4 C 档里，AI 直接改 Go，不再只写配方。
> 下表留着是为了对照「当初为什么缺」。B0 已经做完。

| 优先级 | 补什么 | 解决什么 | 改动量 |
|---|---|---|---|
| ✅ 已落地 | **〔配方 B0〕** `PostListItem` / `PostDetail` 加 `user_id` / `like_count` / `is_liked` | 坑 1、坑 2。后端已写完，前端类型等 T18 | 极小 |
| ✅ 已落地 | `GET /users/:id` 查用户资料（C1） | `UserView` 头部已接（C2）。`UserCard` 已接（C3） | 小 |
| ✅ 已落地 | `GET /users/:id/posts` 查别人发的帖子 | 接口已验收（C4）。`UserView` 的「泡泡」已接（C5） | 小 |
| ★★★★ | `CommentItem` 加 `parent_id` | 二级回复（解锁 S9） | 中（动表结构） |
| ★★★ | 收藏接口 | 解锁 S3 | 中（加表） |
| ★★ | 话题体系 | 解锁 S1 | 大 |
| ★ | 消息 / 私信 | 解锁 S2、S4 | 大 |

> **配方 B0 的后端已经落地**（第 23 轮对过 `post.go` / `feed.go`）。前端类型在第 24 轮补进了 `types\api.ts`。
> 理由见 11.11.4 的「排期铁律」—— 这两个字段决定 `PostItem.vue` 的模板结构。

---

### 11.8 ⚠️ 已知的空枪（**按钮在，功能没接**）

| 位置 | 现象 | 原因 | 什么时候能修 |
|---|---|---|---|
| `HomeView.vue` 的「热门推荐」按钮 | 点了只换高亮，列表不动 | `GET /posts?style=hots` 已按点赞数从高到低排（C13）。页面还没把 `style` 传进去 | C14 |
| `PostDetailView.vue` 评论区的三个排序 Tab | 同上 | `GET /posts/:id/comments` 没有排序参数 | 后端补参数 |
| `ComposeComment.vue` 的图片按钮 | 能传上去，**发布时带不走** | 后端 `CommentReq` 只有 `content` | 后端补字段 |
| `PostItem.vue` 的收藏 / 分享图标 | 只有图标，没有数字 | 后端没有 `collection_count` / `share_count` | 后端补字段 |
| `PostItem.vue` 的「更多」菜单 | 没有私信、加好友、锁定、可见范围 | 关注 / 取关已接（C9）。私信、好友、锁定还没有接口 | C41 / C42 / C51 |
| `PostDetail.vue` 的「更多」菜单 | 别人的帖子只有关注 / 取关。没有私信。自己的帖子只有删除，没有锁定、置顶、亮点、可见范围。管理员看别人的帖子这颗按钮不出现 | 关注 / 取关已接（C10）。其余项作者有，接口还没有 | C41 / C52 |
| `UserView.vue` 的「评论 / 亮点 / 图文 / 喜欢」 | 点了是空态 | 后端没有 `style`。 「泡泡」已经接上 `GET /users/:id/posts` | C55 补参数，C56 接页面 |
| `UserView.vue` 的「好友 / 已关注」标签 | 不显示 | C1 没有好友关系、也没有 `is_following`。`role === 'admin'` 时会显示「管理员」 | 好友和是否已关注等以后的接口 |
| `UserCard.vue` | 头像是空圆；没有「加入时间」；「已关注」仍靠列表页按 Tab 推断 | C1 没有头像、没有注册时间、没有 `is_following`。名字、管理员标签、关注 / 粉丝 / 泡泡数已经来自 `GET /users/:id` | 头像等 C45 / C47。是否已关注等以后的接口 |
| `HomeView.vue` 的标题 | 地址带 `?q=` 时标题会变成「搜索: …」或「#话题」，**列表仍是全部帖子** | 搜索框已接（C12），回车会打开 `/?q=`。`GET /posts?q=` 能搜（C11），但 `HomeView` 还没把 `q` 传给 `getPosts`。`#话题` 仍没有参数 | 列表要接上得另说。话题是 C27 |
| `PostLink.vue` | 广场和详情上都**不再出现**外链。组件文件还在 | T18 选了方案 ②：假链接不能挂在真帖子上。后端仍然没有 `links` | 后端补字段后再挂回去 |

> 这类「样子在、功能挂着」的项，**每发现一个就往这里加一行**，否则三个月后你会以为登录是好的。

---

### 11.9 维护规则（**每轮结束时 AI 自己执行，不用我提醒**）

| 什么时候 | 改哪里 |
|---|---|
| 交付一个任务 | 把 11.4 **C 档**对应行标成 **🚧**，并注明是哪一轮 |
| 收到我回「继续」 | 把上一行的 🚧 改成 **✅**，然后开始取 C 档下一个空行 |
| 我打回某个任务 | 把那一行改回 **🚧**，重做它，**不要往下走** |
| 每发现一个新的空枪 | 加进 11.8 |
| 每发现一处「文档与实现对不上」 | 加进 11.6 |
| 后端补了接口 | 把 C 档里对应的后端行标 ✅。不要再把任务从 B 档挪到 A 档，S 档已停用 |
| **撞到 C 档没列出的后端缺口** | 停下来问我，不要自己加一行就开始写。C 档里已经列出的，按那一行写 Go |
| **每轮交付前** | 先按 §10.7 自检（跑命令 + 真的打开页面看）。自检红了**不要**标 🚧 装作做完了：要么修好，要么如实标 🚧 并在汇报里写清哪一步没过 |

C 档一行标 ✅ 就表示这一行做完了。后端行 ✅ 只表示接口能打通，不表示下一行前端也做了。

### 11.10 轮次记录

第 1–21 轮的逐轮记录已在第 22 轮精简时删除。第 22 轮 = PROJECT.md 精简。第 23–25 轮 = T17、T18。第 26 轮 = 改契约（S 档停用，C1–C68 列出）。
第 27 轮 = **只改文档、没写功能代码**：加了 §10.7 交付前自检、§11.4 取活规则、§11.11.3 后端三步写法、§11.9 自检那一行。
第 28 轮 = C1 `GET /users/:id` 写完，等验收。
第 29 轮 = C1 标 ✅。C2：`UserView.vue` 头部改调 `getUser`，名字、关注、粉丝、泡泡、管理员标签都来自 `GET /users/:id`。帖子列表仍空（C5）。
第 30 轮 = C2 标 ✅。C3：`UserCard.vue` 按 `contact.id` 调 `getUser`。名字、管理员标签、关注 / 粉丝 / 泡泡数来自资料接口；列表里的 name 只在请求回来前兜底。关注 / 取关菜单没改。头像仍空。
第 31 轮 = C3 标 ✅。你回了「可以」之后写下 C4：`GET /users/:id/posts`，分页，每一项是现有 `PostListItem`。用户不存在返回 404。没有 `style`（那是 C55）。没改页面。
第 32 轮 = C4 标 ✅。C5：`UserView.vue` 的「泡泡」用 `getUserPosts` 拉真列表，滚到底继续翻页。另外四个 tab 仍是空态。
第 33 轮 = C5 标 ✅。C6：`GET /posts/me` 的每一项补上 `user_id`、`like_count`、`is_liked`。没有 `author`，也没改 `ProfileView.vue`。
第 34 轮 = C6 标 ✅。C7：`ProfileView.vue` 把这三项交给 `PostItem`。作者名点进 `/user/:id`，爱心显示数字并点赞，不再进详情。
第 35 轮 = C7 标 ✅。你回了「继续」之后写下 C8：`GET /posts`、`GET /feed`、`GET /users/:id/posts`、`GET /posts/:id` 增加 `is_following`。未登录是 false。`GET /posts/me` 没有这个字段。没改页面。
第 36 轮 = C8 标 ✅。C9：`PostItem.vue`「更多」里加上关注 / 取关。用 C8 的 `is_following` 和现有 `followUser` / `unfollowUser`。自己的帖、未登录不出现这项。私信和加好友没做。没改 `PostDetail.vue`（那是 C10）。
第 37 轮 = C9 标 ✅。C10：`PostDetail.vue`「更多」里加上关注 / 取关。作者 `post-detail.vue:331-353`：不是管理员、也不是作者本人，才出现这项；自己的帖子仍只有删除。未登录不出现。私信、锁定、置顶、亮点、可见范围没做。没改 Go。
第 38 轮 = C10 标 ✅。C11：`GET /posts` 增加 `q`，标题或正文包含即命中。空的 `q` 仍是全部帖子。没接搜索引擎，也没做 `style=hots` 和话题。类型放在 `PostListParams`，没改页面。
第 39 轮 = C11 标 ✅。C12：`AppRightbar.vue` 加上搜索框。回车打开 `/?q=`，空词打开广场且不带 `q`。热门话题没画。没改 `HomeView.vue`，所以列表仍不过滤。
第 40 轮 = C12 标 ✅。C13：`GET /posts?style=hots` 按点赞数从高到低，赞数相同再按 id 倒序。不传 style 仍是新帖在前。`q` 可以一起用。没改页面。

---

### 11.11 ★★★ 需要改后端时怎么办（第 26 轮改过）

#### 11.11.1 三条铁律

1. **AI 可以改 `F:\paopao` 的 Go，但只限 C 档当前这一行。** 不要改这一行用不到的文件。写法跟 `internal\service\router.go`、`post.go`、`db\model.go` 一致：没有信封，列表用顶层 `total`，路径不要写成作者的 `/v1/...`。要动表就加进 `AutoMigrate`。
2. **前端行开始之前，它依赖的后端行必须是 ✅。** 不许先画一个点了没反应的按钮。
3. **C 档没列出的缺口，停下来问我。** 不要自己加任务，也不要再写 `后端配方-*.md`。B0 那份配方是历史。

#### 11.11.2 后端那一行交付时必须写清

| 段 | 内容 |
|---|---|
| 改了哪些文件 | 完整路径 |
| 动没动数据库 | 有没有新表 / 新字段，`AutoMigrate` 会不会带上 |
| 怎么验证 | 浏览器或 PowerShell 能直接打的命令。不用登录就能验的先给 |
| 下一行前端要接什么 | 类型在 `types\api.ts` 的哪一个 interface |

#### 11.11.3 后端那一行怎么写（**从零加一个接口的标准三步**）

先读这三个文件，照着它们的写法来，**不要发明新风格**：

| 文件 | 看什么 |
|---|---|
| `F:\paopao\internal\service\service.go` | 公共助手，**先翻这个，能用助手就别自己写**：`GetUserID(c)`（必须登录，取不到直接写 401）、`optionalUserID(c)`（公开路由软取，未登录返回 0）、`GetParamID(c)`（拿路径里的 `:id`）、`parsePage(c)`（拿 `page` / `page_size` / `offset`，上限 100） |
| `F:\paopao\internal\service\router.go` | 路由怎么注册、`Auth` / `OptionalAuth` 挂在哪、handler 在哪 `New` 出来（`:30-36` 那一排） |
| `F:\paopao\internal\service\post.go` | 一个 handler 的完整样子：绑定 → 查库 → `c.JSON` |

**三步：**

1. **表** —— `internal\db\model.go`。要新表 / 新字段才动，不动就跳过。
   加 struct，并在文件末尾 `AutoMigrate(...)` 的参数里补上（`model.go:72-81`）。GORM 会自动建表 / 加列，**不要手写 SQL**。
   联合唯一索引照 `PostLike`（`model.go:44-50`）的写法：两个字段挂同一个 `uniqueIndex` 名。
2. **handler** —— `internal\service\<域>.go`，没有对应文件就新建一个。

   ```go
   func (h *XxxHandler) GetOne(c *gin.Context) {
       id, ok := GetParamID(c)
       if !ok { return }        // 助手已经写过响应了，直接 return
       // ... 查库
       c.JSON(http.StatusOK, 结构体)
   }
   ```

   **返回形状**：没有信封，直接返回结构体；列表返回 `gin.H{"total": n, "list": items}`；
   出错 `gin.H{"error": "中文说明"}`；成功但要说一句用 `gin.H{"status": "中文说明"}`。
   **JSON 字段名必须跟 `src\types\api.ts` 里那个 interface 完全一致** —— 后端改 tag、前端改类型，是**同一行**里的事。
3. **路由** —— `internal\service\router.go`，挂到对应的 `r.Group` 或根上。
   要登录挂 `Auth`，允许未登录挂 `OptionalAuth`，都不挂就是纯公开。新 handler 照 `:30-36` 那一排加一行 `NewXxxHandler(db)`。

**不要做**：动跟这一行无关的 Go 文件；改现成接口的返回结构（会打爆已经写完的前端）；
引入新的第三方库（先问）；把管理员的 `RequireAdmin`（`router.go:29`，现在是注释）打开 —— 那是 C59 的事。

#### 11.11.4 排期

C 档已经按「先后端、再前端」排好了。不要把两行并成一轮。

#### 11.11.5 已经出好的配方

| 编号 | 配方 | 解决什么 | 前端跟着要改什么 | 状态 |
|---|---|---|---|---|
| **B0** | `user_id` / `like_count` / `is_liked` | §11.5 坑 1、坑 2；**T18 的前置条件** | `src\types\api.ts` 已补；`PostItem.vue` / `PostDetail.vue` 已接点赞和作者跳转 | ✅ **后端和前端都已落地**（T18 已验收） |

> **配方正文不写在本文件里** —— B0 那四百多行 Go 代码对「写前端的 AI」没用，只会白占上下文。
> 规矩是：**一条配方一个独立文件**，放在 `F:\paopao\web\后端配方-<编号>.md`（B0 的正文就在 `后端配方-B0.md`）。
> **做后端的时候单独开一个对话，把那个文件整个丢给那边的 AI 就行**；前端这边只需要知道
> 「有 B0 这么个东西、它是 T18 的前置、它会让 `types\api.ts` 多三个字段」。

---

## 12. 禁区清单

**未经我在当轮明确要求，不要改：**

- `src\utils\request.ts` —— 401 处理 + 无信封逻辑已正确，改它容易全线出错
- `src\router\index.ts` 的 **已有路由** —— C 档点名要换占位页时可以改那一条，不要改已经在用的
- `src\App.vue` 的三栏结构
- 已完成页面的外观。C 档点名要改某个已有组件的**功能**时，只加那一项，不要顺手重排样式
- `main.ts` 的 **import 顺序**（有讲究，见第 5 节）
- `vite.config.ts` 的 proxy 配置
- `F:\paopao` 里跟当前 C 档那一行无关的 Go 文件

`src\types\api.ts` 和 `src\api\` 在 C 档后端行里必须改，那不算破禁。

**永远不要：**

- 引入新的第三方依赖（先问）
- 用 `any` / `as any`
- 一轮做两个组件，或同一轮里后端和前端一起交
- 说「其余不变」或给省略号的代码
- 自己发明作者没有的交互
- 把新页面做成作者的 Naive UI 样子。外观跟本项目现有页面走（§0.1）
- 抄作者的 `/v1` 路径和 `{code,data,msg}` 信封

---

**最后一句话：如果你读到某条规矩和你的判断冲突——先说我，不要默默绕过。**
