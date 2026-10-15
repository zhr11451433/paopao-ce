<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MoreFilled } from '@element-plus/icons-vue'
import { BodyOutline, WalkOutline } from '@vicons/ionicons5'
import { followUser, unfollowUser } from '@/api/user'

/**
 * 作者 `components\user-card.vue` 的 props（他的 `:69-77`）：
 *
 *   contact  一条用户信息
 *   type     'contact' | 'follow'，默认 'contact'
 *
 * 他的 `contact` 是 `Item.ContactItemProps`，有
 *   user_id / username / nickname / avatar / is_following / created_on 六个字段。
 * 我们只有两个半，逐条说明为什么：
 *
 *   user_id      → 换成 `id`。本项目后端 `internal\service\follow.go:87-90`
 *                  的关注列表返的是 `{"follow_id", "name"}`，粉丝列表（`:147-150`）
 *                  返的是 `{"fan_id", "name"}` —— 两个字段名不同、值都是**对方的用户 id**，
 *                  所以由页面映射成统一的一个 `id`（见 FollowingView.vue）。
 *   name         → 本项目的 `db.User`（`internal\db\model.go:9-15`）只有
 *                  Name / Email / Password / Role 四个字段，
 *                  **作者那种「用户名 + 昵称」两套名字我们没有**，只有一个 Name。
 *   is_following → 后端不给（关注列表接口的返回里没有「我是否已关注他」），
 *                  改由页面按当前 Tab 推断后传进来（见 PROJECT.md 11.6 的 T12 行）。
 *   avatar       → 没有，头像留空（登记在 11.8）。
 *   created_on   → 没有，「xxx 加入」那一行整行不做。
 */
const props = withDefaults(
  defineProps<{
    contact: {
      id: number
      name: string
      /** 作者是 `contact.is_following`，我们由外面传（后端不返回） */
      isFollowing?: boolean
    }
    /** 作者 `:74-76`，默认 'contact'。Following.vue 传的是 'follow' */
    type?: 'contact' | 'follow'
  }>(),
  {
    type: 'contact',
  },
)

/**
 * 作者 `:79-80`：
 *   showFollowingTag   = type === 'follow'
 *   enableFollowAction = type === 'follow'
 * 我们这里两者都还要**再且上一个「已关注」**才算数 —— 因为 `isFollowing` 是外面给的，
 * 一个 `type="follow"` 但 `isFollowing: false` 的卡片不该被打上「已关注」。
 */
const following = ref(props.contact.isFollowing === true)

const showFollowingTag = computed(() => props.type === 'follow' && following.value)
const enableFollowAction = computed(() => props.type === 'follow')

/** 作者 `:82-85`，只留了我们在做的那一个 */
const emit = defineEmits<{
  (e: 'unfollow-success'): void
}>()

/**
 * 点「关注 @某人」/「取消关注 @某人」（作者 `:95-107` + `composables\useUserAction.ts:14-52`）：
 *
 * 他的流程是「先弹一个确认框，确认了才发请求」，而且**关注和取关都弹**，
 * 弹框用的是 naive 的 `dialog.success`（`useUserAction.ts:16`）——
 * 也就是一个带绿色对勾图标的提示框。EP 的等价物是 `ElMessageBox.confirm` 加 `type: 'success'`。
 *
 * 文案逐字照抄（`useUserAction.ts:18-22`）：
 *   '确定' + (已关注 ? '取消关注 @' : '关注 @') + 用户名 + ' 吗？'，标题「提示」，
 *   按钮「确定」「取消」。
 * 两句话的提示也照抄（`:31` 取关是「操作成功」，`:42` 关注是「关注成功」）。
 *
 * 差异只有两处，都是后端造成的：
 *   1. 作者的接口是按 `user_id` 调的，我们就按 `id`（同一个东西）；
 *   2. 他成功后直接改 `props.contact.is_following`，我们改自己的 `following`
 *      —— 因为我们这边的 `contact` 是页面临时拼出来的对象，改了下一轮请求也带不回来。
 */
