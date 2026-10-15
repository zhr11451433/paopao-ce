<script setup lang="ts">
import { LinkOutline } from '@vicons/ionicons5'

/**
 * 帖子外链项。
 *
 * ⚠️ 这个类型**只写在组件内部**，没有放进 `src\types\api.ts` —— 因为那个文件是
 * 「后端接口契约的唯一出处」，只放后端真的有的东西（PROJECT.md §12）。
 * 而本项目后端**根本没有「链接」这个概念**：`PostListItem`、`PostDetail`、
 * 发帖入参 `ReqPost` 全都没有 links 字段（见 PROJECT.md 11.6 / 11.7）。
 *
 * 所以 T8 按 §10.1「缺什么就进 B 档并登记」归进了 B 档，做法是 §9 那一句
 * 「样式照原版做足，功能挂着，在进度表里标记为阶段 A 完成，等后端」。
 * 后端补上 links 之后：这个类型搬去 `types/api.ts`，本组件一行都不用改。
 */
interface PostLinkItem {
  id: number
  content: string
}

/**
 * 作者 post-link.vue:16-23 的 props，只有一个 `links`，默认空数组。
 * 他那里 `links` 是必填的 `Item.PostItemProps[]`；我们后端还没有这个字段，
 * 所以写成可选 + 默认 `[]`，这样不传也不会炸。
 */
const props = withDefaults(
  defineProps<{
    links?: PostLinkItem[]
  }>(),
  { links: () => [] },
)
</script>

<template>
  <!-- 作者 post-link.vue:2 —— 外层就是唯一的一个 .link-wrap -->
  <div class="link-wrap">
    <!--
      作者 post-link.vue:3 写的是 `v-for="link in props.links"`（他先把 defineProps 的结果
      赋给了 `props` 变量）。我们保持同样的写法，省得以后对照源码时对不上。
    -->
    <div v-for="link in props.links" :key="link.id" class="link-item">
      <!--
        作者 post-link.vue:4 是 `<n-icon class="hash-link">`。
        n-icon → el-icon（PROJECT.md §6）；`hash-link` 这个类是**全局**定义的，
        作者的 main.less 和我们的 main.scss 里都有 `.hash-link { color: #18a058 }`，
        所以链接的绿色两边一致，组件里不用再写颜色。
      -->
      <el-icon class="hash-link"><LinkOutline /></el-icon>

      <div class="link-txt-wrap">
        <!--
          作者 post-link.vue:6 的 `@click.stop` 不能省：卡片根节点上挂着
          `@click="goPostDetail"`（post-item.vue:2），少了 .stop 点链接会**同时跳进详情页**。
          `target="_blank"` 也是原样 —— 新窗口打开，不离开泡泡。
        -->
        <a :href="link.content" class="hash-link" target="_blank" @click.stop>
          <span class="link-txt">{{ link.content }}</span>
        </a>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
/**
 * 作者 post-link.vue:26-49 原样搬，一个字没改。
 *
 * 关键就是 `.link-txt-wrap` 那几条：它是**绝对定位**的，
 * `left: calc(1em + 4px)` 让开左边的图标（1em = 14px，即正文字号），
 * `width: calc(100% - 1em)` 铺满右边剩下的宽度，
 * 再加上 `overflow: hidden` + `white-space: nowrap` + `text-overflow: ellipsis`
 * —— 这样超长的 URL 会变成一行省略号，不会把卡片撑破。
 *
 * 作者这里是 `<style lang="less" scoped>`，本来就有 scoped，所以连 :deep 都不用动。
 */
.link-wrap {
  margin-bottom: 10px;
  position: relative;

  .link-item {
    height: 22px;
    display: flex;
    align-items: center;
    position: relative;

    .link-txt-wrap {
      left: calc(1em + 4px);
      width: calc(100% - 1em);
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
      position: absolute;

      .hash-link {
        .link-txt {
          word-break: break-all;
        }
      }
    }
  }
}
</style>
