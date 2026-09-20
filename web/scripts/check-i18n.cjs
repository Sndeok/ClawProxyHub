// 校验 i18n 消息能被 vue-i18n 正常编译。
//
// 为什么要单独检查：vue-i18n 的消息语法里 @ 是「链接消息」保留符（@:key），
// 所以文案里写 socks5://user:pass@host 会在 t() 时抛
// "Message compilation error: Invalid linked format"。这个错误发生在渲染期，
// 会让抛出位置所在的整个子树渲染失败 —— 表现为「卡片只剩标题、内容空白」，
// 而 vue-tsc / vite build 都不会报错。这里在构建前把每条消息编译一遍。
//
// 需要字面量 @ 时写成 {'@'}。
const fs = require('node:fs')
const path = require('node:path')
const { createI18n } = require('vue-i18n')

const dir = path.join(__dirname, '..', 'src', 'i18n')
const locales = ['zh', 'en']
const i18n = createI18n({ legacy: false, locale: 'check', fallbackLocale: 'check', messages: { check: {} } })

let checked = 0
const failures = []
for (const locale of locales) {
  const src = fs.readFileSync(path.join(dir, `${locale}.ts`), 'utf8')
  // 提取 key: '...' 字面量（i18n 文件只有单引号字符串，转义按 TS 规则）
  for (const m of src.matchAll(/^\s*(?:'?[\w.-]+'?):\s*'((?:[^'\\]|\\.)*)'/gm)) {
    const raw = m[1].replace(/\\'/g, "'").replace(/\\\\/g, '\\')
    checked++
    try {
      i18n.global.setLocaleMessage('check', { probe: raw })
      i18n.global.t('probe')
    } catch (e) {
      const line = src.slice(0, m.index).split('\n').length
      failures.push(`${locale}.ts:${line}  ${e.message.split('\n')[0]}\n      文案: ${raw}`)
    }
  }
}

if (failures.length) {
  console.error(`\n[i18n] ${failures.length}/${checked} 条消息无法编译：\n`)
  for (const f of failures) console.error('  • ' + f)
  console.error('\n提示：字面量 @ 需要写成 {\'@\'}，例如 socks5://user:pass{\'@\'}host:port\n')
  process.exit(1)
}
console.log(`[i18n] ${checked} 条消息编译通过`)