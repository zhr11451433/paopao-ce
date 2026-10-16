import { request } from '@/utils/request'
import type {
  CommentListResult,
  CreatePostPayload,
  CreatePostResult,
  FeedResult,
  LikeCountResult,
  MyLikeItem,
  MyPostListResult,
  PageParams,
  PostDetail,
  PostListParams,
  PostListResult,
  StatusResult,
} from '@/types/api'

/* ---------- 帖子 ---------- */

/** 广场列表：GET /posts（无需登录） */

export function getPosts(params: PostListParams = {}): Promise<PostListResult> {
  return request<PostListResult>({ url: '/posts', method: 'get', params })
}

/** 我的帖子：GET /posts/me（需登录，且没有 author 字段） */
export function getMyPosts(params: PageParams = {}): Promise<MyPostListResult> {
  return request<MyPostListResult>({ url: '/posts/me', method: 'get', params })
}

/** 帖子详情：GET /posts/:id */
export function getPost(id: number): Promise<PostDetail> {
  return request<PostDetail>({ url: `/posts/${id}`, method: 'get' })
}

/** 发帖：POST /posts → 201 {post_id} */
export function createPost(data: CreatePostPayload): Promise<CreatePostResult> {
  return request<CreatePostResult>({ url: '/posts', method: 'post', data })
}

/** 改帖：PUT /posts/:id */
export function updatePost(id: number, data: CreatePostPayload): Promise<StatusResult> {
  return request<StatusResult>({ url: `/posts/${id}`, method: 'put', data })
}

/** 删帖：DELETE /posts/:id */
export function deletePost(id: number): Promise<StatusResult> {
  return request<StatusResult>({ url: `/posts/${id}`, method: 'delete' })
}

/* ---------- 评论 ---------- */

/** 评论列表：GET /posts/:id/comments */
export function getComments(id: number, params: PageParams = {}): Promise<CommentListResult> {
  return request<CommentListResult>({ url: `/posts/${id}/comments`, method: 'get', params })
}

/** 发评论：POST /posts/:id/comments → 201 */
export function createComment(id: number, content: string): Promise<StatusResult> {
  return request<StatusResult>({ url: `/posts/${id}/comments`, method: 'post', data: { content } })
}

/** 删评论：DELETE /comments/:id */
export function deleteComment(commentId: number): Promise<StatusResult> {
  return request<StatusResult>({ url: `/comments/${commentId}`, method: 'delete' })
}

/* ---------- 点赞 ---------- */

/** 点赞：POST /posts/:id/likes —— 首次 201「点赞成功」，重复 200「已点赞」，都算成功 */
export function likePost(id: number): Promise<StatusResult> {
  return request<StatusResult>({ url: `/posts/${id}/likes`, method: 'post' })
}

/** 取消点赞：DELETE /posts/:id/likes */
export function unlikePost(id: number): Promise<StatusResult> {
  return request<StatusResult>({ url: `/posts/${id}/likes`, method: 'delete' })
}

/** 点赞数：GET /posts/:id/likes */
export function getLikeCount(id: number): Promise<LikeCountResult> {
  return request<LikeCountResult>({ url: `/posts/${id}/likes`, method: 'get' })
}

/** 我赞过的帖子：GET /posts/me/likes —— ⚠️ 裸数组，没有 list 包装 */
export function getMyLikes(): Promise<MyLikeItem[]> {
  return request<MyLikeItem[]>({ url: '/posts/me/likes', method: 'get' })
}

/* ---------- Feed 流 ---------- */

/** 关注流：GET /feed（需登录）—— ⚠️ key 是 posts，不是 list */
export function getFeed(params: PageParams = {}): Promise<FeedResult> {
  return request<FeedResult>({ url: '/feed', method: 'get', params })
}