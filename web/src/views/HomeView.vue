<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useMainStore } from '@/store/main'
import { useUserStore } from '@/store/user'
import { getFeed, getPosts } from '@/api/post'
import MainNav from '@/components/MainNav.vue'
import HomeCompose from '@/components/HomeCompose.vue'
import PostItem from '@/components/PostItem.vue'
import PostSkeleton from '@/components/PostSkeleton.vue'
import InfiniteLoadMore from '@/components/InfiniteLoadMore.vue'
import type { PostListItem } from '@/types/api'

type TabKey = 'newest' | 'hots' | 'following'

/** 作者 Home.vue:118-122 的三个筛选项，一个不改 */
const TABS: { key: TabKey; label: string }[] = [
  { key: 'newest', label: '全部' },
  { key: 'hots', label: '热门推荐' },
  { key: 'following', label: '正在关注' },
]

const route = useRoute()
const mainStore = useMainStore()
const userStore = useUserStore()
const { isMobile } = storeToRefs(mainStore)
const { logined } = storeToRefs(userStore)

const activeTab = ref<TabKey>('newest')

/**
 * 当前**已经加载进列表**的帖子。
 * 无限滚动是「往尾巴上追加」，不是整体替换 —— 所以它和下面的 `page` 是一套的
 * （作者 Home.vue 用 `usePagination(20)` 一次拿到 loading/noMore/page/pageSize/totalPage）。
 */
const posts = ref<PostListItem[]>([])

/**
 * 列表真正在请求的来源。和 `activeTab` 分开，是因为「热门推荐」只换高亮、不换数据
 * （后端没有热门排序）。高亮可以停在 hots，下一页仍按这里继续请求。
 */
const listSource = ref<'newest' | 'following'>('newest')

/** 丢掉过期响应。快速连点两个 Tab 时，先发出去的那个不能把后一个的列表盖掉 */
let requestSeq = 0

/**
 * 骨架屏开关。作者 Home.vue:49-51 是这样用的：
 *   <div v-if="loading && list.length === 0" class="skeleton-wrap">
 *       <post-skeleton :num="pageSize" />
 *   </div>
 * 本项目照着写成 `v-if="loading && posts.length === 0"`（见模板）。
 */
const loading = ref(true)

/** 当前是第几页（从 1 开始，和本项目后端的 `page` 参数一致）*/
const page = ref(1)

/** 已经到底了（作者 usePagination 里同名）*/
const noMore = ref(false)

/** 一共几页。`<InfiniteLoadMore>` 靠它决定自己渲不渲染（作者模板 :2 的 `v-if="totalPage > 0"`）*/
const totalPage = ref(0)

/**
 * 每页条数。作者 Home.vue 是 `usePagination(20)`，取 20。
 * 本项目后端默认 `page_size` 是 10（PROJECT.md §7.1），所以取 10。
 * 骨架屏要铺几条，就是照这个数来的（作者 Home.vue:50 写的 `:num="pageSize"`）。
 */
const PAGE_SIZE = 10

/**
 * 作者 Home.vue:258-264 的 showTrendsTag。
 * 重点：里面带了 userInfo.id > 0 —— 未登录时这一行整个不出现。
 * 趋势条（SlideBar）那个开关我们先砍了，没有对应后端。
 */
const showFilterBar = computed(() => logined.value && !isMobile.value)

/** 作者 Home.vue:247-256：标题会跟着搜索/话题查询变 */
const title = computed(() => {
  const q = route.query.q
  if (typeof q === 'string' && q) {
    return route.query.t === 'tag' ? `#${decodeURIComponent(q)}` : `搜索: ${decodeURIComponent(q)}`
  }
  return '泡泡广场'
})

function onFilterClick(key: TabKey): void {
  if (activeTab.value === key) return
  activeTab.value = key
  // 热门推荐没有接口，只换高亮，列表保持原样（PROJECT.md §11.8）
  if (key === 'hots') return
  listSource.value = key
  posts.value = []
  page.value = 1
  noMore.value = false
  totalPage.value = 0
  void loadPosts()
}

/**
 * 加载「当前页」的数据，并**追加**到列表尾部。
 * 首次进页面、以及滚到底加载下一页，走的都是这一个函数
 * —— 作者 Home.vue 同理（首次 `loadPosts`，滚到底是 `handleNextPage` → `nextPage` + `loadPosts`）。
 *
 * 作者把 style 传给同一个 getPosts，再用 rsp.pager.total_rows 算页数（Home.vue:345-367）。
 * 我们的后端没有 style、也没有 pager：全部走 GET /posts（字段 list），
 * 正在关注走 GET /feed（字段 posts），总数都在顶层 total。
 * 地址上的 ?q= 只改标题，不传给接口 —— 列表接口没有搜索参数。
 */
