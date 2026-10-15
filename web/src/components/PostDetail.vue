<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  BookmarkOutline,
  ChatboxOutline,
  EllipsisHorizontal,
  Heart,
  HeartOutline,
  ShareSocialOutline,
  TrashOutline,
} from '@vicons/ionicons5'
import PostImage from '@/components/PostImage.vue'
import { deletePost, likePost, unlikePost } from '@/api/post'
import { useUserStore } from '@/store/user'
import { formatPrettyTime } from '@/utils/formatTime'
import type { PostDetail } from '@/types/api'

/**
 * 作者 `components\post-detail.vue:277-282` 的 props 只有一个 `post`。
 * 我们的 `post` 形状是**本项目后端**的 `PostDetail`（`src\types\api.ts`），
 * 和作者的 `Item.PostProps` 差一大截，见下面每一条的说明。
 *
 * 外链不在这个组件里渲染：后端没有 links，T18 按方案 ② 把 PostLink 从详情上摘掉了。
 * 组件文件 `PostLink.vue` 还留着。
 */
const props = defineProps<{
  post: PostDetail
}>()

/** 作者 `post-detail.vue:314-316`：`(e: 'reload', post_id: number)` */
const emit = defineEmits<{
  (e: 'reload', postId: number): void
}>()

const userStore = useUserStore()

/**
 * 「这条泡泡是不是我发的」。
 *
 * 作者比的是 id：`post-detail.vue:333` 的 `userInfo.value.id != props.post.user.id`。
 * 详情接口现在有 `user_id`，改成比 id。比名字会让同名的两个人互相看到「删除」。
 */
const isOwner = computed(() => {
  const myId = userStore.user?.id
  return typeof myId === 'number' && myId > 0 && myId === props.post.user_id
})

/** 和 PostItem 同一套：本地改数字，不整页重拉。详情接口这两个字段是必有的 */
const likeCount = ref(props.post.like_count)
const liked = ref(props.post.is_liked)
const liking = ref(false)

watch(
  () => [props.post.like_count, props.post.is_liked] as const,
  ([count, isLiked]) => {
    likeCount.value = count
    liked.value = isLiked
  },
)

async function toggleLike(): Promise<void> {
  if (liking.value) return
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
    // 错误文案由 utils/request.ts 统一弹
  } finally {
    liking.value = false
  }
}

/**
 * 作者 `post-detail.vue:329-...` 的 `adminOptions` 菜单里有十几项：
 * 私信 / 关注 / 取消关注 / 加好友 / 删除 / 锁定 / 解锁 / 置顶 / 取消置顶 /
 * 设为亮点 / 取消亮点 / 可见度子菜单（公开·私密·关注可见·好友可见）。
 *
 * 认得出是不是自己了（比 user_id），但菜单里其余项仍依赖后端没有的字段：
 *   is_lock、is_top、is_essence、visibility，
 *   以及好友 / 关注关系（follow.go 有接口，但没有「我和他是什么关系」的查询）。
 *
 * 所以这里只留**唯一能做的**一项：删除 —— 因为后端有 `DELETE /posts/:id`（router.go:47）。
 * 而且他的菜单是「非作者显示关注那几项、作者显示后面那一串」，
 * 我们去掉关注那几项之后，**非作者就没有任何一项可显示**，
 * 所以模板里那颗「…」按钮是 `v-if="isOwner"` 的（他是常驻的）。
 */
const MORE_OPTIONS = [{ key: 'delete', label: '删除', icon: TrashOutline }] as const

/** el-dropdown 的 command 类型是 string | number | object，不是纯 string */
function handleMoreAction(command: string | number | object): void {
  if (command === 'delete') void confirmDelete()
}

/**
 * 作者 `post-detail.vue:67-76` 用的是 `<n-modal preset="dialog" title="提示"
 * content="确定删除该泡泡动态吗？" positive-text="确认" negative-text="取消">`。
 * EP 里没有 `preset="dialog"` 这种东西，对应的就是 `ElMessageBox.confirm` ——
 * 项目里 `AppSidebar.vue` 的退出登录用的也是它，保持一致。
 */
