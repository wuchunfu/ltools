package processmanager

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

// =============================================================================
// 常量定义
// =============================================================================

const (
	PluginID      = "processmanager.builtin"
	PluginName    = "进程管理器"
	PluginVersion = "1.0.0"

	// 性能优化超时设置
	cpuTimeout        = 5 * time.Millisecond  // CPU 获取专用（更严格）
	processTimeout    = 150 * time.Millisecond // 整个进程信息获取超时
	maxConcurrentProc = 15                     // 最大并发进程查询数
	refreshInterval   = 10 * time.Second       // 刷新间隔
	operationTimeout  = 5 * time.Second        // 操作超时
)

// =============================================================================
// 类型定义
// =============================================================================

// ProcessInfo 包含进程的详细信息
type ProcessInfo struct {
	PID            int     `json:"pid"`
	Name           string  `json:"name"`
	ExecutablePath string  `json:"executablePath"`
	CmdLine        string  `json:"cmdLine"`
	CPUPercent     float64 `json:"cpuPercent"`
	MemoryPercent  float32 `json:"memoryPercent"`
	MemoryBytes    uint64  `json:"memoryBytes"`
	MemoryMB       float64 `json:"memoryMb"`
	Status         string  `json:"status"`
	Username       string  `json:"username"`
	CreateTime     int64   `json:"createTime"`
	NumThreads     int     `json:"numThreads"`
	NumFDs         int     `json:"numFds"`
	IsSystem       bool    `json:"isSystem"`
	// 性能优化字段
	PartialData bool   `json:"partialData"`  // 是否为降级数据
	ErrorReason string `json:"errorReason"`  // 错误原因
}

