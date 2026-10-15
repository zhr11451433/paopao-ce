<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import { TrashOutline } from '@vicons/ionicons5'
import { deleteComment } from '@/api/post'
import { useUserStore } from '@/store/user'
import { formatPrettyTime } from '@/utils/formatTime'
import type { CommentItem as CommentItemData } from '@/types/api'

/**
 * 对应作者 `components\comment-item.vue`（303 行）。
 *
 * ⚠️ 我们这一版比原版瘦很多，因为本项目后端的评论**只有四个字段**
 * （`F:\paopao\internal\service\comment.go:69-74` 的 `CommentList`）：
 *
 *     { id, author, created_at, content }
 *
 * 对照作者 `comment-item.vue` 用到的东西，砍掉的是：
 *   · `comment.user.avatar`（`:5`）        → 没有 avatar 字段 → 空头像
 *   · `comment.user.nickname` / `username` → 只有一个 `author` → 只显示它，且**不可点**
 *   · `comment.is_essence` + 精选按钮      → 后端没有「精选评论」这个概念 → 整个不做
 *   · `comment.ip_loc`（`:36`）            → 没有 → **改用 `created_at` 的时间填这个位置**
 *   · `comment.contents` / `texts`（`:90-98`）→ 只有一整段 `content` → 显示纯文本，
 *                                              **没有 `#话题` `@人` 变绿**
 *   · `comment.imgs`（`:101-103`）         → 没有 → 不做
 *   · `comment.replies` + compose-reply + reply-item（`:105-123`）→ 属于 S9 → 不做
 *
 * 另外作者 `:1-2` 的根节点外面没有包 `<n-thing>` 的兄弟结构，我们照 PostItem.vue 的办法
 * 手写两列（`.comment-avatar-col` + `.comment-main`），因为本项目没有 n-thing（§6）。
 */

const props = defineProps<{
  comment: CommentItemData
}>()

/** 作者 `comment-item.vue:149-151` 也是只 emit 一个 `reload` */
const emit = defineEmits<{
  (e: 'reload'): void
}>()

const userStore = useUserStore()

/**
 * 「这条评论是不是我发的」—— **只能比名字**。
 *
 * 作者 `:64-67` 比的是 id：`userInfo.is_admin || userInfo.id === comment.user.id`。
 * 本项目的评论**没有 user_id**，只有 `author` 这个名字，所以只能退而求其次。
 * 后果：同名的两个人会互相看到对方的删除按钮。前端补不了，已登记在 PROJECT.md 11.6 / 11.7。
 * （管理员那一支也不需要 —— 本项目的评论接口不看角色，只有「本人 / 非本人」。）
 */
const isOwner = computed(() => {
  const name = userStore.user?.name
  return typeof name === 'string' && name !== '' && name === props.comment.author
})

/**
 * 作者 `:212-223` 的 `execDelAction()`：
 *   `deleteComment({id})` → `window.$message.success('删除成功')` → **setTimeout 50ms** → `reload()`。
 *
 * 那个 50ms 是照抄的（作者就是这么写的），不是笔误 —— 让提示先渲染出来再刷新列表。
 * `catch` 里保持空：错误提示由 `utils/request.ts` 的响应拦截器统一弹，
 * 这里再弹一次你会看到两条一样的红条。
 */
async function execDelAction(): Promise<void> {
  try {
    await deleteComment(props.comment.id)
    ElMessage.success('删除成功')
    window.setTimeout(() => {
      emit('reload')
    }, 50)
  } catch {
    // 错误提示由 utils/request.ts 的拦截器统一弹
  }
}
</script>

<template>
  <div class="comment-item">
    <!-- 作者 `:4-6`：`<n-avatar round :size="30" :src="comment.user.avatar" />` —— 我们没 avatar，只能空着 -->
    <div class="comment-avatar-col">
      <el-avatar :size="30" />
    </div>

    <div class="comment-main">
      <!--
        作者 `:7-32` 的 `#header`（昵称 + @用户名 + 精选标签）+ `:33-89` 的 `#header-extra`
        （ip_loc + 精选按钮 + 删除按钮）。n-thing 会把这两块放在同一行的左右两端，
        我们用 flex + space-between 还原（和 PostItem.vue 的 `.post-header` 一个写法）。
      -->
      <div class="comment-header">
        <!--
          作者 `:9-19` 这里是一个 `<router-link :to="{name:'user', query:{s:username}}">`。
          我们的评论只有 `author` 一个名字、没有 username 也没有 user_id，
          点进去都不知道该看谁的主页，所以**做成纯文本、不可点**（见文件头的说明）。
        -->
        <span class="nickname-wrap">{{ comment.author }}</span>

        <div class="opt-wrap">
          <!--
            作者 `:35-37` 这里显示的是 `comment.ip_loc`（发评论时的IP归属地）。
            本项目后端没有 ip_loc，这个位置改放**这条评论的时间** ——
            这是本轮唯一的一处内容替换，已登记在 PROJECT.md 11.6。
          -->
          <span class="timestamp">{{ formatPrettyTime(comment.created_at) }}</span>

          <!--
            作者 `:63-87` 用的是 `n-popconfirm`（点一下在按钮旁边冒出一个小气泡确认），
            Element Plus 有同名的 `el-popconfirm`（`#reference` 插槽 = 触发它的那个按钮）。
            条件从作者的 `userInfo.is_admin || userInfo.id === comment.user.id`
            换成了 `isOwner`（只能比名字，见上面）。
          -->
          <el-popconfirm
            v-if="isOwner"
            title="是否删除这条评论？"
            confirm-button-text="确认"
            cancel-button-text="取消"
            @confirm="execDelAction"
          >
            <template #reference>
              <el-button class="action-btn" text circle size="small" title="删除">
                <el-icon><TrashOutline /></el-icon>
              </el-button>
            </template>
          </el-popconfirm>
        </div>
      </div>

      <!-- 作者 `:90-98` 的正文：我们只有一整段 `content`，直接显示（原来的 v-html + parsePostTag 不做） -->
      <span class="comment-text">{{ comment.content }}</span>
    </div>
  </div>
</template>

<style scoped lang="scss">
/**
 * 作者 `comment-item.vue:240-302`。他那个文件本来就是 `<style lang="less" scoped>`，
 * 所以除了把 n-thing 的两列换掉之外，其余逐条照搬。
 */
.comment-item {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;
  display: flex;

  /* 对应作者 `post-item.vue` / `post-detail.vue` 里 n-thing 的头像列宽（我们自己写，见文件头） */
  .comment-avatar-col {
    width: 42px;
    flex: none;
  }

  .comment-main {
    width: calc(100% - 42px);
  }

  /* n-thing 把 header 和 header-extra 摆在同一行左右两端，我们手写 */
  .comment-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .nickname-wrap {
    font-size: 14px;
  }

  .opt-wrap {
    display: flex;
    align-items: center;

    .timestamp {
      opacity: 0.75;
      font-size: 12px;
    }

    .action-btn {
      margin-left: 4px;
    }
  }

  .comment-text {
    display: block;
    text-align: justify;
    overflow: hidden;
    white-space: pre-wrap;
    word-break: break-all;
  }
}

/* 作者 `:295-302` 的暗色：`.dark { .comment-item { background-color: rgba(16,16,20,.75) } } */
:global(.dark) .comment-item {
  background-color: rgba(16, 16, 20, 0.75);
}
</style>
