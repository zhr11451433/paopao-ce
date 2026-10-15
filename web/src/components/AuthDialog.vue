<script setup lang="ts">
/**
 * 登录 / 注册弹窗
 *
 * 对应原版：F:\github_paopao\paopao-ce\web\src\components\auth.vue
 * 原版它是 App.vue 的常驻子组件（不是路由页面），我们同样挂在 App.vue 里。
 *
 * 第 1 轮做完结构与样式；第 2 轮接上接口；
 * 第 2 轮补丁：修掉"一发请求还没回来就又被触发一次"的两个口子。
 */
import { reactive, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useMainStore } from '@/store/main'
import { useUserStore } from '@/store/user'
import { login, register } from '@/api/auth'
import type { LoginPayload, RegisterPayload } from '@/types/api'

const mainStore = useMainStore()
const userStore = useUserStore()

/**
 * authVisible：弹窗开关。storeToRefs 出来的 ref 是可写的，
 * 所以 AppSidebar.vue:132 和 HomeCompose.vue:110 里的 mainStore.openAuth('signup')
 * 能把它打开，弹窗自己点关闭时也写回同一个 ref——全站只有一个真相来源。
 *
 * authTab：受控的当前 Tab。
 * ⚠️ 原版 auth.vue:61 用的是 :default-value="authModelTab"，那是非受控的，
 *    只有首次挂载时读一次；弹窗第二次打开时 Tab 会停在上次的位置。改用 v-model 修掉。
 */
const { authVisible, authTab } = storeToRefs(mainStore)

/**
 * 提交中标记，两个表单共用（原版 auth.vue:311 也是一个 loading）。
 *
 * 它有两个身份：
 *   1. 传给 el-button 的 :loading —— 转圈 + 自动 disabled，挡住"用眼睛看得见的那段时间里的点击"
 *   2. 下面 submit() 里的重入锁 —— 同步判断，挡住需要更精细处理的两种情况：
 *        ① await validate() 那个缝（见 submit 的注释）
 *        ② 长按回车的连发（见 onEnter 的注释）
 */
const loading = ref(false)

/* ========== 登录表单 ========== */

const loginFormRef = ref<FormInstance>()

/**
 * 字段名照【本项目后端】，不是照原版：
 * F:\paopao\internal\service\service.go:15-18 是 {email, password}，
 * 原版用的是 {username, password}。
 */
const loginForm = reactive<LoginPayload>({
  email: '',
  password: '',
})

const loginRules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    // 后端的 email 字段带 binding:"required,email"，前端先拦一道，
    // 否则用户不填邮箱直接点登录，只会收到后端那句看不懂的「解析失败」
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

/* ========== 注册表单 ========== */

const registerFormRef = ref<FormInstance>()

/**
 * 比 RegisterPayload 多一个 repassword：
 * 后端 UserRegisterRequest 只有 {name, email, password}，没有重复密码
 * （F:\paopao\internal\service\service.go:20-24），
 * 所以 repassword 只在本地比对，提交时不会发出去。
 */
const registerForm = reactive<RegisterPayload & { repassword: string }>({
  name: '',
  email: '',
  password: '',
  repassword: '',
})

const registerRules: FormRules = {
  name: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  repassword: [
    { required: true, message: '请再次输入密码', trigger: 'blur' },
    {
      /**
       * 原版 auth.vue:199-205 用的是 password.startsWith(value)（前缀匹配），
       * 所以「123456 / 123」这种明显不一致的组合也能通过，是原版的 bug。
       * 按第 1 轮的约定改成完全相等。
       * 密码本身为空时不在这里报错，交给 password 的 required 规则去说。
       */
      validator: (_rule, value: string, callback) => {
        if (!registerForm.password) {
          callback()
          return
        }
        if (value !== registerForm.password) {
          callback(new Error('两次密码输入不一致'))
          return
        }
        callback()
      },
      trigger: 'input',
    },
  ],
}

/* ========== 提交 ========== */

