package apple

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <IOKit/IOKitLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <string.h>

static double gpuReadDouble(CFTypeRef perfDict, const char *keyStr) {
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

static CFTypeRef gpuCopyPerfStats(io_registry_entry_t entry) {
	CFStringRef key = CFStringCreateWithCString(kCFAllocatorDefault, "PerformanceStatistics", kCFStringEncodingUTF8);
	if (!key) return NULL;
	CFTypeRef result = IORegistryEntrySearchCFProperty(
		entry, kIOServicePlane, key, kCFAllocatorDefault, kIORegistryIterateRecursively);
	CFRelease(key);
	return result;
}
*/
import "C"

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

type GPUStats struct {
	UsagePercent float64
	ActiveMHz    float64
	VRAMUsedMB   uint64
	VRAMTotalMB  uint64
}

const (
	gpuIOKitClass     = "IOGPU"
	gpuAccelClass     = "IOAccelerator"
	gpuKeyUtilization = "Device Utilization %"
	gpuKeyFrequency   = "Current Frequency"
)

func GetGPUUsage() (float64, error) {
	perf, err := gpuPerformanceStats()
	if err != nil {
		return gpuUsageFromIoreg()
	}

	usage := gpuReadPerformanceKey(perf, gpuKeyUtilization)
	if usage > 0 {
		C.CFRelease(perf)
		return usage, nil
	}

	altUsage := gpuReadPerformanceKey(perf, "GPU Core Utilization")
	C.CFRelease(perf)
	if altUsage > 0 {
		return altUsage, nil
	}

	return gpuUsageFromIoreg()
}

func GetGPUFrequency() (float64, error) {
	perf, err := gpuPerformanceStats()
	if err != nil {
		return gpuFrequencyFromIoreg()
	}
	defer C.CFRelease(perf)

	freq := gpuReadPerformanceKey(perf, gpuKeyFrequency)
	if freq > 0 {
		return freq, nil
	}

	altFreq := gpuReadPerformanceKey(perf, "Core Clock")
	if altFreq > 0 {
		return altFreq, nil
	}

	return gpuFrequencyFromIoreg()
}

func gpuPerformanceStats() (C.CFTypeRef, error) {
	cName := C.CString(gpuIOKitClass)
	defer C.free(unsafe.Pointer(cName))
	matcher := C.IOServiceMatching(cName)
	if matcher == 0 {
		return 0, fmt.Errorf("IOServiceMatching %s failed", gpuIOKitClass)
	}

	service := C.IOServiceGetMatchingService(C.kIOMainPortDefault, C.CFDictionaryRef(matcher))
	if service == 0 {
		return 0, fmt.Errorf("GPU service not found via IOKit")
	}
	defer C.IOObjectRelease(C.io_object_t(service))

	perf := C.gpuCopyPerfStats(C.io_registry_entry_t(service))
	if perf == 0 {
		return 0, fmt.Errorf("GPU PerformanceStatistics not found")
	}

	return perf, nil
}

func gpuReadPerformanceKey(perf C.CFTypeRef, key string) float64 {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	return float64(C.gpuReadDouble(perf, cKey))
}

func gpuUsageFromIoreg() (float64, error) {
	output, err := runIoreg(gpuAccelClass)
	if err != nil {
		return 0, err
	}
	return parseIoregFloat(output, gpuKeyUtilization), nil
}

func gpuFrequencyFromIoreg() (float64, error) {
	output, err := runIoreg(gpuAccelClass)
	if err != nil {
		return 0, err
	}
	return parseIoregFloat(output, gpuKeyFrequency), nil
}

func runIoreg(className string) (string, error) {
	cmd := exec.Command("ioreg", "-c", className, "-r", "-d", "2")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ioreg %s failed: %w", className, err)
	}
	return stdout.String(), nil
}

func parseIoregFloat(output, key string) float64 {
	idx := strings.Index(output, key)
	if idx < 0 {
		return 0
	}

	rest := output[idx+len(key):]
	eqIdx := strings.Index(rest, "=")
	if eqIdx < 0 {
		return 0
	}

	valStr := strings.TrimLeft(rest[eqIdx+1:], " ")
	end := strings.IndexAny(valStr, ",}")
	if end < 0 {
		end = len(valStr)
	}
	valStr = strings.TrimSpace(valStr[:end])
	valStr = strings.Trim(valStr, "\"")

	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0
	}
	return val
}
