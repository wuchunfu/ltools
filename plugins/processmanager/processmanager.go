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

const (
	PluginID      = "processmanager.builtin"
	PluginName    = "进程管理器"
	PluginVersion = "1.0.0"
)

// ProcessInfo contains information about a process
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
	NumFDs         int     `json:"numFds"` // File descriptors / handles
	IsSystem       bool    `json:"isSystem"`
}

// ProcessListOptions contains options for querying the process list
type ProcessListOptions struct {
	SearchTerm  string `json:"searchTerm"`
	SortBy      string `json:"sortBy"` // "pid", "name", "cpu", "memory"
	SortDesc    bool   `json:"sortDesc"`
	ShowSystem  bool   `json:"showSystem"`
	Limit       int    `json:"limit"`    // 0 means no limit
	Offset      int    `json:"offset"`   // For pagination
}

// ProcessUpdateEvent represents a process list update event
type ProcessUpdateEvent struct {
	Type     string         `json:"type"`     // "added", "updated", "removed", "full"
	Added    []*ProcessInfo `json:"added"`
	Updated  []*ProcessInfo `json:"updated"`
	Removed  []int          `json:"removed"`  // PIDs of removed processes
	FullList []*ProcessInfo `json:"fullList"`
	Timestamp int64         `json:"timestamp"`
}

// ProcessManagerPlugin provides process management functionality
type ProcessManagerPlugin struct {
	*plugins.BasePlugin
	app              *application.App
	processCache     map[int]*ProcessInfo
	lastSnapshot     map[int]*ProcessInfo
	cacheMutex       sync.RWMutex
	stopChan         chan struct{}
	refreshInterval  time.Duration
	systemUsernames  map[string]bool
	systemPaths      []string
	viewActive       bool          // 是否在前台显示中
	refreshControl   chan struct{} // 用于控制刷新 goroutine 的生命周期
}

// NewProcessManagerPlugin creates a new process manager plugin
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
		Permissions: []plugins.Permission{
			plugins.PermissionProcess,
		},
		Keywords: []string{"进程", "管理", "任务", "process", "task", "manager"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &ProcessManagerPlugin{
		BasePlugin:      base,
		processCache:    make(map[int]*ProcessInfo),
		lastSnapshot:    make(map[int]*ProcessInfo),
		stopChan:        make(chan struct{}),
		refreshInterval: 10 * time.Second, // 增加到10秒，减少频繁刷新
		systemUsernames: make(map[string]bool),
		systemPaths:     getSystemProcessPaths(),
	}
}

// getSystemProcessPaths returns paths that indicate system processes
func getSystemProcessPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/System/",
			"/Library/",
			"/usr/",
			"/bin/",
			"/sbin/",
			"/dev/",
		}
	case "linux":
		return []string{
			"/usr/",
			"/bin/",
			"/sbin/",
			"/lib/",
			"/opt/",
			"/dev/",
		}
	case "windows":
		return []string{
			"\\Windows\\",
			"\\Program Files\\",
			"\\Program Files (x86)\\",
			"\\ProgramData\\",
			"System32",
			"SysWOW64",
		}
	default:
		return []string{}
	}
}

// Init initializes the plugin
func (p *ProcessManagerPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app

	// Detect system usernames
	p.detectSystemUsernames()

	return nil
}

// detectSystemUsernames detects which usernames are considered system users
func (p *ProcessManagerPlugin) detectSystemUsernames() {
	// Add common system usernames
	systemUsers := []string{
		"root",
		"SYSTEM",
		"LOCAL SERVICE",
		"NETWORK SERVICE",
		"daemon",
		"nobody",
		"_spotlight",
		"_mbsetupuser",
	}

	// Get current username
	if currentUser := getCurrentUsername(); currentUser != "" {
		p.systemUsernames[currentUser] = false // Not a system user
	}

	for _, user := range systemUsers {
		p.systemUsernames[user] = true
	}
}

// getCurrentUsername returns the current username
func getCurrentUsername() string {
	if runtime.GOOS == "windows" {
		return getEnv("USERNAME")
	}
	return getEnv("USER")
}

// getEnv gets an environment variable
func getEnv(key string) string {
	return os.Getenv(key)
}

// ServiceStartup is called when the application starts
func (p *ProcessManagerPlugin) ServiceStartup(app *application.App) error {
	if err := p.BasePlugin.ServiceStartup(app); err != nil {
		return err
	}

	// 不再自动启动后台刷新，等待用户打开插件时再启动
	// go p.refreshPeriodically() // <-- 移除这行

	return nil
}

// ServiceShutdown is called when the application shuts down
func (p *ProcessManagerPlugin) ServiceShutdown(app *application.App) error {
	close(p.stopChan)
	return p.BasePlugin.ServiceShutdown(app)
}

// refreshPeriodically refreshes process list periodically and sends incremental updates
func (p *ProcessManagerPlugin) refreshPeriodically() {
	ticker := time.NewTicker(p.refreshInterval)
	defer ticker.Stop()

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
			return // 收到离开视图信号
		}
	}
}