/**
 * "这个表单校验通过了吗"。
 *
 * ⚠️ 这是本项目相对原版最容易写错的一行：
 *    原版（naive-ui）是 `const { errors } = await formRef.validate()`，
 *    Element Plus 的 validate() 则是 **校验失败时 reject**、成功时 resolve(true)。
 *    如果照抄成 `if (errors) return` 或者忘了 .catch，
 *    校验失败会变成一个没人接的 Promise 异常 —— 控制台报红、按钮卡住、什么都不发生。
 */
async function validateForm(formRef: FormInstance | undefined): Promise<boolean> {
  if (!formRef) return false
  const valid = await formRef.validate().catch(() => false)
  return valid === true
}

/** 成功后清空两个表单（原版 auth.vue:238 / 280 也清），免得下次打开还留着上次的密码 */
function resetForms(): void {
  loginFormRef.value?.resetFields()
  registerFormRef.value?.resetFields()
}

/**
 * 一次提交的完整流程：上锁 → 校验 → 干活 → 解锁。
 *
 * 【为什么校验必须放在锁里面】
 *   如果写成  校验 → 上锁 → 干活  ，
 *   那么从"触发"到"上锁"之间就横着一次 `await validate()`，这段时间 loading 还是 false，
 *   第二次触发能直接穿过去，同一个表单就被提交两遍。
 *   原版 auth.vue:211-215 / 247-251 正是这个顺序，所以它也带着这个缝。
 *
 * 【为什么不能只靠按钮的 :loading】
 *   按钮变灰只能挡住"鼠标点击"，挡不住长按回车造成的连发 —— 那种触发是键盘事件，
 *   按钮是不是 disabled 根本不参与。而 loading 是同步读写的：第一次触发进来就置 true，
 *   同一个事件循环里随后到达的触发会立刻被这行挡回去。
 *
 * 【这个锁能保证什么、不能保证什么】
 *   保证：任何时刻最多只有一发请求在飞。
 *   不保证：后端几毫秒就返回 400 时，你一秒内点十下仍会产生十条请求 ——
 *          那不是"重复提交"，是十次货真价实的独立点击，各自都该发。
 *          要连这个也挡，得加"冷却时间"，但那会变成原版没有的行为，需要你先确认。
 */
async function submit(
  formRef: FormInstance | undefined,
  task: () => Promise<void>,
): Promise<void> {
  if (loading.value) return
  loading.value = true
  try {
    if (!(await validateForm(formRef))) return
    await task()
  } catch {
    /**
     * 这里【故意什么都不做】。
     *
     * 错误提示由 utils/request.ts 的响应拦截器统一弹出（"邮箱或密码错误"之类），
     * 如果这里再 ElMessage.error(...) 一次，用户会看到两条一模一样的红条。
     * 原版 auth.vue:242-244 / 284-286 也是空 catch，这不是漏写。
     */
  } finally {
    // 成功、校验不过、请求失败三条路都保证解锁
    loading.value = false
  }
}

/**
 * 回车键的统一入口：**忽略"长按造成的自动重复"**。
 *
 * 按住回车不放时，浏览器会以每秒几十次的频率持续派发 keydown / keyup。
 * 如果不加判断，一次长按就能打出一串提交请求 —— 而用户的主观感受是"我才按了一下"。
 * KeyboardEvent.repeat 为 true 就表示这次事件是自动重复，不是用户真的又按了一次。
 *
 * 参数类型写 Event 而不是 KeyboardEvent：模板里 $event 的推断类型在不同版本间
 * 不完全一致，收窄成 Event 再在里面断言，两边都不会报类型错。
 */
function onEnter(event: Event, action: () => void): void {
  if ((event as KeyboardEvent).repeat) return
  void action()
}

