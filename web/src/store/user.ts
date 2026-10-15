import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { getMe } from '@/api/auth'
import {
  clearToken,
  getCachedUser,
  getToken,
  setCachedUser,
  setToken,
} from '@/utils/token'
import type { CachedUser } from '@/utils/token'

export const useUserStore = defineStore('user', () => {
  /** 当前用户。null = 未登录。初始值从 localStorage 里捞，刷新页面不掉登录态 */
  const user = ref<CachedUser | null>(getCachedUser())
  const role = ref('')

  const logined = computed(() => user.value !== null)

  /** 登录成功后调用 */
  function setLogin(token: string, info: CachedUser): void {
    setToken(token)
    setCachedUser(info)
    user.value = info
  }

  function logout(): void {
    clearToken()
    user.value = null
    role.value = ''
  }

  /**
   * 有 token 时调一次 /me：
   *  - token 还有效 → 顺手拿到 role，和后端认定的 user_id（比缓存的更权威）
   *  - token 已失效 → 后端回 401，拦截器会清 token，这里再把自己也清干净
   *
   * 没有 token 时直接返回，不发这个请求 —— 否则每次打开页面都会弹一个 401 提示。
   */
  async function fetchMe(): Promise<void> {
    if (!getToken()) {
      logout()
      return
    }
    try {
      const res = await getMe()
      role.value = res.role
      if (user.value) {
        user.value.id = res.user_id
      } else {
        // token 在、缓存没了（比如清了 localStorage 的一半），只能先给个空壳
        user.value = { id: res.user_id, name: '', email: '' }
      }
    } catch {
      logout()
    }
  }

  return { user, role, logined, setLogin, logout, fetchMe }
})