// refreshProcesses refreshes the process cache and sends incremental updates
func (p *ProcessManagerPlugin) refreshProcesses() {
	// 添加超时控制，避免操作挂起
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	processes, err := p.getProcessesInternal(ctx, ProcessListOptions{})
	if err != nil {
		p.emitError(fmt.Sprintf("Failed to refresh processes: %v", err))
		return
	}

	p.cacheMutex.Lock()
	defer p.cacheMutex.Unlock()

	// Calculate differences
	event := p.calculateDiff(processes)

	// Update cache
	p.processCache = make(map[int]*ProcessInfo)
	for _, proc := range processes {
		p.processCache[proc.PID] = proc
	}

	// Emit update event if there are changes
	if event != nil {
		p.emitUpdate(event)
	}
}

// calculateDiff calculates the difference between the last snapshot and new processes
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

	// Find added and updated processes
	for pid, newProc := range newProcessMap {
		oldProc, exists := p.lastSnapshot[pid]
		if !exists {
			event.Added = append(event.Added, newProc)
		} else if p.hasProcessChanged(oldProc, newProc) {
			event.Updated = append(event.Updated, newProc)
		}
	}

	// Find removed processes
	for pid := range p.lastSnapshot {
		if _, exists := newProcessMap[pid]; !exists {
			event.Removed = append(event.Removed, pid)
		}
	}

	// If this is the first update or too many changes, send full list
	if len(p.lastSnapshot) == 0 || len(event.Added) > 50 || len(event.Removed) > 50 {
		event.Type = "full"
		event.FullList = newProcesses
		event.Added = nil
		event.Updated = nil
		event.Removed = nil
	}

	// Update last snapshot
	p.lastSnapshot = make(map[int]*ProcessInfo)
	for k, v := range newProcessMap {
		p.lastSnapshot[k] = v
	}

	return event
}

// hasProcessChanged checks if a process has significantly changed
func (p *ProcessManagerPlugin) hasProcessChanged(old, new *ProcessInfo) bool {
	// Check if CPU or memory changed significantly
	cpuChanged := abs(old.CPUPercent-new.CPUPercent) > 1.0
	memoryChanged := abs(float64(old.MemoryPercent)-float64(new.MemoryPercent)) > 0.5
	statusChanged := old.Status != new.Status

	return cpuChanged || memoryChanged || statusChanged
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetProcesses returns the list of processes based on the given options
func (p *ProcessManagerPlugin) GetProcesses(options ProcessListOptions) ([]*ProcessInfo, int, error) {
	p.cacheMutex.RLock()
	cachedProcesses := make([]*ProcessInfo, 0, len(p.processCache))
	for _, proc := range p.processCache {
		cachedProcesses = append(cachedProcesses, proc)
	}
	p.cacheMutex.RUnlock()

	// 应用搜索过滤
	if options.SearchTerm != "" || !options.ShowSystem {
		filtered := make([]*ProcessInfo, 0, len(cachedProcesses))
		for _, proc := range cachedProcesses {
			if options.SearchTerm != "" && !p.matchesSearch(proc, options.SearchTerm) {
				continue
			}
			if !options.ShowSystem && proc.IsSystem {
				continue
			}
			filtered = append(filtered, proc)
		}
		cachedProcesses = filtered
	}

	// 排序
	p.sortProcesses(cachedProcesses, options.SortBy, options.SortDesc)

	total := len(cachedProcesses)

	// Apply pagination
	if options.Limit > 0 {
		start := options.Offset
		if start >= len(cachedProcesses) {
			return []*ProcessInfo{}, total, nil
		}
		end := start + options.Limit
		if end > len(cachedProcesses) {
			end = len(cachedProcesses)
		}
		cachedProcesses = cachedProcesses[start:end]
	}

	return cachedProcesses, total, nil
}

// getProcessesInternal is the internal implementation for getting processes
func (p *ProcessManagerPlugin) getProcessesInternal(ctx context.Context, options ProcessListOptions) ([]*ProcessInfo, error) {
	// Get all PIDs
	pids, err := process.Pids()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %w", err)
	}

	// Use semaphore to limit concurrent goroutines
	sem := make(chan struct{}, 10) // 减少到10个并发，降低系统负载
	var wg sync.WaitGroup
	var mu sync.Mutex
	processes := make([]*ProcessInfo, 0, len(pids))

	for _, pid := range pids {
		wg.Add(1)
		go func(pid int32) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			procInfo := p.getProcessInfo(ctx, pid)
			if procInfo != nil {
				// Apply filters
				if options.SearchTerm != "" && !p.matchesSearch(procInfo, options.SearchTerm) {
					return
				}
				if !options.ShowSystem && procInfo.IsSystem {
					return
				}

				mu.Lock()
				processes = append(processes, procInfo)
				mu.Unlock()
			}
		}(pid)
	}

	wg.Wait()

	// Sort processes
	p.sortProcesses(processes, options.SortBy, options.SortDesc)

	return processes, nil
}