async function confirmDelete(): Promise<void> {
  try {
    await ElMessageBox.confirm('确定删除该泡泡动态吗？', '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
    })
  } catch {
    // 点了「取消」—— 什么都不做
    return
  }

  try {
    await deletePost(props.post.id)
    ElMessage.success('删除成功')
    // 作者的 `post-detail.vue:557` 之后也是回通知外面重新拉一次
    emit('reload', props.post.id)
  } catch {
    // 错误提示由 utils/request.ts 的响应拦截器统一弹，这里不再弹一遍
  }
}

/**
 * 分享 = 复制链接。
 *
 * 作者 `post-detail.vue` 的 `handlePostShare` 用的也是 `copy()`（copy-to-clipboard 包）
 * 拼 `window.location.origin + '/#/post?id=' + id`。本项目不引入那个包（§2），
 * 用浏览器自带的 `navigator.clipboard`；它只在 https 或 localhost 下可用，
 * 所以失败时给一句提示让人手动复制，而不是静默什么都不发生。
 */
async function handleShare(): Promise<void> {
  const url = `${window.location.origin}/#/post/${props.post.id}`
  try {
    await navigator.clipboard.writeText(url)
    ElMessage.success('链接已复制')
  } catch {
    ElMessage.warning(`复制失败，请手动复制：${url}`)
  }
}
</script>