async function handleLogin(): Promise<void> {
  await submit(loginFormRef.value, async () => {
    /**
     * 原版这里要发两次请求：
     *   login({username, password})            → 只拿到 token
     *   get.userInfo(token)                    → 再查一次用户资料
     * 因为作者后端的登录响应里没有用户信息。
     *
     * 我们的后端一次就给全了（F:\paopao\internal\service\service.go:100-107）：
     *   { "token": "eyJ...", "user": { "id", "name", "email" } }
     * 所以这里只发一次 —— 硬抄原版会白发一个 GET /me。
     */
    const res = await login(loginForm)
    userStore.setLogin(res.token, res.user)

    /**
     * 这里【故意不调】userStore.fetchMe()。
     * /login 已经返回了 {id, name, email}，再查一次是多余请求。
     * 代价只有一个：userStore.role 在本次登录后仍是空字符串（role 只有 GET /me 才给）。
     * 而只要用户刷新一次页面，App.vue 里的 fetchMe() 就会把它补上 —— 实际影响很小。
     */
    ElMessage.success('登录成功')

    // 关弹窗：改的是 store 里的 authVisible，不是路由跳转（原版 auth.vue:237 同理）
    mainStore.closeAuth()
    resetForms()
  })
}

async function handleRegister(): Promise<void> {
  await submit(registerFormRef.value, async () => {
    /**
     * 第一步：注册。
     *
     * 注意传的是一个新对象，不是 registerForm 本身 ——
     * registerForm 里带着 repassword 这个纯前端字段，不该发给后端。
     */
    await register({
      name: registerForm.name,
      email: registerForm.email,
      password: registerForm.password,
    })

    /**
     * 第二步：立刻用同一套凭证登录，把 token 换回来。
     *
     * 这一步不是我们多出来的，原版 auth.vue:265-267 也有，而且原因一样：
     * 后端的注册接口【不发 token】（F:\paopao\internal\service\service.go:65-70 只返回
     * {status, id, name, email}），不补一次登录就进不了登录态。
     */
    const res = await login({
      email: registerForm.email,
      password: registerForm.password,
    })
    userStore.setLogin(res.token, res.user)

    ElMessage.success('注册成功')
    mainStore.closeAuth()
    resetForms()
  })
}
</script>

<template>
  <!--
    原版 auth.vue:2-12：
      <n-modal v-model:show preset="card" size="small" :mask-closable="false" :style="{width:'360px'}">
    Element Plus 的对应写法：
      v-model:show      → v-model
      :mask-closable    → :close-on-click-modal（★属性名完全不同，不是改名）
      固定 360px        → width="360px"
    另外原版没有 title，弹窗顶部只有一个关闭按钮。
  -->
  <el-dialog
    v-model="authVisible"
    class="auth-dialog"
    width="360px"
    :close-on-click-modal="false"
  >
    <div class="auth-wrap">
      <div class="auth-inner">
        <el-tabs v-model="authTab">
          <el-tab-pane name="signin" label="登录">
            <!--
              n-form-item-row → el-form-item，两个关键差异：
                1. 原版的 path 在这里叫 prop（不写 prop 就完全不校验）
                2. 原版 n-form-item-row 是"标签在上方"的布局，
                   而 el-form 默认 label-position 是 right，必须显式改
            -->
            <el-form
              ref="loginFormRef"
              :model="loginForm"
              :rules="loginRules"
              label-position="top"
              @submit.prevent
            >
              <el-form-item label="邮箱" prop="email">
                <!--
                  @keyup.enter 走 onEnter 而不是直接调 handleLogin：
                  为的是把"长按回车的自动重复"滤掉，详见 script 里 onEnter 的注释。
                  改之前这里和密码框都写的是 @keyup.enter.prevent="handleLogin"。
                -->
                <el-input
                  v-model="loginForm.email"
                  placeholder="请输入邮箱"
                  @keyup.enter.prevent="onEnter($event, handleLogin)"
                />
              </el-form-item>
              <el-form-item label="密码" prop="password">
                <!--
                  ⚠️ 有意的差异：原版是 show-password-on="mousedown"（按住鼠标显示明文），
                     Element Plus 只有 show-password（点眼睛图标切换），没有等价物。
                -->
                <el-input
                  v-model="loginForm.password"
                  type="password"
                  show-password
                  placeholder="请输入密码"
                  @keyup.enter.prevent="onEnter($event, handleLogin)"
                />
              </el-form-item>
            </el-form>

            <el-button
              type="primary"
              plain
              class="submit-btn"
              :loading="loading"
              @click="handleLogin"
            >
              登录
            </el-button>
          </el-tab-pane>

          <el-tab-pane name="signup" label="注册">
            <el-form
              ref="registerFormRef"
              :model="registerForm"
              :rules="registerRules"
              label-position="top"
              @submit.prevent
            >
              <el-form-item label="用户名" prop="name">
                <el-input
                  v-model="registerForm.name"
                  placeholder="用户名注册后无法修改"
                />
              </el-form-item>

              <!-- 这一项原版没有：后端注册必须给 email -->
              <el-form-item label="邮箱" prop="email">
                <el-input v-model="registerForm.email" placeholder="请输入邮箱" />
              </el-form-item>

              <el-form-item label="密码" prop="password">
                <el-input
                  v-model="registerForm.password"
                  type="password"
                  show-password
                  placeholder="请输入密码"
                  @keyup.enter.prevent="onEnter($event, handleRegister)"
                />
              </el-form-item>
              <el-form-item label="重复密码" prop="repassword">
                <el-input
                  v-model="registerForm.repassword"
                  type="password"
                  show-password
                  placeholder="请再次输入密码"
                  @keyup.enter.prevent="onEnter($event, handleRegister)"
                />
              </el-form-item>
            </el-form>

            <el-button
              type="primary"
              plain
              class="submit-btn"
              :loading="loading"
              @click="handleRegister"
            >
              注册
            </el-button>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>
  </el-dialog>
