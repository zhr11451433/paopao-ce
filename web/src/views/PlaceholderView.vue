<script setup lang="ts">
import { useRoute } from 'vue-router'

/**
 * 临时占位页。
 *
 * 菜单里先把「话题 / 消息 / 收藏 / 好友 / 设置」摆上，是为了让登录前/登录后的
 * 菜单结构和原作者一致；这些页面真正的样式和接口后面一页一页补。
 *
 * 补某一页时，只把 router/index.ts 里那一行的 component
 * 从 PlaceholderView.vue 换成新页面即可，菜单不用改。
 */
const route = useRoute()
</script>

<template>
  <div class="placeholder-page">
    <div class="page-title">{{ route.meta.title }}</div>
    <el-empty :description="`${route.meta.title} —— 样式和接口还没做`" />
  </div>
</template>

<style scoped lang="scss">
.placeholder-page {
  .page-title {
    padding: 12px 20px;
    font-size: 18px;
    font-weight: bold;
    color: var(--el-text-color-primary);
  }
}
</style>