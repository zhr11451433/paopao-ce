<script setup lang="ts">
/**
 * 用户主页。
 *
 * 作者 `views\User.vue` 的 `loadUser`（337-363 行）打资料接口，
 * 把昵称、关注数、粉丝数、泡泡数填进头部。
 * 我们的资料接口是 C1 的 `GET /users/:id`，字段是
 * name / following_count / follower_count / post_count / role。
 * 没有头像、注册时间、是否已关注，这三样继续空着。
 * 帖子列表仍是空态，那是 C5，这一轮不接。
 */
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ElMessage, ElMessageBox } from 'element-plus'
import { BodyOutline, EllipsisHorizontal, SettingsOutline, WalkOutline } from '@vicons/ionicons5'
import MainNav from '@/components/MainNav.vue'
import { followUser, getUser, unfollowUser } from '@/api/user'
import { useUserStore } from '@/store/user'
import { prettyQuoteNum } from '@/utils/count'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const { user: me } = storeToRefs(userStore)

/** 路由是 `/user/:id`（路径传参）。作者那边是 `/u?s=<用户名>`，见 PROJECT.md §11.6 */
const userId = computed(() => Number(route.params.id) || 0)

const userName = ref('')
const followCount = ref(0)
const fanCount = ref(0)
const postCount = ref(0)
/** C1 的 role。作者用 is_admin 显示「管理员」，我们只有 role === 'admin' */
const userRole = ref('')
const userLoading = ref(false)

/** 'post' | 'comment' | 'highlight' | 'media' | 'star'，对齐作者 template 里的 name */
const pageType = ref('post')

/** 我有没有关注他。后端查不到，只能由本轮的操作自己维护（见 §11.8） */
const following = ref(false)

const isMe = computed(() => me.value !== null && me.value.id === userId.value)
const isAdmin = computed(() => userRole.value === 'admin')
const avatarText = computed(() => (userName.value || '泡').slice(0, 1).toUpperCase())

/* ---------------- 数据 ---------------- */

/**
 * 头部。作者 `User.vue:337-358` 把资料接口的结果写进 user。
 * 我们没有 username / avatar / created_on / is_following，只写 C1 给的那几项。
 */
async function loadUser(): Promise<void> {
  if (userId.value === 0) return
  userLoading.value = true
  try {
    const profile = await getUser(userId.value)
    userName.value = profile.name
    userRole.value = profile.role
    followCount.value = profile.following_count
    fanCount.value = profile.follower_count
    postCount.value = profile.post_count
  } catch {
    // 用户不存在时后端 404，拦截器会弹红条。这里把头部清干净。
    userName.value = ''
    userRole.value = ''
    followCount.value = 0
    fanCount.value = 0
    postCount.value = 0
  } finally {
    userLoading.value = false
  }
}

/* ---------------- 交互 ---------------- */

/**
 * 右上角「…」。
 *
 * 作者那里最多有六项：私信（S4）/ 禁言·解封（后端没有）/ 关注·取关（有）/ 加好友·删好友（S4）/ 设置。
 * 本项目只剩两个是能点的：**关注 / 取关**、以及「看自己」时的**设置**。
 * 作者是「登录了才显示这个菜单」（`v-if="userInfo.id > 0"`），我们照抄。
 */
function handleUserAction(command: string): void {
  if (command === 'setting') {
    void router.push({ name: 'setting', query: { t: Date.now() } })
    return
  }
  if (command === 'follow' || command === 'unfollow') {
    void handleFollow()
  }
}

/**
 * 关注 / 取关。
 *
 * 确认框和文案都照 T12 的 `UserCard.vue`（也就是作者 `composables\useUserAction.ts`
 * 那个 `followAction`）—— 关注和取关**都弹**，用的是带绿色对勾的 `type: 'success'`。
 *
 * ⚠️ 和后端对不齐的一点：作者按 `user.is_following` 决定菜单里写「关注」还是「取消关注」，
 * 我们拿不到这个状态，所以**一律先当成「没关注」**。真点下去不会出错 ——
 * 后端 `Follow`（`internal\service\follow.go:22-63`）对「已经关注了」返回 200
 * `{"status":"您已关注该用户"}`，不会 500。此后的状态由本地这次操作维护。
 * 已登记在 PROJECT.md §11.8。
 */
async function handleFollow(): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `确定${following.value ? '取消关注 @' : '关注 @'}${userName.value} 吗？`,
      '提示',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'success' },
    )
  } catch {
    // 点了取消：ElMessageBox 会 reject，这里直接收工，不是错误
    return
  }

  try {
    if (following.value) {
      await unfollowUser(userId.value)
      ElMessage.success('操作成功')
      following.value = false
    } else {
      await followUser(userId.value)
      ElMessage.success('关注成功')
      following.value = true
    }
    // 粉丝数以资料接口为准，避免本地加减和库里对不上
    await loadUser()
  } catch {
    // 401 / 403 这些由 utils\request.ts 的拦截器统一弹红条，这里不再重复提示
  }
}

