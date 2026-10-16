<script setup lang="ts">
/**
 * 我的主页（T15）。
 *
 * 作者对应的是 `F:\github_paopao\paopao-ce\web\src\views\Profile.vue`（425 行）。
 * 他是「一个模板吃两个场景」：自己的主页（`/profile`）和别人的主页（`/user/:s`）
 * 用同一个组件，靠 `isOwner` 区分（T16 会用到这一点）。
 *
 * 我们这一轮只做**自己的主页**，原因是后端只给了「我的」这一类接口：
 *
 *   - `GET /posts/me`（我的泡泡）     → 本轮「泡泡」tab
 *   - `GET /posts/me/likes`（我喜欢） → 本轮「喜欢」tab
 *   - `GET /me`                       → 只有 { user_id, role }
 *   - `GET /users/:id/following` / `followers` → 数关注/粉丝
 *
 * 作者那五个 tab（泡泡 / 评论 / 亮点 / 图文 / 喜欢）里，剩下三个后端没有对应
 * 接口（详见 PROJECT.md §11.7）；而他的 `/user/:s` 那个场景要的
 * 「按 id 查用户资料」接口本后端也没有（`internal\service\router.go` 里没有这条路由），
 * 所以 T16 用户主页的处境要等到那一轮再定。
 */
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { EllipsisHorizontal, SettingsOutline } from '@vicons/ionicons5'
import MainNav from '@/components/MainNav.vue'
import PostItem from '@/components/PostItem.vue'
import PostSkeleton from '@/components/PostSkeleton.vue'
import InfiniteLoadMore from '@/components/InfiniteLoadMore.vue'
import { getMyLikes, getMyPosts } from '@/api/post'
import { getFollowers, getFollowing } from '@/api/user'
import { useMainStore } from '@/store/main'
import { useUserStore } from '@/store/user'
import { prettyQuoteNum } from '@/utils/count'
import type { MyLikeItem, PostListItem } from '@/types/api'

const PAGE_SIZE = 20

const router = useRouter()
const mainStore = useMainStore()
const userStore = useUserStore()
const { isMobile } = storeToRefs(mainStore)
const { user } = storeToRefs(userStore)

/** 'post' = 泡泡，'star' = 喜欢（对齐作者 template 里的 name） */
const pageType = ref('post')

const loading = ref(false)
const postList = ref<PostListItem[]>([])
const starList = ref<MyLikeItem[]>([])
const page = ref(1)
const totalPage = ref(0)
const noMore = ref(false)

const followCount = ref(0)
const fanCount = ref(0)
const postCount = ref(0)

/**
 * 头像是空的。
 *
 * `internal\db\model.go:9-15` 的 User 表只有 `Name / Email / Password / Role`，
 * 没有头像字段；`GET /me` 也只有一个 user_id。所以这里和 T12 的 UserCard 一样，
 * 摆一个空头像占位，用名字首字母当内容，免得是一块纯灰。
 */
const avatarText = computed(() => (user.value?.name || '泡').slice(0, 1).toUpperCase())

/**
 * 作者显示的是 `userInfo.is_admin`，我们后端没有 `is_admin` 这个布尔，
 * 只有 JWT 里的 `role` 字符串（`GET /me` 会返回）。
 */
const isAdmin = computed(() => userStore.role === 'admin')

const listEmpty = computed(() =>
  pageType.value === 'post' ? postList.value.length === 0 : starList.value.length === 0,
)

/* ---------------- 数据 ---------------- */

/**
 * 我的泡泡（`GET /posts/me`）。
 *
 * 作者 `views\Profile.vue` 把列表交给 `post-item`，卡片自己读作者 id 和点赞。
 * 我们的 `PostItem` 也一样：有 `user_id` 作者名才能点进 `/user/:id`；
 * 有 `like_count` 才显示数字，点爱心才是点赞，否则点击会冒泡进详情。
 *
 * C6 的接口已经返回 `user_id` / `like_count` / `is_liked`，这里原样摊进去。
 * 仍然没有 `author`，名字用当前登录用户的。
 */
