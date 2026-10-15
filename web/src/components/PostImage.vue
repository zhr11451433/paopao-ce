<script setup lang="ts">
import { computed } from 'vue'
import type { ImageItem } from '@/types/api'

const props = withDefaults(
  defineProps<{
    imgs: ImageItem[]
  }>(),
  {
    imgs: () => [],
  },
)

/**
 * 作者用 `<n-image-group>` 把同一组图片串成一个灯箱（左右箭头能翻页）——
 * 他的 post-image.vue 里 8 个分支每个外面都套了一层（:3、:20、:37 …）。
 *
 * Element Plus **没有 image-group 这个组件**（对照表见 PROJECT.md §6），
 * 它把同一件事做在 `el-image` 自己的两个 prop 上：
 *   :preview-src-list = 整组 URL 列表
 *   :initial-index   = 我这张在列表里排第几
 * 所以下面每张图都收到同一个列表，只有 initial-index 不同，效果与 n-image-group 一致。
 */
const previewList = computed(() => props.imgs.map((img) => img.url))

/**
 * 作者用 n-grid 的 `:cols="N"` + `:x-gap="4" :y-gap="4"` 铺网格。
 * EP 没有 n-grid（对照表同见 §6），用 CSS grid 等价顶替：
 * 列数用 grid-template-columns 给，间距用 gap（下面 scss 里的 .images-grid）。
 */
function cols(n: number): Record<string, string> {
  return { gridTemplateColumns: `repeat(${n}, 1fr)` }
}
</script>

<template>
  <div class="images-wrap">
    <!--
      下面 8 个分支**逐个照抄**作者 post-image.vue:3-221 的结构
      （哪几张图分几列、分几个网格、每张用哪个高度档，一个都没改）。
      唯一变掉的是组件本身：
        n-image-group  →  el-image 的 :preview-src-list + :initial-index
        n-grid / n-gi  →  <div class="images-grid"> + grid-template-columns
        n-image        →  el-image（object-fit="cover" → fit="cover"）

      另外作者有两处我们不搬（后方标了原因）：
        :src="img.content + thumbnail"  →  我们没有图片缩略图服务，直接给原图 URL
        @error="() => (img.content = defaultImg)"  →  这个 env 常量我们没有；
          而且它在改 props（Vue 里不该这么写）。加载失败交给 el-image 自带的占位。
    -->

    <!-- 1 张：占 2 列中的 1 列，高度 174px -->
    <div v-if="[1].includes(props.imgs.length)" class="images-grid" :style="cols(2)">
      <el-image
        v-for="(img, idx) in props.imgs"
        :key="img.id"
        class="post-img x1"
        fit="cover"
        :src="img.url"
        :preview-src-list="previewList"
        :initial-index="idx"
        :preview-teleported="true"
        @click.stop
      />
    </div>

    <!-- 2 ~ 3 张：占 3 列，高度 112px -->
    <div v-if="[2, 3].includes(props.imgs.length)" class="images-grid" :style="cols(3)">
      <el-image
        v-for="(img, idx) in props.imgs"
        :key="img.id"
        class="post-img x2"
        fit="cover"
        :src="img.url"
        :preview-src-list="previewList"
        :initial-index="idx"
        :preview-teleported="true"
        @click.stop
      />
    </div>

    <!-- 4 张：占 4 列，高度 100px -->
    <div v-if="[4].includes(props.imgs.length)" class="images-grid" :style="cols(4)">
      <el-image
        v-for="(img, idx) in props.imgs"
        :key="img.id"
        class="post-img x3"
        fit="cover"
        :src="img.url"
        :preview-src-list="previewList"
        :initial-index="idx"
        :preview-teleported="true"
        @click.stop
      />
    </div>

    <!-- 5 张：上 3 张（3 列，112px）+ 下 2 张（2 列，174px） -->
    <template v-if="[5].includes(props.imgs.length)">
      <div class="images-grid" :style="cols(3)">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx < 3"
            class="post-img x2"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
      <div class="images-grid" :style="cols(2)" style="margin-top: 4px">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx >= 3"
            class="post-img x1"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
    </template>

    <!-- 6 张：上下各 3 张（都是 3 列，112px） -->
    <template v-if="[6].includes(props.imgs.length)">
      <div class="images-grid" :style="cols(3)">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx < 3"
            class="post-img x2"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
      <div class="images-grid" :style="cols(3)" style="margin-top: 4px">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx >= 3"
            class="post-img x2"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
    </template>

    <!-- 7 张：上 4 张（4 列，100px）+ 下 3 张（3 列，112px） -->
    <template v-if="props.imgs.length === 7">
      <div class="images-grid" :style="cols(4)">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx < 4"
            class="post-img x3"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
      <div class="images-grid" :style="cols(3)" style="margin-top: 4px">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx >= 4"
            class="post-img x2"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
    </template>

    <!-- 8 张：上下各 4 张（都是 4 列，100px） -->
    <template v-if="props.imgs.length === 8">
      <div class="images-grid" :style="cols(4)">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx < 4"
            class="post-img x3"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
      <div class="images-grid" :style="cols(4)" style="margin-top: 4px">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx >= 4"
            class="post-img x3"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
    </template>

    <!-- 9 张：三行各 3 张（都是 3 列，112px）—— 九宫格 -->
    <template v-if="props.imgs.length === 9">
      <div class="images-grid" :style="cols(3)">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx < 3"
            class="post-img x2"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
      <div class="images-grid" :style="cols(3)" style="margin-top: 4px">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx >= 3 && idx < 6"
            class="post-img x2"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
      <div class="images-grid" :style="cols(3)" style="margin-top: 4px">
        <template v-for="(img, idx) in props.imgs" :key="img.id">
          <el-image
            v-if="idx >= 6"
            class="post-img x2"
            fit="cover"
            :src="img.url"
            :preview-src-list="previewList"
            :initial-index="idx"
            :preview-teleported="true"
            @click.stop
          />
        </template>
      </div>
    </template>
  </div>
</template>

<style scoped lang="scss">
/**
 * 作者 post-image.vue:240-280 的样式，逐条搬过来。
 * 作者写的是**全局样式**（`<style lang="less">` 没有 scoped），因为要能改到
 * n-image 内部的 `<img>`。本项目所有组件的样式都是 scoped（PROJECT.md §8），
 * 所以那条 `img { ... }` 改写成 `:deep(img)`，效果一样但不会漏到别的组件去。
 */
.images-wrap {
  margin-top: 10px;
}

/* n-grid 的替代品：列数由上面 :style 的 grid-template-columns 给，间距固定 4px */
.images-grid {
  display: grid;
  gap: 4px;
}

.post-img {
  display: flex;
  margin: 0;
  border-radius: 3px;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.1);
  border: 1px solid #eee;

  :deep(img) {
    width: 100%;
    height: 100%;
  }
}

/* 三档高度，对应作者 .x1 / .x2 / .x3 */
.x1 {
  height: 174px;
}

.x2 {
  height: 112px;
}

.x3 {
  height: 100px;
}

/* 作者 post-image.vue:265-269 的暗色写法 */
.dark .post-img {
  border: 1px solid #333;
}

/* 作者 post-image.vue:270-280：窄屏三档高度全部缩小 */
@media screen and (max-width: 821px) {
  .x1 {
    height: 100px;
  }

  .x2 {
    height: 70px;
  }

  .x3 {
    height: 50px;
  }
}
</style>
