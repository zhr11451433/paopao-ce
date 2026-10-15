<script setup lang="ts">
import { ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { ChevronBackOutline, MenuOutline, MoonOutline, SunnyOutline } from '@vicons/ionicons5'
import { useMainStore } from '@/store/main'
import MobileNav from '@/components/MobileNav.vue'

/**
 * 作者 `components\main-nav.vue:21-27` 的 props 是三个：`title`、`back?`、`theme?`（默认 true）。
 *
 * - `title`  —— 顶部卡片上的字（「泡泡广场」「泡泡详情」……）
 * - `back`   —— 要不要在左边显示一个返回箭头。详情页传 true（`views\Post.vue:3`），
 *               列表页不传（`views\Home.vue:5` 只传了 title）。
 * - `theme`  —— 要不要显示右边的主题开关。**本项目没有做这个 prop**，
 *               因为我们的主题开关是全局唯一的，还没有哪个页面需要把它藏起来。
 *               等真遇到（作者的移动端抽屉里会藏起来）再补，先记在 PROJECT.md 11.6。
 */
withDefaults(defineProps<{ title?: string; back?: boolean }>(), {
  title: '',
  back: false,
})

const mainStore = useMainStore()
const { theme, isMobile } = storeToRefs(mainStore)
const router = useRouter()

/**
 * 移动端抽屉开着没有。
 *
 * 作者把这个状态放在**导航条组件自己**里（`main-nav.vue:80` 的 `activeDrawerRef`），
 * 抽屉也画在这个组件的模板里 —— 我们也一样，所以没往 store 里加东西。
 */
const drawerVisible = ref(false)

/**
 * 照抄作者 `main-nav.vue` 的 `goBack()`（162 行那个文件里的写法）：
 *
 *   window.history.length <= 1 ? router.push({ path: '/' }) : router.go(-1)
 *
 * 为什么不能直接 `router.go(-1)`：**用户在地址栏里直接打开 /#/post/3、或者从收藏夹
 * 点进来时，浏览器的历史里没有上一页**，`go(-1)` 会什么都不做（或者把人送出站），
 * 页面就卡在详情页出不去了。`history.length <= 1` 就是用来识别这种情况的，
 * 那时候退回到广场页。
 */
function goBack(): void {
  if (window.history.length <= 1) {
    void router.push({ path: '/' })
    return
  }
  router.go(-1)
}
</script>

<template>
  <!--
    作者 `main-nav.vue:2-13`：抽屉就画在导航条组件自己的模板里（里面装的是 `<sidebar />`）。
    桌面端不渲染它 —— 作者的条件是 `drawerModelShow`，我们换成 `isMobile`，同一件事。
  -->
  <MobileNav v-if="isMobile" v-model="drawerVisible" />

  <div class="nav-title-card">
    <div class="navbar">
      <!--
        作者 `main-nav.vue:17-28`：移动端在最左边放一个汉堡按钮，点开抽屉。
        条件是 `drawerModelShow && !back` —— **详情页（back=true）左边只留返回箭头**，
        不再放汉堡。这一点照抄，没有自作主张。
        图标：作者用的 `dehaze-round` 来自 `@vicons/material`（本项目没装那个包，§2），
        换成 `@vicons/ionicons5` 的 `MenuOutline` —— 同样是「三条横线」。
      -->
      <el-button
        v-if="isMobile && !back"
        class="drawer-btn"
        text
        circle
        size="small"
        title="菜单"
        @click="drawerVisible = true"
      >
        <el-icon :size="20"><MenuOutline /></el-icon>
      </el-button>

      <!--
        作者 `main-nav.vue:29-40`：
          <n-button class="back-btn" v-if="back" @click="goBack" quaternary circle size="small">
              <template #icon><n-icon><chevron-left-round /></n-icon></template>
          </n-button>
        他那个 `chevron-left-round` 来自 `@vicons/material`，本项目没装那个包（§2），
        换成 `@vicons/ionicons5` 的 `ChevronBackOutline` —— 同一个「向左的尖括号」。
      -->
      <el-button v-if="back" class="back-btn" text circle size="small" @click="goBack">
        <el-icon><ChevronBackOutline /></el-icon>
      </el-button>

      <span class="nav-title">{{ title }}</span>

      <el-switch
        class="theme-switch-wrap"
        size="small"
        inline-prompt
        :model-value="theme === 'dark'"
        :active-icon="MoonOutline"
        :inactive-icon="SunnyOutline"
        @update:model-value="(v: string | number | boolean) => mainStore.triggerTheme(v ? 'dark' : 'light')"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
.nav-title-card {
  z-index: 99;
  width: 100%;
  position: sticky;
  top: 0;
  box-sizing: border-box;
  border: 1px solid var(--el-border-color);
  border-radius: 0;
  background-color: rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(12px);
}

.navbar {
  height: 54px;              /* 作者：30px 行高 + n-card 上下各 12px；我们不用 el-card，直接写总高 */
  box-sizing: border-box;
  padding: 0 20px;
  position: relative;
  display: flex;
  align-items: center;
}

/* 作者 `main-nav.vue:145-148`：`.drawer-btn, .back-btn { margin-right: 8px }` */
.drawer-btn,
.back-btn {
  margin-right: 8px;
}

.nav-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.theme-switch-wrap {
  position: absolute;
  right: 20px;
  top: calc(50% - 9px);
}
</style>
