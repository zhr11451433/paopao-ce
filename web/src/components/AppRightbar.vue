<script setup lang="ts">
/**
 * 右栏。
 *
 * 作者原版有三块：搜索框、热门话题、版权。
 * 其中「热门话题 / 关注话题」需要 GET /v1/tags 一类接口，你的后端没有，
 * 所以这里只保留版权，并在原位置留一个注释说明将来补什么。
 * 宽度 240px、定位方式都照搬作者的 rightbar.vue。
 */
const year = new Date().getFullYear()
</script>

<template>
  <aside class="rightbar-wrap">
    <!--
      将来补「热门话题」的位置：
      <el-card shadow="never" class="panel"> ... </el-card>
      数据源需要一个"取热门标签"的接口，后端目前没有。
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