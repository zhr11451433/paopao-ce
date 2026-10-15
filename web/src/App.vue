<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useMainStore } from '@/store/main'
import { useUserStore } from '@/store/user'
import { setUnauthorizedHandler } from '@/utils/request'
import AppSidebar from '@/components/AppSidebar.vue'
import AppRightbar from '@/components/AppRightbar.vue'
import AuthDialog from '@/components/AuthDialog.vue'

const mainStore = useMainStore()
const userStore = useUserStore()
const { isMobile } = storeToRefs(mainStore)

/**
 * 把 request.ts 那个"插座"接上：
 * 任何请求被判 401 时，这里把自己这边也清干净，界面立刻回到未登录状态。
 * request.ts 不认识 store，是 store 这边主动接线的。
 */
setUnauthorizedHandler(() => {
  userStore.logout()
})

function handleResize(): void {
  mainStore.updateIsMobile()
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
  // 带着本地 token 进来时，验一次票（顺便拿 role）
  userStore.fetchMe()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="app-container">
    <div class="main-wrap">
      <!-- 左栏：作者用 v-if="desktopModelShow"，我们语义化成 !isMobile -->
      <AppSidebar v-if="!isMobile" />

      <div class="content-wrap">
        <!--
          作者的 keep-alive 写法，照搬：
          <keep-alive> 不能直接裹 <router-view>（那样会把所有页面都缓存），
          要用 v-slot 拿到当前组件，只缓存 meta.keepAlive 为 true 的那个。
        -->
        <router-view v-slot="{ Component }">
          <keep-alive>
            <component :is="Component" v-if="$route.meta.keepAlive" />
          </keep-alive>
          <component :is="Component" v-if="!$route.meta.keepAlive" />
        </router-view>
      </div>

      <AppRightbar v-if="!isMobile" />
    </div>

    <!--
      常驻的登录/注册弹窗，对应原版 App.vue 里那个 <auth />。
      放在 .main-wrap 外面：它是 position: fixed 的遮罩，不属于三栏布局。
      「登录按钮点了没反应」这个空枪（PROJECT.md 第 11 节）就是靠这一行接上的。
    -->
    <AuthDialog />
  </div>
</template>

<style scoped></style>
