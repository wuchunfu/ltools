package main

import (
	"net/http"
	"strings"
)

// CombinedAssetHandler 组合资源处理器
// 先尝试代理请求，如果不是代理请求则交给默认处理器
type CombinedAssetHandler struct {
	proxyHandler   http.Handler
	defaultHandler http.Handler
}

// NewCombinedAssetHandler 创建组合资源处理器
func NewCombinedAssetHandler(proxyHandler, defaultHandler http.Handler) *CombinedAssetHandler {
	return &CombinedAssetHandler{
		proxyHandler:   proxyHandler,
		defaultHandler: defaultHandler,
	}
}

// ServeHTTP 实现 http.Handler 接口
func (h *CombinedAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 检查是否是代理请求
	if strings.HasPrefix(path, "/proxy/") {
		h.proxyHandler.ServeHTTP(w, r)
		return
	}

	// 其他请求交给默认处理器
	h.defaultHandler.ServeHTTP(w, r)
}
