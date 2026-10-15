<script setup lang="ts">
/**
 * 404 页（T17）。
 *
 * 作者对应的是 `F:\github_paopao\paopao-ce\web\src\views\404.vue`（42 行），
 * 是整个项目里最小的一页：顶栏 + 一个居中的结果块 + 一个「回主页」按钮。
 *
 * 路由那一边已经接好了（`src\router\index.ts`）：
 *   - `/404` 这条路由指向本文件；
 *   - 最后一条 `{ path: '/:pathMatch(.*)*', redirect: '/404' }` 兜住所有没匹配上的地址。
 * 所以随便敲一个 `/#/不存在的地址` 就会落到这里。
 */
import { useRouter } from 'vue-router'
import MainNav from '@/components/MainNav.vue'

const router = useRouter()

function goHome(): void {
  void router.push({ path: '/' })
}
</script>

<template>
  <div class="not-found-page">
    <!-- 作者 `:3` 是 `<main-nav title="404" />`，注意**没有 :back** -->
    <MainNav title="404" />

    <div class="main-content-wrap wrap404">
      <!--
        作者 `:6-14` 用的是 naive-ui 的 `<n-result status="404">`，
        它会画一张**带 404 字样的插画**（那个插画是 naive 内置的资源，EP 没有对应物）。
        Element Plus 的 `el-result` 只有四种图标：success / warning / info / error。
        这里用它的 `#icon` 插槽自己写一个大的「404」顶上去，视觉位置和原版一致。

        另外两个插槽名也对不上：作者的 `title` / `description` / `#footer`
        → EP 是 `title` / `sub-title` / `#extra`。
      -->
      <el-result title="404 资源不存在" sub-title="再看看其他的吧">
        <template #icon>
          <span class="result-404">404</span>
        </template>
        <template #extra>
          <el-button @click="goHome">回主页</el-button>
        </template>
      </el-result>
    </div>
  </div>
</template>

<style scoped lang="scss">
/**
 * `.main-content-wrap` 在本项目里是**每个页面各自 scoped 定义**的
 * （HomeView / PostDetailView / FollowingView / ProfileView / UserView 都是这么写的）。
 * 404 这一页作者只套了底色、没有写 `.list-item` 那套分隔线，这里跟着他。
 */
.main-content-wrap {
  border: 1px solid var(--el-border-color);
  border-top: none;
  border-radius: 0;
  background-color: var(--el-bg-color);
}

/* 作者 `:31-36` */
.wrap404 {
  min-height: 500px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 替代 naive-ui 那张内置的 404 插画 */
.result-404 {
  font-size: 72px;
  font-weight: 700;
  line-height: 1;
  color: var(--el-color-primary);
}

/**
 * 作者 `:37-41` 这段写的是 `.dark { .main-content-wra { ... } }` ——
 * **少了一个字母 p**，选择器根本匹配不到任何东西，所以他那边 404 页在暗色下
 * 是一块白底。这是笔误不是风格，这里写成了正确的 `.main-content-wrap`
 * （PROJECT.md §11.6 登记了这一处「故意不像」）。
 */
:global(.dark) {
  .main-content-wrap {
    background-color: rgba(16, 16, 20, 0.75);
  }
}
</style>
