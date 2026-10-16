<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from '@vicons/ionicons5'

/**
 * 右栏。
 *
 * 作者 `rightbar.vue` 有三块：搜索框、热门话题、版权。
 * 热门话题 / 关注话题要 GET /tags，那是 C31，这里先不画。
 * 宽度 240px、定位方式和作者一样。
 */

const router = useRouter()
const year = new Date().getFullYear()

/**
 * 作者 `rightbar.vue:102` 的 keyword，回车走 `handleSearch`（`:160-167`）：
 * `router.push({ name: 'home', query: { q: keyword } })`。
 * 他不带 `t=tag`，那是话题链接才加的。空词就打开广场，不带 q。
 */
const keyword = ref('')

function handleSearch(): void {
  const q = keyword.value.trim()
  void router.push({
    name: 'home',
    query: q ? { q } : {},
  })
}
</script>

<template>
  <aside class="rightbar-wrap">
    <!--
      作者 rightbar.vue:3-15：圆角、可清空、前缀放大镜，回车才搜。
      没有单独的搜索按钮。n-input 换成 el-input。
    -->
    <div class="search-wrap">
      <el-input
        v-model="keyword"
        clearable
        placeholder="搜一搜..."
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
    </div>

    <!--
      将来补「热门话题」的位置（C31）：
      <el-card shadow="never" class="panel"> ... </el-card>
    -->

    <el-card shadow="never" class="panel about-card">
      <div class="about-title">关于本站</div>
      <p class="about-text">
        这是一个模仿 paopao-ce 的练手项目：Go + Gin 写后端，Vue 3 + Element Plus 写前端。
      </p>
    </el-card>

    <el-card shadow="never" class="panel copyright-card">
      <div class="copyright">&copy; {{ year }} 泡泡</div>
      <a class="hash-link" href="https://github.com/rocboss/paopao-ce" target="_blank" rel="noopener">
        paopao-ce
      </a>
    </el-card>
  </aside>
</template>

<style scoped lang="scss">
.rightbar-wrap {
  width: 240px;
  position: fixed;
  /* 和左栏镜像：贴着中间栏右边 */
  left: calc(50% + var(--content-main) / 2 + 10px);
  max-height: 100vh;
  overflow: auto;

  &::-webkit-scrollbar {
    width: 0;
    height: 0;
  }

  /* 作者 rightbar.vue:224-226 */
  .search-wrap {
    margin: 12px 0;
  }

  .panel {
    margin-top: 12px;
    border-radius: 12px;
  }

  .about-title {
    font-weight: bold;
    margin-bottom: 8px;
  }

  .about-text {
    font-size: 13px;
    line-height: 1.6;
    opacity: 0.75;
  }

  .copyright-card {
    font-size: 12px;
    opacity: 0.75;
  }
}
</style>