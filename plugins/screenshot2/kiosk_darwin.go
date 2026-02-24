//go:build darwin

package screenshot2

/*
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework Cocoa

#include <Cocoa/Cocoa.h>

// 使用 s2_ 前缀避免与 screenshot 插件的符号冲突
void s2_enterKioskMode() {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSApplication *app = [NSApplication sharedApplication];
        app.presentationOptions =
            NSApplicationPresentationHideDock |
            NSApplicationPresentationHideMenuBar |
            NSApplicationPresentationDisableProcessSwitching |
            NSApplicationPresentationDisableForceQuit;
    });
}

void s2_exitKioskMode() {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSApplication *app = [NSApplication sharedApplication];
        app.presentationOptions = NSApplicationPresentationDefault;
    });
}
*/
import "C"

import "log"

// EnterKioskMode 进入 macOS Kiosk 模式
func EnterKioskMode() {
	log.Printf("[Screenshot2] Entering kiosk mode...")
	C.s2_enterKioskMode()
}

// ExitKioskMode 退出 macOS Kiosk 模式
func ExitKioskMode() {
	log.Printf("[Screenshot2] Exiting kiosk mode...")
	C.s2_exitKioskMode()
}
