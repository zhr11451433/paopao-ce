import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { getToken } from '@/utils/token'
import { useMainStore } from '@/store/main'

/**
 * 给路由的 meta 补上类型。
 *
 * vue-router 默认把 meta 定成 Record<string | symbol, unknown>，
 * 于是 `to.meta.title` 是 unknown，拼字符串时 TS 会报错。
 * 这个 declare module 是官方的扩展方式，写一次到处受益。
 */
declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    /** true 表示这个页面要缓存（见 App.vue 的 <keep-alive>） */
    keepAlive?: boolean
    /** true 表示需要登录 */
    requiresAuth?: boolean
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue'),
    // 广场是唯一被 keep-alive 缓存的页面：
    // 从广场点进详情、再返回，列表和滚动位置都还在，不会重新请求一遍
    meta: { title: '广场', keepAlive: true },
  },

  // ↓↓↓ 下面 5 条是这次新加的，都先指向占位页 ↓↓↓
  {
    path: '/topic',
    name: 'topic',
    component: () => import('@/views/PlaceholderView.vue'),
    meta: { title: '话题' }, // 未登录也能看，和原作者一致
  },
  {
    path: '/messages',
    name: 'messages',
    component: () => import('@/views/PlaceholderView.vue'),
    meta: { title: '消息', requiresAuth: true },
  },
  {
    path: '/collection',
    name: 'collection',
    component: () => import('@/views/PlaceholderView.vue'),
    meta: { title: '收藏', requiresAuth: true },
  },
  {
    path: '/contacts',
    name: 'contacts',
    component: () => import('@/views/PlaceholderView.vue'),
    meta: { title: '好友', requiresAuth: true },
  },
  {
    path: '/setting',
    name: 'setting',
    component: () => import('@/views/PlaceholderView.vue'),
    meta: { title: '设置', requiresAuth: true },
  },
  // ↑↑↑ 新加的结束 ↑↑↑

  {
    // 作者的详情页是 /post?id=xxx（query 传参），我们改成 /post/:id（路径传参）。
    // 理由：RESTful 后端本来就是 GET /posts/:id，路径传参语义更直白，刷新也不丢。
    path: '/post/:id',
    name: 'post',
    component: () => import('@/views/PostDetailView.vue'),
    meta: { title: '泡泡详情' },
  },
  {
    path: '/profile',
    name: 'profile',
    component: () => import('@/views/ProfileView.vue'),
    meta: { title: '我的主页', requiresAuth: true },
  },
  {
    path: '/user/:id',
    name: 'user',
    component: () => import('@/views/UserView.vue'),
    meta: { title: '用户主页' },
  },
  {
    // 作者 `router\index.ts:77-84`：路径和 name 都叫 following，标题「关注」，
    // **没有 requiresAuth** —— 关注/粉丝列表本身是公开的，未登录也能看。
    // 只是未登录时点「关注」会 401（由 utils\request.ts 统一处理）。
    path: '/following',
    name: 'following',
    component: () => import('@/views/FollowingView.vue'),
    meta: { title: '关注' },
  },
  {
    path: '/404',
    name: 'notFound',
    component: () => import('@/views/NotFoundView.vue'),
    meta: { title: '404' },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404',
  },
]

const router = createRouter({
  /**
   * 用 hash 模式（URL 形如 /#/post/1），和作者一致。
   *
   * 为什么不用 createWebHistory？
   * 历史模式下 /post/1 是真的路径，刷新页面时浏览器会向服务器要 /post/1，
   * 而 Go 后端没配 SPA 兜底路由，会直接 404。hash 模式刷新永远只要 /，
   * 不需要后端配合。
   */
  history: createWebHashHistory(),
  routes,
})

router.beforeEach((to, from) => {
  document.title = `${to.meta.title ?? '泡泡'} | 泡泡 - 一个清新文艺的微社区`

  /**
   * 登录守卫（T3）。
   *
   * ⚠️ 先说清楚：**作者原版没有这个东西**。
   * 他的 `router/index.ts:128-131` 里 beforeEach 只写 document.title 然后 next()，
   * 从不拦人；未登录进不去「主页 / 消息」是因为他的左栏只渲染「广场 / 话题」两项
   * （`components/sidebar.vue:135-223`），人根本点不到，而不是路由拦下来的。
   * 所以这条守卫是**本项目自己加的防线**，拦的是「手敲地址 / 后退按钮 / 旧收藏夹」这几条路。
   *
   * 两个容易踩的坑：
   *
   * 1) `useMainStore()` 必须写在守卫**函数体里面**，不能提到模块顶层。
   *    pinia 的 store 需要一个「当前激活的 pinia 实例」才能创建，而模块顶层执行时
   *    main.ts 还没走到 `app.use(createPinia())`，会直接抛
   *    "getActivePinia() was called but there was no active Pinia"。
   *    守卫是真正发生导航时才运行的，那时 pinia 早装好了。
   *
   * 2) 冷启动时必须给个去处，不能只 `return false`。
   *    如果直接刷新 /#/profile，这是本次会话的**首次导航**，被拦掉的话
   *    app 里一个路由组件都不会渲染 —— 左栏右栏还在，中栏会是一片空白。
   *    首次导航时 vue-router 传进来的 `from` 是 START_LOCATION，它的 `name` 是 undefined，
   *    正好可以用它区分「站内点击被拦」和「冷启动被拦」。
   */
  if (to.meta.requiresAuth && !getToken()) {
    useMainStore().openAuth('signin')
    // 站内点击：留在当前页（本来就是能看的页面），只把登录框弹出来
    // 冷启动：退回广场，否则中栏空白
    return from.name ? false : { name: 'home' }
  }
})

export default router