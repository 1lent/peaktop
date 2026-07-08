package apple

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <IOKit/IOKitLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <string.h>

static double iokitDictGetDouble(CFTypeRef dict, const char *keyStr) {
	CFStringRef key = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!key) return 0.0;
	double value = 0.0;
	if (CFGetTypeID(dict) == CFDictionaryGetTypeID()) {
		CFNumberRef num = (CFNumberRef)CFDictionaryGetValue((CFDictionaryRef)dict, key);
		if (num && CFGetTypeID(num) == CFNumberGetTypeID()) {
			CFNumberGetValue(num, kCFNumberDoubleType, &value);
		}
	}
	CFRelease(key);
	return value;
}

static int64_t iokitDictGetInt64(CFTypeRef dict, const char *keyStr) {
	CFStringRef key = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!key) return 0;
	int64_t value = 0;
	if (CFGetTypeID(dict) == CFDictionaryGetTypeID()) {
		CFNumberRef num = (CFNumberRef)CFDictionaryGetValue((CFDictionaryRef)dict, key);
		if (num && CFGetTypeID(num) == CFNumberGetTypeID()) {
			CFNumberGetValue(num, kCFNumberSInt64Type, &value);
		}
	}
	CFRelease(key);
	return value;
}

static CFTypeRef iokitCopyProperty(io_registry_entry_t entry, const char *keyStr) {
	CFStringRef cfKey = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!cfKey) return NULL;
	CFTypeRef result = IORegistryEntrySearchCFProperty(
		entry,
		kIOServicePlane,
		cfKey,
		kCFAllocatorDefault,
		kIORegistryIterateRecursively);
	CFRelease(cfKey);
	return result;
}
*/
import "C"

import (
	"unsafe"
)

// GetGPUStats delegates to gpu.go for GPU-specific IOKit access.
// See gpu.go for GetGPUUsage() and GetGPUFrequency().
func GetGPUStats() (*GPUStats, error) {
	usage, err := GetGPUUsage()
	if err != nil {
		return nil, err
	}
	freq, _ := GetGPUFrequency()
	vramUsed, vramTotal := getGPUVRAM()
	return &GPUStats{
		UsagePercent: usage,
		ActiveMHz:    freq,
		VRAMUsedMB:   vramUsed,
		VRAMTotalMB:  vramTotal,
	}, nil
}

func getGPUVRAM() (uint64, uint64) {
	cName := C.CString("IOGPU")
	defer C.free(unsafe.Pointer(cName))
	matcher := C.IOServiceMatching(cName)
	if matcher == 0 {
		return 0, 0
	}
	service := C.IOServiceGetMatchingService(C.kIOMainPortDefault, C.CFDictionaryRef(matcher))
	if service == 0 {
		return 0, 0
	}
	defer C.IOObjectRelease(C.io_object_t(service))

	perfKey := C.CString("PerformanceStatistics")
	defer C.free(unsafe.Pointer(perfKey))
	perfDict := C.iokitCopyProperty(C.io_registry_entry_t(service), perfKey)
	if perfDict == 0 {
		return 0, 0
	}
	defer C.CFRelease(perfDict)

	vramTotalKey := C.CString("vramTotalBytes")
	defer C.free(unsafe.Pointer(vramTotalKey))
	vramUsedKey := C.CString("vramUsedBytes")
	defer C.free(unsafe.Pointer(vramUsedKey))
	vramTotal := uint64(C.iokitDictGetInt64(perfDict, vramTotalKey)) / (1024 * 1024)
	vramUsed := uint64(C.iokitDictGetInt64(perfDict, vramUsedKey)) / (1024 * 1024)
	return vramUsed, vramTotal
}

// Battery and Disk IOKit access moved to battery.go and disk.go respectively.
