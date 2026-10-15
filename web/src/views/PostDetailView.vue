<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import MainNav from '@/components/MainNav.vue'
import PostDetail from '@/components/PostDetail.vue'
import PostSkeleton from '@/components/PostSkeleton.vue'
import CommentItem from '@/components/CommentItem.vue'
import ComposeComment from '@/components/ComposeComment.vue'
import InfiniteLoadMore from '@/components/InfiniteLoadMore.vue'
import { getComments, getPost } from '@/api/post'
import type { CommentItem as CommentItemData, PostDetail as PostDetailData } from '@/types/api'

/**
 * 对应作者 `views\Post.vue`（331 行）。
 *
 * ⚠️ 这个文件以前是**空的**（0 字节），而 PROJECT.md §11.3 一直把它列在「已完成的页面」里，
 * 所以点任何一条帖子，中栏都是一片空白。第 12 轮发现，第 13 轮（T14）补上「正文」那一段，
 * 第 14 轮（T10，本轮）补上「评论区」。
 */

const route = useRoute()

/**
 * 我们的路由是 `/post/:id`（**路径参数**），作者是 `/post?id=`（query，`views\Post.vue:72`）。
 * 这是本项目从一开始就定下的差异（PROJECT.md 11.6 有记录），不改。
 */
const postId = computed(() => Number(route.params.id))

/* ------------------------------ 帖子正文 ------------------------------ */

const post = ref<PostDetailData | null>(null)
const loading = ref(false)

/**
 * 作者 `views\Post.vue:167-183` 的 `loadPost()` 大致就是这几步：
 * 置 loading → 请求 → 写 post → 收 loading。
 *
 * 差别在**取数方式**：作者是 `getPost({id})`，本项目的 `src\api\post.ts` 是 `getPost(id)`。
 * 还有一点照抄了作者：他也没有写 `catch` 里的错误提示 —— 那是 `utils/request.ts`
 * 的响应拦截器统一负责的，这里再弹一遍就会看到两条一样的红条。
 */
async function loadPost(): Promise<void> {
  if (!(postId.value > 0)) return

  post.value = null
  loading.value = true
  try {
    post.value = await getPost(postId.value)
  } catch {
    // 错误提示由 utils/request.ts 的拦截器统一弹
  } finally {
    loading.value = false
  }
}

/** `PostDetail.vue` 删掉帖子之后会喊一声 `reload`，这里重新拉一次——作者也是这个流程 */
function reloadPost(): void {
  void loadPost()
}

/* -------------------------------- 评论 -------------------------------- */

/**
 * 作者 `views\Post.vue:75` 是 `const pageSize = 20`，评论和帖子共用这一个常量。
 * 我们分开写：帖子的分页在 HomeView 里（`PAGE_SIZE = 10`，对齐后端默认），
 * 评论这里对齐作者的 20。
 */
const COMMENT_PAGE_SIZE = 20

const comments = ref<CommentItemData[]>([])
const commentLoading = ref(false)
const commentPage = ref(1)
const commentTotal = ref(0)
const commentNoMore = ref(false)

/** 作者 `views\Post.vue:17` 的三个 Tab：推荐(default) / 热门(hots) / 最新(newest) */
const commentTab = ref('default')

const commentTotalPage = computed(() => Math.ceil(commentTotal.value / COMMENT_PAGE_SIZE))

/**
 * 作者 `views\Post.vue:184-...` 的 `loadComments()`。
 *
 * 两处必须改（§0.3 后端为准）：
 *   ① 作者取的是 `rsp.pager.total_rows`，我们后端是**顶层 `total`**；
 *   ② 作者按 `defaultCommentsSort / hotsNoMore / newestNoMore` 三套状态分别记「没有更多」，
 *      因为我们后端**没有排序参数**（见下面 commentTab 的说明），所以只需要一套 `commentNoMore`。
 *
 * `mapKey` 这个参数是给「刷新」用的：作者在删评论/发评论成功后调 `reloadComments()`，
 * 我让它回到第 1 页重新拉，而不是往后追加。
 */
