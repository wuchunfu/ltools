package musicplayer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// LXClient LX Music 客户端（基于 stdin/stdout 通信）
type LXClient struct {
	processManager *ProcessManager

	// 请求管理
	requestID      int
	pendingReqs    map[int]chan *LXResponse
	pendingMutex   sync.RWMutex

	// 响应读取器
	responseReader *bufio.Scanner
	readerMutex    sync.Mutex

	// 配置
	requestTimeout time.Duration
}

// LXRequest 请求结构
type LXRequest struct {
	ID     int         `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

// LXResponse 响应结构
type LXResponse struct {
	ID    int             `json:"id"`
	Code  int             `json:"code"`
	Data  json.RawMessage `json:"data"`
	Error *LXError        `json:"error"`
}

// LXError 错误信息
type LXError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// NewLXClient 创建 LX 客户端
func NewLXClient(processManager *ProcessManager) (*LXClient, error) {
	if processManager == nil {
		return nil, fmt.Errorf("process manager is required")
	}

	client := &LXClient{
		processManager: processManager,
		pendingReqs:    make(map[int]chan *LXResponse),
		requestTimeout: 10 * time.Second,
	}

	// 启动响应读取 goroutine
	go client.readResponses()

	return client, nil
}

// Call 通用调用方法
func (c *LXClient) Call(ctx context.Context, method string, params interface{}) (*LXResponse, error) {
	// 生成请求 ID
	c.pendingMutex.Lock()
	c.requestID++
	id := c.requestID
	c.pendingMutex.Unlock()

	// 构建请求
	req := &LXRequest{
		ID:     id,
		Method: method,
		Params: params,
	}

	// 序列化请求
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建响应通道
	respChan := make(chan *LXResponse, 1)
	c.pendingMutex.Lock()
	c.pendingReqs[id] = respChan
	c.pendingMutex.Unlock()

	// 清理
	defer func() {
		c.pendingMutex.Lock()
		delete(c.pendingReqs, id)
		c.pendingMutex.Unlock()
	}()

	// 发送请求（添加换行符）
	stdin := c.processManager.GetStdin()
	if stdin == nil {
		return nil, fmt.Errorf("process stdin is not available")
	}

	c.readerMutex.Lock()
	_, err = stdin.Write(append(jsonData, '\n'))
	c.readerMutex.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// 等待响应或超时
	select {
	case resp := <-respChan:
		return resp, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
	case <-time.After(c.requestTimeout):
		return nil, fmt.Errorf("request timeout after %v", c.requestTimeout)
	}
}

// readResponses 读取响应的 goroutine
func (c *LXClient) readResponses() {
	stdout := c.processManager.GetStdout()
	if stdout == nil {
		log.Printf("[LXClient] stdout is not available")
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Bytes()

		var resp LXResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			log.Printf("[LXClient] Failed to parse response: %v (line: %s)", err, string(line))
			continue
		}

		// 查找对应的请求
		c.pendingMutex.RLock()
		if ch, ok := c.pendingReqs[resp.ID]; ok {
			ch <- &resp
		} else {
			log.Printf("[LXClient] Received response for unknown request ID: %d", resp.ID)
		}
		c.pendingMutex.RUnlock()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[LXClient] Error reading responses: %v", err)
	}
}

// Search 搜索歌曲
func (c *LXClient) Search(ctx context.Context, keyword, source string, page, limit int) (*SearchResult, error) {
	params := map[string]interface{}{
		"keyword": keyword,
		"source":  source,
		"page":    page,
		"limit":   limit,
	}

	resp, err := c.Call(ctx, "search", params)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("search failed: %s", resp.Error.Message)
	}

	var result SearchResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search result: %w", err)
	}

	return &result, nil
}

// GetMusicURL 获取播放 URL
func (c *LXClient) GetMusicURL(ctx context.Context, source string, musicInfo map[string]interface{}, quality string) (*MusicURLResult, error) {
	params := map[string]interface{}{
		"source":    source,
		"musicInfo": musicInfo,
		"quality":   quality,
	}

	resp, err := c.Call(ctx, "getMusicUrl", params)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("get music URL failed: %s", resp.Error.Message)
	}

	var result MusicURLResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse music URL result: %w", err)
	}

	return &result, nil
}

// GetMusicURLBatch 批量获取播放 URL（多源聚合）
func (c *LXClient) GetMusicURLBatch(ctx context.Context, songID, songName, singer string, duration int, sources []string, quality string) (*BatchURLResult, error) {
	params := map[string]interface{}{
		"songId":   songID,
		"songName": songName,
		"singer":   singer,
		"duration": duration,
		"sources":  sources,
		"quality":  quality,
	}

	resp, err := c.Call(ctx, "getMusicUrlBatch", params)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("batch get music URL failed: %s", resp.Error.Message)
	}

	var result BatchURLResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse batch URL result: %w", err)
	}

	return &result, nil
}

// GetLyric 获取歌词
func (c *LXClient) GetLyric(ctx context.Context, source string, musicInfo map[string]interface{}) (*LyricResult, error) {
	params := map[string]interface{}{
		"source":    source,
		"musicInfo": musicInfo,
	}

	resp, err := c.Call(ctx, "getLyric", params)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("get lyric failed: %s", resp.Error.Message)
	}

	var result LyricResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse lyric result: %w", err)
	}

	return &result, nil
}

// GetPic 获取封面图片
func (c *LXClient) GetPic(ctx context.Context, source string, musicInfo map[string]interface{}) (*PicResult, error) {
	params := map[string]interface{}{
		"source":    source,
		"musicInfo": musicInfo,
	}

	resp, err := c.Call(ctx, "getPic", params)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("get pic failed: %s", resp.Error.Message)
	}

	var result PicResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse pic result: %w", err)
	}

	return &result, nil
}

// HealthCheck 健康检查
func (c *LXClient) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	resp, err := c.Call(ctx, "health", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("health check failed: %s", resp.Error.Message)
	}

	var status HealthStatus
	if err := json.Unmarshal(resp.Data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse health status: %w", err)
	}

	return &status, nil
}

// SetRequestTimeout 设置请求超时时间
func (c *LXClient) SetRequestTimeout(timeout time.Duration) {
	c.requestTimeout = timeout
}

// ===== 数据结构定义 =====

// SearchResult 搜索结果
type SearchResult struct {
	Songs []LXSong `json:"songs"`
	Total int      `json:"total"`
	Page  int      `json:"page"`
}

// LXSong LX Music 歌曲格式
type LXSong struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Singer   string `json:"singer"`
	Source   string `json:"source"`
	Interval string `json:"interval"` // "03:45"
	Meta     struct {
		SongID    string `json:"songId"`
		AlbumName string `json:"albumName"`
		PicURL    string `json:"picUrl"`
	} `json:"meta"`
}

// MusicURLResult 播放 URL 结果
type MusicURLResult struct {
	URL     string `json:"url"`
	Quality string `json:"quality"`
	Size    int64  `json:"size"`
}

// BatchURLResult 批量 URL 结果
type BatchURLResult struct {
	URLs []SongURLOption `json:"urls"`
}

// SongURLOption URL 选项
type SongURLOption struct {
	URL      string `json:"url"`
	Source   string `json:"source"`
	Quality  string `json:"quality"`
	Priority int    `json:"priority"`
}

// LyricResult 歌词结果
type LyricResult struct {
	Lyric  string `json:"lyric"`
	TLyric string `json:"tlyric"`
	RLyric string `json:"rlyric"`
	LXLyric string `json:"lxlyric"`
}

// PicResult 封面图片结果
type PicResult struct {
	URL string `json:"url"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status   string       `json:"status"`
	Uptime   float64      `json:"uptime"`
	Requests int64        `json:"requests"`
	Errors   int64        `json:"errors"`
	Sources  []SourceInfo `json:"sources"`
}

// SourceInfo 音源信息
type SourceInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Available bool `json:"available"`
}

// Ensure LXClient implements io.Closer
var _ io.Closer = (*LXClient)(nil)

// Close 关闭客户端
func (c *LXClient) Close() error {
	// 客户端本身不需要关闭，由 ProcessManager 管理进程生命周期
	return nil
}
