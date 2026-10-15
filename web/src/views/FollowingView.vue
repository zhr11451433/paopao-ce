<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import MainNav from '@/components/MainNav.vue'
import PostSkeleton from '@/components/PostSkeleton.vue'
import UserCard from '@/components/UserCard.vue'
import InfiniteLoadMore from '@/components/InfiniteLoadMore.vue'
import { getFollowers, getFollowing } from '@/api/user'

/**
 * 作者 `views\Following.vue` 是一个页面装两个列表：
 *   顶上一个 `n-tabs`，两个 Tab 分别是「正在关注」和「我的粉丝」，
 *   切 Tab 只换请求，不分页。
 * 我们照这个结构做，但**传参方式改了**（见 PROJECT.md 11.6 的 T12 行）：
 *
 *   作者                     我们                 为什么
 *   ?s=<用户名>              ?id=<用户 id>        本项目后端的关注/粉丝接口是
 *                                                 `GET /users/:id/following`（按数字 id），
 *                                                 不是他的 `?username=`
 *   ?n=<昵称>                ?n=<名字>            沿用，只用来当顶栏标题
 *   ?t=follows|followings    ?t=follows|followings 沿用，只用来定初始 Tab
 */
const route = useRoute()

/** 作者 `:50-52` */
const nickname = (route.query.n as string) || '粉丝详情'
const userId = Number(route.query.id) || 0
const tabler = ref<'follows' | 'followings'>(
  route.query.t === 'followings' ? 'followings' : 'follows',
)

/** 作者 `:56` `usePagination(20)`。本项目没有 composables 目录，照 PostDetailView 手写 */
const PAGE_SIZE = 20

/**
 * 作者列表里的一项是他的 `Item.ContactItemProps`；
 * 我们的后端只给 `{follow_id, name}` / `{fan_id, name}`，
 * 所以在页面里统一成这个更瘦的形状再往 `UserCard` 传。
 * `isFollowing` 是**我们补的**：后端不返回「我是否已关注他」（见 11.6 的 T12 行）。
 */
interface UserCardContact {
  id: number
  name: string
  isFollowing: boolean
}

const list = ref<UserCardContact[]>([])
const loading = ref(false)
const page = ref(1)
const totalPage = ref(0)
const noMore = ref(false)

/** 作者 `:67-73`：到底之后显示的话，两个 Tab 不一样 */
const completeStr = computed(() =>
  tabler.value === 'follows' ? '没有更多关注了' : '没有更多粉丝了',
)

/**
 * 作者 `loadFollows` / `loadFollowings`（`:100-166`）是两个几乎一样的函数，
 * 我们合成一个，行为逐条对齐：
 *   · `list.length === 0` 时才显示骨架（翻页时不闪骨架）；
 *   · 第 1 页是替换、后面几页是追加；
 *   · 返回空数组 = 到底了；
 *   · `totalPage = ceil(total / pageSize)`；
 *   · 失败时如果是翻页，把页码退回去（`:126-131`），否则用户滚到底之后再也拉不动。
 * 多出来的一条：`page >= totalPage` 也判定到底 —— 我们的 `InfiniteLoadMore` 靠 `noMore`
 * 决定「还在转圈」还是「没有更多了」，而作者的 `v3-infinite-loading` 自己会看空数组。
 */
async function loadPage(reset = false): Promise<void> {
  if (userId === 0) {
    return
  }

  if (reset) {
    page.value = 1
    noMore.value = false
    list.value = []
  }

  if (list.value.length === 0) {
    loading.value = true
  }

  try {
    if (tabler.value === 'follows') {
      const res = await getFollowing(userId, { page: page.value, page_size: PAGE_SIZE })
      // 能出现在「正在关注」这一页里的人，按定义都是已关注
      const rows = res.follows.map((item) => ({
        id: item.follow_id,
        name: item.name,
        isFollowing: true,
      }))
      list.value = page.value > 1 ? [...list.value, ...rows] : rows
      totalPage.value = Math.ceil(res.total / PAGE_SIZE)
      if (rows.length === 0 || page.value >= totalPage.value) {
        noMore.value = true
      }
    } else {
      const res = await getFollowers(userId, { page: page.value, page_size: PAGE_SIZE })
      // ⚠️ 粉丝这一页**不知道我有没有关注对方**（后端不返回 is_following），
      //    只能一律当成「没关注」—— 点「关注」时如果本来就已经关注着，
      //    后端会返回 200「您已关注该用户」，不会报错。
      const rows = res.fans.map((item) => ({
        id: item.fan_id,
        name: item.name,
        isFollowing: false,
      }))
      list.value = page.value > 1 ? [...list.value, ...rows] : rows
      totalPage.value = Math.ceil(res.total / PAGE_SIZE)
      if (rows.length === 0 || page.value >= totalPage.value) {
        noMore.value = true
      }
    }
  } catch {
    if (page.value > 1) {
      page.value -= 1
    }
  } finally {
    loading.value = false
  }
}