</template>

<style scoped lang="scss">
/**
 * 原版 auth.vue:307-309 是 .auth-wrap { margin-top: -30px }。
 * n-modal preset="card" 即使没有标题也会渲染一个只有关闭按钮的空 header，
 * 作者用负 margin 把这 30px 吃掉，让内容顶到卡片顶部。el-dialog 结构同理，所以抄过来。
 *
 * ⚠️ 两个框架的 header 高度不一定相同，这是本项目唯一需要靠肉眼校准的地方。
 *    如果弹窗内容整体偏下，把 -30px 调大（-36px / -40px）；偏上就调小。
 */
.auth-wrap {
  margin-top: -30px;

  /*
   * 原版内层还有一个 <n-card>，它的默认 padding 撑出了四周留白。
   * 按第 1 轮选定的方案 A，这里用一个普通 div 自己给留白。
   * el-dialog 的 body 自带 16px 左右内边距，所以这里只补一点点。
   */
  .auth-inner {
    padding: 0 6px 6px;
  }

  /*
   * 原版 auth.vue:63 用了 n-tabs 的 justify-content="space-evenly"（两个标题各占一半），
   * el-tabs 没有这个参数，只能用 CSS 实现。
   * .el-tabs__nav / .el-tabs__item 是 Element Plus 的内部结构，必须 :deep()；
   * 锚点 .auth-wrap 是本组件自己的元素、带 scoped 属性，选择器才成立。
   */
  :deep(.el-tabs__nav) {
    width: 100%;
    display: flex;
  }

  :deep(.el-tabs__item) {
    flex: 1;
    padding: 0;
    text-align: center;
    font-size: 16px; /* 对应原版的 size="large" */
  }

  /* 原版 n-tab-pane 自带一点上间距，el-tab-pane 没有，补一点；嫌大改这个值 */
  :deep(.el-tab-pane) {
    padding-top: 16px;
  }
}

/*
 * 原版三个提交按钮都是 <n-button block secondary strong>。
 * Element Plus 的 el-button 既没有 block 也没有 strong，用 CSS 补：
 *   block   → width: 100%
 *   strong  → font-weight: 600
 *   secondary → 已经在模板上用 plain 对应
 */
.submit-btn {
  width: 100%;
  font-weight: 600;
  margin-top: 4px;
}
</style>