// getProcessInfo gets information about a single process
func (p *ProcessManagerPlugin) getProcessInfo(ctx context.Context, pid int32) *ProcessInfo {
	proc, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		// Process may have exited or we don't have permission
		return nil
	}

	info := &ProcessInfo{PID: int(pid)}

	// Get name
	if name, err := proc.Name(); err == nil {
		info.Name = name
	}

	// Get executable path
	if exe, err := proc.Exe(); err == nil {
		info.ExecutablePath = exe
	}

	// Get command line
	if cmdline, err := proc.CmdlineSlice(); err == nil && len(cmdline) > 0 {
		info.CmdLine = strings.Join(cmdline, " ")
	}

	// Get CPU percent - 使用非常短的超时时间避免阻塞
	// 跳过CPU使用率获取以提高性能，或者使用更短的时间
	cpuCtx, cpuCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cpuCancel()
	if cpuPercent, err := proc.CPUPercentWithContext(cpuCtx); err == nil {
		info.CPUPercent = cpuPercent
	}

	// Get memory info
	if memInfo, err := proc.MemoryInfo(); err == nil {
		info.MemoryBytes = memInfo.RSS
		info.MemoryMB = float64(memInfo.RSS) / 1024 / 1024
	}

	if memPercent, err := proc.MemoryPercent(); err == nil {
		info.MemoryPercent = memPercent
	}

	// Get status
	if status, err := proc.Status(); err == nil && len(status) > 0 {
		info.Status = status[0]
	}

	// Get username
	if username, err := proc.Username(); err == nil {
		// Extract just the username part (before the domain)
		if idx := strings.Index(username, "\\"); idx != -1 {
			info.Username = username[idx+1:]
		} else {
			info.Username = username
		}
	}

	// Get create time
	if create_time, err := proc.CreateTime(); err == nil {
		info.CreateTime = create_time
	}

	// Get num threads
	if numThreads, err := proc.NumThreads(); err == nil {
		info.NumThreads = int(numThreads)
	}

	// Get num file descriptors/handles
	if numFDs, err := proc.NumFDs(); err == nil {
		info.NumFDs = int(numFDs)
	}

	// Determine if it's a system process
	info.IsSystem = p.isSystemProcess(info)

	return info
}

// isSystemProcess determines if a process is a system process
func (p *ProcessManagerPlugin) isSystemProcess(info *ProcessInfo) bool {
	// Check by username
	if p.systemUsernames[info.Username] {
		return true
	}

	// Check by path
	for _, sysPath := range p.systemPaths {
		if strings.Contains(info.ExecutablePath, sysPath) {
			return true
		}
	}

	return false
}

// matchesSearch checks if a process matches the search term
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

// sortProcesses sorts the process list
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
		case "pid":
			fallthrough
		default:
			less = processes[i].PID < processes[j].PID
		}

		if desc {
			return !less
		}
		return less
	})
}

// GetProcessDetail returns detailed information about a single process
func (p *ProcessManagerPlugin) GetProcessDetail(pid int) (*ProcessInfo, error) {
	ctx := context.Background()
	_, err := process.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return nil, fmt.Errorf("process not found: %w", err)
	}

	info := p.getProcessInfo(ctx, int32(pid))
	if info == nil {
		return nil, fmt.Errorf("failed to get process info")
	}

	return info, nil
}

// KillProcess terminates a process
func (p *ProcessManagerPlugin) KillProcess(pid int) error {
	ctx := context.Background()
	proc, err := process.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	// Try graceful termination first
	if err := proc.Terminate(); err != nil {
		// If that fails, try force kill
		if killErr := proc.Kill(); killErr != nil {
			return fmt.Errorf("failed to kill process: %w", killErr)
		}
	}

	p.emitKill(pid)
	return nil
}

// ForceKillProcess forcefully terminates a process using SIGKILL
func (p *ProcessManagerPlugin) ForceKillProcess(pid int) error {
	ctx := context.Background()
	proc, err := process.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	// Force kill using SIGKILL
	if err := proc.Kill(); err != nil {
		return fmt.Errorf("failed to force kill process: %w", err)
	}

	p.emitKill(pid)
	return nil
}

// ForceRefresh forces an immediate refresh of the process list
func (p *ProcessManagerPlugin) ForceRefresh() {
	p.refreshProcesses()
}

// GetSystemInfo returns system information for process management
func (p *ProcessManagerPlugin) GetSystemInfo() map[string]interface{} {
	result := make(map[string]interface{})

	// Get host info
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

// Event emission helpers
func (p *ProcessManagerPlugin) emitUpdate(event *ProcessUpdateEvent) {
	if p.app != nil {
		// For now, just emit a simple timestamp
		// The full event data can be sent via JSON serialization if needed
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

// OnViewEnter starts background refresh when user opens the plugin
func (p *ProcessManagerPlugin) OnViewEnter(app *application.App) error {
	p.viewActive = true
	p.refreshControl = make(chan struct{})

	// 立即执行一次刷新
	go p.refreshProcesses()

	// 启动后台定时刷新
	go p.refreshPeriodically()

	return nil
}

// OnViewLeave stops background refresh when user leaves the plugin
func (p *ProcessManagerPlugin) OnViewLeave(app *application.App) error {
	p.viewActive = false
	if p.refreshControl != nil {
		close(p.refreshControl)
		p.refreshControl = nil
	}
	return nil
}
