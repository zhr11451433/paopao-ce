<script setup lang="ts">
import { computed } from 'vue'
import AppSidebar from '@/components/AppSidebar.vue'

/**
 * 移动端抽屉导航。作者 `components\main-nav.vue:2-13`：
 *
 *   <div v-if="drawerModelShow">
 *     <n-drawer v-model:show="activeDrawerRef" :width=212 :placement="placementRef" resizable>
 *       <n-drawer-content><sidebar /></n-drawer-content>
 *     </n-drawer>
 *   </div>
 *
 * 三件事照原版：
 * - **抽屉里装的还是那个左栏组件**，不是另写一套菜单 —— 作者桌面端的左栏和移动端的抽屉
 *   共用 `components\sidebar.vue`。所以我们也是 `<AppSidebar drawer />`：同一份菜单数据、
 *   同一份样式，只是排布从「fixed 钉在中间栏左边」换成「在抽屉里正常排」。
 * - 宽度 212（作者的 `:width=212`），从左边滑出来（`:placement="left"`）。
 * - 桌面端根本不渲染它（作者是 `v-if="drawerModelShow"`，在我们这里是 MainNav 上的
 *   `v-if="isMobile"`，见 MainNav.vue）。
 *
 * 一处没搬：作者的 `resizable`（可以拖着改抽屉宽度），EP 的 el-drawer 没有这个能力，
 * 登记在 PROJECT.md 11.6。
 */
const props = defineProps<{ modelValue: boolean }>()
const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void }>()

const visible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})
</script>

<template>
  <!--
    `append-to-body` 在这里是必须的，不是可选的美化项：

    MainNav 的根节点 `.nav-title-card` 上有 `backdrop-filter: blur(12px)`，而
    backdrop-filter 会让这个元素成为**它内部所有 position: fixed 元素的包含块**。
    el-drawer 的遮罩和面板恰好就是 fixed 定位的，不 append 到 body 的话，
    整个抽屉会被塞进那条 54px 高的导航条里（看起来就是「点了一下，什么都没发生」）。

    作者不用写这一句，是因为 naive-ui 的 n-drawer 默认就 teleport 到 body。
  -->
  <el-drawer
    v-model="visible"
    class="mobile-nav-drawer"
    direction="ltr"
    size="212px"
    :with-header="false"
    append-to-body
  >
    <AppSidebar drawer />
  </el-drawer>
</template>

<style lang="scss">
/**
 * 故意**不加** scoped：抽屉被 teleport 到了 body 上，scoped 的 `[data-v-xxx]`
 * 选择器够不到 `.el-drawer__body`（它已经不是这个组件的子节点了）。
 *
 * EP 的抽屉正文默认带 20px 内边距，而左栏组件自己已经有 `padding: 12px 0`
 * 和 logo 的 `margin-left`，再加 20px 全部会对不齐
 * （作者的 n-drawer-content 本身没有额外内边距）。
 */
.mobile-nav-drawer .el-drawer__body {
  padding: 0;
  overflow: hidden;
}
</style>
