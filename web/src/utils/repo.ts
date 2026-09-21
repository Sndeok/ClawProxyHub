// 仓库地址集中配置。
//
// 本仓库是 Sndeok/ClawProxyHub 的二开 fork：顶栏 GitHub 入口与默认发布页都指向 fork，
// 需要改回上游（ShadowSmallBaby）或换到别的镜像，只改这个文件即可。
export const REPO_OWNER = 'Sndeok'
export const REPO_NAME = 'ClawProxyHub'
export const REPO_URL = `https://github.com/${REPO_OWNER}/${REPO_NAME}`
export const REPO_RELEASES_URL = `${REPO_URL}/releases`

// 插件市场默认仓库（后端内置默认值同源，插件页展示用）
export const PLUGIN_REPO_URL = 'https://github.com/Sndeok/ClawProxyHubPlugins'