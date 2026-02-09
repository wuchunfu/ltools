package screenshot

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

// SetPNGToClipboard sets PNG image data to the clipboard on macOS
BOOL SetPNGToClipboard(const void *data, size_t length) {
    @autoreleasepool {
        // Create NSData from the PNG data
        NSData *pngData = [NSData dataWithBytes:data length:length];

        // Create NSImage from PNG data
        NSImage *image = [[NSImage alloc] initWithData:pngData];
        if (!image) {
            return NO;
        }

        // Get the pasteboard
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];

        // Clear the pasteboard and declare types
        [pasteboard clearContents];
        NSArray *types = @[@"public.png", @"PNG data"];
        [pasteboard declareTypes:types owner:nil];

        // Write the image to the pasteboard
        BOOL success = [pasteboard setData:pngData forType:@"public.png"];
        if (!success) {
            // Try alternative type
            success = [pasteboard setData:pngData forType:@"PNG data"];
        }

        return success;
    }
}
*/
import "C"
import (
	"log"
)

// setImageMac sets image on macOS clipboard using native Cocoa API
func (c *Clipboard) setImageMac(imgData []byte) error {
	log.Printf("[Screenshot] setImageMac called with %d bytes", len(imgData))

	if len(imgData) == 0 {
		log.Printf("[Screenshot] Error: empty image data")
		return &InvalidImageError{}
	}

	// 验证 PNG 签名
	if len(imgData) >= 8 {
		log.Printf("[Screenshot] PNG signature: %x", imgData[:8])
	}

	// Convert Go slice to C pointer
	log.Printf("[Screenshot] Converting data to C pointer...")
	data := C.CBytes(imgData)
	defer C.free(data)

	// Call the C function
	log.Printf("[Screenshot] Calling SetPNGToClipboard...")
	success := C.SetPNGToClipboard(data, C.size_t(len(imgData)))

	if success == 0 {
		log.Printf("[Screenshot] ✗ Failed to set image to clipboard")
		return &ClipboardSetError{}
	}

	log.Printf("[Screenshot] ✓ Successfully set image to clipboard")
	return nil
}
