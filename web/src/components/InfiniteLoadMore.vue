<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Loading } from '@element-plus/icons-vue'

/**
 * 作者 components\infinite-load-more.vue:21-27 的 props，一个不改：
 *   totalPage    总页数（他的模板 :2 挂着 `v-if="totalPage > 0"`，0 页时整个不渲染）
 *   noMore       已经到底了
 *   completeText 到底之后显示的话，默认「没有更多了」
 */
withDefaults(
  defineProps<{
    totalPage: number
    noMore: boolean
    completeText?: string
  }>(),
  {
    completeText: '没有更多了',
  },
)

/** 作者 :29-31，一个不改 */
const emit = defineEmits<{
  (e: 'load-more'): void
}>()

/**
 * ★ 这里没有 `v3-infinite-loading` 的替代品可用：PROJECT.md §2 定死了不引入新依赖，
 *   所以用浏览器自带的 IntersectionObserver 自己实现「快滚到底了就喊一声」。
 *
 * 提前多远开始加载：v3-infinite-loading 也有个同样的参数（它叫 distance），
 * 但作者的 node_modules 里**没有装这个包**，我读不到它的默认值，所以自己取 200px
 * ——滚到离底部还有 200px 就开始加载，用户不会看到明显的空白停顿。
 * 这个数不合适的话改这一个常量就行。
 */
const PRELOAD_DISTANCE = 200

/** 哨兵元素：就是底部那一行「加载更多」。它一露头，就说明该加载下一页了 */
const sentinelRef = ref<HTMLElement | null>(null)

let observer: IntersectionObserver | null = null

function disconnect(): void {
  observer?.disconnect()
  observer = null
}

function observe(el: HTMLElement): void {
  disconnect()
  observer = new IntersectionObserver(
    (entries) => {
      // noUncheckedIndexedAccess 开着，entries[0] 的类型是「可能 undefined」，必须判
      if (entries[0]?.isIntersecting) emit('load-more')
    },
    { rootMargin: `0px 0px ${PRELOAD_DISTANCE}px 0px` },
  )
  observer.observe(el)
}

/**
 * 为什么用 watch 盯着 ref，而不是只在 onMounted 里挂一次：
 * `totalPage` 一开始是 0，那时 `v-if` 让这个 div **根本不存在**，ref 是 null。
 * 等首页数据回来、totalPage 变成 1，div 才第一次被渲染出来。
 * 只在 onMounted 里挂，就永远挂不上。
 */
watch(
  sentinelRef,
  (el) => {
    if (el) observe(el)
    else disconnect()
  },
  { flush: 'post' },
)

onMounted(() => {
  if (sentinelRef.value) observe(sentinelRef.value)
})

// 组件销毁必须断开，否则这个 observer 会一直盯着一个已经不存在的节点
onBeforeUnmount(disconnect)
</script>

<template>
  <!--
    作者 :2 是 `<n-space v-if="totalPage > 0" justify="center">`。
    n-space 在 Element Plus 里没有对应组件（PROJECT.md §6），它在这儿只干一件事：
    把里面的东西水平居中 —— 一行 display:flex + justify-content:center 就等价了，
    所以省掉了外面那层。`.load-more` 这个类名和它的 margin 照作者 :39-40 来。
  -->
  <div v-if="totalPage > 0" ref="sentinelRef" class="load-more">
    <div class="load-more-wrap">
      <!--
        作者 :10 是 `<n-spin :size="14" v-if="!noMore" />`。
        EP 没有「14px 的 inline spinner」这种形态，等价物是
        `<el-icon class="is-loading">` 里放一个 Loading 图标
        —— `is-loading` 这个类名就是 EP 自己的旋转动画。
      -->
      <el-icon v-if="!noMore" class="is-loading" :size="14">
        <Loading />
      </el-icon>
      <!-- 作者 :11 的三元表达式，一个不改 -->
      <span class="load-more-spinner">{{ noMore ? completeText : '加载更多' }}</span>
    </div>
  </div>
</template>

<style scoped lang="scss">
/* 作者 infinite-load-more.vue:38-54 逐条照抄；
   只有 display / justify-content 这两条是从他顶上那层 n-space 挪下来的 */
.load-more {
  margin: 20px;
  display: flex;
  justify-content: center;

  .load-more-wrap {
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
    gap: 14px;

    .load-more-spinner {
      font-size: 14px;
      opacity: 0.65;
    }
  }
}
</style>