/** 作者 `:75-77` `nextPage(loadPage)` */
function handleNextPage(): void {
  if (noMore.value || loading.value) {
    return
  }
  page.value += 1
  void loadPage()
}

/** 作者 `:79-82` `changeTab`：换 Tab 等于回到第 1 页重拉 */
function changeTab(tab: string | number): void {
  tabler.value = tab === 'followings' ? 'followings' : 'follows'
  void loadPage(true)
}

/**
 * 作者 `:92-98`：取关成功后**只在「正在关注」这一页刷新**。
 * 粉丝页不刷 —— 在粉丝页取关别人本来就不是这一页的语义。
 */
function handleUnfollowSuccess(): void {
  if (tabler.value === 'follows') {
    void loadPage(true)
  }
}

/** 作者 `:168-170` */
onMounted(() => {
  void loadPage()
})
</script>

<template>
  <div>
    <!-- 作者 `:3`：`<main-nav :title="nickname" :back="true" />` -->
    <MainNav :title="nickname" :back="true" />

    <!--
      作者 `:5` 是 `<n-list class="main-content-wrap" bordered>`，
      `.main-content-wrap` 在他那边还有 `padding: 20px`（`:175-177`）。
      本项目没有 n-list（§6），照 HomeView / PostDetailView 的办法手写边框 + 分隔线，
      并保留他那 20px 内边距。
    -->
    <div class="main-content-wrap">
      <!-- 作者 `:6-9`。他还有 `animated`，EP 的 el-tabs 本来就是滑动下划线 -->
      <el-tabs v-model="tabler" class="follow-tabs" @tab-change="changeTab">
        <el-tab-pane label="正在关注" name="follows" />
        <el-tab-pane label="我的粉丝" name="followings" />
      </el-tabs>

      <!-- 作者 `:10-12`：只在「一条都还没有」时铺骨架，`:num="pageSize"` 是他写的 -->
      <div v-if="loading && list.length === 0" class="skeleton-wrap">
        <PostSkeleton :num="PAGE_SIZE" />
      </div>

      <div v-else>
        <!-- 作者 `:14-16` -->
        <div v-if="list.length === 0" class="empty-wrap">
          <el-empty description="暂无数据" :image-size="120" />
        </div>

        <!-- 作者 `:18-20` -->
        <div v-for="contact in list" :key="contact.id" class="list-item">
          <UserCard
            type="follow"
            :contact="contact"
            @unfollow-success="handleUnfollowSuccess"
          />
        </div>
      </div>
    </div>

    <!--
      作者 `:26-35` 是 `v3-infinite-loading`，本项目统一走自己的 `InfiniteLoadMore.vue`（T7），
      它内部已经带 `v-if="totalPage > 0"` 和「加载中」的样子。
    -->
    <InfiniteLoadMore
      :total-page="totalPage"
      :no-more="noMore"
      :complete-text="completeStr"
      @load-more="handleNextPage"
    />
  </div>
</template>

<style scoped lang="scss">
/* 作者 `:175-177` 的 padding + 本项目 HomeView / PostDetailView 的边框写法 */
.main-content-wrap {
  padding: 20px;
  border: 1px solid var(--el-border-color);
  border-top: none;
  border-radius: 0;
  background-color: var(--el-bg-color);

  > .list-item + .list-item {
    border-top: 1px solid var(--el-border-color);
  }
}

/*
 * `el-tabs` 的 header 自带下边距，作者那边（`n-tabs type="line"`）也有一条，
 * 但压在 20px 内边距里会显得空 —— 这里只把下边距压掉，行为不变。
 */
.follow-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 12px;
  }
}

/* 沿用 HomeView.vue 里那一套空状态尺寸（作者 main.less 的 .empty-wrap） */
.empty-wrap {
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 作者 `:196-200` 的暗色块 */
:global(.dark) {
  .main-content-wrap,
  .empty-wrap,
  .skeleton-wrap {
    background-color: rgba(16, 16, 20, 0.75);
  }
}
</style>
