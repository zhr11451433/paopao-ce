<script setup lang="ts">
import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'
import {
  ImageOutline,
  VideocamOutline,
  AttachOutline,
  CompassOutline,
  EyeOutline,
  CloseOutline,
} from '@vicons/ionicons5'
import { ElMessage } from 'element-plus'
import type { UploadInstance, UploadRawFile, UploadRequestOptions } from 'element-plus'
import { useMainStore } from '@/store/main'
import { useUserStore } from '@/store/user'
import { createPost } from '@/api/post'
import { uploadImages } from '@/api/upload'
import type { ImageItem, UploadResult } from '@/types/api'

const mainStore = useMainStore()
const userStore = useUserStore()
const { logined } = storeToRefs(userStore)

/** 作者从站点配置 profile.defaultTweetMaxLength 取上限，我们暂时写死。只用来画字数环，不截断 */
const MAX_LENGTH = 1400
/**
 * 后端 ReqPost.Title 是 binding:"required,max=100"（post_image.go:363），
 * 但原版发布框根本没有标题框。所以把正文前 100 字当标题发出去。
 * 广场页帖子卡片显示的就是 title，截断后的观感和"原版显示正文开头"是一致的。
 */
const TITLE_MAX_LENGTH = 100
/** 后端 post_image.go:199 的单张上限 */
const MAX_IMAGE_SIZE = 5 * 1024 * 1024
/** 后端 post_image.go:167 的张数上限 */
const MAX_IMAGE_COUNT = 9
/** 白名单照抄作者 compose.vue:446-452；后端还有一道 magic bytes 校验（post_image.go:226） */
const ALLOWED_IMAGE_TYPES = ['image/webp', 'image/png', 'image/jpg', 'image/jpeg', 'image/gif']

/** 原版 compose.vue:299-301 对外抛 post-success，广场页将来靠它刷新列表 */
const emit = defineEmits<{
  (e: 'post-success', postId: number): void
}>()

const content = ref('')
const submitting = ref(false)

const uploadRef = ref<UploadInstance>()
/** 已经上传成功的图片：id 发帖时要带给后端，url 拿来做预览 */
const images = ref<ImageItem[]>([])
/** 正在飞的张数：>0 时「图片」按钮转圈，顺便挡住重复点 */
const uploadingCount = ref(0)

/** 就是截图里那个小圆圈：字数进度环 */
const percent = computed(() =>
  Math.min((content.value.length / MAX_LENGTH) * 100, 100),
)
const isUploading = computed(() => uploadingCount.value > 0)
const imageButtonDisabled = computed(
  () => images.value.length + uploadingCount.value >= MAX_IMAGE_COUNT,
)

/** 原版五个工具按钮的顺序去掉「图片」—— 图片那个单独摘出来，它要挂真正的上传 */
const TOOLS = [
  { key: 'video', label: '视频', icon: VideocamOutline },
  { key: 'attach', label: '附件', icon: AttachOutline },
  { key: 'link', label: '链接', icon: CompassOutline },
  { key: 'eye', label: '可见范围', icon: EyeOutline },
]

function onToolClick(key: string): void {
  // 这四个后端都没有对应字段（见 PROJECT.md §9：没后端的业务做完阶段 A 就停）
  ElMessage.info(`「${key}」还没接`)
}

/** el-upload 的 before-upload 钩子，每张图都会走一次；返回 false 就跳过这张 */
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
 * 真正的上传。EP 的 http-request 会拿到一个 options，我们只关心 options.file。
 *
 * 【为什么不照抄作者的 :action】
 * 作者用的是 naive n-upload 内建 XHR，token 从 localStorage 手拼（compose.vue:332-334），
 * 等于绕过了 utils/request.ts —— 401 自动清 token、超时、统一错误提示全都没了。
 * 这里改成调已有的 uploadImages()，让它走同一条鉴权 / 错误通路。
 *
 * 【EP 会自动收尾】upload-content.mjs:109 是 request.then(options.onSuccess, options.onError)，
 * 所以我们只管把 promise 返回出去，成功/失败的标记由 EP 自己打。
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

/** 缩略图右上角那个叉 */
function removeImage(index: number): void {
  images.value.splice(index, 1)
}

/** 发布成功后把发布框恢复原样（对应作者 compose.vue:616-626 那段置空） */
function resetCompose(): void {
  content.value = ''
  images.value = []
  uploadRef.value?.clearFiles()
}

async function submitPost(): Promise<void> {
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
    const res = await createPost({
      title: text.slice(0, TITLE_MAX_LENGTH),
      content: text,
      images: images.value.map((img) => img.id),
    })
    ElMessage.success('发布成功')
    resetCompose()
    emit('post-success', res.post_id)
  } catch {
    // 错误提示已经由 utils/request.ts 的响应拦截器统一弹过了，这里只把转圈收回来
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="compose-wrap">
    <!-- ========== 已登录（作者 compose.vue:3） ========== -->
    <template v-if="logined">
      <div class="compose-line">
        <div class="compose-user">
          <el-avatar :size="30" />
        </div>
        <el-input
          v-model="content"
          class="compose-input"
          type="textarea"
          resize="none"
          :autosize="{ minRows: 1, maxRows: 8 }"
          placeholder="说说您的新鲜事..."
        />
      </div>

      <div class="compose-line compose-options">
        <div class="attachment">
          <!--
            图片：真正的上传入口。
            el-upload 只包住这个小图标 —— EP 在 list-type="picture-card" 时会把它
            画成网格里的大格子（upload.vue_...mjs:52-111），那就跑出工具行了。
            所以我们用 :show-file-list="false" 关掉 EP 的列表，缩略图自己画在下面。
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

          <!-- 其余四个：后端没有对应能力，保持"还没接" -->
          <el-button
            v-for="tool in TOOLS"
            :key="tool.key"
            text
            circle
            type="primary"
            :title="tool.label"
            @click="onToolClick(tool.key)"
          >
            <el-icon :size="20"><component :is="tool.icon" /></el-icon>
          </el-button>
        </div>

        <div class="submit-wrap">
          <el-tooltip :content="`已输入 ${content.length} 字`" placement="bottom">
            <el-progress
              class="text-statistic"
              type="circle"
              :percentage="percent"
              :width="20"
              :stroke-width="10"
              :show-text="false"
              color="var(--el-color-primary)"
            />
          </el-tooltip>
          <el-button type="primary" plain round :loading="submitting" @click="submitPost">
            发布
          </el-button>
        </div>
      </div>

      <!--
        缩略图行：位置照作者 compose.vue:192-210 的 .attachment-list-wrap
        （工具行下面、左缩进 42px，跟输入框左边缘对齐）
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

    <!-- ========== 未登录（作者 compose.vue:238） ========== -->
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

  /* 把 EP textarea 自带的边框/阴影/内边距全抹掉，还原作者 :bordered="false" 的效果 */
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

    /* el-upload 会渲染一层 div，把它压成 inline-flex 才能和其余图标同一行 */
    .attachment-upload {
      display: inline-flex;
      align-items: center;
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

    .text-statistic {
      margin-right: 8px;
      width: 20px;
      height: 20px;
      transform: rotate(180deg);
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
</style>