async function loadComments(reset = false): Promise<void> {
  if (!(postId.value > 0)) return
  if (commentLoading.value) return

  if (reset) {
    commentPage.value = 1
    comments.value = []
    commentNoMore.value = false
  }

  commentLoading.value = true
  try {
    const res = await getComments(postId.value, {
      page: commentPage.value,
      page_size: COMMENT_PAGE_SIZE,
    })
    comments.value = [...comments.value, ...res.comments]
    commentTotal.value = res.total
    commentNoMore.value = commentPage.value >= commentTotalPage.value
  } catch {
    // 错误提示由 utils/request.ts 的拦截器统一弹
  } finally {
    commentLoading.value = false
  }
}

/** 评论的「加载更多」——对应作者 `views\Post.vue:44-55` 那段 InfiniteLoading */
function loadMoreComments(): void {
  if (commentNoMore.value || commentLoading.value) return
  commentPage.value += 1
  void loadComments()
}

/** 作者 `views\Post.vue:26,40` 的 `reloadComments()` —— 发评论 / 删评论成功后回来重拉第一页 */
function reloadComments(): void {
  commentLoading.value = false
  void loadComments(true)
}

/* -------------------------------- 生命周期 -------------------------------- */

onMounted(() => {
  void loadPost()
  void loadComments(true)
})

/**
 * 作者 `views\Post.vue:289-293` 的 watch 一模一样：
 *   从广场点进**另一条**帖子时，组件不会重新挂载（同一个路由），只能靠 id 变化来触发重新加载。
 *   `route.name === 'post'` 那个判断也是照抄的 —— 防止详情页还没离开时，别的页面改了 id。
 */
watch(postId, () => {
  if (route.name !== 'post') return
  commentPage.value = 1
  commentNoMore.value = false
  comments.value = []
  commentTotal.value = 0
  void loadPost()
  void loadComments(true)
})
</script>

