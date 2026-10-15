/**
 * 数字的简写显示。
 *
 * 作者对应的是 `F:\github_paopao\paopao-ce\web\src\utils\count.ts`（8 行），
 * 逐行照抄，**包括他自己那个 bug**：
 *
 *   - `num >= 1000` 先命中，所以 `num >= 10000` 那个分支永远走不到，
 *     「万」这两个字在界面上根本出不来。按 PROJECT.md §0.1「像比好重要」，
 *     照抄不改 —— 改了界面就和原版不一样了。
 *
 * 目前只有「我的主页」（T15）的 UID 旁边那三个计数用到它。
 */
export function prettyQuoteNum(num: number): number | string {
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + '千'
  } else if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num
}
