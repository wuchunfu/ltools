package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ProxyAssetHandler 自定义 Asset Handler，用于代理音频和图片请求
type ProxyAssetHandler struct {
	defaultHandler http.Handler
	urlMapping     map[string]string // 本地路径 -> 远程 URL
	mutex          sync.RWMutex
}

// NewProxyAssetHandler 创建代理 Asset Handler
func NewProxyAssetHandler(defaultHandler http.Handler) *ProxyAssetHandler {
	return &ProxyAssetHandler{
		defaultHandler: defaultHandler,
		urlMapping:     make(map[string]string),
	}
}

// RegisterAudioURL 注册音频 URL 到代理映射
func (h *ProxyAssetHandler) RegisterAudioURL(resourceID, remoteURL string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.urlMapping["/proxy/audio/"+resourceID] = remoteURL
	log.Printf("[ProxyAssetHandler] Registered audio: %s -> %s", resourceID, remoteURL)
}

// RegisterImageURL 注册图片 URL 到代理映射
func (h *ProxyAssetHandler) RegisterImageURL(resourceID, remoteURL string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.urlMapping["/proxy/image/"+resourceID] = remoteURL
	log.Printf("[ProxyAssetHandler] Registered image: %s -> %s", resourceID, remoteURL)
}

// ServeHTTP 实现 http.Handler 接口
func (h *ProxyAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 检查是否是代理请求
	if strings.HasPrefix(path, "/proxy/audio/") || strings.HasPrefix(path, "/proxy/image/") {
		h.handleProxy(w, r, path)
		return
	}

	// 其他请求交给默认处理器
	h.defaultHandler.ServeHTTP(w, r)
}

// handleProxy 处理代理请求
func (h *ProxyAssetHandler) handleProxy(w http.ResponseWriter, r *http.Request, path string) {
	h.mutex.RLock()
	remoteURL, ok := h.urlMapping[path]
	h.mutex.RUnlock()

	if !ok {
		http.Error(w, "Resource not found", http.StatusNotFound)
		log.Printf("[ProxyAssetHandler] Resource not found: %s", path)
		return
	}

	log.Printf("[ProxyAssetHandler] Proxying: %s -> %s", path, remoteURL)

	// 创建代理请求
	proxyReq, err := http.NewRequest(r.Method, remoteURL, nil)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		log.Printf("[ProxyAssetHandler] Failed to create request: %v", err)
		return
	}

	// 复制请求头
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// 设置 User-Agent
	proxyReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 手动处理重定向
		},
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("[ProxyAssetHandler] Failed to fetch: %v", err)
		http.Error(w, "Failed to fetch resource", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 处理重定向
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		location := resp.Header.Get("Location")
		if location != "" {
			log.Printf("[ProxyAssetHandler] Following redirect to: %s", location)
			redirectReq, err := http.NewRequest(r.Method, location, nil)
			if err != nil {
				http.Error(w, "Failed to create redirect request", http.StatusInternalServerError)
				return
			}
			redirectReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

			redirectResp, err := client.Do(redirectReq)
			if err != nil {
				http.Error(w, "Failed to follow redirect", http.StatusBadGateway)
				return
			}
			defer redirectResp.Body.Close()
			resp = redirectResp
		}
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		http.Error(w, fmt.Sprintf("Server returned status %d", resp.StatusCode), resp.StatusCode)
		return
	}

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 添加 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Range")

	// 写入响应状态码
	w.WriteHeader(resp.StatusCode)

	// 流式传输响应体
	written, err := io.Copy(w, resp.Body)
	if err != nil {
		// 忽略 broken pipe 错误（客户端提前关闭连接）
		if !isBrokenPipeError(err) {
			log.Printf("[ProxyAssetHandler] Failed to write response: %v", err)
		}
		return
	}

	log.Printf("[ProxyAssetHandler] Proxied %d bytes for %s", written, path)
}

// isBrokenPipeError 检查是否是 broken pipe 错误
func isBrokenPipeError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "connection reset by peer")
}
