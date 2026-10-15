<script setup lang="ts">
/**
 * 作者 post-skeleton.vue:14-21：
 *   const props = withDefaults(defineProps<{ num: number }>(), { num: 1 })
 *
 * 原版把返回值赋给了 `props`，但整个文件从未用过它 —— 模板里直接用 `num`。
 * 这样写会触发 TS 的「声明了但没用到」告警，所以这里不 assign。
 * 不 assign 也能在模板里用 `num`：`defineProps` 是编译宏，它会把 props
 * 注册成模板绑定，跟返回值有没有被接住无关。
 */
withDefaults(
  defineProps<{
    num: number
  }>(),
  {
    num: 1,
  },
)
</script>

<template>
  <!--
    作者 post-skeleton.vue:2 的写法是 `v-for="i in new Array(num)"`。

    ⚠️ 这里改成了 `v-for="i in num"`，是本次唯一一处偏离原版的地方，原因：
      `new Array(3)` 是一个**稀疏数组**，三个位置的值都是 undefined，
      所以作者那句 `:key="i"` 拿到的三个 key 全是 undefined —— 控制台会刷
      「Duplicate keys found」警告。改成 `v-for="i in num"`（源是数字时
      i 依次取 1、2、3）**渲染结果一模一样**，但 key 是对的。
      属于 PROJECT.md §0.9「原版的 bug」例外，已记进 11.6。

    n-skeleton → el-skeleton 的映射（PROJECT.md §6）：
      <n-skeleton circle size="small" />  →  <el-skeleton-item variant="circle" />
      <n-skeleton text />                 →  <el-skeleton-item variant="text" />

    ⚠️ 那三行文字为什么不用 <el-skeleton :rows="3">：
      EP 的 el-skeleton 在 rows > 1 时会**自作主张把最后一行缩成 60% 宽**，
      而作者这里是「3 条整行 + 另外 1 条 60% 宽」，对不上。
      逐条写 el-skeleton-item 才能和作者渲染出完全一样的东西。
  -->
  <div v-for="i in num" :key="i" class="skeleton-item">
    <div class="user">
      <el-skeleton-item variant="circle" class="skeleton-avatar" />
    </div>
    <div class="content">
      <el-skeleton-item variant="text" />
      <el-skeleton-item variant="text" />
      <el-skeleton-item variant="text" />
      <el-skeleton-item variant="text" style="width: 60%" />
    </div>
  </div>
</template>

<style scoped lang="scss">
/* 作者 post-skeleton.vue:25-36 逐条搬（他那个本来就是 scoped，连 :deep 都不用加） */
.skeleton-item {
  padding: 12px;
  display: flex;

  .user {
    width: 42px;
  }

  .content {
    width: calc(100% - 42px);
  }
}

/**
 * 作者写的是 `circle size="small"`。naive 的 small 圆形骨架直径是 28px，
 * 而我们的帖子头像是 30px（PostItem.vue 里的 <el-avatar :size="30" />）。
 * 取值 30px，是为了骨架换成真内容的那一刻**不跳一下**。
 * （这是本文件第二处没有照抄原版数值的地方，理由就是这句。）
 */
.skeleton-avatar {
  width: 30px;
  height: 30px;
}

/**
 * naive 的 n-skeleton 每行自带上下间距，EP 的 el-skeleton-item 光秃秃的没有，
 * 不补的话四行会糊成一整块。8px 是让总高度接近真实卡片正文的估值。
 * `:deep()` 是因为 el-skeleton-item 是子组件，要穿透到它内部的那个 div。
 */
.content {
  :deep(.el-skeleton__item) {
    margin-bottom: 8px;

    &:last-child {
      margin-bottom: 0;
    }
  }
}

/* 作者 post-skeleton.vue:37-41 的暗色写法，原样搬。
   本项目的 .dark 打在 <html> 上（PROJECT.md §5）；scoped 会把它编译成
   `.dark .skeleton-item[data-v-x]` —— `.dark` 是祖先、`.skeleton-item` 带作用域属性，两者都满足。 */
.dark .skeleton-item {
  background-color: rgba(16, 16, 20, 0.75);
}
</style>
