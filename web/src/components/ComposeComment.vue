<script setup lang="ts">
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { CloseOutline, ImageOutline } from '@vicons/ionicons5'
import { ElMessage } from 'element-plus'
import type { UploadInstance, UploadRawFile, UploadRequestOptions } from 'element-plus'
import { useMainStore } from '@/store/main'
import { useUserStore } from '@/store/user'
import { createComment } from '@/api/post'
import { uploadImages } from '@/api/upload'
import type { ImageItem, UploadResult } from '@/types/api'

/**
 * 对应作者 `components\compose-comment.vue`（479 行）—— 帖子详情页里的发评论框。
 *
 * 它和广场页那个发帖框（`compose.vue` / 我们的 `HomeCompose.vue`）是**两个不同的组件**，
 * 长得像但不是一回事，差别有三处：
 *   ① 这里的工具行只有「图片」一个按钮，没有 视频/附件/链接/可见范围；
 *   ② **工具行是聚焦输入框之后才出来的**（作者 `:34` 的 `v-if="showBtn"`，
 *      `showBtn` 在 `@focus` 里才置 true）—— 广场那个是一直显示；
 *   ③ 多一个「取消」按钮。
 *
 * 本版砍掉的东西（本项目后端没有）：
 *   · `n-mention` 的 @提及功能 —— 作者 `:12-30` 用的是 naive 的 `n-mention`，
 *     输入 @ 会去 `Api.v1.suggest.get.users` 拉用户列表。本项目**没有「搜索用户」接口**，
 *     Element Plus 也没有 mention 组件，所以退化成普通的 `el-input type="textarea"`，
 *     **@ 只是个普通字符**（和 `HomeCompose.vue` 的做法一致）。
 *   · `props.lock`（`:21` 的 `:disabled="props.lock === 1"` 和锁定时的 placeholder）——
 *     本项目后端没有「锁定帖子」这个字段。
 *   · `import.meta.env.VITE_DEFAULT_COMMENT_MAX_LENGTH`（`:211-213`）——
 *     作者从环境变量取上限，我们直接用**后端的真上限**：
 *     `CommentReq.Content` 是 `binding:"required,max=512"`（`internal\service\comment.go:28`）。
 *   · `allowUserRegister` 那个开关（`:208-210`）—— 我们两个按钮都显示（同 HomeCompose.vue）。
 */

const props = defineProps<{
  /** 作者 `:184-193` 的 props 是两个（`lock` / `postId`），我们只留 `postId`（见文件头） */
  postId: number
}>()

/** 作者 `:181-183`：只抛 `post-success`，详情页收到后重拉评论第一页 */
const emit = defineEmits<{
  (e: 'post-success'): void
}>()

const mainStore = useMainStore()
const userStore = useUserStore()
const { logined } = storeToRefs(userStore)

/** 后端硬上限，见文件头 */
const MAX_LENGTH = 512
/** 单张图上限，和 `post_image.go:199` 一致（评论和帖子走的是同一个 `/upload`） */
const MAX_IMAGE_SIZE = 5 * 1024 * 1024
/** 张数上限，和 `post_image.go:167` 一致 */
const MAX_IMAGE_COUNT = 9
/** 白名单照抄作者 `compose-comment.vue:276-284`（比发帖框少了 webp，我们统一成一套） */
const ALLOWED_IMAGE_TYPES = ['image/webp', 'image/png', 'image/jpg', 'image/jpeg', 'image/gif']

const content = ref('')
const submitting = ref(false)

/**
 * 作者 `:200` 的 `showBtn` —— 输入框聚焦后才把「图片 / 字数环 / 取消 / 发布」那一行露出来。
 * 名字我换成 `showOptions`，因为这个 ref 控制的是整行选项，不只是"按钮"。
 */
const showOptions = ref(false)

const uploadRef = ref<UploadInstance>()
/** 已经上传成功的图片：id 要带给后端，url 拿来做预览 */
const images = ref<ImageItem[]>([])
/** 正在飞的张数：>0 时「图片」按钮转圈，顺便挡住重复点 */
const uploadingCount = ref(0)

/** 作者 `:92` 那个圆圈：`(content.length / defaultCommentMaxLength) * 100` */
const percent = computed(() => Math.min((content.value.length / MAX_LENGTH) * 100, 100))
const isUploading = computed(() => uploadingCount.value > 0)
const imageButtonDisabled = computed(
  () => images.value.length + uploadingCount.value >= MAX_IMAGE_COUNT,
)

/** 作者 `:336-338` 的 `focusComment()` */
function onFocus(): void {
  showOptions.value = true
}

