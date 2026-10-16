<script setup lang="ts">
import { computed, ref, watch, type Component } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatPrettyDate } from '@/utils/formatTime'
import { likePost, unlikePost } from '@/api/post'
import { followUser, unfollowUser } from '@/api/user'
import PostImage from '@/components/PostImage.vue'
import { useUserStore } from '@/store/user'
import type { PostListItem } from '@/types/api'
import {
  BodyOutline,
  Heart,
  HeartOutline,
  ChatboxOutline,
  BookmarkOutline,
  EllipsisHorizontal,
  ShareSocialOutline,
  WalkOutline,
} from '@vicons/ionicons5'

const props = withDefaults(
  defineProps<{
    post: PostListItem
    /**
     * 作者把 isMobile 做成了 prop（post-item.vue:184，默认 false），
     * 而不是在组件里自己读 store。我们照做，保持一致。
     */
    isMobile?: boolean
  }>(),
  { isMobile: false },
)

const router = useRouter()
const userStore = useUserStore()

/**
 * 点赞的本地数字。
 *
 * 作者 `post-item.vue:291-307` 点完之后改的是本地的 `upvote_count`，不重新拉列表。
 * 我们照做。`null` 表示这条数据里根本没有点赞字段（我的主页用的 `GET /posts/me` 就没有），
 * 这时不显示数字，点下去也不发请求。
 */
const likeCount = ref<number | null>(
  typeof props.post.like_count === 'number' ? props.post.like_count : null,
)
const liked = ref(props.post.is_liked === true)
const liking = ref(false)

watch(
  () => [props.post.like_count, props.post.is_liked] as const,
  ([count, isLiked]) => {
    likeCount.value = typeof count === 'number' ? count : null
    liked.value = isLiked === true
  },
)

/**
 * 作者 `post-item.vue:215-228` 用 `post.user.is_following` 决定菜单写「关注」还是「取消关注」。
 * C8 把这个布尔值放到了列表项顶层 `is_following`。翻页换了一批帖子时，props 会变，所以要 watch。
 */
const following = ref(props.post.is_following === true)
watch(
  () => props.post.is_following,
  (value) => {
    following.value = value === true
  },
)

/**
 * 作者 `Home.vue:59` 传入 `:isOwner="userInfo.id == post.user_id"`，
 * 再配合 `addFollowAction`（`post-item.vue:215`）才画出关注项。
 * C9 只改这一个组件，所以自己用 store 里的登录 id 判断「是不是我的帖」，
 * 有 `user_id`、已登录、且不是自己，才显示关注 / 取关。
 * 私信、加好友作者也有（`:209-244`），对应 C41 / C42，这一轮不加。
 */
const canFollow = computed(() => {
  const authorId = props.post.user_id
  const myId = userStore.user?.id
  return Boolean(authorId) && Boolean(myId) && authorId !== myId
})

type MoreOption = {
  key: 'copyLink' | 'follow' | 'unfollow'
  label: string
  icon: Component
}

const moreOptions = computed((): MoreOption[] => {
  const options: MoreOption[] = []
  if (canFollow.value) {
    if (following.value) {
      options.push({
        key: 'unfollow',
        label: '取消关注 @' + props.post.author,
        icon: WalkOutline,
      })
    } else {
      options.push({
        key: 'follow',
        label: '关注 @' + props.post.author,
        icon: BodyOutline,
      })
    }
  }
  options.push({ key: 'copyLink', label: '复制链接', icon: ShareSocialOutline })
  return options
})

function goPostDetail(): void {
  router.push({ name: 'post', params: { id: props.post.id } })
}

/**
 * 作者的点赞是一个来回切换的接口（`post-item.vue:291` 的 `postStar`，靠返回的 status 判断加一还是减一）。
 * 我们的后端是两个：`POST /posts/:id/likes` 点赞，`DELETE /posts/:id/likes` 取消。
 * `is_liked` 决定打哪一个。
 * 赞过之后换成红色实心，照作者详情页 `post-detail.vue:180-181`：
 *   `<heart-outline v-if="!hasStarred" />`
 *   `<heart v-if="hasStarred" color="red" />`
 * 广场卡片作者没做这个切换，这里按你的要求补上，和详情页用同一套。
 */
