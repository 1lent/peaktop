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

static int64_t iokitReadPropertyInt(io_registry_entry_t entry, const char *keyStr) {
	CFStringRef cfKey = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!cfKey) return 0;
	int64_t value = 0;
	CFTypeRef result = IORegistryEntrySearchCFProperty(
		entry, kIOServicePlane, cfKey, kCFAllocatorDefault, kIORegistryIterateRecursively);
	if (result) {
		if (CFGetTypeID(result) == CFNumberGetTypeID()) {
			CFNumberGetValue((CFNumberRef)result, kCFNumberSInt64Type, &value);
		}
		CFRelease(result);
	}
	CFRelease(cfKey);
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
	classes := []string{"IOGPU", "IOAccelerator", "AGXAccelerator", "AGXAcceleratorG17P"}
	for _, className := range classes {
		cName := C.CString(className)
		matcher := C.IOServiceMatching(cName)
		C.free(unsafe.Pointer(cName))
		if matcher == 0 {
			continue
		}
		service := C.IOServiceGetMatchingService(C.kIOMainPortDefault, C.CFDictionaryRef(matcher))
		if service == 0 {
			continue
		}
		defer C.IOObjectRelease(C.io_object_t(service))

		entry := C.io_registry_entry_t(service)

		totalKey := C.CString("VRAM,totalMB")
		vramTotal := uint64(C.iokitReadPropertyInt(entry, totalKey))
		C.free(unsafe.Pointer(totalKey))

		if vramTotal == 0 {
			continue
		}

		perfKey := C.CString("PerformanceStatistics")
		perfDict := C.iokitCopyProperty(entry, perfKey)
		C.free(unsafe.Pointer(perfKey))

		var vramFree uint64
		if perfDict != 0 {
			freeKey := C.CString("vramFreeBytes")
			vramFree = uint64(C.iokitDictGetInt64(perfDict, freeKey)) / (1024 * 1024)
			C.free(unsafe.Pointer(freeKey))
			C.CFRelease(perfDict)
		}

		vramUsed := vramTotal - vramFree
		return vramUsed, vramTotal
	}
	return 0, 0
}

// Battery and Disk IOKit access moved to battery.go and disk.go respectively.
