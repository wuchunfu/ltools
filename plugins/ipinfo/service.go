package ipinfo

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Service IP信息服务
type Service struct {
	plugin *Plugin
	app    *application.App
	cache  *IPInfo
	mu     sync.RWMutex
}

// NewService 创建新的IP信息服务
func NewService(plugin *Plugin, app *application.App) *Service {
	return &Service{
		plugin: plugin,
		app:    app,
	}
}

// GetIPInfo 获取当前IP信息
func (s *Service) GetIPInfo() (*IPInfo, error) {
	// 检查缓存（5分钟有效）
	s.mu.RLock()
	if s.cache != nil && time.Since(s.cache.FetchedAt) < 5*time.Minute {
		info := *s.cache
		s.mu.RUnlock()
		return &info, nil
	}
	s.mu.RUnlock()

	// 使用 ip-api.com API（免费，每分钟45次请求限制）
	resp, err := http.Get("http://ip-api.com/json/?lang=zh-CN")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IP info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Status      string  `json:"status"`
		Message     string  `json:"message,omitempty"`
		Query       string  `json:"query"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"regionName"`
		City        string  `json:"city"`
		ISP         string  `json:"isp"`
		Org         string  `json:"org"`
		Timezone    string  `json:"timezone"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	info := &IPInfo{
		IP:          result.Query,
		Country:     result.Country,
		CountryCode: result.CountryCode,
		Region:      result.Region,
		City:        result.City,
		ISP:         result.ISP,
		Org:         result.Org,
		Timezone:    result.Timezone,
		Lat:         result.Lat,
		Lon:         result.Lon,
		Query:       result.Query,
		FetchedAt:   time.Now(),
	}

	// 更新缓存
	s.mu.Lock()
	s.cache = info
	s.mu.Unlock()

	return info, nil
}

// Refresh 强制刷新IP信息
func (s *Service) Refresh() (*IPInfo, error) {
	s.mu.Lock()
	s.cache = nil
	s.mu.Unlock()
	return s.GetIPInfo()
}

// ClearCache 清除缓存
func (s *Service) ClearCache() {
	s.mu.Lock()
	s.cache = nil
	s.mu.Unlock()
}

// GetLocalIPs 获取本地IP地址列表
func (s *Service) GetLocalIPs() []LocalIPInfo {
	var result []LocalIPInfo

	interfaces, err := net.Interfaces()
	if err != nil {
		return result
	}

	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		info := LocalIPInfo{
			Interface: iface.Name,
			MAC:       iface.HardwareAddr.String(),
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			if ipNet.IP.IsLinkLocalUnicast() || ipNet.IP.IsLinkLocalMulticast() {
				continue
			}

			if ipNet.IP.To4() != nil {
				info.IP = ipNet.IP.String()
			} else if ipNet.IP.To16() != nil {
				info.IPv6 = ipNet.IP.String()
			}
		}

		// 只添加有IP地址的接口
		if info.IP != "" || info.IPv6 != "" {
			result = append(result, info)
		}
	}

	return result
}

// GetHostname 获取主机名
func (s *Service) GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// GetMACAddress 获取主MAC地址（第一个非回环接口的MAC）
func (s *Service) GetMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		// 跳过回环接口和未启用的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 跳过虚拟接口（通常以 veth、docker、bridge 开头）
		name := strings.ToLower(iface.Name)
		if strings.HasPrefix(name, "veth") ||
			strings.HasPrefix(name, "docker") ||
			strings.HasPrefix(name, "bridge") ||
			strings.HasPrefix(name, "virbr") ||
			strings.HasPrefix(name, "utun") ||
			strings.HasPrefix(name, "awdl") ||
			strings.HasPrefix(name, "llw") ||
			strings.HasPrefix(name, "anpi") {
			continue
		}

		if iface.HardwareAddr != nil && iface.HardwareAddr.String() != "" {
			return iface.HardwareAddr.String()
		}
	}

	return ""
}
