// settings.go — 系统设置 API（网关全局参数）。
package admin

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ShadowSmallBaby/ClawProxyHub/internal/setting"
)

// getSettings GET /admin/settings
func (s *Server) getSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"settings": map[string]interface{}{
			"first_event_timeout": int(s.settings.FirstEventTimeout().Seconds()),
			"github_proxy":        s.settings.GitHubProxy(),
			"marketplace_url":     s.marketURL(),
			"market_proxy":        s.settings.MarketProxy(),
			"log_retention_days":  s.settings.LogRetentionDays(),
		},
	})
}

// putSettings PUT /admin/settings — body: {first_event_timeout, github_proxy,
// marketplace_url, market_proxy, log_retention_days}。
func (s *Server) putSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FirstEventTimeout int    `json:"first_event_timeout"`
		GitHubProxy       string `json:"github_proxy"`
		MarketplaceURL    string `json:"marketplace_url"`
		MarketProxy       string `json:"market_proxy"`
		LogRetentionDays  int    `json:"log_retention_days"`
	}
	if !readBody(w, r, &body) {
		return
	}
	if body.FirstEventTimeout < 5 || body.FirstEventTimeout > 3600 {
		http.Error(w, `{"error":"首事件超时需在 5–3600 秒之间"}`, http.StatusBadRequest)
		return
	}
	proxy := strings.TrimSuffix(strings.TrimSpace(body.GitHubProxy), "/")
	if proxy != "" && !strings.HasPrefix(proxy, "http://") && !strings.HasPrefix(proxy, "https://") {
		http.Error(w, `{"error":"GitHub 代理需以 http:// 或 https:// 开头（如 https://ghproxy.com），留空则直连"}`, http.StatusBadRequest)
		return
	}
	marketURL := strings.TrimSpace(body.MarketplaceURL)
	if marketURL != "" && !strings.HasPrefix(marketURL, "http://") && !strings.HasPrefix(marketURL, "https://") {
		http.Error(w, `{"error":"插件市场地址需以 http:// 或 https:// 开头（指向 index.json），留空则用默认地址"}`, http.StatusBadRequest)
		return
	}
	marketProxy := strings.TrimSpace(body.MarketProxy)
	if !validMarketProxy(marketProxy) {
		http.Error(w, `{"error":"代理地址无效：支持 socks5://host:port、socks5://user:pass@host:port、http://host:port（省略协议头按 socks5 处理）"}`, http.StatusBadRequest)
		return
	}
	if body.LogRetentionDays < 0 || body.LogRetentionDays > 3650 {
		http.Error(w, `{"error":"日志保留天数需在 0–3650 之间（0 = 永久保留）"}`, http.StatusBadRequest)
		return
	}
	s.settings.Set(setting.KeyFirstEventTimeout, strconv.Itoa(body.FirstEventTimeout))
	s.settings.Set(setting.KeyGitHubProxy, proxy)
	s.settings.Set(setting.KeyMarketplaceURL, marketURL)
	s.settings.Set(setting.KeyMarketProxy, marketProxy)
	s.settings.Set(setting.KeyLogRetentionDays, strconv.Itoa(body.LogRetentionDays))
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// testMarket POST /admin/settings/test-market — 用给定（或当前生效）的地址与代理
// 试拉一次市场索引，返回条数 / 耗时 / 插件清单，便于在页面上确认代理是否通。
// 请求体可省略；传入的值不落库。
func (s *Server) testMarket(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MarketplaceURL string `json:"marketplace_url"`
		MarketProxy    string `json:"market_proxy"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body)

	rawURL := strings.TrimSpace(body.MarketplaceURL)
	if rawURL == "" {
		rawURL = s.marketURL()
	}
	rawProxy := strings.TrimSpace(body.MarketProxy)
	if rawProxy == "" {
		rawProxy = s.settings.MarketProxy()
	}
	target := s.withGitHubProxy(rawURL)

	client, err := marketClient(rawProxy, marketIndexTimeout)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": false, "error": err.Error(), "url": target, "proxy": rawProxy,
		})
		return
	}
	started := time.Now()
	entries, err := fetchIndex(client, target)
	elapsed := time.Since(started).Milliseconds()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok": false, "error": err.Error(), "url": target, "proxy": rawProxy, "elapsed_ms": elapsed,
		})
		return
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name+"@"+e.Version)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true, "url": target, "proxy": rawProxy,
		"elapsed_ms": elapsed, "count": len(entries), "plugins": names,
	})
}