/** `el-upload` 的 before-upload 钩子，每张图都会走一次；返回 false 就跳过这张 */
function beforeUpload(rawFile: UploadRawFile): boolean {
  if (!ALLOWED_IMAGE_TYPES.includes(rawFile.type)) {
    ElMessage.warning('图片仅允许 webp/png/jpg/gif 格式')
    return false
  }
  if (rawFile.size > MAX_IMAGE_SIZE) {
    ElMessage.warning('图片大小不能超过 5MB')
    return false
  }
  if (images.value.length >= MAX_IMAGE_COUNT) {
    ElMessage.warning(`一次最多上传 ${MAX_IMAGE_COUNT} 张图片`)
    return false
  }
  return true
}

/**
 * 上传走 `uploadImages()`，**不用** `el-upload` 自带的 `:action`。
 *
 * 作者 `:33-53` 是让 naive 的 n-upload 自己发 XHR、token 从 localStorage 手拼（`:216-218`），
 * 等于绕过了 `utils/request.ts` —— 401 自动清 token、超时、统一错误提示全都没了。
 * 我们的 `HomeCompose.vue` 已经改成走同一条鉴权通路，这里保持一致。
 */
async function doUpload(options: UploadRequestOptions): Promise<UploadResult> {
  uploadingCount.value++
  try {
    const res = await uploadImages([options.file])
    // noUncheckedIndexedAccess 开着，res.images[0] 的类型是 ImageItem | undefined，必须判一下
    const first = res.images[0]
    if (first) {
      images.value.push(first)
    }
    return res
  } finally {
    // 成功失败都要减，否则按钮会一直转圈
    uploadingCount.value--
  }
}

function removeImage(index: number): void {
  images.value.splice(index, 1)
}

/**
 * 作者 `:339-346` 的 `cancelComment()`：收起选项行 + 三样全部置空。
 * 有一个地方不一样：他的评论**可以带图**（`contents` 里 type=3），
 * 但我们后端的 `CommentReq` **只收一个 `content` 字符串**（`comment.go:27-29`），
 * 评论图根本传不进去（见 PROJECT.md 11.6）——
 * 所以这里的图片按钮虽然做了，上传结果**发不出去**，已登记在 11.8。
 */
function cancelComment(): void {
  showOptions.value = false
  content.value = ''
  images.value = []
  uploadRef.value?.clearFiles()
}

