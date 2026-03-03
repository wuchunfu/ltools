package musicplayer

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

// ProxyService 音乐代理服务
type ProxyService struct {
	server     *http.Server
	port       int
	baseURL    string
	urlMapping map[string]string // 本地路径 -> 远程 URL
	mutex      sync.RWMutex
}

// NewProxyService 创建代理服务
func NewProxyService() (*ProxyService, error) {
	// 查找可用端口
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	ps := &ProxyService{
		port:       port,
		baseURL:    fmt.Sprintf("http://127.0.0.1:%d", port),
		urlMapping: make(map[string]string),
	}

	// 创建 HTTP 服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/proxy/audio/", ps.handleProxyAudio)
	mux.HandleFunc("/proxy/image/", ps.handleProxyImage)

	ps.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// 启动服务器
	go func() {
		log.Printf("[ProxyService] Starting proxy server on port %d", port)
		if err := ps.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[ProxyService] Server error: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	return ps, nil
}

// handleProxyAudio 处理代理音频请求
func (ps *ProxyService) handleProxyAudio(w http.ResponseWriter, r *http.Request) {
	ps.proxyRequest(w, r, "/proxy/audio/")
}

// handleProxyImage 处理代理图片请求
func (ps *ProxyService) handleProxyImage(w http.ResponseWriter, r *http.Request) {
	ps.proxyRequest(w, r, "/proxy/image/")
}

// proxyRequest 通用的代理请求处理
func (ps *ProxyService) proxyRequest(w http.ResponseWriter, r *http.Request, prefix string) {
	// 获取资源 ID
	resourceID := r.URL.Path[len(prefix):]

	ps.mutex.RLock()
	remoteURL, ok := ps.urlMapping[resourceID]
	ps.mutex.RUnlock()

	if !ok {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	log.Printf("[ProxyService] Proxying: %s -> %s", resourceID, remoteURL)

	// 创建代理请求
	proxyReq, err := http.NewRequest(r.Method, remoteURL, nil)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
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
		log.Printf("[ProxyService] Failed to fetch: %v", err)
		http.Error(w, "Failed to fetch resource", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 处理重定向
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		location := resp.Header.Get("Location")
		if location != "" {
			log.Printf("[ProxyService] Following redirect to: %s", location)
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

	// 设置 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Range, Content-Type")

	// 写入状态码
	w.WriteHeader(resp.StatusCode)

	// 流式传输数据
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("[ProxyService] Failed to stream data: %v", err)
	}
}

// RegisterAudioURL 注册音频 URL 并返回本地代理 URL
func (ps *ProxyService) RegisterAudioURL(remoteURL string) string {
	return ps.registerURL(remoteURL, "audio")
}

// RegisterImageURL 注册图片 URL 并返回本地代理 URL
func (ps *ProxyService) RegisterImageURL(remoteURL string) string {
	return ps.registerURL(remoteURL, "image")
}

// RegisterAudioURLSimple 简化的音频 URL 注册（暴露给前端，使用 song ID）
func (ps *ProxyService) RegisterAudioURLSimple(songID, remoteURL string) string {
	ps.mutex.Lock()
	ps.urlMapping[songID] = remoteURL
	ps.mutex.Unlock()

	localURL := fmt.Sprintf("%s/proxy/audio/%s", ps.baseURL, songID)
	log.Printf("[ProxyService] Registered audio (simple): %s -> %s", localURL, remoteURL)
	return localURL
}

// RegisterImageURLSimple 简化的图片 URL 注册（暴露给前端，使用 pic ID）
func (ps *ProxyService) RegisterImageURLSimple(picID, remoteURL string) string {
	ps.mutex.Lock()
	ps.urlMapping[picID] = remoteURL
	ps.mutex.Unlock()

	localURL := fmt.Sprintf("%s/proxy/image/%s", ps.baseURL, picID)
	log.Printf("[ProxyService] Registered image (simple): %s -> %s", localURL, remoteURL)
	return localURL
}

// registerURL 通用的 URL 注册方法
func (ps *ProxyService) registerURL(remoteURL, resourceType string) string {
	// 生成简单的 ID
	resourceID := fmt.Sprintf("%d", time.Now().UnixNano())

	ps.mutex.Lock()
	ps.urlMapping[resourceID] = remoteURL
	ps.mutex.Unlock()

	localURL := fmt.Sprintf("%s/proxy/%s/%s", ps.baseURL, resourceType, resourceID)
	log.Printf("[ProxyService] Registered %s: %s -> %s", resourceType, localURL, remoteURL)

	return localURL
}

// Close 关闭代理服务
func (ps *ProxyService) Close() error {
	if ps.server != nil {
		return ps.server.Close()
	}
	return nil
}
