package apple

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <IOKit/IOKitLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <string.h>

static double batteryDictGetDouble(CFTypeRef dict, const char *keyStr) {
	CFStringRef key = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!key) return 0.0;
	double value = 0.0;
	if (dict && CFGetTypeID(dict) == CFDictionaryGetTypeID()) {
		CFNumberRef num = (CFNumberRef)CFDictionaryGetValue((CFDictionaryRef)dict, key);
		if (num && CFGetTypeID(num) == CFNumberGetTypeID()) {
			CFNumberGetValue(num, kCFNumberDoubleType, &value);
		}
	}
	CFRelease(key);
	return value;
}

static int64_t batteryDictGetInt64(CFTypeRef dict, const char *keyStr) {
	CFStringRef key = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!key) return 0;
	int64_t value = 0;
	if (dict && CFGetTypeID(dict) == CFDictionaryGetTypeID()) {
		CFNumberRef num = (CFNumberRef)CFDictionaryGetValue((CFDictionaryRef)dict, key);
		if (num && CFGetTypeID(num) == CFNumberGetTypeID()) {
			CFNumberGetValue(num, kCFNumberSInt64Type, &value);
		}
	}
	CFRelease(key);
	return value;
}

static int64_t batteryReadInt(io_registry_entry_t entry, const char *keyStr) {
	CFStringRef cfKey = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!cfKey) return 0;
	int64_t value = 0;
	CFTypeRef result = IORegistryEntrySearchCFProperty(
		entry, kIOServicePlane, cfKey, kCFAllocatorDefault, kIORegistryIterateRecursively);
	if (result && CFGetTypeID(result) == CFNumberGetTypeID()) {
		CFNumberGetValue((CFNumberRef)result, kCFNumberSInt64Type, &value);
	}
	if (result) CFRelease(result);
	CFRelease(cfKey);
	return value;
}

static int batteryReadBool(io_registry_entry_t entry, const char *keyStr) {
	CFStringRef cfKey = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!cfKey) return 0;
	int value = 0;
	CFTypeRef result = IORegistryEntrySearchCFProperty(
		entry, kIOServicePlane, cfKey, kCFAllocatorDefault, kIORegistryIterateRecursively);
	if (result && CFGetTypeID(result) == CFBooleanGetTypeID()) {
		value = CFBooleanGetValue((CFBooleanRef)result);
	}
	if (result) CFRelease(result);
	CFRelease(cfKey);
	return value;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const batteryIOKitClass = "IOPMPowerSource"

func GetBatteryInfo() (percent float64, charging bool, cycleCount int, maxCapacity int, designCapacity int, timeRemaining string, hasBattery bool, voltageMV int, currentMA int, err error) {
	cName := C.CString(batteryIOKitClass)
	defer C.free(unsafe.Pointer(cName))
	matcher := C.IOServiceMatching(cName)
	if matcher == 0 {
		return 0, false, 0, 0, 0, "", false, 0, 0, nil
	}

	service := C.IOServiceGetMatchingService(C.kIOMainPortDefault, C.CFDictionaryRef(matcher))
	if service == 0 {
		return 0, false, 0, 0, 0, "", false, 0, 0, nil
	}
	defer C.IOObjectRelease(C.io_object_t(service))

	entry := C.io_registry_entry_t(service)

	curKey := C.CString("CurrentCapacity")
	defer C.free(unsafe.Pointer(curKey))
	percent = float64(C.batteryReadInt(entry, curKey))

	chargeKey := C.CString("IsCharging")
	defer C.free(unsafe.Pointer(chargeKey))
	charging = C.batteryReadBool(entry, chargeKey) != 0

	pctNotZero := percent > 0
	if !pctNotZero {
		chargingCheck := C.batteryReadInt(entry, chargeKey)
		if chargingCheck != 0 {
			charging = true
		}
		_ = chargingCheck
	}

	cycleKey := C.CString("CycleCount")
	defer C.free(unsafe.Pointer(cycleKey))
	cycleCount = int(C.batteryReadInt(entry, cycleKey))

	maxKey := C.CString("MaxCapacity")
	defer C.free(unsafe.Pointer(maxKey))
	maxCapacity = int(C.batteryReadInt(entry, maxKey))

	designKey := C.CString("DesignCapacity")
	defer C.free(unsafe.Pointer(designKey))
	designCapacity = int(C.batteryReadInt(entry, designKey))

	batteryIsPresent := percent > 0 || maxCapacity > 0 || designCapacity > 0
	if !batteryIsPresent {
		return 0, false, 0, 0, 0, "", false, 0, 0, nil
	}

	hasBattery = true

	voltKey := C.CString("Voltage")
	defer C.free(unsafe.Pointer(voltKey))
	voltageMV = int(C.batteryReadInt(entry, voltKey))

	ampKey := C.CString("InstantAmperage")
	defer C.free(unsafe.Pointer(ampKey))
	currentMA = int(C.batteryReadInt(entry, ampKey))

	timeRemainingKey := C.CString("TimeRemaining")
	defer C.free(unsafe.Pointer(timeRemainingKey))
	timeRemainingInt := int(C.batteryReadInt(entry, timeRemainingKey))
	if timeRemainingInt > 0 {
		hours := timeRemainingInt / 60
		mins := timeRemainingInt % 60
		timeRemaining = fmt.Sprintf("%dh%dm", hours, mins)
	}

	return percent, charging, cycleCount, maxCapacity, designCapacity, timeRemaining, hasBattery, voltageMV, currentMA, nil
}