onMounted(() => {
  void loadUser()
})

/** 从一个用户主页点到另一个用户主页（同路由、只换参数）时组件不会重建，得自己重拉 */
watch(userId, () => {
  userName.value = ''
  userRole.value = ''
  followCount.value = 0
  fanCount.value = 0
  postCount.value = 0
  following.value = false
  pageType.value = 'post'
  void loadUser()
})
</script>

<template>
  <div class="user-page">
    <!-- 作者 `:3` 是 `<main-nav title="用户详情" />`，注意**没有 :back** -->
    <MainNav title="用户详情" />

    <div class="main-content-wrap profile-wrap">
      <div v-loading="userLoading" class="profile-baseinfo">
        <div class="avatar">
          <el-avatar :size="72">{{ avatarText }}</el-avatar>
        </div>

        <div class="base-info">
          <div class="username">
            <strong>{{ userName || '未知用户' }}</strong>
            <el-tag v-if="isAdmin" class="top-tag" type="danger" size="small" round>
              管理员
            </el-tag>
          </div>

          <div class="userinfo">
            <span class="info-item">UID. {{ userId }}</span>
          </div>

          <div class="userinfo">
            <router-link
              class="following-link info-item"
              :to="{ name: 'following', query: { id: userId, n: userName, t: 'follows' } }"
            >
              关注&nbsp;&nbsp;{{ prettyQuoteNum(followCount) }}
            </router-link>
            <router-link
              class="following-link info-item"
              :to="{ name: 'following', query: { id: userId, n: userName, t: 'followings' } }"
            >
              粉丝&nbsp;&nbsp;{{ prettyQuoteNum(fanCount) }}
            </router-link>
            <span class="info-item">泡泡&nbsp;&nbsp;{{ prettyQuoteNum(postCount) }}</span>
          </div>
        </div>

        <!-- 作者是「登录了才显示这个菜单」 -->
        <div v-if="me" class="user-opts">
          <el-dropdown placement="bottom-end" trigger="click" @command="handleUserAction">
            <el-button text circle size="small" title="更多">
              <el-icon :size="16"><EllipsisHorizontal /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <!-- 看自己的时候只有「设置」，和作者的判断一致 -->
                <el-dropdown-item v-if="isMe" command="setting">
                  <el-icon><SettingsOutline /></el-icon>设置
                </el-dropdown-item>
                <template v-else>
                  <el-dropdown-item v-if="following" command="unfollow">
                    <el-icon><WalkOutline /></el-icon>取消关注 {{ userName }}
                  </el-dropdown-item>
                  <el-dropdown-item v-else command="follow">
                    <el-icon><BodyOutline /></el-icon>关注 {{ userName }}
                  </el-dropdown-item>
                </template>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!--
        作者这里也是五个 tab（和 Profile.vue 同一套）。
        而本项目**一个都没得做** —— 他的数据来自
        `GET /users/:username/posts?style=post|comment|highlight|media|star`，
        本后端没有这个接口（也没有 `style` 参数）。
        所以五个 tab 全留着（照原版），点了只会换高亮，下面永远是空态。
      -->
      <el-tabs v-if="!userLoading" v-model="pageType" class="profile-tabs-wrap">
        <el-tab-pane label="泡泡" name="post" />
        <el-tab-pane label="评论" name="comment" />
        <el-tab-pane label="亮点" name="highlight" />
        <el-tab-pane label="图文" name="media" />
        <el-tab-pane label="喜欢" name="star" />
      </el-tabs>

      <div class="empty-wrap">
        <el-empty description="暂无数据" :image-size="120" />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
/**
 * `.main-content-wrap` 在本项目里是**每个页面各自 scoped 定义**的
 * （HomeView / PostDetailView / FollowingView / ProfileView 都是这么写的），
 * 这里照抄同一套边框 + 底色写法。
 */
.main-content-wrap {
  border: 1px solid var(--el-border-color);
  border-top: none;
  border-radius: 0;
  background-color: var(--el-bg-color);
}

/* ---- 头部（作者 `:584-628`）---- */
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
    /**
     * 作者在 User.vue 里写的是 `position: absolute; top:16px; right:16px`
     * （他在 Profile.vue 里写的是 `position: relative`）—— 同一个东西两种写法，
     * 我们统一用 Profile.vue 那种：`.profile-baseinfo` 是个 flex 行，
     * 用 relative 就让「…」自然落在这一行的最右边，不依赖任何祖先定位。
     */
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

.empty-wrap {
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 作者 `:647-651`：暗色下整块底色压深 */
:global(.dark) {
  .profile-wrap {
    background-color: rgba(16, 16, 20, 0.75);
  }
}
</style>
