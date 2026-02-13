package main

import (
	"embed"
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"ltools/internal/plugins"
	"ltools/plugins/applauncher"
	"ltools/plugins/calculator"
	"ltools/plugins/clipboard"
	"ltools/plugins/datetime"
	"ltools/plugins/hosts"
	"ltools/plugins/jsoneditor"
	"ltools/plugins/password"
	"ltools/plugins/processmanager"
	"ltools/plugins/qrcode"
	"ltools/plugins/screenshot"
	"ltools/plugins/sysinfo"
	"ltools/plugins/tunnel"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var systrayIcon []byte

func init() {
	// Register custom events for the datetime plugin
	application.RegisterEvent[string]("datetime:current")
	application.RegisterEvent[string]("datetime:time")
	application.RegisterEvent[string]("datetime:date")
	application.RegisterEvent[string]("datetime:datetime")
	application.RegisterEvent[string]("datetime:weekday")
	application.RegisterEvent[int]("datetime:year")
	application.RegisterEvent[int]("datetime:month")
	application.RegisterEvent[int]("datetime:day")
	application.RegisterEvent[int]("datetime:hour")
	application.RegisterEvent[int]("datetime:minute")
	application.RegisterEvent[int]("datetime:second")

	// Register custom events for the calculator plugin
	application.RegisterEvent[string]("calculator:result")
	application.RegisterEvent[string]("calculator:error")
	application.RegisterEvent[string]("calculator:history")

	// Register custom events for the clipboard plugin
	application.RegisterEvent[string]("clipboard:new")
	application.RegisterEvent[string]("clipboard:cleared")
	application.RegisterEvent[string]("clipboard:deleted")
	application.RegisterEvent[string]("clipboard:count")
	application.RegisterEvent[string]("clipboard:heartbeat")
	application.RegisterEvent[string]("clipboard:permission:requested")

	// Register custom events for the system info plugin
	application.RegisterEvent[string]("sysinfo:updated")
	application.RegisterEvent[string]("sysinfo:cpu")
	application.RegisterEvent[string]("sysinfo:uptime")
	application.RegisterEvent[string]("sysinfo:gc")
	application.RegisterEvent[string]("sysinfo:maxprocs")

	// Register custom events for the shortcut service
	application.RegisterEvent[string]("shortcut:triggered")

	// Register event for permission requirement (sends map with title, message, platform)
	// We need to define the event type - using map[string]interface{} for flexibility
	// Note: Wails v3 events need a specific type, so we'll use string for the data
	application.RegisterEvent[string]("shortcut:permission-required")

	// Register custom events for the screenshot plugin
	application.RegisterEvent[string]("screenshot:started")
	application.RegisterEvent[string]("screenshot:captured")
	application.RegisterEvent[string]("screenshot:saved")
	application.RegisterEvent[string]("screenshot:copied")
	application.RegisterEvent[string]("screenshot:cancelled")
	application.RegisterEvent[string]("screenshot:error")

	// Register custom events for the JSON editor plugin
	application.RegisterEvent[string]("jsoneditor:formatted")
	application.RegisterEvent[string]("jsoneditor:error")
	application.RegisterEvent[bool]("jsoneditor:validated")

	// Register custom events for the process manager plugin
	application.RegisterEvent[string]("processmanager:updated")
	application.RegisterEvent[string]("processmanager:killed")
	application.RegisterEvent[string]("processmanager:error")

	// Register custom events for qrcode plugin
	application.RegisterEvent[string]("qrcode:generated")
	application.RegisterEvent[string]("qrcode:copied")
	application.RegisterEvent[string]("qrcode:saved")

	// Register custom events for the hosts plugin
	application.RegisterEvent[string]("hosts:scenario:created")
	application.RegisterEvent[string]("hosts:scenario:updated")
	application.RegisterEvent[string]("hosts:scenario:deleted")
	application.RegisterEvent[string]("hosts:scenario:switched")
	application.RegisterEvent[string]("hosts:backup:created")
	application.RegisterEvent[string]("hosts:backup:restored")
	application.RegisterEvent[string]("hosts:backup:deleted")
	application.RegisterEvent[string]("hosts:entry:added")
	application.RegisterEvent[string]("hosts:entry:updated")
	application.RegisterEvent[string]("hosts:entry:removed")
	application.RegisterEvent[string]("hosts:error")

	// Register custom events for tunnel plugin
	application.RegisterEvent[string]("tunnel:install:progress")
	application.RegisterEvent[string]("tunnel:created")
	application.RegisterEvent[string]("tunnel:updated")
	application.RegisterEvent[string]("tunnel:deleted")
	application.RegisterEvent[string]("tunnel:started")
	application.RegisterEvent[string]("tunnel:stopped")
	application.RegisterEvent[any]("tunnel:error")
	application.RegisterEvent[string]("tunnel:restarting")
	application.RegisterEvent[any]("tunnel:url")
	application.RegisterEvent[any]("tunnel:log")
	application.RegisterEvent[string]("tunnel:options:updated")

	// Register notification event for user notifications
	application.RegisterEvent[string]("notification:show")

	// Register custom events for the search window service
	application.RegisterEvent[string]("search:opened")
	application.RegisterEvent[string]("search:closed")
	application.RegisterEvent[[]*plugins.SearchResult]("search:results")
}

// main function serves as the application's entry point. It initializes the application, creates a window,
// registers plugins, and runs the application.
func main() {

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Services' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "ltools",
		Description: "A plugin-based desktop toolbox",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false, // 保持应用在窗口关闭后运行
		},
	})

	// 创建系统托盘
	systray := app.SystemTray.New()
	systray.SetIcon(systrayIcon)
	systray.SetLabel("LTools")

	// 保存主窗口引用，用于系统托盘菜单
	var mainWindow *application.WebviewWindow

	// 创建系统托盘菜单
	menu := app.NewMenu()
	menu.Add("显示窗口").OnClick(func(_ *application.Context) {
		if mainWindow != nil {
			mainWindow.Show()
		}
	})
	menu.Add("隐藏窗口").OnClick(func(_ *application.Context) {
		if mainWindow != nil {
			mainWindow.Hide()
		}
	})
	menu.AddSeparator()
	menu.Add("关于 LTools").OnClick(func(_ *application.Context) {
		log.Println("LTools - 插件式桌面工具箱 v1.0.0")
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(_ *application.Context) {
		app.Quit()
	})

	systray.SetMenu(menu)

	// Get user data directory for plugin storage
	userDataDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal("Failed to get user config dir:", err)
	}
	dataDir := filepath.Join(userDataDir, "ltools")

	// Create plugin manager
	pluginManager, err := plugins.NewManager(app, dataDir)
	if err != nil {
		log.Fatal("Failed to create plugin manager:", err)
	}

	// Create and register the datetime plugin
	datetimePlugin := datetime.NewDateTimePlugin()
	if err := pluginManager.Register(datetimePlugin); err != nil {
		log.Fatal("Failed to register datetime plugin:", err)
	}

	// Create and register password plugin
	passwordPlugin := password.NewPasswordPlugin()
	if err := pluginManager.Register(passwordPlugin); err != nil {
		log.Fatal("Failed to register password plugin:", err)
	}

	// Create and register the calculator plugin
	calculatorPlugin := calculator.NewCalculatorPlugin()
	if err := pluginManager.Register(calculatorPlugin); err != nil {
		log.Fatal("Failed to register calculator plugin:", err)
	}

	// Create and register the clipboard plugin
	clipboardPlugin := clipboard.NewClipboardPlugin()
	if err := pluginManager.Register(clipboardPlugin); err != nil {
		log.Fatal("Failed to register clipboard plugin:", err)
	}

	// Create and register the system info plugin
	sysInfoPlugin := sysinfo.NewSysInfoPlugin()
	if err := pluginManager.Register(sysInfoPlugin); err != nil {
		log.Fatal("Failed to register sysinfo plugin:", err)
	}

	// Create and register the JSON editor plugin
	jsonEditorPlugin := jsoneditor.NewJSONEditorPlugin()
	if err := pluginManager.Register(jsonEditorPlugin); err != nil {
		log.Fatal("Failed to register jsoneditor plugin:", err)
	}

	// Create and register the process manager plugin
	processManagerPlugin := processmanager.NewProcessManagerPlugin()
	if err := pluginManager.Register(processManagerPlugin); err != nil {
		log.Fatal("Failed to register processmanager plugin:", err)
	}

	// Create and register the screenshot plugin
	screenshotPlugin := screenshot.NewScreenshotPlugin()
	if err := pluginManager.Register(screenshotPlugin); err != nil {
		log.Fatal("Failed to register screenshot plugin:", err)
	}

	// Create and register the app launcher plugin
	appLauncherPlugin := applauncher.NewAppLauncherPlugin()
	if err := pluginManager.Register(appLauncherPlugin); err != nil {
		log.Fatal("Failed to register applauncher plugin:", err)
	}
	// 设置数据目录
	if err := appLauncherPlugin.SetDataDir(dataDir); err != nil {
		log.Printf("[Main] Failed to set data dir for app launcher: %v", err)
	}

	// Create and register qrcode plugin
	qrcodePlugin := qrcode.NewQrcodePlugin()
	if err := pluginManager.Register(qrcodePlugin); err != nil {
		log.Fatal("Failed to register qrcode plugin:", err)
	}

	// Create and register hosts plugin
	hostsPlugin := hosts.NewHostsPlugin()
	if err := pluginManager.Register(hostsPlugin); err != nil {
		log.Fatal("Failed to register hosts plugin:", err)
	}
	// Set data directory for hosts plugin
	if err := hostsPlugin.SetDataDir(dataDir); err != nil {
		log.Printf("[Main] Failed to set data dir for hosts plugin: %v", err)
	}

	// Create and register tunnel plugin
	tunnelPlugin := tunnel.NewTunnelPlugin()
	if err := pluginManager.Register(tunnelPlugin); err != nil {
		log.Fatal("Failed to register tunnel plugin:", err)
	}
	// Set data directory for tunnel plugin
	if err := tunnelPlugin.SetDataDir(dataDir); err != nil {
		log.Printf("[Main] Failed to set data dir for tunnel plugin: %v", err)
	}

	// Start all enabled plugins - this calls ServiceStartup() on each enabled plugin
	// This is crucial for plugins like clipboard that need to start background monitoring
	if err := pluginManager.StartupAll(); err != nil {
		log.Fatal("Failed to startup plugins:", err)
	}

	// Create plugin service to expose plugin management to frontend
	pluginService := plugins.NewPluginService(pluginManager, app)

	// Create datetime service to expose datetime functionality to frontend
	datetimeService := datetime.NewDateTimeService(datetimePlugin)

	// Create password service to expose password functionality to frontend
	passwordService := password.NewPasswordService(passwordPlugin, app)

	// Create calculator service to expose calculator functionality to frontend
	calculatorService := calculator.NewCalculatorService(calculatorPlugin, app)

	// Create clipboard service to expose clipboard functionality to frontend
	clipboardService := clipboard.NewClipboardService(clipboardPlugin, app)

	// Create system info service to expose system info functionality to frontend
	sysInfoService := sysinfo.NewSysInfoService(sysInfoPlugin, app)

	// Create JSON editor service to expose JSON editor functionality to frontend
	jsonEditorService := jsoneditor.NewJSONEditorService(jsonEditorPlugin, app)

	// Create process manager service to expose process manager functionality to frontend
	processManagerService := processmanager.NewProcessManagerService(processManagerPlugin, app)

	// Create screenshot window service (manages independent screenshot editor window)
	screenshotWindowService := screenshot.NewScreenshotWindowService(screenshotPlugin, app)

	// Create screenshot service to expose screenshot functionality to frontend (for plugin info)
	screenshotService := screenshot.NewScreenshotService(screenshotPlugin, app)
	// Set window service reference so frontend can trigger screenshots
	screenshotService.SetWindowService(screenshotWindowService)

	// Create app launcher service to expose app launcher functionality to frontend
	appLauncherService := applauncher.NewAppLauncherService(app, appLauncherPlugin)

	// Create qrcode service to expose qrcode functionality to frontend
	qrcodeService := qrcode.NewQrcodeService(qrcodePlugin, app)

	// Create hosts service to expose hosts functionality to frontend
	hostsService := hosts.NewHostsService(hostsPlugin, app, dataDir)

	// Create tunnel service to expose tunnel functionality to frontend
	tunnelService := tunnel.NewTunnelService(tunnelPlugin, app, dataDir)

	// Create shortcut service to expose keyboard shortcut management to frontend
	shortcutService, err := plugins.NewShortcutService(app, dataDir)
	if err != nil {
		log.Fatal("Failed to create shortcut service:", err)
	}

	// Create search window service for global search functionality
	searchWindowService := plugins.NewSearchWindowService(app, pluginService, shortcutService)
	// Set app launcher service for app search integration
	searchWindowService.SetAppLauncherService(appLauncherService)

	// Register services
	app.RegisterService(application.NewService(pluginService))
	app.RegisterService(application.NewService(datetimeService))
	app.RegisterService(application.NewService(passwordService))
	app.RegisterService(application.NewService(calculatorService))
	app.RegisterService(application.NewService(clipboardService))
	app.RegisterService(application.NewService(sysInfoService))
	app.RegisterService(application.NewService(jsonEditorService))
	app.RegisterService(application.NewService(processManagerService))
	app.RegisterService(application.NewService(screenshotService))
	app.RegisterService(application.NewService(appLauncherService))
	app.RegisterService(application.NewService(qrcodeService))
	app.RegisterService(application.NewService(hostsService))
	app.RegisterService(application.NewService(tunnelService))
	app.RegisterService(application.NewService(shortcutService))
	app.RegisterService(application.NewService(searchWindowService))

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	mainWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "LTools",
		Width:  1400,
		Height:  900,
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBar{
				HideTitle:          true,    // 隐藏标题文字
				AppearsTransparent: true,    // 标题栏透明
				FullSizeContent:    true,    // 内容延伸到标题栏区域（关键！）
			},
			Backdrop: application.MacBackdropTransparent,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// 监听主窗口关闭事件，隐藏窗口而非销毁（系统托盘应用模式）
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		mainWindow.Hide()
		e.Cancel() // 阻止窗口真正关闭
	})

	// Start the search window service
	searchWindowService.ServiceStartup(app)

	// Start the shortcut service AFTER window is created
	// This is crucial for Wails v3 KeyBinding API to work properly
	if err := shortcutService.ServiceStartup(app); err != nil {
		log.Printf("Failed to start shortcut service: %v", err)
	}

	// Set the main window reference in ShortcutService so it can focus the window
	shortcutService.SetMainWindow(mainWindow)

	// Set the main window reference in SearchWindowService so it can show the main window
	searchWindowService.SetMainWindow(mainWindow)

	// Set the main window reference in ScreenshotWindowService so it can hide/show the main window
	screenshotWindowService.SetMainWindow(mainWindow)

	// Register default search window hotkey (Cmd+5 / Ctrl+5)
	// Using Cmd+5 which might work better with gohook
	// Note: This will override the sysinfo plugin's shortcut
	searchHotkey := "cmd+5" // macOS
	if runtime.GOOS != "darwin" {
		searchHotkey = "ctrl+5" // Windows/Linux
	}

	// Remove old search window shortcuts if they exist
	// Also remove cmd+5 which is currently used by sysinfo plugin
	oldShortcuts := []string{"cmd+space", "ctrl+space", "cmd+shift+p", "ctrl+shift+p", "alt+space", "cmd+k", "ctrl+k", "cmd+9", "ctrl+9", "cmd+l", "ctrl+l", "cmd+6", "ctrl+6", "cmd+5", "ctrl+5"}
	for _, oldShortcut := range oldShortcuts {
		if err := shortcutService.RemoveShortcut(oldShortcut); err != nil {
			// Ignore errors - the shortcut might not exist
			log.Printf("[Main] Note: %s (may not exist): %v", oldShortcut, err)
		}
	}

	// Try to set the default search shortcut
	// If it conflicts, we'll let the user know
	if err := shortcutService.SetShortcut(searchHotkey, "search.window.builtin"); err != nil {
		log.Printf("[Main] Failed to set default search hotkey %s: %v (may conflict with existing shortcut)", searchHotkey, err)
	} else {
		log.Printf("[Main] Default search hotkey registered: %s", searchHotkey)
	}

	// Register default screenshot hotkey (Cmd+Shift+S / Ctrl+Shift+S)
	// Using S for Screenshot - stable letter key combination
	// Note: Avoid cmd+shift+6 as it's unstable (see docs/keycode/safe-shortcuts.md)
	screenshotHotkey := "cmd+shift+s"
	if runtime.GOOS != "darwin" {
		screenshotHotkey = "ctrl+shift+s" // Windows/Linux
	}

	if err := shortcutService.SetShortcut(screenshotHotkey, "screenshot.window.builtin"); err != nil {
		log.Printf("[Main] Failed to set screenshot hotkey %s: %v (may conflict with existing shortcut)", screenshotHotkey, err)
	} else {
		log.Printf("[Main] Default screenshot hotkey registered: %s", screenshotHotkey)
	}

	// Start the screenshot window service
	if err := screenshotWindowService.ServiceStartup(app); err != nil {
		log.Printf("Failed to start screenshot window service: %v", err)
	}

	// Run the application. This blocks until the application has been exited.
	err = app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
