import axios from "axios"
import type { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { clearToken, getToken } from './token'

/** 后端的错误体：主要是 {error}，少数几个接口用 {status} */
interface ApiErrorBody {
  error?: string
  status?: string
}

const service = axios.create({
    /**
   * 写 '/api' 而不是后端地址：请求会先打到 vite dev server，
   * 由你在 vite.config.ts 里配的 proxy 去掉 /api 前缀再转发到 8080。
   *
   * 好处：开发时靠代理绕过跨域，上线时前后端同源 —— 两边都不用改代码。
   */
    baseURL:'/api',
    timeout:15000,
})

/**
 * 【关键设计】401 之后要做什么，由"外面"决定，这里只留一个插座。
 *
 * 为什么不在这个文件里直接 import store？
 *   store/user.ts → api/auth.ts → utils/request.ts，
 *   如果 request.ts 再回头 import store，就是循环依赖（见 token.ts 里的注释）。
 * 所以这里反过来：谁需要知道"掉线了"，谁自己把处理函数塞进来（App.vue 里塞）。
 */

type UnauthorizedHandler = () => void
let unauthorizedHandler: UnauthorizedHandler | null = null

export function setUnauthorizedHandler(handler: UnauthorizedHandler): void {
  unauthorizedHandler = handler
}

/* 请求拦截器：每个请求自动带上 token */
service.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

/* 响应拦截器：把 axios 的"套娃"拆掉 */
service.interceptors.response.use(
  /**
   * 成功分支（2xx 都算，包括 201）。
   *
   * 作者的后端有 {code, data, msg} 信封，所以他写的是
   *   const { data = {}, code = 0 } = response?.data || {}
   *   if (+code === 0) return data
   * 而你的后端没有信封 —— 响应的 body 本身就是数据。
   * 所以我们直接 return response.data，调用方拿到的就是后端那串 JSON。
   */
  (response: AxiosResponse) => response.data,

  /* 失败分支（4xx / 5xx / 网络错误） */
  (error: AxiosError<ApiErrorBody>) => {
    const status = error.response?.status
    const body = error.response?.data
    const message = body?.error || body?.status || error.message || '请求失败'

    /**
     * 【为什么是 status === 401 && hadToken，而不是单纯的 status === 401】
     *
     * 作者的后端不用 401 表达业务错误（他的错误走 {code, msg} 信封 + HTTP 200），
     * 所以在他的项目里 401 只有一个含义：未登录 / token 过期。
     *
     * 但我们的后端把"邮箱或密码错误"也放在了 401 上
     * （F:\paopao\internal\service\service.go:90-93）。
     * 如果不加区分，用户在登录框里输错密码，看到的就是
     * "登录已过期，请重新登录" —— 风马牛不相及，还会顺手清掉已有的登录态。
     *
     * 判断依据来自上面的请求拦截器：它的规则是"有 token 才加 Authorization"，
     * 所以这次请求带没带 token，就等价于"这次请求是否代表了登录态"：
     *   带 token 还收到 401  → token 真的失效了，走下线流程
     *   没带 token 收到 401  → 是接口把业务错误放在了 401 上，当普通错误提示
     * （未登录时才会调的接口 —— /login、/register —— 永远走 else 分支，
     *   于是后端 body 里的 error 文案就能原样显示出来。）
     */
    const hadToken = !!getToken()
    if (status === 401 && hadToken) {
      clearToken()
      unauthorizedHandler?.()
      ElMessage.warning('登录已过期，请重新登录')
    } else {
      ElMessage.error(message)
    }

    // 继续把错误往下抛，页面里的 .catch 仍然能拿到
    return Promise.reject(error)
  },
)

/**
 * 所有 api/ 下的函数都走这一个口子。
 *
 * 泛型 T 是"响应体的类型"，由调用方指定：
 *   request<LoginResult>({ url: '/login', method: 'post', data })
 * 但 axios 自己认为它返回的是 AxiosResponse，类型对不上，
 * 所以最后要 as unknown as 掰一下 —— 作者也是这么处理的。
 */
export function request<T>(config: AxiosRequestConfig): Promise<T> {
  return service.request(config) as unknown as Promise<T>
}

export default service