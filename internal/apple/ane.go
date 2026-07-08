package apple

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <IOKit/IOKitLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <string.h>

static double aneReadDouble(CFTypeRef perfDict, const char *keyStr) {
	CFStringRef key = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!key) return 0.0;
	double value = 0.0;
	if (perfDict && CFGetTypeID(perfDict) == CFDictionaryGetTypeID()) {
		CFNumberRef num = (CFNumberRef)CFDictionaryGetValue((CFDictionaryRef)perfDict, key);
		if (num && CFGetTypeID(num) == CFNumberGetTypeID()) {
			CFNumberGetValue(num, kCFNumberDoubleType, &value);
		}
	}
	CFRelease(key);
	return value;
}

static CFTypeRef aneCopyPerfStats(io_registry_entry_t entry) {
	CFStringRef key = CFStringCreateWithCString(kCFAllocatorDefault, "PerformanceStatistics", kCFStringEncodingUTF8);
	if (!key) return NULL;
	CFTypeRef result = IORegistryEntrySearchCFProperty(
		entry, kIOServicePlane, key, kCFAllocatorDefault, kIORegistryIterateRecursively);
	CFRelease(key);
	return result;
}

static io_registry_entry_t aneFindService(const char *className) {
	CFStringRef cfClass = CFStringCreateWithCString(kCFAllocatorDefault, className, kCFStringEncodingUTF8);
	if (!cfClass) return 0;
	CFMutableDictionaryRef matcher = IOServiceMatching(className);
	if (!matcher) {
		CFRelease(cfClass);
		return 0;
	}
	io_registry_entry_t service = IOServiceGetMatchingService(kIOMainPortDefault, matcher);
	if (service) return service;

	CFStringRef cfResources = CFStringCreateWithCString(kCFAllocatorDefault, "IOResources", kCFStringEncodingUTF8);
	CFMutableDictionaryRef props = CFDictionaryCreateMutable(kCFAllocatorDefault, 1, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	CFDictionarySetValue(props, CFSTR("IOProviderClass"), cfResources);
	CFRelease(cfResources);

	io_iterator_t iter;
	kern_return_t kr = IOServiceGetMatchingServices(kIOMainPortDefault, props, &iter);
	if (kr != KERN_SUCCESS) return 0;

	io_registry_entry_t found = 0;
	while ((service = IOIteratorNext(iter)) != 0) {
		CFStringRef ioclass = (CFStringRef)IORegistryEntrySearchCFProperty(
			service, kIOServicePlane, CFSTR("IOClass"), kCFAllocatorDefault, 0);
		if (ioclass) {
			char buf[128];
			if (CFStringGetCString(ioclass, buf, sizeof(buf), kCFStringEncodingUTF8)) {
				const char *p = buf;
				const char *s = className;
				while (*p && *s && *p == *s) { p++; s++; }
				if (*s == '\0') {
					found = service;
					CFRelease(ioclass);
					break;
				}
			}
			CFRelease(ioclass);
		}
		IOObjectRelease(service);
	}
	IOObjectRelease(iter);
	return found;
}
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

type ANEStats struct {
	UsagePercent float64
}

func GetANEUsage() (float64, error) {
	classNames := aneClassNames()
	var anyFound bool

	for _, className := range classNames {
		cName := C.CString(className)
		service := C.aneFindService(cName)
		C.free(unsafe.Pointer(cName))
		if service == 0 {
			continue
		}
		anyFound = true

		perf := C.aneCopyPerfStats(service)
		C.IOObjectRelease(service)
		if perf == 0 {
			continue
		}
		defer C.CFRelease(perf)

		keys := []string{
			"ANE Utilization",
			"ANE Utilization %",
			"Neural Engine Utilization",
			"Neural Engine Utilization %",
		}

		for _, key := range keys {
			cKey := C.CString(key)
			val := float64(C.aneReadDouble(perf, cKey))
			C.free(unsafe.Pointer(cKey))
			if val > 0 {
				return val, nil
			}
		}

		return 0, fmt.Errorf("ANE service found but no utilization data available")
	}

	if anyFound {
		return 0, fmt.Errorf("ANE service found but no performance statistics exposed")
	}

	return 0, fmt.Errorf("ANE not available on this device")
}

func aneClassNames() []string {
	names := []string{"AppleNeuralEngine"}

	name, err := ReadSysctlString("machdep.cpu.brand_string")
	if err == nil {
		name = strings.TrimSpace(name)
		name = strings.TrimPrefix(name, "Apple ")
		if len(name) > 0 {
			chip := chipToTNumber(name)
			if chip != "" {
				names = append([]string{"Apple" + chip + "ANEHAL"}, names...)
			}
		}
	}

	names = append(names,
		"AppleT8140ANEHAL",
		"AppleT8130ANEHAL",
		"AppleT8120ANEHAL",
		"AppleT8110ANEHAL",
		"AppleT8103ANEHAL",
		"AppleT8101ANEHAL",
		"AppleT6000ANEHAL",
		"AppleT6001ANEHAL",
		"AppleT6002ANEHAL",
	)

	return names
}

func chipToTNumber(chipName string) string {
	switch {
	case strings.Contains(chipName, "A18 Pro") || strings.Contains(chipName, "A18Pro"):
		return "T8140"
	case strings.Contains(chipName, "A18"):
		return "T8130"
	case strings.Contains(chipName, "A17 Pro") || strings.Contains(chipName, "A17Pro"):
		return "T8120"
	case strings.Contains(chipName, "A16"):
		return "T8110"
	case strings.Contains(chipName, "A15"):
		return "T8110"
	case strings.Contains(chipName, "A14"):
		return "T8101"
	case strings.Contains(chipName, "M4"):
		return "T8140"
	case strings.Contains(chipName, "M3"):
		return "T8122"
	case strings.Contains(chipName, "M2"):
		return "T8112"
	case strings.Contains(chipName, "M1"):
		return "T8103"
	}
	return ""
}

func GetANEStats() (*ANEStats, error) {
	usage, err := GetANEUsage()
	if err != nil {
		return nil, fmt.Errorf("ANE: %w", err)
	}
	return &ANEStats{UsagePercent: usage}, nil
}
