/**
 * 后端所有接口的出入参类型。
 * 一份"契约"集中放这里，api/ 下的函数只引用类型、不重复定义。
 * 注意：文件里全是 interface，所以任何地方 import 它都必须用 `import type`。
 */

/* ========== 通用 ========== */

/** 所有列表接口都支持的分页参数（默认 page=1、page_size=10，上限 100） */
export interface PageParams {
  page?: number
  page_size?: number
}

/**
 * 广场列表 GET /posts 的查询参数。
 * `q` 按标题和正文包含匹配。`style=hots` 按点赞数排序。
 * 别的列表接口没有这两个参数，所以不放进 PageParams。
 */
export interface PostListParams extends PageParams {
  /** 按标题和正文包含匹配。不传或空字符串就是全部帖子 */
  q?: string
  /**
   * `hots` 按点赞数从高到低。不传就是按发帖时间从新到旧。
   * 正在关注不走这个参数，它是 GET /feed。
   */
  style?: 'hots'
}

/** 图片：上传接口返回的、以及帖子列表里带的，是同一个形状 */
export interface ImageItem {
  id: number
  url: string
}

/* ========== 帖子 ========== */

/**
 * 帖子列表项 —— GET /posts、GET /feed。
 *
 * `user_id` / `like_count` / `is_liked` 广场、关注流、`GET /users/:id/posts`
 * 以及 `GET /posts/me` 都有。`GET /posts/me` 的正式类型是下面的 `MyPostListItem`，
 * 没有 `author`。我的主页补上作者名后再当成 `PostListItem` 用，
 * 所以这三项在这里仍是可选的，避免手写的列表项必须填点赞数。
 */
export interface PostListItem {
  id: number
  title: string
  /** Go 的 time.Time 序列化成 RFC3339 字符串，前端要自己格式化 */
  created_at: string
  author: string
  comment_count: number
  /** ⚠️ 可能是 null：Go 的 nil 切片序列化出来就是 null，不是 [] */
  images: ImageItem[] | null
  /** 作者 id。有它才能点进 `/user/:id` */
  user_id?: number
  /** 点赞人数 */
  like_count?: number
  /** 当前登录用户赞过没有。未登录时后端给 false */
  is_liked?: boolean
  /**
   * 当前登录用户是否关注了作者。
   * 广场、关注流、别人的帖子、详情都有。我的主页自己拼的卡片不带，所以可选。
   */
  is_following?: boolean
}

/** 我的帖子 —— GET /posts/me。没有 author，名字由「我的主页」自己填 */
export interface MyPostListItem {
  id: number
  title: string
  created_at: string
  user_id: number
  comment_count: number
  like_count: number
  is_liked: boolean
  images: ImageItem[] | null
}

/** 帖子详情 —— GET /posts/:id。这三个字段详情接口是必有的（配方 B0） */
export interface PostDetail {
  id: number
  title: string
  content: string
  created_at: string
  author: string
  /** 作者 id，用来跳主页，也用来判断「这条是不是我发的」 */
  user_id: number
  /** 点赞人数 */
  like_count: number
  /** 当前登录用户赞过没有。未登录时后端给 false */
  is_liked: boolean
  /** 当前登录用户是否关注了作者。未登录或看自己的帖子时是 false */
  is_following: boolean
  comments: CommentItem[] | null
  images: ImageItem[] | null
}

/**
 * 广场 `GET /posts` 和某人的帖子 `GET /users/:id/posts` 用同一个外壳。
 * ⚠️ 六个接口的列表 key 各不相同，所以下面四个外壳不能合并成一个泛型。
 */
export interface PostListResult {
  total: number
  page: number
  page_size: number
  list: PostListItem[]
}

export interface MyPostListResult {
  total: number
  page: number
  page_size: number
  list: MyPostListItem[]
}

/** GET /feed 的 key 是 posts，不是 list */
export interface FeedResult {
  total: number
  page: number
  page_size: number
  posts: PostListItem[]
}

export interface CommentListResult {
  total: number
  page: number
  page_size: number
  comments: CommentItem[]
}

export interface CommentItem {
  id: number
  author: string
  created_at: string
  content: string
}

/** GET /posts/:id/likes —— 这是"点赞数"，不是点赞的人名单 */
export interface LikeCountResult {
  id: number
  title: string
  like_count: number
}

/** GET /posts/me/likes —— 裸数组，没有 list 包装 */
export interface MyLikeItem {
  author: string
  title: string
  id: number
}

/* ========== 用户与关注 ========== */

export interface LoginPayload {
  email: string
  password: string
}

export interface RegisterPayload {
  name: string
  email: string
  password: string
}

/** ⚠️ 登录响应是嵌套的：{token, user:{id,name,email}}，不是平的 */
export interface LoginResult {
  token: string
  user: {
    id: number
    name: string
    email: string
  }
}

export interface RegisterResult {
  status: string
  id: number
  name: string
  email: string
}

/** GET /me 只给 user_id 和 role，不给昵称邮箱 */
export interface MeResult {
  user_id: number
  role: string
}

/**
 * GET /users/:id 的公开资料（C1）。
 * 没有邮箱。关注数是他关注了多少人，粉丝数是多少人关注他。
 */
export interface UserProfile {
  id: number
  name: string
  role: string
  following_count: number
  follower_count: number
  post_count: number
}

export interface FollowItem {
  follow_id: number
  name: string
}

export interface FanItem {
  fan_id: number
  name: string
}

export interface FollowingResult {
  total: number
  page: number
  page_size: number
  follows: FollowItem[]
}

export interface FollowerResult {
  total: number
  page: number
  page_size: number
  fans: FanItem[]
}

/* ========== 写操作 ========== */

export interface CreatePostPayload {
  title: string
  content: string
  /** 图片 id 数组。不传 / null = 保留原图；[] = 解绑全部；[1,2] = 换成这两张 */
  images?: number[] | null
}

export interface CreatePostResult {
  post_id: number
}

/**
 * 后端有一批接口成功时只回一句中文状态，而且形状还不一样：
 * 多数是 {"status": "..."}，少数是 {"error": "..."} 但状态码仍是 200。
 * 统一用这个类型接。
 */
export interface StatusResult {
  status?: string
  error?: string
}

export interface UploadResult {
  images: ImageItem[]
}

