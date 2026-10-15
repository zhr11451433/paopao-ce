/**
 * 登录令牌的本地存储。
 *
 * 【为什么要单独一个文件】request.ts 需要读 token，store/user.ts 也需要读写 token。
 * 如果 request.ts 直接 import store，而 store 又要 import api、api 又要 import request，
 * 就形成了循环依赖（util → store → api → util），打包时很容易变成"用的时候是 undefined"。
 * 把"钥匙放哪"这件事抽成谁都不依赖的最小模块，环就断了。
 */

export const TOKEN_KEY = 'PAOPAO_TOKEN'
/** 缓存登录用户，刷新页面后不用等 /me 返回就能显示昵称 */
export const USER_KEY = 'PAOPAO_USER'

export interface CachedUser {
    id: number
  name: string
  email: string
}

export function getToken():string | null{
    return localStorage.getItem(TOKEN_KEY) //浏览器自带的一个“本地持久化存储”工具。它是一个纯键值对（Key-Value）数据库，数据存在用户的硬盘上。只要用户不主动清理浏览器缓存，关掉浏览器再打开，数据依然存在
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function getCachedUser():CachedUser|null{
    const raw = localStorage.getItem(USER_KEY)
    if (!raw){
        return null
    }
    try{
        //json.Unmarshal()（把 JSON 字符串 -> 结构体对象
        return JSON.parse(raw) as CachedUser
    }catch{
        // localStorage 里的东西可能是脏的（用户手改过、或旧版本写进去的），解析失败就当没有
        return null
    }
}

export function setCachedUser(user: CachedUser): void {
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}