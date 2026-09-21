<template>
  <t-layout class="layout">
    <t-aside :width="collapsed ? '76px' : '232px'" class="aside">
      <div class="brand" @click="router.push('/dashboard')">
        <img class="brand-badge" src="/logo.png" alt="ClawProxyHub" />
        <div v-if="!collapsed" class="brand-text">
          <span class="brand-name">Claw<span>ProxyHub</span></span>
          <span class="brand-sub">{{ $t('common.brandSub') }}</span>
        </div>
      </div>

      <nav class="nav">
        <template v-for="group in menuGroups" :key="group.key">
          <div v-if="!collapsed" class="nav-label">{{ $t('menuGroup.' + group.key) }}</div>
          <div v-else class="nav-label nav-label--dot"></div>
          <t-menu
            :value="route.path"
            :collapsed="collapsed"
            :width="collapsed ? '76px' : '232px'"
            class="nav-menu"
            @change="(v: string) => router.push(v)"
          >
            <t-menu-item v-for="item in group.items" :key="item.value" :value="item.value">
              <template #icon><component :is="item.icon" /></template>{{ $t(item.label) }}
            </t-menu-item>
          </t-menu>
        </template>
      </nav>

      <div class="aside-footer" @click="collapsed = !collapsed">
        <chevron-left-icon v-if="!collapsed" />
        <chevron-right-icon v-else />
        <span v-if="!collapsed" class="aside-footer-text">{{ $t('common.collapse') }}</span>
      </div>
    </t-aside>

    <t-layout class="main">
      <t-header class="header">
        <div class="header-left">
          <div class="header-title">{{ $t(currentPage.label) }}</div>
          <div class="header-desc">{{ $t(currentPage.desc) }}</div>
        </div>
        <div class="header-right">
          <t-tooltip v-if="version" :content="updateAvailable ? $t('common.hasUpdate') : $t('common.checkUpdate')">
            <div class="ver-chip" :class="{ 'has-update': updateAvailable, checking }" @click="checkVersion(true)">
              <t-loading v-if="checking" size="12px" />
              <span class="ver-text num">v{{ version }}</span>
              <span v-if="updateAvailable && !checking" class="ver-dot" />
            </div>
          </t-tooltip>
          <t-tooltip :content="$t('common.github')">
            <t-button variant="text" shape="square" theme="default" @click="openGithub">
              <logo-github-icon />
            </t-button>
          </t-tooltip>
          <t-popup trigger="click">
            <t-button variant="text" shape="square" theme="default">
              <translate-icon />
            </t-button>
            <template #content>
              <div class="lang-menu">
                <div
                  v-for="l in langs"
                  :key="l"
                  class="lang-menu-item"
                  :class="{ active: locale === l }"
                  @click="setLocale(l)"
                >
                  {{ $t('common.' + l) }}
                  <check-icon v-if="locale === l" />
                </div>
              </div>
            </template>
          </t-popup>
          <t-tooltip :content="dark ? $t('common.light') : $t('common.dark')">
            <t-button variant="text" shape="square" theme="default" @click="dark = !dark">
              <moon-icon v-if="!dark" />
              <sunny-icon v-else />
            </t-button>
          </t-tooltip>
          <t-popup trigger="click">
            <div class="user-chip">
              <t-avatar size="26px" theme="light" class="user-avatar">
                <template #icon><user-icon /></template>
              </t-avatar>
              <span v-if="username" class="user-name">{{ username }}</span>
            </div>
            <template #content>
              <div class="user-menu">
                <div class="user-menu-head">
                  <t-avatar size="34px" theme="light">
                    <template #icon><user-icon /></template>
                  </t-avatar>
                  <div>
                    <div class="user-menu-name">{{ username || $t('common.admin') }}</div>
                    <div class="user-menu-sub">{{ roleLabel }}</div>
                  </div>
                </div>
                <div class="user-menu-item" @click="logout">
                  <poweroff-icon /> {{ $t('common.logout') }}
                </div>
              </div>
            </template>
          </t-popup>
        </div>
      </t-header>

      <t-content class="content">
        <router-view />
      </t-content>
    </t-layout>
  </t-layout>
</template>

<script setup lang="ts">
import { computed, ref, watch, type Component } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  DashboardIcon, AppIcon, UserIcon, FolderIcon, InternetIcon,
  LockOnIcon, TimeIcon, FileIcon, RootListIcon, SettingIcon,
  ChevronLeftIcon, ChevronRightIcon, TranslateIcon, MoonIcon, SunnyIcon,
  PoweroffIcon, CheckIcon, LogoGithubIcon,
} from 'tdesign-icons-vue-next'
import { MessagePlugin } from 'tdesign-vue-next'
import { api, clearToken, getToken } from '../api/client'
import i18n, { setLocale as applyLocale, type Locale } from '../i18n'
import { REPO_URL, REPO_RELEASES_URL } from '../utils/repo'