async function handleFollowUser(): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `确定${following.value ? '取消关注 @' : '关注 @'}${props.contact.name} 吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'success',
      },
    )
  } catch {
    // 点了「取消」（ElMessageBox 用 reject 表示取消），什么都不做
    return
  }

  try {
    if (following.value) {
      await unfollowUser(props.contact.id)
      following.value = false
      ElMessage.success('操作成功')
      // 作者 `:98-102`：只有「本来关注 → 现在不关注」才通知外面刷新列表
      emit('unfollow-success')
    } else {
      await followUser(props.contact.id)
      following.value = true
      ElMessage.success('关注成功')
    }
  } catch {
    // 失败提示由 `utils\request.ts` 的响应拦截器统一弹，这里不再重复弹一次
  }
}

/**
 * 作者的 `handleAction`（`:137-162`）有三个分支：follow / unfollow / whisper。
 * 我们只做前两个 —— **「私信」整项砍掉了**：本项目后端没有私信接口，
 * 他的 `whisper` 分支还要造一个 `Item.UserInfo` 塞给私信组件（`:144-156`），
 * 那个组件我们也没有（属于 PROJECT.md 的 S8）。
 * 选项只剩一个，但下拉框还留着 —— 照原版的样子。
 */
function handleAction(command: string | number | object): void {
  if (command === 'follow' || command === 'unfollow') {
    void handleFollowUser()
  }
}
</script>

<template>
  <div class="user-card">
    <!--
      作者 `:4-6` 是 `n-thing` 的 `#avatar` 插槽：`<n-avatar :size="54" :src="contact.avatar" />`。
      我们没有 avatar 字段，留空占位（同 PostDetail.vue 的处理）。
      n-thing content-indented 会把右侧内容缩进「头像宽 + 间距」，这里用手写的两列还原：
      54 + 16 = 70px。
    -->
    <div class="user-avatar-col">
      <el-avatar :size="54" />
    </div>

    <div class="user-main">
      <div class="user-header">
        <div class="user-name-wrap">
          <!--
            作者 `:8-19`：昵称是个 router-link，点击进他的用户主页。
            我们的路由是 `/user/:id`（不是作者的 `?s=用户名`），所以按 id 跳。
            ⚠️ `UserView.vue` 目前还是 0 字节（T16 才做），点进去中栏是空的 —— 已登记在 11.8。
            `@click.stop` 照抄 `:10`：避免点名字顺带把整张卡片的点击事件也触发了。
          -->
          <span class="nickname-wrap">
            <router-link
              class="username-link"
              :to="{ name: 'user', params: { id: contact.id } }"
              @click.stop
            >
              {{ contact.name }}
            </router-link>
          </span>

          <!-- 作者 `:21-25`，「已关注」绿标签 -->
          <el-tag v-if="showFollowingTag" class="top-tag" type="success" size="small" round>
            已关注
          </el-tag>
        </div>

        <!-- 作者 `:35-53` 的 `#header-extra`：一个「⋯」下拉 -->
        <div class="item-header-extra">
          <el-dropdown placement="bottom-end" trigger="click" @command="handleAction">
            <el-button text circle>
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-if="enableFollowAction"
                  :command="following ? 'unfollow' : 'follow'"
                >
                  <!-- 作者 `:123` / `:128` 两个图标 -->
                  <el-icon>
                    <component :is="following ? WalkOutline : BodyOutline" />
                  </el-icon>
                  {{ (following ? '取消关注 @' : '关注 @') + contact.name }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!-- 作者 `:26-33`：`UID. xxx` + `xxx 加入`。加入时间我们没有，只留 UID -->
      <div class="user-info">
        <span class="info-item">UID. {{ contact.id }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
/* 作者 `:166-200` 逐条照抄，只有两列布局那几条是手写的 */
.user-card {
  width: 100%;
  box-sizing: border-box;
  padding: 12px 16px;
  display: flex;
  align-items: flex-start;

  &:hover {
    background: #f7f9f9;
  }
}

/* n-thing content-indented：54px 头像 + 16px 间距 */
.user-avatar-col {
  width: 70px;
  flex: none;
}

.user-main {
  width: calc(100% - 70px);
}

/* n-thing 把 #header 和 #header-extra 摆在一行的左右两端，手写还原 */
.user-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-name-wrap {
  display: flex;
  align-items: center;
}

.nickname-wrap {
  line-height: 16px;
  font-size: 16px;
}

.top-tag {
  transform: scale(0.75);
}

.user-info {
  .info-item {
    font-size: 14px;
    line-height: 14px;
    margin-right: 8px;
    opacity: 0.75;
  }
}

.item-header-extra {
  display: flex;
  align-items: center;
  opacity: 0.75;
}

/* 作者 `:202-209` */
:global(.dark) .user-card {
  background-color: rgba(16, 16, 20, 0.75);

  &:hover {
    background: #18181c;
  }
}
</style>
