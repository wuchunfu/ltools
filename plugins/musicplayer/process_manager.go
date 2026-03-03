package musicplayer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// ProcessManager Node.js 进程管理器
type ProcessManager struct {
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	stdout       io.Reader
	stderr       io.Reader
	processMutex sync.RWMutex

	// 进程状态
	isRunning bool
	startTime time.Time

	// 配置
	nodePath   string // Node.js 可执行文件路径
	serviceDir string // lx-music-service 目录路径

	// 健康检查
	healthCheckInterval time.Duration
	lastHealthCheck     time.Time
	healthCheckErrors   int
}

// ProcessManagerConfig 进程管理器配置
type ProcessManagerConfig struct {
	NodePath           string // Node.js 路径（如果为空，使用 "node"）
	ServiceDir         string // lx-music-service 目录路径
	HealthCheckInterval time.Duration
}

// NewProcessManager 创建进程管理器
func NewProcessManager(config *ProcessManagerConfig) (*ProcessManager, error) {
	// 默认配置
	if config.NodePath == "" {
		config.NodePath = "node"
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 30 * time.Second
	}

	// 如果未指定服务目录，使用默认路径
	if config.ServiceDir == "" {
		// 获取当前可执行文件目录
		execPath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable path: %w", err)
		}

		// 推算 lx-music-service 目录路径
		// 假设结构为: <app>/lx-music-service/
		appDir := filepath.Dir(execPath)
		config.ServiceDir = filepath.Join(appDir, "lx-music-service")

		// 开发环境：可能在项目根目录
		if _, err := os.Stat(config.ServiceDir); os.IsNotExist(err) {
			// 尝试从当前工作目录查找
			wd, _ := os.Getwd()
			config.ServiceDir = filepath.Join(wd, "lx-music-service")
		}
	}

	// 检查服务目录是否存在
	if _, err := os.Stat(config.ServiceDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("lx-music-service directory not found: %s", config.ServiceDir)
	}

	// 检查 Node.js 是否安装
	if err := checkNodeInstalled(config.NodePath); err != nil {
		return nil, fmt.Errorf("Node.js not found: %w. Please install Node.js >= 16.0.0", err)
	}

	return &ProcessManager{
		nodePath:            config.NodePath,
		serviceDir:          config.ServiceDir,
		healthCheckInterval: config.HealthCheckInterval,
	}, nil
}

// Start 启动进程
func (pm *ProcessManager) Start() error {
	pm.processMutex.Lock()
	defer pm.processMutex.Unlock()

	if pm.isRunning {
		return fmt.Errorf("process is already running")
	}

	// 构建启动命令
	serverPath := filepath.Join(pm.serviceDir, "server.js")
	pm.cmd = exec.Command(pm.nodePath, serverPath)
	pm.cmd.Dir = pm.serviceDir

	// 获取 stdin/stdout/stdout 管道
	stdin, err := pm.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	pm.stdin = stdin

	stdout, err := pm.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	pm.stdout = stdout

	stderr, err := pm.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	pm.stderr = stderr

	// 启动进程
	if err := pm.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	pm.isRunning = true
	pm.startTime = time.Now()

	// 启动 goroutine 读取 stderr（日志）
	go pm.readStderr()

	log.Printf("[ProcessManager] Process started (PID: %d)", pm.cmd.Process.Pid)

	// 等待一小段时间，确保服务器初始化完成
	time.Sleep(500 * time.Millisecond)

	return nil
}

// Stop 停止进程
func (pm *ProcessManager) Stop() error {
	pm.processMutex.Lock()
	defer pm.processMutex.Unlock()

	if !pm.isRunning {
		return nil
	}

	// 关闭 stdin
	if pm.stdin != nil {
		pm.stdin.Close()
	}

	// 发送 SIGTERM 信号
	if pm.cmd != nil && pm.cmd.Process != nil {
		if err := pm.cmd.Process.Signal(os.Interrupt); err != nil {
			log.Printf("[ProcessManager] Failed to send SIGTERM: %v", err)
		}

		// 等待进程退出（最多 5 秒）
		done := make(chan error, 1)
		go func() {
			done <- pm.cmd.Wait()
		}()

		select {
		case <-time.After(5 * time.Second):
			// 强制杀死进程
			log.Printf("[ProcessManager] Force killing process")
			pm.cmd.Process.Kill()
		case err := <-done:
			if err != nil {
				log.Printf("[ProcessManager] Process exited with error: %v", err)
			}
		}
	}

	pm.isRunning = false
	log.Printf("[ProcessManager] Process stopped")

	return nil
}

// Restart 重启进程
func (pm *ProcessManager) Restart() error {
	if err := pm.Stop(); err != nil {
		log.Printf("[ProcessManager] Failed to stop process during restart: %v", err)
	}

	time.Sleep(1 * time.Second)

	return pm.Start()
}

// IsRunning 检查进程是否运行中
func (pm *ProcessManager) IsRunning() bool {
	pm.processMutex.RLock()
	defer pm.processMutex.RUnlock()
	return pm.isRunning
}

// GetStdin 获取 stdin 管道
func (pm *ProcessManager) GetStdin() io.WriteCloser {
	pm.processMutex.RLock()
	defer pm.processMutex.RUnlock()
	return pm.stdin
}

// GetStdout 获取 stdout 管道
func (pm *ProcessManager) GetStdout() io.Reader {
	pm.processMutex.RLock()
	defer pm.processMutex.RUnlock()
	return pm.stdout
}

// readStderr 读取 stderr（日志输出）
func (pm *ProcessManager) readStderr() {
	scanner := bufio.NewScanner(pm.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[LXService] %s", line)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[ProcessManager] Error reading stderr: %v", err)
	}
}

// WaitForExit 等待进程退出
func (pm *ProcessManager) WaitForExit() error {
	if pm.cmd == nil {
		return fmt.Errorf("process not started")
	}

	return pm.cmd.Wait()
}

// GetUptime 获取进程运行时长
func (pm *ProcessManager) GetUptime() time.Duration {
	pm.processMutex.RLock()
	defer pm.processMutex.RUnlock()

	if !pm.isRunning {
		return 0
	}

	return time.Since(pm.startTime)
}

// checkNodeInstalled 检查 Node.js 是否安装
func checkNodeInstalled(nodePath string) error {
	cmd := exec.Command(nodePath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	log.Printf("[ProcessManager] Node.js version: %s", string(output))
	return nil
}

// HealthCheckWithContext 健康检查（带超时）
func (pm *ProcessManager) HealthCheckWithContext(ctx context.Context) error {
	// 这里只是简单检查进程是否运行
	// 实际的健康检查由 LXClient 实现
	if !pm.IsRunning() {
		return fmt.Errorf("process is not running")
	}

	return nil
}