const route = useRoute()
const router = useRouter()
const dark = ref(localStorage.getItem('cph-theme') === 'dark')
const collapsed = ref(localStorage.getItem('cph-sidebar') === 'collapsed')
// token 形如 "user:password"，展示用户名部分
const username = computed(() => getToken().split(':')[0])
// 角色展示名（管理员 / 访客，后端 /admin/me 提供）
const role = ref('')
const roleLabel = computed(() =>
  i18n.global.t(role.value === 'guest' ? 'common.guest' : 'common.admin'),
)
api.get<{ role: string }>('/admin/me').then((r) => (role.value = r.role)).catch(() => {})

// ---------- 多语言（zh / en，vue-i18n 全局响应式） ----------
const locale = computed<Locale>(() => i18n.global.locale.value as Locale)
const langs: Locale[] = ['zh', 'en']
function setLocale(l: Locale) {
  applyLocale(l)
}

// 项目主页：二开 fork 的仓库地址（需要改回上游只改 utils/repo.ts）
function openGithub() {
  window.open(REPO_URL, '_blank')
}

// ---------- 版本信息（头部徽标；点击重查，首次进页自动查一次） ----------
const version = ref('')
const latest = ref('')
const updateAvailable = ref(false)
const checking = ref(false)
const releaseUrl = ref(REPO_RELEASES_URL)

// checkVersion 拉取本机/远端版本对比；manual=true 时弹结果提示（手动点击）。
async function checkVersion(manual = false) {
  if (checking.value) return
  checking.value = true
  try {
    const r = await api.get<{ version: string; latest?: string; update_available?: boolean; release_url?: string }>('/admin/version')
    version.value = r.version
    latest.value = r.latest ?? ''
    updateAvailable.value = !!r.update_available
    if (r.release_url) releaseUrl.value = r.release_url
    if (manual) {
      if (updateAvailable.value) {
        MessagePlugin.info(i18n.global.t('common.updateFound', { v: latest.value }))
        window.open(releaseUrl.value, '_blank')
      } else if (latest.value) {
        MessagePlugin.success(i18n.global.t('common.upToDate'))
      } else {
        MessagePlugin.warning(i18n.global.t('common.checkFailed'))
      }
    }
  } catch {
    if (manual) MessagePlugin.warning(i18n.global.t('common.checkFailed'))
  } finally {
    checking.value = false
  }
}
checkVersion() // 首次进页自动查一次

interface MenuItem {
  value: string
  label: string // i18n key（menu.xxx）
  desc: string // i18n key（menuDesc.xxx）
  icon: Component
}

// 菜单按使用频率分四组：先看总览，再配资源，然后管流量，最后做运维
const menuGroups: { key: string; items: MenuItem[] }[] = [
  {
    key: 'overview',
    items: [
      { value: '/dashboard', label: 'menu.dashboard', desc: 'menuDesc.dashboard', icon: DashboardIcon },
    ],
  },
  {
    key: 'resource',
    items: [
      { value: '/plugins', label: 'menu.plugins', desc: 'menuDesc.plugins', icon: AppIcon },
      { value: '/accounts', label: 'menu.accounts', desc: 'menuDesc.accounts', icon: UserIcon },
      { value: '/groups', label: 'menu.groups', desc: 'menuDesc.groups', icon: FolderIcon },
      { value: '/proxies', label: 'menu.proxies', desc: 'menuDesc.proxies', icon: RootListIcon },
    ],
  },
  {
    key: 'traffic',
    items: [
      { value: '/routes', label: 'menu.routes', desc: 'menuDesc.routes', icon: InternetIcon },
      { value: '/keys', label: 'menu.keys', desc: 'menuDesc.keys', icon: LockOnIcon },
      { value: '/logs', label: 'menu.logs', desc: 'menuDesc.logs', icon: FileIcon },
    ],
  },
  {
    key: 'ops',
    items: [
      { value: '/tasks', label: 'menu.tasks', desc: 'menuDesc.tasks', icon: TimeIcon },
      { value: '/settings', label: 'menu.settings', desc: 'menuDesc.settings', icon: SettingIcon },
    ],
  },
]

const flatMenu = menuGroups.flatMap((g) => g.items)

// 当前菜单（含子路径前缀匹配）
const currentPage = computed(
  () => flatMenu.find((m) => route.path.startsWith(m.value)) ?? flatMenu[0],
)

// 主题切换与收起状态持久化
watch(dark, (v) => {
  const mode = v ? 'dark' : 'light'
  document.documentElement.setAttribute('theme-mode', mode)
  localStorage.setItem('cph-theme', mode)
})
watch(collapsed, (v) => {
  localStorage.setItem('cph-sidebar', v ? 'collapsed' : 'expanded')
})

// 初始主题
document.documentElement.setAttribute('theme-mode', dark.value ? 'dark' : 'light')

function logout() {
  clearToken()
  router.push('/login')
}
</script>

<style scoped>
.layout {
  height: 100%;
  min-width: 0;
}

