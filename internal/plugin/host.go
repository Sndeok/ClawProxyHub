// host.go — 核心侧 ClawHost 服务：插件经 broker 反调的统一入口。
// 每个插件实例持有独立的 HostService，KV 按 plugin 隔离并持久化到 plugin_storage 表。
package plugin

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"gorm.io/gorm"

	"google.golang.org/grpc"

	"github.com/ShadowSmallBaby/ClawProxyHub/internal/model"
	pb "github.com/ShadowSmallBaby/ClawProxyHub/sdk/proto/cphv1"
)

// HostService ClawHost gRPC 实现。按插件隔离 key 空间，持久化到 plugin_storage 表。
type HostService struct {
	pb.UnimplementedClawHostServer

	db         *gorm.DB
	pluginName atomic.Value // string，握手完成后设置
	mu         sync.RWMutex
	stores     map[string]map[string][]byte // plugin name → key → value（内存缓存）
}

func NewHostService(db *gorm.DB) *HostService {
	return &HostService{db: db, stores: map[string]map[string][]byte{}}
}

// SetPluginName 握手完成后设置插件名（此后 KV 操作按该插件隔离）。
func (h *HostService) SetPluginName(name string) {
	h.pluginName.Store(name)
}

// plugin 获取当前插件名（未设置时用空串，兼容握手期调用）。
func (h *HostService) plugin() string {
	if v, ok := h.pluginName.Load().(string); ok {
		return v
	}
	return ""
}

func (h *HostService) Log(ctx context.Context, e *pb.LogEntry) (*pb.Empty, error) {
	log.Printf("[plugin:%s] %s: %s", h.plugin(), e.Level, e.Message)
	return &pb.Empty{}, nil
}

func (h *HostService) StoreGet(ctx context.Context, r *pb.StoreGetRequest) (*pb.StoreGetResponse, error) {
	plugin := h.plugin()

	// 先查内存缓存
	h.mu.RLock()
	if m, ok := h.stores[plugin]; ok {
		if v, ok := m[r.Key]; ok {
			h.mu.RUnlock()
			return &pb.StoreGetResponse{Value: v, Found: true}, nil
		}
	}
	h.mu.RUnlock()

	// 缓存未命中 → 查库
	var rec model.PluginStorage
	if err := h.db.Where("plugin = ? AND key = ?", plugin, r.Key).First(&rec).Error; err != nil {
		return &pb.StoreGetResponse{Found: false}, nil
	}

	// 回填缓存
	h.mu.Lock()
	if h.stores[plugin] == nil {
		h.stores[plugin] = map[string][]byte{}
	}
	h.stores[plugin][r.Key] = rec.Value
	h.mu.Unlock()

	return &pb.StoreGetResponse{Value: rec.Value, Found: true}, nil
}

func (h *HostService) StorePut(ctx context.Context, r *pb.StorePutRequest) (*pb.Empty, error) {
	plugin := h.plugin()

	// 写库（upsert）
	if err := h.db.Exec(
		`INSERT INTO plugin_storage (plugin, key, value, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(plugin, key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
		plugin, r.Key, r.Value,
	).Error; err != nil {
		return nil, fmt.Errorf("plugin storage write: %w", err)
	}

	// 更新内存缓存
	h.mu.Lock()
	if h.stores[plugin] == nil {
		h.stores[plugin] = map[string][]byte{}
	}
	h.stores[plugin][r.Key] = r.Value
	h.mu.Unlock()

	return &pb.Empty{}, nil
}

func (h *HostService) GetProxy(ctx context.Context, r *pb.GetProxyRequest) (*pb.ProxyConfig, error) {
	var links []model.GroupProxy
	if err := h.db.Where("group_id = ?", r.GroupId).Order("proxy_id").Find(&links).Error; err != nil || len(links) == 0 {
		return nil, fmt.Errorf("no proxy bound to group %s", r.GroupId)
	}
	var proxy model.Proxy
	if err := h.db.First(&proxy, links[0].ProxyID).Error; err != nil {
		return nil, fmt.Errorf("proxy record missing")
	}
	return &pb.ProxyConfig{
		Scheme: proxy.Scheme, Host: proxy.Host, Port: proxy.Port,
		Username: proxy.Username, Password: proxy.Password,
	}, nil
}

// GetSettings 读插件设置（管理界面在线修改，保存即生效）。
func (h *HostService) GetSettings(ctx context.Context, r *pb.GetSettingsRequest) (*pb.GetSettingsResponse, error) {
	var p model.Plugin
	if err := h.db.Select("settings_json").Where("name = ?", r.Plugin).First(&p).Error; err != nil {
		return &pb.GetSettingsResponse{Values: []byte("{}")}, nil // 记录缺失按空配置处理
	}
	if p.SettingsJSON == "" {
		p.SettingsJSON = "{}"
	}
	return &pb.GetSettingsResponse{Values: []byte(p.SettingsJSON)}, nil
}

// ServeHost 在 broker 上挂出宿主服务（由 ClawPluginPlugin.GRPCClient 调用）。
func (h *HostService) ServeHost(broker interface {
	AcceptAndServe(id uint32, f func([]grpc.ServerOption) *grpc.Server)
}) {
	broker.AcceptAndServe(hostBrokerID, func(opts []grpc.ServerOption) *grpc.Server {
		srv := grpc.NewServer(opts)
		pb.RegisterClawHostServer(srv, h)
		return srv
	})
}