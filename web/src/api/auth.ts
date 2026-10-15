import { request } from '@/utils/request'
import type {
  LoginPayload,
  LoginResult,
  MeResult,
  RegisterPayload,
  RegisterResult,
} from '@/types/api'

/** 登录：POST /login → {token, user:{id,name,email}} */
export function login(data: LoginPayload): Promise<LoginResult> {
  return request<LoginResult>({ url: '/login', method: 'post', data })
}

/** 注册：POST /register → {status, id, name, email} */
export function register(data: RegisterPayload): Promise<RegisterResult> {
  return request<RegisterResult>({ url: '/register', method: 'post', data })
}

/** 当前登录者：GET /me → {user_id, role} */
export function getMe(): Promise<MeResult> {
  return request<MeResult>({ url: '/me', method: 'get' })
}