async function toggleLike(): Promise<void> {
  if (likeCount.value === null || liking.value) return
  liking.value = true
  try {
    if (liked.value) {
      await unlikePost(props.post.id)
      liked.value = false
      likeCount.value = Math.max(0, likeCount.value - 1)
    } else {
      await likePost(props.post.id)
      liked.value = true
      likeCount.value += 1
    }
  } catch {
    // 错误文案由 utils/request.ts 统一弹，这里不再弹一遍
  } finally {
    liking.value = false
  }
}

/** 没有点赞字段时不拦截点击，让它冒泡进详情；有字段时停住，只做点赞 */
function onLikeClick(event: MouseEvent): void {
  if (likeCount.value === null) return
  event.stopPropagation()
  void toggleLike()
}

/**
 * 作者点关注会走 `useUserAction.ts:14-51` 的 `followAction`：
 * 先弹「确定关注 @名字 吗？」/「确定取消关注 @名字 吗？」，确定后再打接口。
 * 文案和 `UserCard.vue` / `UserView.vue` 已经对齐过：关注成功「关注成功」，取关成功「操作成功」。
 */
async function handleFollowAction(): Promise<void> {
  const authorId = props.post.user_id
  if (!authorId) return
  try {
    await ElMessageBox.confirm(
      `确定${following.value ? '取消关注 @' : '关注 @'}${props.post.author} 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'success',
      },
    )
  } catch {
    return
  }
  try {
    if (following.value) {
      await unfollowUser(authorId)
      following.value = false
      ElMessage.success('操作成功')
    } else {
      await followUser(authorId)
      following.value = true
      ElMessage.success('关注成功')
    }
  } catch {
    // 401 / 400 由 utils/request.ts 统一弹
  }
}

function copyPostLink(): void {

  /**
   * 作者 post-item.vue:263-266 用的是 copy-to-clipboard 这个第三方包，
   * 链接拼成 `/#/post?id=${id}&share=copy_link&t=...`（他的详情页用 query 传参）。
   *
   * 我们两处不同：
   *   - 路由是 /post/:id（路径传参），链接形状不一样；
   *   - 不引第三方包（PROJECT.md §8：要加依赖先问），改用浏览器原生剪贴板 API。
   *     localhost 属于安全上下文，navigator.clipboard 是可用的。
   */
  const url = `${window.location.origin}/#/post/${props.post.id}`
  navigator.clipboard
    .writeText(url)
    .then(() => ElMessage.success('链接已复制到剪贴板'))
    .catch(() => ElMessage.error('复制失败，请手动复制'))
}

/** el-dropdown 的 command 类型是 string | number | object，不是纯 string */
function handleMoreAction(command: string | number | object): void {
  if (command === 'follow' || command === 'unfollow') {
    void handleFollowAction()
    return
  }
  if (command === 'copyLink') {
    copyPostLink()
  }
}
</script>

<template>
  <!--
    作者 post-item.vue:2 的根节点：整张卡片都是可点区域，点哪都进详情页，
    卡片内部的链接 / 按钮再用 @click.stop 把事件挡住。
  -->
  <div class="post-item" @click="goPostDetail">
    <!--
      作者这里用的是 naive 的 <n-thing content-indented>（post-item.vue:3）。
      Element Plus **没有对应组件**（PROJECT.md §6 列的就是「用 div 自己写」），
      所以下面按原版视觉自己搭：左边一列放头像，右边一列装头部 / 正文 / 附件 / 操作栏。

      实测作者仓库里的官方预览图（.assets\readme\preview-light.jpeg）：
      正文、图片、操作栏**全都缩进到头像右边**，操作栏横向铺满整列。
    -->
    <div class="post-avatar-col">
      <!-- 作者是 :src="post.user.avatar"（post-item.vue:5）。我们没有头像字段，只能默认占位 -->
      <el-avatar :size="30" />
    </div>

    <div class="post-main">
      <div class="post-header">
        <!--
          作者 post-item.vue:9-18 是 router-link，query 用用户名（?s=）。
          我们的用户页是 /user/:id，所以用 user_id。
          @click.stop 照抄他的第 10 行：点名字不能再冒泡去开详情。
          没有 user_id 的列表（我的主页 GET /posts/me）退回不可点的文字。
        -->
        <span class="nickname-wrap">
          <router-link
            v-if="post.user_id"
            class="username-link"
            :to="{ name: 'user', params: { id: post.user_id } }"
            @click.stop
          >
            {{ post.author }}
          </router-link>
          <template v-else>{{ post.author }}</template>
        </span>

        <div class="item-header-extra">
          <!-- 作者：`{{ post.ip_loc }} · {{ formatPrettyDate(...) }}`。我们没有 ip_loc，只显示时间 -->
          <span v-if="!isMobile" class="timestamp">{{ formatPrettyDate(post.created_at) }}</span>

          <!-- 作者的 n-dropdown（post-item.vue:60-74）：桌面端悬停展开、移动端点击展开 -->
          <el-dropdown
            :trigger="isMobile ? 'click' : 'hover'"
            placement="bottom-end"
            @click.stop
            @command="handleMoreAction"
          >
            <el-button text circle>
              <!--
                作者用的是 @vicons/material 的 MoreHorizFilled。
                本项目没装 material 那套图标（PROJECT.md §2 只列了 ionicons5 / tabler），
                所以换成 ionicons5 的 EllipsisHorizontal —— 三条横点，视觉一致。
              -->
              <el-icon><EllipsisHorizontal /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="opt in moreOptions" :key="opt.key" :command="opt.key">
                  <el-icon><component :is="opt.icon" /></el-icon>
                  {{ opt.label }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- 作者在移动端把时间挪到名字下面（post-item.vue:48-52） -->
      <div v-if="isMobile" class="timestamp-mobile">{{ formatPrettyDate(post.created_at) }}</div>

      <!--
        正文。作者的正文在 post.texts[].content 里，而且用 preparePost() 把 #话题 / @某人
        渲染成可点链接（post-item.vue:77-94）。
        我们的后端把正文存在 title 字段里（没有单独的 content 字段），也没有话题体系，
        所以这里就是纯文本；pre-wrap 保留用户敲的换行，和作者的 .post-text 一致。
      -->
      <div v-if="post.title" class="post-text">{{ post.title }}</div>

      <!--
        图片（T5）。作者这个位置依次是 post-attachment / post-image / post-video / post-link
        （post-item.vue:96-114），我们现在有图片和外链两种，顺序也照他排 —— 外链在最后。

        ⚠️ 必须写成 `post.images?.length` 而不是直接 v-for ——
        后端是 Go，nil 切片序列化成 `null` 而不是 `[]`（PROJECT.md §7.5 第 1 条），
        直接 v-for 会在没有图的帖子上抛 "null is not iterable"。
      -->
      <PostImage v-if="post.images?.length" :imgs="post.images" />

      <!--
        外链组件 PostLink.vue 还在，但这轮从卡片上摘掉了。
        后端没有 links 字段，假链接不能再挂到真帖子上（T18 开工说明选的方案 ②）。
      -->

      <!--
        操作栏。作者用 <n-space justify="space-between">（post-item.vue:116）把指标横向铺满整列，
        这里用 flex + justify-content: space-between 等价实现。
        点赞数字对应他的 upvote_count（post-item.vue:121）。0 也显示。
      -->
      <div class="post-action">
        <div class="opt-item hover" @click="onLikeClick">
          <el-icon class="opt-item-icon" :size="18">
            <HeartOutline v-if="!liked" />
            <Heart v-else color="red" />
          </el-icon>
          <template v-if="likeCount !== null">{{ likeCount }}</template>
        </div>
        <div class="opt-item hover" @click.stop="goPostDetail">
          <el-icon class="opt-item-icon" :size="18"><ChatboxOutline /></el-icon>
          {{ post.comment_count }}
        </div>
        <div class="opt-item hover">
          <el-icon class="opt-item-icon" :size="18"><BookmarkOutline /></el-icon>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
/**
 * 样式照抄作者 post-item.vue:376-445，只去掉已经不存在的两条：
 *   - `.n-thing-avatar`、`.n-thing-header` 是 naive 内部类名，EP 下没有，
 *     换成我们自己搭的 `.post-avatar-col`、`.post-header`。
 * 颜色值原样保留（作者写死的 #f7f9f9 / #18181c / rgba(16,16,20,.75)），
 * 这样亮暗两套底色和他的观感一致 —— PROJECT.md §0.1：像比好重要。
 */
.post-item {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
  display: flex; /* n-thing 的默认布局：头像一列、内容一列 */
  align-items: flex-start;

  /**
   * 头像列宽 42px = 头像 30px + 12px 间距。
   * 这个 42 不是我编的：作者的 compose.vue 里 `.compose-user { width: 42px }`、
   * 头像 `:size="30"`、`.compose-options { padding-left: 42px }` 用的就是它，
   * 所以发帖框和帖子卡片的正文左边缘在**同一条竖线**上 —— 原版就是这么对齐的。
   */
  .post-avatar-col {
    width: 42px;
    flex-shrink: 0;
  }

  .post-main {
    flex: 1;
    min-width: 0; /* 不加这句，长文本会把 flex 子项撑破 */
  }

  .post-header {
    display: flex;
    align-items: center;
    /* 作者覆盖 naive 的 .n-thing-header：line-height 16px、下边距 8px */
    line-height: 16px;
    margin-bottom: 8px;
  }

  .nickname-wrap {
    font-size: 14px;
  }

  .username-wrap {
    font-size: 14px;
    opacity: 0.75;
  }

  .top-tag {
    transform: scale(0.75);
  }

  .item-header-extra {
    display: flex;
    align-items: center;
    /* naive 的 .n-thing-header 自己就是 space-between，所以作者没写这句；
       我们手搭的 .post-header 要自己把「时间 + 更多」推到右边 */
    margin-left: auto;
    opacity: 0.75;

    .timestamp {
      font-size: 12px;
    }
  }

  .timestamp-mobile {
    margin-top: 2px;
    opacity: 0.75;
    font-size: 11px;
  }

  .post-text {
    text-align: justify;
    overflow: hidden;
    white-space: pre-wrap;
    word-break: break-all;
  }

  /**
   * ⚠️ 这条 margin-top 是**我加的**，作者没有。
   * 原因：naive 的 n-thing 内部会给 footer / action 一截上边距，而 naive-ui
   * 没装在作者仓库里（`node_modules/naive-ui` 不存在），我读不到它的具体数值，
   * 只能给一个 12px。原版截图也看不出这一处（帖子被裁掉了）。
   * 这是本轮唯一需要你对着原版截图校一下的数。
   */
  .post-action {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 12px;
  }

  .opt-item {
    display: flex;
    align-items: center;
    opacity: 0.7;

    .opt-item-icon {
      margin-right: 10px;
    }
  }

  &:hover {
    background: #f7f9f9;
  }

  /**
   * 作者写的是 `&.hover { cursor: pointer }`，但整张卡片的根节点只有 `post-item`
   * 一个类、没有 `hover`，而带 `hover` 的是**里面的** `.post-text` 和 `.opt-item` ——
   * 它们不是根节点，匹配不上这条规则。
   * 也就是说：原版里鼠标移到卡片上是**普通箭头**，不是小手。
   * 我保持原样，没有替你"修好"（PROJECT.md §0.1）。
   */
  &.hover {
    cursor: pointer;
  }
}

/* 作者 post-item.vue:438-445 的暗色写法，原样搬过来 */
.dark .post-item {
  &:hover {
    background: #18181c;
  }

  background-color: rgba(16, 16, 20, 0.75);
}
</style>