<template>
  <div>
    <!-- 作者 `views\Post.vue:3`：`<main-nav title="泡泡详情" :back="true" />` -->
    <MainNav title="泡泡详情" :back="true" />

    <!--
      作者 `views\Post.vue:5` 是 `<n-list class="main-content-wrap" bordered>`，
      他里面每个 `<n-list-item>` 的 padding 被 `assets\css\main.less` 归零了。
      本项目没有 n-list（§6），照 HomeView.vue 的办法手写：外层带边框，块与块之间加分隔线。
    -->
    <div class="main-content-wrap">
      <!-- :6-15 -->
      <div v-loading="loading" class="list-item">
        <div v-if="post" class="detail-wrap">
          <PostDetail :post="post" @reload="reloadPost" />
        </div>
        <div v-else class="empty-wrap">
          <el-empty description="暂无数据" :image-size="120" />
        </div>
      </div>

      <!-- :16-25 评论排序 Tab -->
      <div v-if="post" class="comment-opts-wrap">
        <span class="comment-title-item">评论</span>

        <!--
          ⚠️ 三个 Tab **点了只会换高亮，列表不会变**（已登记在 PROJECT.md 11.8）。
          原因：本项目的 `GET /posts/:id/comments` **没有排序参数**
          （`internal\service\comment.go` 固定按 created_at 升序返回），
          所以「推荐 / 热门 / 最新」这三档在后端根本无从区分。
          作者的三个 Tab 对应他后端的三种排序，我们这里先照原版把样子做出来。

          ⚠️ 还有一处结构差异：作者是 `<n-tabs>` 的 `#prefix` 插槽里放「评论」两个字
          （`views\Post.vue:18-20`，还要配 `tab-style="margin-left:-24px"` 把 Tab 往左拉）。
          **Element Plus 的 `el-tabs` 没有 prefix 插槽**，所以这里改成 flex：
          左边一个 `<span>`、右边一个 `<el-tabs>`，视觉结果一致。
        -->
        <el-tabs v-model="commentTab" class="comment-tabs">
          <el-tab-pane label="推荐" name="default" />
          <el-tab-pane label="热门" name="hots" />
          <el-tab-pane label="最新" name="newest" />
        </el-tabs>
      </div>

      <!--
        作者 `views\Post.vue:26-28` 的发评论框（T11）：
          `<compose-comment :lock="post.is_lock" :post-id="post.id" @post-success="reloadComments" />`
        作者的 `:lock="post.is_lock"` 我们没搬 —— 后端没有「锁定帖子」这个字段（见 11.6）。
        仍然包一层 `.list-item`：列表的分隔线靠 `> .list-item + .list-item` 画，
        不包的话发评论框和第一条评论之间会少一条线。
      -->
      <div v-if="post" class="list-item">
        <ComposeComment :post-id="post.id" @post-success="reloadComments" />
      </div>

      <!-- :30-43 评论列表 -->
      <div v-if="post">
        <!-- :31-33 评论加载中铺 5 条骨架（作者 `:32` 写的就是 `:num="5"`） -->
        <div v-if="commentLoading && comments.length === 0" class="skeleton-wrap">
          <PostSkeleton :num="5" />
        </div>

        <div v-else>
          <!-- :35-37 -->
          <div v-if="comments.length === 0" class="empty-wrap">
            <el-empty description="暂无评论，快来抢沙发" :image-size="120" />
          </div>

          <!-- :39-41：一个评论一块，块与块之间的分隔线由 .list-item + .list-item 规则画 -->
          <div v-for="comment in comments" :key="comment.id" class="list-item">
            <CommentItem :comment="comment" @reload="reloadComments" />
          </div>
        </div>
      </div>

      <!--
        :44-55 评论的「加载更多」。
        作者那边用的是 `v3-infinite-loading`，本项目统一走自己的 `InfiniteLoadMore.vue`（T7）。
        他那个 `#spinner` 插槽里塞了六个互相排斥的 `<span>`，还留了一句注释「注意一定要保留这里」
        —— 那是 `v3-infinite-loading` 的插槽机制逼出来的写法，我们用 IntersectionObserver 自己实现，
        不需要它。另外作者的 guard 是 `v-if="comments.length >= pageSize"`（上一页装满才显示），
        我们的 `InfiniteLoadMore` 内部已经会用「没有更多数据了」顶掉加载中的样子，所以这里只判非空。
      -->
      <InfiniteLoadMore
        v-if="comments.length > 0"
        :total-page="commentTotalPage"
        :no-more="commentNoMore"
        complete-text="没有更多数据了"
        @load-more="loadMoreComments"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
/* 作者 `views\Post.vue:297-299` */
.detail-wrap {
  min-height: 100px;
}

/*
 * 作者 `:5` 用的是 `<n-list bordered>`，边框在 naive 那边；我们手写。
 * 和 HomeView.vue 的 `.main-content-wrap` 保持同一套写法（边框 + 底色）。
 */
.main-content-wrap {
  border: 1px solid var(--el-border-color);
  border-top: none;
  border-radius: 0;
  background-color: var(--el-bg-color);

  > .list-item + .list-item {
    border-top: 1px solid var(--el-border-color);
  }
}

/* 作者 `views\Post.vue:309-314` */
.comment-opts-wrap {
  padding: 6px 16px 0;
  opacity: 0.75;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.comment-title-item {
  padding-top: 4px;
  font-size: 16px;
  text-align: center;
}

/*
 * `el-tabs` 自带的底部下划线会横穿整行，而作者那边（`n-tabs type="bar"`）也有一条，
 * 只是他在 `.comment-opts-wrap` 里只占右边一小块。这里把 header 的下边距压掉，
 * 免得它和下面评论块之间的分隔线叠成两条。
 */
.comment-tabs {
  :deep(.el-tabs__header) {
    margin: 0;
  }
}

/* 沿用 HomeView.vue 里那一套空状态尺寸（作者 main.less 的 .empty-wrap） */
.empty-wrap {
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