/* ---------- 侧栏 ---------- */
.aside {
  flex-shrink: 0; /* TDesign sider 默认参与收缩，会把侧栏挤没 */
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: var(--cph-bg-aside);
  border-right: 1px solid var(--cph-border);
  transition: width 0.22s cubic-bezier(0.22, 0.61, 0.36, 1);
  overflow: hidden;
}
.brand {
  height: 62px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 18px;
  cursor: pointer;
  overflow: hidden;
}
.brand-badge {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  flex-shrink: 0;
  object-fit: cover;
  box-shadow: 0 2px 8px rgba(61, 114, 245, 0.25);
}
.brand-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.brand-name {
  font-size: 15px;
  font-weight: 700;
  white-space: nowrap;
  color: var(--cph-text-1);
  letter-spacing: 0.2px;
}
.brand-name span {
  font-weight: 400;
  color: var(--td-brand-color);
}
.brand-sub {
  font-size: 10px;
  color: var(--cph-text-3);
  white-space: nowrap;
  letter-spacing: 0.4px;
  text-transform: uppercase;
}

.nav {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 4px 10px 12px;
}
.nav-label {
  padding: 12px 10px 4px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.6px;
  color: var(--cph-text-3);
  text-transform: uppercase;
  white-space: nowrap;
}
.nav-label--dot {
  height: 12px;
  padding: 0;
  margin: 6px 12px;
  border-top: 1px solid var(--cph-border);
}
/* TDesign 菜单默认高度 100%，多个菜单叠起来会把后面几组挤到屏幕外，这里放开 */
.nav-menu {
  background: transparent;
  width: 100% !important;
  height: auto !important;
  min-height: 0 !important;
}
/* 明暗两种模式下都保证菜单文字对比度 */
.nav-menu :deep(.t-menu__item) {
  color: var(--cph-text-2);
  margin: 2px 0;
}
.nav-menu :deep(.t-menu__item:hover) {
  color: var(--cph-text-1);
  background: var(--cph-surface-2);
}
.nav-menu :deep(.t-menu__item.t-is-active) {
  color: var(--td-brand-color);
  font-weight: 600;
}
.nav-menu :deep(.t-menu__item.t-is-active .t-icon) {
  color: var(--td-brand-color);
}

.aside-footer {
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  cursor: pointer;
  color: var(--cph-text-3);
  border-top: 1px solid var(--cph-border);
  flex-shrink: 0;
  font-size: 12px;
  transition: color 0.16s ease, background-color 0.16s ease;
}
.aside-footer:hover {
  color: var(--td-brand-color);
  background: var(--cph-surface-2);
}
.aside-footer-text {
  white-space: nowrap;
}

/* ---------- 主区 ---------- */
.main {
  min-width: 0;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  height: 62px;
  padding: 0 22px;
  flex-shrink: 0;
}
.header-left {
  min-width: 0;
}
.header-title {
  font-size: 17px;
  font-weight: 650;
  line-height: 1.25;
}
.header-desc {
  font-size: 12px;
  color: var(--cph-text-3);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

/* 版本 chip：默认灰字，有更新时描边高亮 + 红点 */
.ver-chip {
  display: flex;
  align-items: center;
  gap: 5px;
  height: 26px;
  padding: 0 10px;
  border-radius: 13px;
  border: 1px solid transparent;
  font-size: 12px;
  color: var(--cph-text-3);
  cursor: pointer;
  transition: all 0.15s ease;
}
.ver-chip:hover {
  color: var(--cph-text-1);
  background: var(--cph-surface-2);
}
.ver-chip.checking {
  cursor: progress;
}
.ver-chip.has-update {
  color: var(--td-warning-color);
  border-color: var(--td-warning-color-3);
  background: var(--td-warning-color-1);
}
.ver-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--td-warning-color);
}

/* 语言菜单 */
.lang-menu {
  min-width: 140px;
}
.lang-menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 12px;
  font-size: 13px;
  color: var(--cph-text-1);
  cursor: pointer;
  white-space: nowrap;
  border-radius: 6px;
  transition: background-color 0.15s ease;
}
.lang-menu-item:hover {
  background: var(--cph-surface-2);
}
.lang-menu-item.active {
  color: var(--td-brand-color);
}

/* 用户胶囊 */
.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 10px;
  border-radius: 18px;
  transition: background-color 0.15s ease;
}
.user-chip:hover {
  background: var(--cph-surface-2);
}
.user-avatar {
  color: var(--td-brand-color);
}
.user-name {
  font-size: 13px;
  color: var(--cph-text-1);
  max-width: 120px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.user-menu {
  min-width: 180px;
}
.user-menu-head {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px 12px;
  border-bottom: 1px solid var(--cph-border);
  margin-bottom: 4px;
}
.user-menu-head .t-avatar {
  color: var(--td-brand-color);
}
.user-menu-name {
  font-size: 14px;
  font-weight: 600;
}
.user-menu-sub {
  font-size: 12px;
  color: var(--cph-text-3);
}
.user-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 13px;
  color: var(--cph-text-1);
  cursor: pointer;
  transition: background-color 0.15s ease;
}
.user-menu-item:hover {
  background: var(--cph-surface-2);
}

/* 内容区：双向滚动（横向溢出不再被侧栏/视口裁掉） */
.content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  height: 0;
  overflow: auto;
}
</style>