// ProcessListOptions 进程列表查询选项
type ProcessListOptions struct {
	SearchTerm string `json:"searchTerm"`
	SortBy     string `json:"sortBy"` // "pid", "name", "cpu", "memory"
	SortDesc   bool   `json:"sortDesc"`
	ShowSystem bool   `json:"showSystem"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// ProcessUpdateEvent 进程更新事件
type ProcessUpdateEvent struct {
	Type      string         `json:"type"` // "added", "updated", "removed", "full"
	Added     []*ProcessInfo `json:"added"`
	Updated   []*ProcessInfo `json:"updated"`
	Removed   []int          `json:"removed"`
	FullList  []*ProcessInfo `json:"fullList"`
	Timestamp int64          `json:"timestamp"`
}

// processBlacklist 进程黑名单（用于跳过问题进程）
type processBlacklist struct {
	mu          sync.RWMutex
	skipCPU     map[int32]string // PID -> 跳过CPU查询的原因
	skipMemory  map[int32]string // PID -> 跳过内存查询的原因
	skipAll     map[int32]string // PID -> 完全跳过的原因
	knownZombie map[string]bool  // 已知的僵尸进程名称
}

// ProcessManagerPlugin 进程管理器插件
type ProcessManagerPlugin struct {
	*plugins.BasePlugin
	app             *application.App
	processCache    map[int]*ProcessInfo
	lastSnapshot    map[int]*ProcessInfo
	cacheMutex      sync.RWMutex
	stopChan        chan struct{}
	refreshControl  chan struct{}
	systemUsernames map[string]bool
	systemPaths     []string
	viewActive      bool
	blacklist       *processBlacklist
}

// =============================================================================
// 构造函数
// =============================================================================

// NewProcessManagerPlugin 创建进程管理器插件实例
func NewProcessManagerPlugin() *ProcessManagerPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "查看和管理系统运行中的进程",
		Icon:        "process",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Permissions: []plugins.Permission{plugins.PermissionProcess},
		Keywords:    []string{"进程", "管理", "任务", "process", "task", "manager"},
	}

	return &ProcessManagerPlugin{
		BasePlugin:      plugins.NewBasePlugin(metadata),
		processCache:    make(map[int]*ProcessInfo),
		lastSnapshot:    make(map[int]*ProcessInfo),
		stopChan:        make(chan struct{}),
		systemUsernames: make(map[string]bool),
		systemPaths:     getSystemProcessPaths(),
		blacklist:       newBlacklist(),
	}
}

// newBlacklist 创建黑名单实例
func newBlacklist() *processBlacklist {
	return &processBlacklist{
		skipCPU:    make(map[int32]string),
		skipMemory: make(map[int32]string),
		skipAll:    make(map[int32]string),
		knownZombie: map[string]bool{
			"(zombie)": true,
			"defunct":  true,
		},
	}
}

// =============================================================================
// 生命周期方法
// =============================================================================

// Init 初始化插件
func (p *ProcessManagerPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	p.detectSystemUsernames()
	return nil
}

// ServiceStartup 服务启动
func (p *ProcessManagerPlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

// ServiceShutdown 服务关闭
func (p *ProcessManagerPlugin) ServiceShutdown(app *application.App) error {
	close(p.stopChan)
	return p.BasePlugin.ServiceShutdown(app)
}

// OnViewEnter 进入视图时启动后台刷新
func (p *ProcessManagerPlugin) OnViewEnter(app *application.App) error {
	p.viewActive = true
	p.refreshControl = make(chan struct{})

	// 立即执行一次刷新
	go p.refreshProcesses()

	// 启动后台定时刷新
	go p.refreshPeriodically()

	return nil
}

// OnViewLeave 离开视图时停止后台刷新
func (p *ProcessManagerPlugin) OnViewLeave(app *application.App) error {
	p.viewActive = false
	if p.refreshControl != nil {
		close(p.refreshControl)
		p.refreshControl = nil
	}
	return nil
}

// =============================================================================
// 公共服务方法（前端 API）
// =============================================================================

// GetProcesses 获取进程列表
func (p *ProcessManagerPlugin) GetProcesses(options ProcessListOptions) ([]*ProcessInfo, int, error) {
	p.cacheMutex.RLock()
	cachedProcesses := make([]*ProcessInfo, 0, len(p.processCache))
	for _, proc := range p.processCache {
		cachedProcesses = append(cachedProcesses, proc)
	}
	p.cacheMutex.RUnlock()

	// 应用过滤
	cachedProcesses = p.filterProcesses(cachedProcesses, options)

	// 排序
	p.sortProcesses(cachedProcesses, options.SortBy, options.SortDesc)

	total := len(cachedProcesses)

	// 分页
	if options.Limit > 0 {
		cachedProcesses = p.paginate(cachedProcesses, options.Offset, options.Limit)
	}

	return cachedProcesses, total, nil
}

// GetProcessDetail 获取单个进程详情
func (p *ProcessManagerPlugin) GetProcessDetail(pid int) (*ProcessInfo, error) {
	ctx := context.Background()
	_, err := process.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return nil, fmt.Errorf("进程不存在: %w", err)
	}

	info := p.getProcessInfo(ctx, int32(pid))
	if info == nil {
		return nil, fmt.Errorf("获取进程信息失败")
	}
	return info, nil
}

// KillProcess 终止进程（先尝试优雅终止，失败则强制）
func (p *ProcessManagerPlugin) KillProcess(pid int) error {
	ctx := context.Background()
	proc, err := process.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return fmt.Errorf("进程不存在: %w", err)
	}

	if err := proc.Terminate(); err != nil {
		if killErr := proc.Kill(); killErr != nil {
			return fmt.Errorf("终止进程失败: %w", killErr)
		}
	}

	p.emitKill(pid)
	return nil
}

// ForceKillProcess 强制终止进程
func (p *ProcessManagerPlugin) ForceKillProcess(pid int) error {
	ctx := context.Background()
	proc, err := process.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return fmt.Errorf("进程不存在: %w", err)
	}

	if err := proc.Kill(); err != nil {
		return fmt.Errorf("强制终止进程失败: %w", err)
	}

	p.emitKill(pid)
	return nil
}

// ForceRefresh 强制刷新进程列表
func (p *ProcessManagerPlugin) ForceRefresh() {
	p.refreshProcesses()
}

// GetSystemInfo 获取系统信息
func (p *ProcessManagerPlugin) GetSystemInfo() map[string]interface{} {
	result := make(map[string]interface{})

	if hostInfo, err := host.Info(); err == nil {
		result["hostname"] = hostInfo.Hostname
		result["uptime"] = hostInfo.Uptime
		result["bootTime"] = hostInfo.BootTime
		result["procCount"] = hostInfo.Procs
	}

	result["os"] = runtime.GOOS
	result["arch"] = runtime.GOARCH

	return result
}

// =============================================================================
// 内部方法 - 刷新逻辑
// =============================================================================

// refreshPeriodically 定期刷新进程列表
func (p *ProcessManagerPlugin) refreshPeriodically() {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	// 延迟首次刷新，避免启动时卡顿
	time.Sleep(3 * time.Second)

	for {
		select {
		case <-ticker.C:
			if p.Enabled() && p.viewActive {
				p.refreshProcesses()
			}
		case <-p.stopChan:
			return
		case <-p.refreshControl:
			return
		}
	}
}

// refreshProcesses 刷新进程缓存并发送更新事件
func (p *ProcessManagerPlugin) refreshProcesses() {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	processes, err := p.getProcessesInternal(ctx, ProcessListOptions{})
	if err != nil {
		p.emitError(fmt.Sprintf("刷新进程列表失败: %v", err))
		return
	}

	p.cacheMutex.Lock()
	defer p.cacheMutex.Unlock()

	// 计算差异并发送事件
	event := p.calculateDiff(processes)

	// 更新缓存
	p.processCache = make(map[int]*ProcessInfo)
	for _, proc := range processes {
		p.processCache[proc.PID] = proc
	}

	if event != nil {
		p.emitUpdate(event)
	}
}

// calculateDiff 计算进程变化差异
func (p *ProcessManagerPlugin) calculateDiff(newProcesses []*ProcessInfo) *ProcessUpdateEvent {
	event := &ProcessUpdateEvent{
		Type:      "incremental",
		Added:     make([]*ProcessInfo, 0),
		Updated:   make([]*ProcessInfo, 0),
		Removed:   make([]int, 0),
		Timestamp: time.Now().UnixMilli(),
	}

	newProcessMap := make(map[int]*ProcessInfo)
	for _, proc := range newProcesses {
		newProcessMap[proc.PID] = proc
	}

	// 查找新增和更新的进程
	for pid, newProc := range newProcessMap {
		oldProc, exists := p.lastSnapshot[pid]
		if !exists {
			event.Added = append(event.Added, newProc)
		} else if p.hasProcessChanged(oldProc, newProc) {
			event.Updated = append(event.Updated, newProc)
		}
	}

	// 查找移除的进程
	for pid := range p.lastSnapshot {
		if _, exists := newProcessMap[pid]; !exists {
			event.Removed = append(event.Removed, pid)
		}
	}

	// 首次更新或变化过多时发送完整列表
	if len(p.lastSnapshot) == 0 || len(event.Added) > 50 || len(event.Removed) > 50 {
		event.Type = "full"
		event.FullList = newProcesses
		event.Added = nil
		event.Updated = nil
		event.Removed = nil
	}

	// 更新快照
	p.lastSnapshot = make(map[int]*ProcessInfo)
	for k, v := range newProcessMap {
		p.lastSnapshot[k] = v
	}

	return event
}

// hasProcessChanged 检查进程是否有显著变化
func (p *ProcessManagerPlugin) hasProcessChanged(old, new *ProcessInfo) bool {
	return abs(old.CPUPercent-new.CPUPercent) > 1.0 ||
		abs(float64(old.MemoryPercent)-float64(new.MemoryPercent)) > 0.5 ||
		old.Status != new.Status
}

// =============================================================================
// 内部方法 - 进程信息获取
// =============================================================================

// getProcessesInternal 内部获取进程列表实现
func (p *ProcessManagerPlugin) getProcessesInternal(ctx context.Context, options ProcessListOptions) ([]*ProcessInfo, error) {
	pids, err := process.Pids()
	if err != nil {
		return nil, fmt.Errorf("获取进程列表失败: %w", err)
	}

	sem := make(chan struct{}, maxConcurrentProc)
	var wg sync.WaitGroup
	var mu sync.Mutex
	processes := make([]*ProcessInfo, 0, len(pids))

	for _, pid := range pids {
		wg.Add(1)
		go func(pid int32) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			procInfo := p.getProcessInfo(ctx, pid)
			if procInfo != nil {
				mu.Lock()
				processes = append(processes, procInfo)
				mu.Unlock()
			}
		}(pid)
	}

	wg.Wait()
	p.sortProcesses(processes, options.SortBy, options.SortDesc)
	return processes, nil
}

// getProcessInfo 获取单个进程信息（带超时控制和降级策略）
func (p *ProcessManagerPlugin) getProcessInfo(ctx context.Context, pid int32) *ProcessInfo {
	// 检查黑名单
	p.blacklist.mu.RLock()
	skipCPU := p.blacklist.skipCPU[pid]
	skipMemory := p.blacklist.skipMemory[pid]
	p.blacklist.mu.RUnlock()

	// 创建带超时的上下文
	procCtx, cancel := context.WithTimeout(ctx, processTimeout)
	defer cancel()

	proc, err := process.NewProcessWithContext(procCtx, pid)
	if err != nil {
		return nil
	}

	info := &ProcessInfo{PID: int(pid)}

	// 获取进程名
	if name, err := proc.Name(); err == nil {
		info.Name = name
		// 检查僵尸进程
		p.blacklist.mu.RLock()
		isZombie := p.blacklist.knownZombie[strings.ToLower(name)]
		p.blacklist.mu.RUnlock()
		if isZombie {
			info.PartialData = true
			info.ErrorReason = "僵尸进程"
			return info
		}
	}

	// 获取状态
	if status, err := proc.Status(); err == nil && len(status) > 0 {
		info.Status = status[0]
		if info.Status == "Z" {
			info.PartialData = true
			info.ErrorReason = "僵尸进程"
			p.blacklist.mu.Lock()
			p.blacklist.skipCPU[pid] = "僵尸进程"
			p.blacklist.skipMemory[pid] = "僵尸进程"
			p.blacklist.mu.Unlock()
			return info
		}
	}

	// 获取其他信息
	info.ExecutablePath, _ = proc.Exe()
	if cmdline, err := proc.CmdlineSlice(); err == nil {
		info.CmdLine = strings.Join(cmdline, " ")
	}

	// CPU（带黑名单检查）
	if skipCPU == "" {
		cpuCtx, cpuCancel := context.WithTimeout(context.Background(), cpuTimeout)
		if cpuPercent, err := proc.CPUPercentWithContext(cpuCtx); err == nil {
			info.CPUPercent = cpuPercent
		} else {
			p.blacklist.mu.Lock()
			p.blacklist.skipCPU[pid] = "CPU查询超时"
			p.blacklist.mu.Unlock()
		}
		cpuCancel()
	}

	// 内存（带黑名单检查）
	if skipMemory == "" {
		if memInfo, err := proc.MemoryInfo(); err == nil {
			info.MemoryBytes = memInfo.RSS
			info.MemoryMB = float64(memInfo.RSS) / 1024 / 1024
		}
		if memPercent, err := proc.MemoryPercent(); err == nil {
			info.MemoryPercent = memPercent
		}
	}

	// 其他属性
	if username, err := proc.Username(); err == nil {
		if idx := strings.Index(username, "\\"); idx != -1 {
			info.Username = username[idx+1:]
		} else {
			info.Username = username
		}
	}
	info.CreateTime, _ = proc.CreateTime()
	if numThreads, err := proc.NumThreads(); err == nil {
		info.NumThreads = int(numThreads)
	}
	if numFDs, err := proc.NumFDs(); err == nil {
		info.NumFDs = int(numFDs)
	}

	info.IsSystem = p.isSystemProcess(info)
	return info
}

// =============================================================================
// 辅助方法
// =============================================================================

// filterProcesses 过滤进程列表
func (p *ProcessManagerPlugin) filterProcesses(processes []*ProcessInfo, options ProcessListOptions) []*ProcessInfo {
	if options.SearchTerm == "" && options.ShowSystem {
		return processes
	}

	filtered := make([]*ProcessInfo, 0, len(processes))
	for _, proc := range processes {
		if options.SearchTerm != "" && !p.matchesSearch(proc, options.SearchTerm) {
			continue
		}
		if !options.ShowSystem && proc.IsSystem {
			continue
		}
		filtered = append(filtered, proc)
	}
	return filtered
}

// paginate 分页
func (p *ProcessManagerPlugin) paginate(processes []*ProcessInfo, offset, limit int) []*ProcessInfo {
	if offset >= len(processes) {
		return []*ProcessInfo{}
	}
	end := offset + limit
	if end > len(processes) {
		end = len(processes)
	}
	return processes[offset:end]
}

// matchesSearch 检查进程是否匹配搜索词
func (p *ProcessManagerPlugin) matchesSearch(info *ProcessInfo, term string) bool {
	term = strings.ToLower(term)
	searchStrings := []string{
		strings.ToLower(info.Name),
		strings.ToLower(info.ExecutablePath),
		strings.ToLower(info.CmdLine),
		fmt.Sprintf("%d", info.PID),
	}
	for _, s := range searchStrings {
		if strings.Contains(s, term) {
			return true
		}
	}
	return false
}

// sortProcesses 排序进程列表
func (p *ProcessManagerPlugin) sortProcesses(processes []*ProcessInfo, sortBy string, desc bool) {
	sort.Slice(processes, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "name":
			less = processes[i].Name < processes[j].Name
		case "cpu":
			less = processes[i].CPUPercent < processes[j].CPUPercent
		case "memory":
			less = processes[i].MemoryMB < processes[j].MemoryMB
		default:
			less = processes[i].PID < processes[j].PID
		}
		if desc {
			return !less
		}
		return less
	})
}

// isSystemProcess 判断是否为系统进程
func (p *ProcessManagerPlugin) isSystemProcess(info *ProcessInfo) bool {
	if p.systemUsernames[info.Username] {
		return true
	}
	for _, sysPath := range p.systemPaths {
		if strings.Contains(info.ExecutablePath, sysPath) {
			return true
		}
	}
	return false
}

// detectSystemUsernames 检测系统用户名
func (p *ProcessManagerPlugin) detectSystemUsernames() {
	systemUsers := []string{"root", "SYSTEM", "LOCAL SERVICE", "NETWORK SERVICE", "daemon", "nobody", "_spotlight", "_mbsetupuser"}
	for _, user := range systemUsers {
		p.systemUsernames[user] = true
	}
	if currentUser := getCurrentUsername(); currentUser != "" {
		p.systemUsernames[currentUser] = false
	}
}

// getSystemProcessPaths 获取系统进程路径
func getSystemProcessPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"/System/", "/Library/", "/usr/", "/bin/", "/sbin/", "/dev/"}
	case "linux":
		return []string{"/usr/", "/bin/", "/sbin/", "/lib/", "/opt/", "/dev/"}
	case "windows":
		return []string{"\\Windows\\", "\\Program Files\\", "\\Program Files (x86)\\", "\\ProgramData\\", "System32", "SysWOW64"}
	default:
		return []string{}
	}
}

// getCurrentUsername 获取当前用户名
func getCurrentUsername() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERNAME")
	}
	return os.Getenv("USER")
}

// abs 绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// =============================================================================
// 事件发射
// =============================================================================

func (p *ProcessManagerPlugin) emitUpdate(event *ProcessUpdateEvent) {
	if p.app != nil {
		p.app.Event.Emit("processmanager:updated", fmt.Sprintf("%d", event.Timestamp))
	}
}

func (p *ProcessManagerPlugin) emitKill(pid int) {
	if p.app != nil {
		p.app.Event.Emit("processmanager:killed", fmt.Sprintf("%d", pid))
	}
}

func (p *ProcessManagerPlugin) emitError(message string) {
	if p.app != nil {
		p.app.Event.Emit("processmanager:error", message)
	}
}
