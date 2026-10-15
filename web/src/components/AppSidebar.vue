<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ElMessageBox } from 'element-plus'
import {
  BookmarksOutline,
  ChatbubblesOutline,
  HomeOutline,
  LeafOutline,
  LogOutOutline,
  PeopleOutline,
  SettingsOutline,
} from '@vicons/ionicons5'
// 话题的 # 图标：原作者用的就是 tabler 的 Hash（sidebar.vue:5、:204）
import { Hash } from '@vicons/tabler'
import { useMainStore } from '@/store/main'
import { useUserStore } from '@/store/user'

/**
 * `drawer` —— 同一个组件既当桌面端左栏、又当移动端抽屉里的内容。
 * 作者就是这么复用的：`components\main-nav.vue:10` 抽屉里的 `<sidebar />`
 * 和桌面端用的是同一个 `components\sidebar.vue`。
 *
 * 两种场合只差排布：桌面上它是 `position: fixed` 钉在「视口中心往左偏」的位置，
 * 抽屉里没有这个概念，要回到正常文档流。所以不用两套模板，只切一个类名。
 */
withDefaults(defineProps<{ drawer?: boolean }>(), { drawer: false })

const route = useRoute()
const router = useRouter()
const mainStore = useMainStore()
const userStore = useUserStore()
const { user, logined } = storeToRefs(userStore)

/**
 * 菜单。作者 sidebar.vue:135-223 是分登录态的：
 *   未登录 → 广场 / 话题
 *   登录后 → 广场 / 话题 / [公告] / 主页 / 消息 / 收藏 / [好友] / [钱包] / 设置 / [系统配置]
 * 方括号是受配置开关控制的，我们按演示站取中间那 7 项。
 *
 * ⚠️ 这两个数组故意写在模块级、是普通数组：
 * 组件对象一旦进了 ref()/reactive()，Vue 会套 Proxy 并警告
 * "Vue received a Component that was made a reactive object"。
 * 下面用 computed 在两者之间切 —— computed 只把 getter 的返回值原样交出来，
 * 不做深包装，所以 icon 不会被 Proxy 掉，是安全的。
 */
const PUBLIC_MENU = [
  { name: 'home', label: '广场', icon: HomeOutline },
  { name: 'topic', label: '话题', icon: Hash },
]

const USER_MENU = [
  { name: 'home', label: '广场', icon: HomeOutline },
  { name: 'topic', label: '话题', icon: Hash },
  { name: 'profile', label: '主页', icon: LeafOutline },
  { name: 'messages', label: '消息', icon: ChatbubblesOutline },
  { name: 'collection', label: '收藏', icon: BookmarksOutline },
  { name: 'contacts', label: '好友', icon: PeopleOutline },
  { name: 'setting', label: '设置', icon: SettingsOutline },
]

const menuItems = computed(() => (logined.value ? USER_MENU : PUBLIC_MENU))

/** el-menu 高亮项：跟着当前路由的 name 走 */
const activeName = computed(() => String(route.name ?? ''))

/** 没有头像图时，拿昵称/邮箱首字母当占位 */
const avatarText = computed(() => {
  const name = user.value?.name || user.value?.email || ''
  return name.charAt(0).toUpperCase() || '泡'
})

/**
 * 跳转。这是原作者 sidebar.vue:258-266 的一个小技巧：
 * 每次都往 query 上挂一个当前时间戳 t。
 *
 * 为什么？菜单项指向的路由可能就是你当前所在的页，不加 t 的话
 * vue-router 认为"目标一样"，直接忽略这次点击，页面没有任何反应。
 * 加了 t 每次都是新地址，导航一定会触发，页面就能借这次导航重新拉数据。
 */
function goRouter(name: string): void {
  router.push({ name, query: { t: Date.now() } })
}

/** 点 logo 回广场；已经在广场上就当作"刷新一下" */
function goHome(): void {
  goRouter('home')
}

async function handleLogout(): Promise<void> {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '退出',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    // 点了取消，ElMessageBox 会 reject，什么都不用做
    return
  }
  userStore.logout()
  goRouter('home')
}
</script>

<template>
  <aside class="sidebar-wrap" :class="{ 'is-drawer': drawer }">
    <div class="logo-wrap">
      <!-- public/ 里的文件用根路径引用，不经 vite 打包 -->
      <img
        class="logo-img"
        src="/paopao-logo.png"
        width="36"
        alt="泡泡"
        @click="goHome"
      />
    </div>

    <el-menu class="sidebar-menu" :default-active="activeName" @select="goRouter">
      <el-menu-item v-for="item in menuItems" :key="item.name" :index="item.name">
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.label }}</span>
      </el-menu-item>
    </el-menu>

    <div class="user-wrap">
      <template v-if="logined">
        <el-avatar :size="34" class="user-avatar">{{ avatarText }}</el-avatar>
        <div class="user-info">
          <div class="nickname">
            <span class="nickname-txt">{{ user?.name || '泡泡用户' }}</span>
            <el-button class="logout-btn" text circle size="small" @click="handleLogout">
              <el-icon :size="18"><LogOutOutline /></el-icon>
            </el-button>
          </div>
          <div class="email">{{ user?.email }}</div>
        </div>
      </template>

      <template v-else>
        <el-button type="primary" round @click="mainStore.openAuth('signin')">登录</el-button>
        <el-button round @click="mainStore.openAuth('signup')">注册</el-button>
      </template>
    </div>
  </aside>
