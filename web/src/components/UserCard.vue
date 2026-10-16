<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MoreFilled } from '@element-plus/icons-vue'
import { BodyOutline, WalkOutline } from '@vicons/ionicons5'
import { followUser, getUser, unfollowUser } from '@/api/user'
import type { UserProfile } from '@/types/api'
import { prettyQuoteNum } from '@/utils/count'

/**
 * 作者 `components\user-card.vue` 不自己请求。
 * 他的 `contact`（`:69-77`，`Item.ContactItemProps`）里已经有
 * user_id / username / nickname / avatar / is_following / created_on，
 * 模板直接渲染。
 *
 * 我们的关注 / 粉丝列表只有 `{id, name}`（`follow.go` 的 FollowItem / FanItem）。
 * C1 的 `GET /users/:id` 才有真名字、role、关注数、粉丝数、帖子数。
 * 这一轮卡片自己去拉这份资料，不再把列表里的 name 当成资料。
 * 头像、加入时间、是否已关注，C1 仍然没有：头像留空，加入时间不画，
 * 「已关注」继续用外面传进来的 `isFollowing`。
 */
const props = withDefaults(
  defineProps<{
    contact: {
      id: number
      name: string
      /** 作者是 `contact.is_following`。C1 不返回这个字段，仍由列表页按 Tab 推断 */
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
 * 我们这里两者都还要**再且上一个「已关注」**才显示标签 ——
 * `type="follow"` 但 `isFollowing: false`（粉丝页）不该被打上「已关注」。
 */
const following = ref(props.contact.isFollowing === true)

const showFollowingTag = computed(() => props.type === 'follow' && following.value)
const enableFollowAction = computed(() => props.type === 'follow')

/** C1 的资料。请求没回来之前是 null，名字先用列表里那份，避免空白一闪 */
const profile = ref<UserProfile | null>(null)

const displayName = computed(() => profile.value?.name || props.contact.name)
/** 作者用户主页用 is_admin 打「管理员」。C1 只有 role，和 UserView 一样认 'admin' */
const isAdmin = computed(() => profile.value?.role === 'admin')

/** 作者 `:82-85`，只留了我们在做的那一个。私信是 C41，这一轮不发 */
const emit = defineEmits<{
  (e: 'unfollow-success'): void
}>()

/**
 * 按当前这个人拉资料。
 * 列表用 `:key="contact.id"`，同一个人滚分页时组件会被复用，所以盯着 id。
 * 响应回来时如果 id 已经换成下一个人，丢掉这份结果。
 */
async function loadProfile(): Promise<void> {
  const id = props.contact.id
  if (id === 0) {
    profile.value = null
    return
  }
  try {
    const data = await getUser(id)
    if (props.contact.id !== id) {
      return
    }
    profile.value = data
  } catch {
    if (props.contact.id === id) {
      profile.value = null
    }
  }
}

watch(
  () => props.contact.id,
  () => {
    profile.value = null
    following.value = props.contact.isFollowing === true
    void loadProfile()
  },
  { immediate: true },
)

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
 * 关注成功后重拉一次资料，粉丝数跟着库里走，不在本地加减。
 */
async function handleFollowUser(): Promise<void> {
  try {
    await ElMessageBox.confirm(
      `确定${following.value ? '取消关注 @' : '关注 @'}${displayName.value} 吗？`,
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
    void loadProfile()
  } catch {
    // 失败提示由 `utils\request.ts` 的响应拦截器统一弹，这里不再重复弹一次
  }
}

/**
 * 作者的 `handleAction`（`:137-162`）有三个分支：follow / unfollow / whisper。
 * 我们只做前两个 —— **「私信」整项砍掉了**：本项目后端没有私信接口（C40 / C41）。
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
      C1 没有 avatar，留空占位。头像组件是 C47，这一轮不换。
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
            显示的名字用资料接口的 name；接口还没回来时先用列表里的 name。
            `@click.stop` 照抄 `:10`。
          -->
          <span class="nickname-wrap">
            <router-link
              class="username-link"
              :to="{ name: 'user', params: { id: contact.id } }"
              @click.stop
            >
              {{ displayName }}
            </router-link>
          </span>

          <el-tag v-if="isAdmin" class="top-tag" type="danger" size="small" round>
            管理员
          </el-tag>

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
                  {{ (following ? '取消关注 @' : '关注 @') + displayName }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <!--
        作者 `:26-33`：`UID. xxx` + `xxx 加入`。
        加入时间 C1 没有。关注 / 粉丝 / 泡泡是资料接口给的，数字格式和用户主页同一套。
        这三样在作者的卡片上没有（他的卡片不拉资料接口），放在 UID 下面是因为
        本项目这张卡以前只有一个名字，C3 就是要把这几项补上。
        不做链接：作者的卡片上这几项也不是链接，链接在用户主页。
      -->
      <div class="user-info">
        <span class="info-item">UID. {{ contact.id }}</span>
      </div>
      <div v-if="profile" class="user-info user-counts">
        <span class="info-item">关注&nbsp;&nbsp;{{ prettyQuoteNum(profile.following_count) }}</span>
        <span class="info-item">粉丝&nbsp;&nbsp;{{ prettyQuoteNum(profile.follower_count) }}</span>
        <span class="info-item">泡泡&nbsp;&nbsp;{{ prettyQuoteNum(profile.post_count) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
/* 作者 `:166-200` 的尺寸。两列布局是手写的，因为没有 n-thing */
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

.user-counts {
  margin-top: 8px;
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
