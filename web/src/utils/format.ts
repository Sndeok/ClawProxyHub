// 展示层格式化工具：数字、时间等在多个页面/组件间共用的转换。

// fmtNum 数值字符串 → 最多两位小数；空值给 "-"。
export function fmtNum(v: string | number | undefined | null): string {
  if (v === undefined || v === null || v === '') return '-'
  const n = Number(v)
  return Number.isFinite(n) ? String(Math.round(n * 100) / 100) : String(v)
}

// fmtCompact 大数压缩：12345 → 12.3K
export function fmtCompact(n?: number): string {
  const v = n || 0
  if (v < 1000) return String(v)
  if (v < 1000000) return `${(v / 1000).toFixed(1).replace(/\.0$/, '')}K`
  return `${(v / 1000000).toFixed(2)}M`
}

// fmtTime 后端时间戳 → "YYYY-MM-DD HH:mm:ss"。
export function fmtTime(t: string | null | undefined): string {
  return t ? t.replace('T', ' ').slice(0, 19) : '-'
}
