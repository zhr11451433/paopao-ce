import { ref } from 'vue'
import { defineStore } from 'pinia'

/** 821px 这个断点是作者定的（sidebar.vue / rightbar.vue 里的 media query 都是它），沿用 */
const MOBILE_BREAKPOINT = 821

/** 暗色模式记在 localStorage 的这个键里。作者 sidebar.vue 的 triggerTheme 用的也是 'PAOPAO_THEME' */
const THEME_KEY = 'PAOPAO_THEME'

/** 只有两个值，写成联合类型而不是 string，这样 triggerTheme('darkk') 在编译期就会报错 */
type Theme = 'light' | 'dark'

export const useMainStore = defineStore('main', () => {
  /** 窄屏（手机）模式：左栏和右栏都藏起来，中间栏铺满 */
  const isMobile = ref(document.body.clientWidth <= MOBILE_BREAKPOINT)

  /** 登录/注册弹窗。作者的做法：登录不是独立页面，而是弹窗，所以路由表里没有 /login */
  const authVisible = ref(false)
  const authTab = ref<'signin' | 'signup'>('signin')

  /**
   * 当前主题。初值从 localStorage 读，所以刷新页面不会掉回亮色。
   *
   * ⚠️ 这个 ref 必须写在 setup 里面（T1 修的就是这件事）。
   * 之前它和下面的 applyTheme / triggerTheme 一起写在模块顶层，那它就不属于 store：
   * pinia devtools 看不到它、HMR 不会重建它，而且将来只要出现第二个 pinia 实例
   * （写单元测试时很常见），两个 store 会共用同一个 ref。
   * 从外面读 `mainStore.theme` 一切正常，所以这个错在界面上完全看不出来。
   */
  const theme = ref<Theme>((localStorage.getItem(THEME_KEY) as Theme) || 'light')

  /**
   * 把主题落到 DOM 上。
   * EP 的暗色变量挂在 <html class="dark"> 上（见 element-plus/theme-chalk/dark/css-vars.css），
   * 作者是挂在自己的 .app-container 上 —— 所以必须是 documentElement，不能是 app-container，
   * 否则 EP 那些组件（弹窗、下拉）根本读不到暗色变量（它们 teleport 到 body）。
   */
  function applyTheme(): void {
    document.documentElement.classList.toggle('dark', theme.value === 'dark')
  }

  function triggerTheme(next: Theme): void {
    theme.value = next
    localStorage.setItem(THEME_KEY, next)
    applyTheme()
  }

  function updateIsMobile(): void {
    isMobile.value = document.body.clientWidth <= MOBILE_BREAKPOINT
  }

  function openAuth(tab: 'signin' | 'signup' = 'signin'): void {
    authTab.value = tab
    authVisible.value = true
  }

  function closeAuth(): void {
    authVisible.value = false
  }

  /**
   * store 一建好就先应用一次。
   *
   * 搬进来之后这一行从"模块加载时执行"变成了"第一次 useMainStore() 时执行"，
   * 但你不会看出区别：第一次调用它的是 App.vue 的 setup，那时才 .mount()，
   * 而模块加载和 mount 之间没有浏览器重绘的机会。
   */
  applyTheme()

  return {
    isMobile,
    authVisible,
    authTab,
    theme,
    updateIsMobile,
    openAuth,
    closeAuth,
    triggerTheme,
  }
})