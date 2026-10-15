/**
 * 日期格式化。
 *
 * 作者对应的是 web\src\utils\formatTime.ts —— 他那里是 moment 的一层包装，
 * 比如 formatPrettyDate(time)：输入是**秒级 unix 时间戳**（他后端给的是数字），
 * 输出按三种情况分支（formatTime.ts:32-41）：
 *
 *   1. 不是今年           → 'YYYY-MM-DD'
 *   2. 同年、但差超过 3 个月 → 'MM-DD'
 *   3. 其余              → moment 的 fromNow()，即「3 小时前」这种相对时间
 *
 * 本项目不能照抄，有两个硬差异：
 *
 *   - **不引入 moment**（PROJECT.md §2、§8：不新增第三方依赖）。所以下面自己实现，
 *     行为尽量对齐 moment(zh-cn) 的 fromNow() 分档。
 *   - **输入不是时间戳而是字符串**（PROJECT.md §7.5 第 2 条）：
 *     我们后端是 Go 的 time.Time，序列化出来是 RFC3339，形如
 *     "2026-02-18T10:30:00+08:00"，new Date() 能直接解析。
 */

const MINUTE = 60
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

/** 补零：8 → '08' */
function pad(n: number): string {
  return String(n).padStart(2, '0')
}

/**
 * 两个日期相差几个**整月**，对齐 moment.diff(x, 'month') 的口径：
 * 先按年月算出月数，再看「日」够不够，不够就退一个月。
 * 例：3月20日 → 6月19日 算 2 个月，3月20日 → 6月20日 才算 3 个月。
 */
function diffInMonths(later: Date, earlier: Date): number {
  let months =
    (later.getFullYear() - earlier.getFullYear()) * 12 + (later.getMonth() - earlier.getMonth())
  if (later.getDate() < earlier.getDate()) months -= 1
  return months
}

/**
 * 相对时间，对齐 moment(zh-cn) 的 fromNow()。
 * moment 默认的分档阈值是 45 秒 / 45 分 / 22 小时 / 26 天 / 11 个月，
 * 这里照搬，这样界面上出现的字和原版一模一样（「几秒前」「3 小时前」……）。
 */
function fromNow(diffSeconds: number): string {
  if (diffSeconds < 45) return '几秒前'

  const minutes = Math.round(diffSeconds / MINUTE)
  if (minutes < 45) return `${minutes} 分钟前`

  const hours = Math.round(diffSeconds / HOUR)
  if (hours < 22) return `${hours} 小时前`

  const days = Math.round(diffSeconds / DAY)
  if (days < 26) return `${days} 天前`

  const months = Math.round(days / 30)
  if (months < 11) return `${months} 个月前`

  return `${Math.round(months / 12)} 年前`
}

/**
 * 帖子上显示的时间（作者界面上就是「3 小时前」这种）。
 *
 * @param rfc3339 后端给的 created_at，形如 "2026-02-18T10:30:00+08:00"
 */
export function formatPrettyDate(rfc3339: string): string {
  const target = new Date(rfc3339)
  // 后端理论上不会给坏数据；真给了也不能让整个列表崩掉，返回空串让它自己消失
  if (Number.isNaN(target.getTime())) return ''

  const now = new Date()

  // 作者的第一个分支：不是今年，直接给完整日期
  if (target.getFullYear() !== now.getFullYear()) {
    return `${target.getFullYear()}-${pad(target.getMonth() + 1)}-${pad(target.getDate())}`
  }

  // 作者的第二个分支：同年但超过 3 个月，只给月日
  if (diffInMonths(now, target) > 3) {
    return `${pad(target.getMonth() + 1)}-${pad(target.getDate())}`
  }

  // 作者直接用 fromNow()（formatTime.ts:40）
  return fromNow((now.getTime() - target.getTime()) / 1000)
}

/**
 * 作者的 `formatPrettyTime`（formatTime.ts:21-30）。
 *
 * 它和上面的 `formatPrettyDate` **只差两个分支的格式**，第三个分支完全一样：
 *
 *   1. 不是今年            → 'YYYY-MM-DD HH:mm'（Date 版是 'YYYY-MM-DD'）
 *   2. 同年、但差超过 3 个月 → 'MM-DD HH:mm'     （Date 版是 'MM-DD'）
 *   3. 其余               → 同样是 fromNow()，即「3 小时前」
 *
 * 多出来的那个 HH:mm 用在**帖子详情页**底部那行「发布于 ……」——
 * 作者 `components\post-detail.vue` 的 `#footer` 里调的是 `formatPrettyTime`，
 * 不是 `formatPrettyDate`（列表卡片才用 Date 版）。两个函数都留着，谁也别删。
 *
 * 作者那边 `mt.utc(true)` 的意思是「按本地时间取年月日时分」，和 `new Date()` 的
 * `getHours()/getMinutes()` 取本地时间是同一件事，所以直接这样写。
 *
 * @param rfc3339 后端给的 created_at，形如 "2026-02-18T10:30:00+08:00"
 */
export function formatPrettyTime(rfc3339: string): string {
  const target = new Date(rfc3339)
  if (Number.isNaN(target.getTime())) return ''

  const now = new Date()
  const hhmm = `${pad(target.getHours())}:${pad(target.getMinutes())}`

  if (target.getFullYear() !== now.getFullYear()) {
    return `${target.getFullYear()}-${pad(target.getMonth() + 1)}-${pad(target.getDate())} ${hhmm}`
  }

  if (diffInMonths(now, target) > 3) {
    return `${pad(target.getMonth() + 1)}-${pad(target.getDate())} ${hhmm}`
  }

  return fromNow((now.getTime() - target.getTime()) / 1000)
}
