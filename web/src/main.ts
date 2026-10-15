import { createApp } from 'vue'
import App from './App.vue'
import { createPinia } from 'pinia'
import router from './router/index.ts'

import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
// Element Plus 的样式：它自己也定义了一套 :root 变量，必须排在前面
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'   // ← 新增：暗色变量套
import './assets/reset.css'
// 我们的主题：一定要放在 element-plus/dist/index.css **后面**
// 同权重（都是 :root）时后出现的赢，顺序反了就还是 EP 的蓝
import './assets/css/main.scss'

const app = createApp(App)

// 顺序有讲究：pinia 必须在 router 之前装上，
// 否则路由守卫里 useXxxStore() 会因为"还没有 active pinia"而报错
app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })

app.mount('#app')