/** 作者 `:349-392` 的 `submitPost()`，我们这版只剩下「校验 + 一个 content」 */
async function submitComment(): Promise<void> {
  const text = content.value.trim()
  if (text.length === 0) {
    ElMessage.warning('请输入内容哦')
    return
  }
  if (uploadingCount.value > 0) {
    ElMessage.warning('图片还在上传中，稍等一下再发布')
    return
  }

  submitting.value = true
  try {
    await createComment(props.postId, text)
    ElMessage.success('发布成功')
    emit('post-success')
    cancelComment()
  } catch {
    // 错误提示已经由 utils/request.ts 的响应拦截器统一弹过了，这里只把转圈收回来
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="compose-wrap">
    <!-- ========== 已登录（作者 `:3-127`） ========== -->
    <template v-if="logined">
      <div class="compose-line">
        <div class="compose-user">
          <!-- 作者 `:6-10` 是 `:src="userInfo.avatar"`；我们没有头像字段，空着 -->
          <el-avatar :size="30" />
        </div>

        <!--
          作者 `:12-30` 的 `n-mention`。换成 EP 的 `el-input type="textarea"`（见文件头）。
          两处对应：
            · `:21`  `:disabled="props.lock === 1"`        → 我们没有 lock，去掉
            · `:25-29` placeholder「快来评论两句吧...」      → 原样
            · `:248-254` changeContent 超长就截断          → 用 `:maxlength`（EP 原生，行为一样）
        -->
        <el-input
          v-model="content"
          class="compose-input"
          type="textarea"
          resize="none"
          :autosize="{ minRows: 1, maxRows: 8 }"
          :maxlength="MAX_LENGTH"
          placeholder="快来评论两句吧..."
          @focus="onFocus"
        />
      </div>

      <!-- 作者 `:34` 的 `v-if="showBtn"`：聚焦之后这一行才出现 -->
      <div v-if="showOptions" class="compose-line compose-options">
        <div class="attachment">
          <!--
            作者 `:56-82` 是一个 `n-upload-trigger` 包着图片按钮；
            `el-upload` + `:show-file-list="false"` 是等价写法（同 HomeCompose.vue）。
          -->
          <el-upload
            ref="uploadRef"
            class="attachment-upload"
            :show-file-list="false"
            :multiple="true"
            :auto-upload="true"
            accept="image/webp,image/png,image/jpg,image/jpeg,image/gif"
            :before-upload="beforeUpload"
            :http-request="doUpload"
          >
            <el-button
              text
              circle
              type="primary"
              title="图片"
              :loading="isUploading"
              :disabled="imageButtonDisabled"
            >
              <el-icon :size="20"><ImageOutline /></el-icon>
            </el-button>
          </el-upload>

          <!-- 作者 `:84-96` 的字数环：`n-tooltip` + `n-progress type="circle"` -->
          <el-tooltip :content="`${content.length} / ${MAX_LENGTH}`" placement="bottom">
            <el-progress
              class="text-statistic"
              type="circle"
              :percentage="percent"
              :width="18"
              :stroke-width="10"
              :show-text="false"
              status="success"
            />
          </el-tooltip>
        </div>

        <div class="submit-wrap">
          <!-- 作者 `:100-109`：`n-button quaternary round type="tertiary" size="small"`「取消」 -->
          <el-button class="cancel-btn" text round size="small" @click="cancelComment">
            取消
          </el-button>
          <!-- 作者 `:110-119`：`type="primary" secondary size="small" round`「发布」 -->
          <el-button type="primary" plain round size="small" :loading="submitting" @click="submitComment">
            发布
          </el-button>
        </div>
      </div>

      <!--
        作者 `:123-125` 用 `<n-upload-file-list />` 画缩略图。
        我们关掉了 EP 的列表，自己画（位置照他的 `.attachment-list-wrap`：工具行下面、左缩进 42px）。
      -->
      <div v-if="images.length > 0" class="attachment-list-wrap">
        <div v-for="(img, index) in images" :key="img.id" class="attachment-item">
          <el-image class="attachment-thumb" :src="img.url" fit="cover" />
          <button class="attachment-remove" type="button" title="移除" @click="removeImage(index)">
            <el-icon :size="12"><CloseOutline /></el-icon>
          </button>
        </div>
      </div>
    </template>

    <!-- ========== 未登录（作者 `:129-164`） ========== -->
    <template v-else>
      <div class="login-wrap">
        <span class="login-banner">登录后，精彩更多</span>
      </div>
      <div class="login-wrap">
        <el-button type="primary" plain round @click="mainStore.openAuth('signin')">登录</el-button>
        <el-button type="info" plain round @click="mainStore.openAuth('signup')">注册</el-button>
      </div>
    </template>
  </div>
</template>

<style scoped lang="scss">
/**
 * 作者 `compose-comment.vue:399-478`。和 `HomeCompose.vue` 基本一样，
 * 所以下面直接沿用那一套（两个发布框的观感必须一致），
 * 只有 `.text-statistic` 的尺寸按作者这里写的 18px（广场那个是 20px）。
 */
.compose-wrap {
  width: 100%;
  padding: 16px;
  box-sizing: border-box;

  .compose-line {
    display: flex;
    flex-direction: row;

    .compose-user {
      width: 42px;
      height: 42px;
      flex: none;
      display: flex;
      align-items: center;
    }

    &.compose-options {
      margin-top: 6px;
      padding-left: 42px;          /* ← 作者那句，图标对齐输入框 */
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }

  /* 把 EP textarea 自带的边框/阴影/内边距全抹掉，还原作者 `:bordered="false"` 的效果 */
  .compose-input {
    flex: 1;

    :deep(.el-textarea__inner) {
      box-shadow: none;
      border: none;
      background-color: transparent;
      padding: 0;
      font-size: 16px;
      line-height: 1.5;
      color: var(--el-text-color-primary);
    }
  }

  .attachment {
    display: flex;
    align-items: center;

    /* el-upload 会渲染一层 div，把它压成 inline-flex 才能和字数环同一行 */
    .attachment-upload {
      display: inline-flex;
      align-items: center;
    }

    /* 作者 `:457-462` */
    .text-statistic {
      margin-left: 8px;
      width: 18px;
      height: 18px;
      transform: rotate(180deg);
    }
  }

  .attachment-list-wrap {
    margin-top: 12px;
    margin-left: 42px;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .attachment-item {
    position: relative;
    width: 72px;
    height: 72px;
    border-radius: 6px;
    overflow: hidden;
    border: 1px solid var(--el-border-color);

    .attachment-thumb {
      width: 100%;
      height: 100%;
      display: block;
    }

    .attachment-remove {
      position: absolute;
      top: 0;
      right: 0;
      width: 18px;
      height: 18px;
      padding: 0;
      border: none;
      border-radius: 0 0 0 6px;
      background-color: rgba(0, 0, 0, 0.5);
      color: #fff;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;

      &:hover {
        background-color: var(--el-color-danger);
      }
    }
  }

  .submit-wrap {
    display: flex;
    align-items: center;

    .cancel-btn {
      margin-right: 8px;
    }

    /* 作者的按钮是 naive 的 secondary（有底色无边框），EP 的 plain 多一条边，抹掉 */
    :deep(.el-button) {
      border-color: transparent;
    }
  }

  .login-wrap {
    display: flex;
    justify-content: center;
    width: 100%;

    .login-banner {
      margin-bottom: 12px;
      opacity: 0.8;
    }
  }
}

/* 作者 `:471-478` 的暗色 */
:global(.dark) .compose-wrap {
  background-color: rgba(16, 16, 20, 0.75);
}
</style>