async function loadPostList(reset = false): Promise<void> {
  if (reset) {
    page.value = 1
    noMore.value = false
  }
  // 只有列表还是空的时候才让骨架顶上，翻页时不能把已有内容盖掉
  if (postList.value.length === 0) loading.value = true
  try {
    const res = await getMyPosts({ page: page.value, page_size: PAGE_SIZE })
    const rows: PostListItem[] = res.list.map((item) => ({
      id: item.id,
      title: item.title,
      created_at: item.created_at,
      comment_count: item.comment_count,
      images: item.images,
      author: user.value?.name ?? '',
      user_id: item.user_id,
      like_count: item.like_count,
      is_liked: item.is_liked,
    }))
    postList.value = page.value > 1 ? [...postList.value, ...rows] : rows
    postCount.value = res.total
    totalPage.value = Math.ceil(res.total / PAGE_SIZE)
    if (rows.length === 0 || page.value >= totalPage.value) noMore.value = true
  } catch {
    // 这一页没拿到：退回去，别让用户卡在一个空页上
    if (page.value > 1) {
      page.value -= 1
    } else {
      postList.value = []
    }
  } finally {
    loading.value = false
  }
}

/**
 * 我喜欢的（`GET /posts/me/likes`）。
 *
 * 这个接口很瘦：直接返回一个**裸数组**（不是 `{total, page, list}`），
 * 每一项只有 `{ author, title, id }` —— 没有 created_at、没有 images、没有 comment_count，
 * 所以画不出 `PostItem` 那种卡片，只能画一行「作者 + 标题」。
 * 它也没有分页参数，一次全给，所以不存在「下一页」。
 */
async function loadStarList(): Promise<void> {
  loading.value = true
  try {
    starList.value = await getMyLikes()
    totalPage.value = 0
    noMore.value = true
  } catch {
    starList.value = []
  } finally {
    loading.value = false
  }
}

/** 关注数 / 粉丝数。数不出来就保持 0，不能因为一个计数把整页拦住 */
async function loadCounts(): Promise<void> {
  if (!user.value) return
  const id = user.value.id
  try {
    // page_size 给 1：只要 total，不想把名单拉下来
    const [following, followers] = await Promise.all([
      getFollowing(id, { page: 1, page_size: 1 }),
      getFollowers(id, { page: 1, page_size: 1 }),
    ])
    followCount.value = following.total
    fanCount.value = followers.total
  } catch {
    // 保持 0
  }
}

function loadPage(reset = false): void {
  if (pageType.value === 'post') {
    void loadPostList(reset)
  } else {
    void loadStarList()
  }
}

/* ---------------- 交互 ---------------- */

function changeTab(): void {
  loadPage(true)
}

/** InfiniteLoadMore 让去下一页时调 */
function handleNextPage(): void {
  if (noMore.value || loading.value) return
  page.value += 1
  void loadPostList()
}

function goPost(id: number): void {
  void router.push({ name: 'post', params: { id } })
}

/**
 * 右上角「…」。
 *
 * 作者这里是五项（设置 / 举报 / 拉黑 / 删除账号……），我们后端一个都没有，
 * 留下唯一一个点了有东西的：跳设置页。
 * `query.t` 那个时间戳是照作者的手法（`views\Profile.vue:265` 附近）——
 * 同一个路由重复 push 时，靠参数变化触发组件的 watch 刷新。
 */
function handleUserAction(command: string): void {
  if (command === 'setting') {
    void router.push({ name: 'setting', query: { t: Date.now() } })
  }
}

onMounted(() => {
  void loadCounts()
  loadPage(true)
})
</script>