async function loadPosts(): Promise<void> {
  const seq = ++requestSeq
  loading.value = true
  try {
    const params = { page: page.value, page_size: PAGE_SIZE }
    let list: PostListItem[]
    let total: number
    if (listSource.value === 'following') {
      const res = await getFeed(params)
      // Go 的 nil 切片会变成 null，不能直接展开
      list = res.posts ?? []
      total = res.total
    } else {
      const res = await getPosts(params)
      list = res.list ?? []
      total = res.total
    }
    if (seq !== requestSeq) return
    posts.value = [...posts.value, ...list]
    totalPage.value = Math.ceil(total / PAGE_SIZE)
    // 这一页是空的也算到底，避免 total 和列表对不上时不停请求
    noMore.value = list.length === 0 || page.value >= totalPage.value
  } catch {
    if (seq !== requestSeq) return
    // 作者 Home.vue:371-373：这一页失败了就把页码退回去，下次还能再试
    if (page.value > 1) page.value -= 1
  } finally {
    if (seq === requestSeq) loading.value = false
  }
}

/** `<InfiniteLoadMore>` 滚到快到底时 emit 的就是这个 */
function loadMore(): void {
  // totalPage 是 0 时组件压根不渲染；这两个判断是防「一次滚动触发两回」
  if (noMore.value || loading.value) return
  page.value += 1
  void loadPosts()
}

onMounted(() => {
  void loadPosts()
})
</script>

<template>
  <div>
    <MainNav :title="title" />

    <div class="main-content-wrap">
      <!-- 每块包一层 .list-item，靠相邻兄弟选择器画分隔线（作者是 n-list 的 bordered 干的） -->
      <div class="list-item">
        <HomeCompose />
      </div>

      <div v-if="showFilterBar" class="list-item style-wrap">
        <button
          v-for="tab in TABS"
          :key="tab.key"
          class="filter-item"
          :class="{ 'is-active': activeTab === tab.key }"
          @click="onFilterClick(tab.key)"
        >
          {{ tab.label }}
        </button>
      </div>

      <!--
        帖子列表。作者 Home.vue 用的是 `<n-list bordered>`，n-list 在 Element Plus 里
        没有对应组件（PROJECT.md §6 →「用 div 自己写分隔线」），所以每块包一层 .list-item，
        由下面 scss 里那句 `> .list-item + .list-item` 负责画线。
      -->
      <!--
        骨架屏（T6）。作者 Home.vue:49-51 原样：
          <div v-if="loading && list.length === 0" class="skeleton-wrap">
              <post-skeleton :num="pageSize" />
          </div>
        作者的 pageSize 是「每页几条」（Home.vue:185 的 usePagination(20)）。
        本项目后端默认 page_size 是 10，所以传 PAGE_SIZE。
      -->
      <div v-if="loading && posts.length === 0" class="skeleton-wrap">
        <PostSkeleton :num="PAGE_SIZE" />
      </div>

      <template v-if="posts.length > 0">
        <div v-for="post in posts" :key="post.id" class="list-item">
          <PostItem :post="post" :is-mobile="isMobile" />
        </div>
      </template>

      <!--
        列表为空才显示「暂无数据」（作者 Home.vue:265-270 就是这个条件）。
        ⚠️ 这里**不能**再挂 showFilterBar —— 未登录时同样是没数据，也要看到「暂无数据」，
        作者自己的未登录广场页就是这样。原来的写法把它和 Tab 行绑在了一起，是个 bug，第 6 轮顺手修正。
      -->
      <div v-else class="list-item">
        <div class="empty-wrap">
          <el-empty description="暂无数据" :image-size="120" />
        </div>
      </div>
    </div>

    <!--
      无限滚动（T7）。作者 Home.vue:73-82 是**直接内联**在页面里的
      （他另有一个 `components\infinite-load-more.vue` 组件，但只有 Contacts.vue / Messages.vue 在用，
      广场页却没用）。我们统一走组件，省得以后每个页面抄一遍。

      `completeText` 取作者 Home.vue:74 的原文「没有更多泡泡了」——
      组件自己的默认值是「没有更多了」，好友页那边作者传的是「没有更多好友了」。
    -->
    <InfiniteLoadMore
      :total-page="totalPage"
      :no-more="noMore"
      complete-text="没有更多泡泡了"
      @load-more="loadMore"
    />
  </div>
</template>

<style scoped lang="scss">
.main-content-wrap {
  border: 1px solid var(--el-border-color);
  border-top: none;              /* 作者 main.less:30 就是这句，顶栏已经画了上边 */
  border-radius: 0;
  background-color: var(--el-bg-color);

  > .list-item + .list-item {
    border-top: 1px solid var(--el-border-color);
  }
}

.empty-wrap {
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 作者 Home.vue:527-537 的 .style-wrap，margin 改成 padding 好让上边框贴着 */
.style-wrap {
  padding: 10px 0 4px 16px;
  opacity: 0.8;
  display: flex;
  gap: 12px;

  /* naive 的 n-button secondary round：有底色、无边框、选中态上色。
     EP 没有一一对应的形态，这里手写反而最准。 */
  .filter-item {
    height: 28px;
    padding: 0 14px;
    border: none;
    border-radius: 14px;
    font-size: 14px;
    font-family: inherit;
    cursor: pointer;
    background-color: var(--el-fill-color-light);
    color: var(--el-text-color-primary);
    transition: background-color 0.2s;

    &:hover { background-color: var(--el-fill-color); }

    &.is-active {
      background-color: var(--el-color-primary-light-9);
      color: var(--el-color-primary);
    }
  }
}
</style>