<template>
  <!--
    作者 `post-detail.vue:2` 的根节点是 `<div class="detail-item" @click="goPostDetail(post.id)">`
    —— **整块可点、点了跳详情**。但详情页自己就是详情，再跳一次没有意义，
    所以详情页这个组件**不搬那个 @click**（列表卡片才需要）。
  -->
  <div class="detail-item">
    <!--
      作者用的 `n-thing`（`post-detail.vue:3`）和 `post-item.vue` 是同一个组件。
      我们在 `PostItem.vue` 里已经手工搭过它的两列骨架（头像 42px + 右侧内容），
      这里**照同一套写**，两个页面的版式才是同一个模子出来的。
    -->
    <div class="post-avatar-col">
      <!-- 作者的 `:src="post.user.avatar"`；我们没有 avatar 字段，用 EP 的默认占位 -->
      <el-avatar :size="30" />
    </div>

    <div class="post-main">
      <div class="post-header">
        <!--
          作者 `post-detail.vue:8-18` 在这里放的是
            <router-link class="username-link" :to="{name:'user',query:{s:username}}">{{ nickname }}</router-link>
            <span class="username-wrap"> @{{ username }} </span>
          我们没有 nickname / username 之分，只有 author 一个名字。
          用户页是 /user/:id，所以用 user_id 跳，不再用作者的 ?s=用户名。
        -->
        <span class="nickname">
          <router-link
            v-if="post.user_id"
            class="username-link"
            :to="{ name: 'user', params: { id: post.user_id } }"
          >
            {{ post.author }}
          </router-link>
          <template v-else>{{ post.author }}</template>
        </span>

        <div class="item-header-extra">
          <!--
            作者 `post-detail.vue:49-63` 的 n-dropdown：trigger="click"（列表卡片才是 hover）
          -->
          <el-dropdown v-if="isOwner" trigger="click" @command="handleMoreAction">
            <el-button text circle size="small">
              <el-icon><EllipsisHorizontal /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="opt in MORE_OPTIONS" :key="opt.key" :command="opt.key">
                  <el-icon><component :is="opt.icon" /></el-icon>
                  {{ opt.label }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!--
        作者 `post-detail.vue` 的 #description 是 `v-for` 的 `post.texts`（一个帖子可以有多段），
        我们的 `PostDetail.content` 就是**一整段字符串**，所以不用循环。
        也**没有** `v-html="parsePostTag(...)"` —— 我们没有 `utils/content.ts`，
        #话题 和 @人 都不会变绿、不可点（同 11.6 登记的那一类）。

        另外：本项目的 `post.title`（发帖时取正文前 100 字）**详情页不显示** ——
        作者的详情页压根没有标题这一栏（他只有 `texts`），显示了会和正文头上重复一遍。
      -->
      <div class="post-text">{{ post.content }}</div>

      <!--
        作者 `post-detail.vue` 的 #footer 顺序是：
          附件 / 图片 / 视频 / 链接
        我们只有「图片」和「外链」两样（其余三样后端没有，见 PROJECT.md 11.4 的 S11）。
      -->
      <PostImage v-if="post.images?.length" :imgs="post.images" />

      <!--
        作者 `post-detail.vue:758-762` 的 `.timestamp`：`发布于 {{ formatPrettyTime(post.created_on) }}`
        他用的是 **formatPrettyTime**（带 HH:mm），不是列表卡片那个 formatPrettyDate。
      -->
      <div class="timestamp">发布于 {{ formatPrettyTime(post.created_at) }}</div>

      <!--
        作者 `post-detail.vue` 的 #action：`n-space justify="space-between"` 里四个 .opt-item ——
        点赞（heart + upvote_count）/ 评论（chatbox + comment_count）/
        收藏（bookmark，已收藏时换成实心且染成 #ff7600）/ 分享（share + share_count）。

        点赞数字用 like_count（作者写的是 upvote_count）。
        评论数、收藏数、分享数后端仍然没有，这三个图标旁边继续不显示数字。
        收藏点了仍然不发请求。分享是复制链接。
      -->
      <div class="opts-wrap">
        <div class="opt-item hover" @click.stop="toggleLike">
          <!-- 作者 post-detail.vue:180-181：没赞空心，赞过红色实心 -->
          <el-icon class="opt-item-icon" :size="20">
            <HeartOutline v-if="!liked" />
            <Heart v-else color="red" />
          </el-icon>
          {{ likeCount }}
        </div>
        <div class="opt-item hover" @click.stop>
          <el-icon class="opt-item-icon" :size="20"><ChatboxOutline /></el-icon>
        </div>
        <div class="opt-item hover" @click.stop>
          <el-icon class="opt-item-icon" :size="20"><BookmarkOutline /></el-icon>
        </div>
        <div class="opt-item hover" @click.stop="handleShare">
          <el-icon class="opt-item-icon" :size="20"><ShareSocialOutline /></el-icon>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
/*
 * 作者 `post-detail.vue:712-769` 的样式，逐条搬。
 *
 * ⚠️ 注意它和 `post-item.vue` 的**三处不同**（不是笔误，是作者就是这么分的）：
 *   1. `.detail-item` 有底色 `background: #f7f9f9`（卡片是鼠标悬停时才变色）
 *   2. `.post-text` 是 `font-size: 16px`（卡片上没写，所以是正文默认的 14px）
 *   3. 操作栏叫 `.opts-wrap`、`margin-top: 20px`（卡片上叫 `.post-action`，间距也不同）
 */
.detail-item {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
  background: #f7f9f9;

  /* 手工搭 n-thing 的两列骨架，和 PostItem.vue 用同一组数值 */
  display: flex;
}

.post-avatar-col {
  width: 42px;
  flex: none;
}

.post-main {
  width: calc(100% - 42px);
}

.nickname {
  font-size: 14px;
}

.post-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  line-height: 16px;
  margin-bottom: 8px;
}

.item-header-extra {
  display: flex;
  align-items: center;
  opacity: 0.75;
}

.post-text {
  font-size: 16px;
  text-align: justify;
  overflow: hidden;
  white-space: pre-wrap;
  word-break: break-all;
}

.timestamp {
  opacity: 0.75;
  font-size: 12px;
  margin-top: 10px;
}

.opts-wrap {
  margin-top: 20px;
  display: flex;
  justify-content: space-between;

  .opt-item {
    display: flex;
    align-items: center;
    opacity: 0.7;

    /*
     * 作者这里是给「图标后面的那个数字」留的间距（`{{ post.upvote_count }}` 等）。
     * 我们没有数字可显示，这条其实**什么都不影响**，但按 §0.1 原样留着，
     * 等后端把计数补上就不用再回来改样式了。
     */
    .opt-item-icon {
      margin-right: 10px;
    }

    &.hover {
      cursor: pointer;
    }
  }
}

/* 作者 `post-detail.vue:764-767`：`.dark .detail-item { background: #18181c }` */
:global(.dark) .detail-item {
  background: #18181c;
}
</style>