<template>
  <div class="profile-page">
    <!-- 作者 `:3` 就是 `<main-nav title="主页" />`，注意**没有 :back** -->
    <MainNav title="主页" />

    <div v-if="user" class="main-content-wrap profile-wrap">
      <div class="profile-baseinfo">
        <div class="avatar">
          <el-avatar :size="72">{{ avatarText }}</el-avatar>
        </div>

        <div class="base-info">
          <div class="username">
            <strong>{{ user.name }}</strong>
            <el-tag v-if="isAdmin" class="top-tag" type="danger" size="small" round>
              管理员
            </el-tag>
          </div>

          <!--
            作者这里还有一行 `{{ formatDate(userInfo.created_on) }} 加入`。
            我们后端没有注册时间（`GET /me` 只有 user_id 和 role），
            为了不留一行空白的「加入」，整行去掉了（登记在 PROJECT.md §11.6）。
          -->
          <div class="userinfo">
            <span class="info-item">UID. {{ user.id }}</span>
          </div>

          <div class="userinfo">
            <!--
              作者三个计数都用 `prettyQuoteNum` 包着。
              `t=follows` 是「正在关注」、`t=followings` 是「粉丝」——
              名字和直觉反着来，这是作者的原样（见 T12 那一轮的说明）。
            -->
            <router-link
              class="following-link info-item"
              :to="{ name: 'following', query: { id: user.id, n: user.name, t: 'follows' } }"
            >
              关注&nbsp;&nbsp;{{ prettyQuoteNum(followCount) }}
            </router-link>
            <router-link
              class="following-link info-item"
              :to="{ name: 'following', query: { id: user.id, n: user.name, t: 'followings' } }"
            >
              粉丝&nbsp;&nbsp;{{ prettyQuoteNum(fanCount) }}
            </router-link>
            <span class="info-item">泡泡&nbsp;&nbsp;{{ prettyQuoteNum(postCount) }}</span>
          </div>
        </div>

        <div class="user-opts">
          <el-dropdown placement="bottom-end" trigger="click" @command="handleUserAction">
            <el-button text circle size="small" title="更多">
              <el-icon :size="16"><EllipsisHorizontal /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="setting">
                  <el-icon><SettingsOutline /></el-icon>设置
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!--
        作者是五个 tab：泡泡 / 评论 / 亮点 / 图文 / 喜欢。
        后三个（评论、亮点、图文）后端没有对应接口，砍掉；
        剩下的两个正好各有一个接口。
      -->
      <el-tabs v-model="pageType" class="profile-tabs-wrap" @tab-change="changeTab">
        <el-tab-pane label="泡泡" name="post" />
        <el-tab-pane label="喜欢" name="star" />
      </el-tabs>

      <div v-if="loading && listEmpty" class="skeleton-wrap">
        <PostSkeleton :num="PAGE_SIZE" />
      </div>

      <template v-else>
        <div v-if="listEmpty" class="empty-wrap">
          <el-empty description="暂无数据" :image-size="120" />
        </div>

        <!-- 泡泡：和广场页共用同一个卡片组件 -->
        <template v-else-if="pageType === 'post'">
          <div v-for="post in postList" :key="post.id" class="list-item">
            <PostItem :post="post" :is-mobile="isMobile" />
          </div>
        </template>

        <!-- 喜欢：接口只给 {author, title, id}，画不了卡片，画一行 -->
        <template v-else>
          <div
            v-for="like in starList"
            :key="like.id"
            class="list-item star-item"
            @click="goPost(like.id)"
          >
            <span class="star-author">{{ like.author }}</span>
            <span class="star-title">{{ like.title }}</span>
          </div>
        </template>
      </template>

      <!-- 「喜欢」是一次性给的裸数组，没有下一页，所以只给「泡泡」挂加载条 -->
      <InfiniteLoadMore
        v-if="pageType === 'post' && !listEmpty"
        :total-page="totalPage"
        :no-more="noMore"
        complete-text="没有更多泡泡了"
        @load-more="handleNextPage"
      />
    </div>

    <!--
      正常走不到这里：`router\index.ts:78` 的 `/profile` 挂着 requiresAuth，
      没 token 会被守卫弹去登录框。但 token 过期时会先进来再登出，
      那时候 user 就是 null，得给个东西挡住。
    -->
    <div v-else class="main-content-wrap">
      <div class="empty-wrap">
        <el-empty description="请先登录" :image-size="120" />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
/**
 * `.main-content-wrap` 在本项目里是**每个页面各自 scoped 定义**的
 * （HomeView / PostDetailView / FollowingView 都是这么写的，main.scss 里没有全局版），
 * 这里照抄同一套边框 + 底色写法。
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

/* ---- 头部（作者 `:360-400`）---- */
.profile-baseinfo {
  display: flex;
  padding: 16px;

  .avatar {
    width: 72px;
  }

  .base-info {
    position: relative;
    margin-left: 12px;
    width: calc(100% - 84px);

    .username {
      line-height: 16px;
      font-size: 16px;
    }

    .userinfo {
      font-size: 14px;
      line-height: 14px;
      margin-top: 10px;
      opacity: 0.75;

      .info-item {
        margin-right: 12px;
      }
    }

    .top-tag {
      transform: scale(0.75);
    }
  }

  .user-opts {
    position: relative;
    opacity: 0.75;
  }
}

.profile-tabs-wrap {
  padding: 0 16px;

  :deep(.el-tabs__header) {
    margin-bottom: 12px;
  }
}

/* ---- 「喜欢」那一行 ---- */
.star-item {
  padding: 12px 16px;
  cursor: pointer;

  &:hover {
    background-color: var(--el-fill-color-light);
  }

  .star-author {
    font-size: 14px;
    opacity: 0.75;
    margin-right: 8px;
  }

  .star-title {
    font-size: 15px;
  }
}

.empty-wrap,
.skeleton-wrap {
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 作者 `:420-424`：暗色下整块底色压深 */
:global(.dark) {
  .profile-wrap {
    background-color: rgba(16, 16, 20, 0.75);
  }
}
</style>
