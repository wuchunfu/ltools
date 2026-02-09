// +build darwin

package screenshot

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

// 设置最后一个创建的窗口到 CGShieldingWindowLevel
// 这应该是我们刚创建的截图窗口
void setLastWindowToShieldingLevel() {
    NSApplication *app = [NSApplication sharedApplication];
    NSArray *windows = [app windows];

    if ([windows count] > 0) {
        // 获取最后一个窗口（最新创建的）
        NSWindow *window = [windows lastObject];
        // 设置到 CGShieldingWindowLevel
        [window setLevel:CGShieldingWindowLevel()];
        NSLog(@"Set last window to CGShieldingWindowLevel: %@", window);
    }
}

// 设置所有窗口到 CGShieldingWindowLevel
void setAllWindowsToShieldingLevel() {
    NSApplication *app = [NSApplication sharedApplication];
    NSArray *windows = [app windows];

    for (NSWindow *window in windows) {
        [window setLevel:CGShieldingWindowLevel()];
        NSLog(@"Set window to CGShieldingWindowLevel: %@", window);
    }
}

*/
import "C"
import (
	"log"
	"runtime"
)

// SetLastWindowToShieldingLevel 设置最后一个窗口（截图窗口）到 CGShieldingWindowLevel
func SetLastWindowToShieldingLevel() error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	log.Printf("[Screenshot] Setting last window to CGShieldingWindowLevel...")
	C.setLastWindowToShieldingLevel()
	log.Printf("[Screenshot] Last window set to CGShieldingWindowLevel successfully")
	return nil
}

// SetAllWindowsToShieldingLevel 设置所有窗口到 CGShieldingWindowLevel
func SetAllWindowsToShieldingLevel() error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	log.Printf("[Screenshot] Setting all windows to CGShieldingWindowLevel...")
	C.setAllWindowsToShieldingLevel()
	log.Printf("[Screenshot] All windows set to CGShieldingWindowLevel successfully")
	return nil
}
