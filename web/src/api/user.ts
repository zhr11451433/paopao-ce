import { request } from '@/utils/request'
import type {
  FollowerResult,
  FollowingResult,
  PageParams,
  StatusResult,
  UserProfile,
} from '@/types/api'

/** 某人的公开资料：GET /users/:id。不用登录 */
export function getUser(id: number): Promise<UserProfile> {
  return request<UserProfile>({ url: `/users/${id}`, method: 'get' })
}

/** 某人关注了谁：GET /users/:id/following */
export function getFollowing(id: number, params: PageParams = {}): Promise<FollowingResult> {
  return request<FollowingResult>({ url: `/users/${id}/following`, method: 'get', params })
}

/** 谁关注了某人：GET /users/:id/followers */
export function getFollowers(id: number, params: PageParams = {}): Promise<FollowerResult> {
  return request<FollowerResult>({ url: `/users/${id}/followers`, method: 'get', params })
}

/** 关注：POST /users/:id/follow —— 首次 201，重复 200「您已关注该用户」 */
export function followUser(id: number): Promise<StatusResult> {
  return request<StatusResult>({ url: `/users/${id}/follow`, method: 'post' })
}

/** 取关：DELETE /users/:id/follow */
export function unfollowUser(id: number): Promise<StatusResult> {
  return request<StatusResult>({ url: `/users/${id}/follow`, method: 'delete' })
}