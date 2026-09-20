// Package setting — 系统设置 KV（settings 表）读写，带进程内缓存。
package setting

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/ShadowSmallBaby/ClawProxyHub/internal/model"
)

// KeyFirstEventTimeout 网关首事件超时（秒）。
const KeyFirstEventTimeout = "gateway.first_event_timeout"

// KeyGitHubProxy GitHub 代理前缀（ghproxy 风格，加速插件市场访问；空 = 直连）。
const KeyGitHubProxy = "network.github_proxy"

// KeyMarketplaceURL 插件市场索引地址（用户自建；空 = 默认）。
const KeyMarketplaceURL = "network.marketplace_url"

// KeyMarketProxy 插件市场的出站代理（拉索引 + 下载插件包共用；空 = 直连）。
// 与 network.github_proxy 的区别：后者是「URL 前缀改写」（ghproxy 风格），
// 只对 GitHub 域名生效；本项是传输层代理，支持 socks5 / socks5h / http(s)。
const KeyMarketProxy = "network.market_proxy"

// KeyLogRetentionDays 调用日志保留天数（0 = 保留全部，不自动清理）。
const KeyLogRetentionDays = "logs.retention_days"

// DefaultMarketplaceURL 默认插件市场索引地址（初始化时写入设置）。
// 二开：改为自建插件仓库（Sndeok/ClawProxyHubPlugins）的 index.json。
const DefaultMarketplaceURL = "https://raw.githubusercontent.com/Sndeok/ClawProxyHubPlugins/main/index.json"

const defaultFirstEventTimeout = 90

// Store 设置存储。
type Store struct {
	db    *gorm.DB
	mu    sync.RWMutex
	cache map[string]string
}

func New(db *gorm.DB) *Store {
	return &Store{db: db, cache: map[string]string{}}
}

// Get 读设置，缺省返回 def。
func (s *Store) Get(key, def string) string {
	s.mu.RLock()
	v, ok := s.cache[key]
	s.mu.RUnlock()
	if ok {
		return v
	}
	var rec model.Setting
	if err := s.db.Where("key = ?", key).First(&rec).Error; err != nil {
		return def
	}
	s.mu.Lock()
	s.cache[key] = rec.Value
	s.mu.Unlock()
	return rec.Value
}

// Set 写设置（upsert + 刷新缓存）。
func (s *Store) Set(key, value string) {
	s.db.Exec(`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`, key, value)
	s.mu.Lock()
	s.cache[key] = value
	s.mu.Unlock()
}

// FirstEventTimeout 网关首事件超时；非法值回退默认。
func (s *Store) FirstEventTimeout() time.Duration {
	n, err := strconv.Atoi(s.Get(KeyFirstEventTimeout, strconv.Itoa(defaultFirstEventTimeout)))
	if err != nil || n <= 0 {
		n = defaultFirstEventTimeout
	}
	return time.Duration(n) * time.Second
}

// GitHubProxy GitHub 代理前缀（以 / 结尾与否均可；空 = 直连）。
func (s *Store) GitHubProxy() string {
	return s.Get(KeyGitHubProxy, "")
}

// LogRetentionDays 调用日志保留天数；非法值按 0（保留全部）处理。
func (s *Store) LogRetentionDays() int {
	n, err := strconv.Atoi(s.Get(KeyLogRetentionDays, "0"))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// MarketplaceURL 用户自建市场地址；空 = 未配置（回退默认）。
func (s *Store) MarketplaceURL() string {
	return s.Get(KeyMarketplaceURL, "")
}

// MarketProxy 插件市场出站代理；空 = 直连（或跟随进程环境代理）。
func (s *Store) MarketProxy() string {
	return strings.TrimSpace(s.Get(KeyMarketProxy, ""))
}

// EnsureDefault key 尚未写入时落默认值（仅初始化场景使用，不覆盖已有配置）。
func (s *Store) EnsureDefault(key, def string) {
	var rec model.Setting
	if err := s.db.Where("key = ?", key).First(&rec).Error; err != nil {
		s.Set(key, def)
	}
}