</template>

<style scoped lang="scss">
/**
 * 定位方式和作者逐行一致（sidebar.vue:293-303）：
 * 不参与 flex 排版，而是 fixed 钉在"视口中心往左偏"的位置，
 *   右边缘 = 50% + 中间栏一半 + 10px
 * 这样中间栏永远居中，左栏自动贴着它左边，窗口怎么缩都不错位。
 */
.sidebar-wrap {
  z-index: 99;
  width: 200px;
  height: 100vh;
  position: fixed;
  right: calc(50% + var(--content-main) / 2 + 10px);
  padding: 12px 0;
  box-sizing: border-box;
  max-height: 100vh;
  overflow: auto;

  &::-webkit-scrollbar {
    width: 0;
    height: 0;
  }

  /* 作者写法：logo 由 .logo-img 自己 margin-left: 24px（sidebar.vue:310-322） */
  .logo-wrap {
    display: flex;
    justify-content: flex-start;
    margin-bottom: 12px;

    .logo-img {
      margin-left: 24px;

      &:hover {
        cursor: pointer;
      }
    }
  }

  /**
   * 尺寸全部对齐作者那边的实测值（n-menu-item-content = 200 × 42，
   * padding: 0 18px 0 32px）和源码里的 :icon-size="24"：
   *
   *   高度 42  ← 作者 42
   *   左 32    ← 作者 32（EP 默认只有 20，图标会贴着左边缘）
   *   右 18    ← 作者 18
   *   margin 0 ← 作者项与项之间没有间距，naive-ui 靠 21px 圆角形成视觉间隔
   *   字号 14  ← naive-ui 中号菜单默认值，作者没有改过
   */
  .sidebar-menu {
    border-right: none;
    background-color: transparent;

    :deep(.el-menu-item) {
      height: 42px;
      line-height: 42px;
      padding: 0 18px 0 32px;
      margin: 0;
      border-radius: 21px;
      font-size: 14px;
    }

    /* 图标 24px，对齐作者的 :icon-size="24"（sidebar.vue:6） */
    :deep(.el-menu-item .el-icon) {
      font-size: 24px;
    }

    :deep(.el-menu-item.is-active) {
      background-color: var(--el-color-primary-light-9);
      font-weight: bold;
    }
  }

  .user-wrap {
    display: flex;
    align-items: center;
    position: absolute;
    bottom: 12px;
    left: 12px;
    right: 12px;

    .user-avatar {
      margin-right: 8px;
    }

    .user-info {
      flex: 1;
      min-width: 0; /* flex 子项要能收缩，里面的 ellipsis 才会生效 */
      display: flex;
      flex-direction: column;

      .nickname {
        display: flex;
        align-items: center;
        height: 16px;
        line-height: 16px;
        font-size: 15px;
        font-weight: bold;
        margin-bottom: 2px;

        .nickname-txt {
          max-width: 90px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }

        .logout-btn {
          margin-left: 6px;
        }
      }

      .email {
        font-size: 13px;
        line-height: 16px;
        height: 16px;
        max-width: 120px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        opacity: 0.75;
      }
    }
  }

  /**
   * 抽屉里的排布（T13，配套 MobileNav.vue）。
   *
   * 上面那条 `position: fixed; right: calc(50% + var(--content-main) / 2 + 10px)`
   * 是桌面端的定位法：不参与 flex 排版，而是钉在「中间栏左边 10px」，
   * 这样中间栏永远居中、左栏自动贴着它。抽屉里根本没有「中间栏」这个概念，
   * 那条就失效了，换成正常文档流：撑满抽屉、从左上角开始排。
   *
   * 「用户块贴在抽屉底部」不用另外写：`.user-wrap` 是 absolute + bottom: 12px，
   * 这里把容器改成 position: relative 之后，它自然就贴抽屉底部了。
   */
  &.is-drawer {
    position: relative;
    right: auto;
    width: 100%;
    height: 100%;

    /*
     * 作者 `sidebar.vue:406-416` 的 `@media screen and (max-width: 821px)` 里有一条
     * `.logo-img { margin-left: 12px !important }` —— 窄屏时 logo 左边距从 24 收到 12。
     * 我们桌面端左栏在窄屏下是直接不渲染的（App.vue 里 `v-if="!isMobile"`），
     * 所以那条只在抽屉里才有意义，搬到这里。
     */
    .logo-img {
      margin-left: 12px;
    }
  }
}
</style>