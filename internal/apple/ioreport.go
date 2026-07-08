package apple

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <dlfcn.h>
#include <stdlib.h>
#include <stdint.h>

typedef void* IOReportRef;

static int ioreportAvailability = -1;

static int _ioreportCheckAvailable(void) {
	if (ioreportAvailability >= 0) return ioreportAvailability;

	void *handle = dlopen("/System/Library/Frameworks/IOKit.framework/IOKit", RTLD_LAZY);
	if (!handle) {
		ioreportAvailability = 0;
		return 0;
	}

	void *sym = dlsym(handle, "IOReporterOpen");
	dlclose(handle);

	if (sym) {
		ioreportAvailability = 1;
	} else {
		ioreportAvailability = 0;
	}
	return ioreportAvailability;
}
*/
import "C"

import "fmt"

type IOReportChannel struct {
	Group    string
	SubGroup string
	Channel  uint64
	Value    float64
}

const ioReportUnavailable = "IOReport requires com.apple.private.io-report entitlement (not available for third-party apps)"

func IsIOReportAvailable() bool {
	return C._ioreportCheckAvailable() == 1
}

func GetIOReportPower() ([]IOReportChannel, error) {
	if !IsIOReportAvailable() {
		return nil, fmt.Errorf(ioReportUnavailable)
	}

	return []IOReportChannel{}, fmt.Errorf("IOReport power channels not implemented via dlsym")
}

func GetIOReportANE() (float64, error) {
	if !IsIOReportAvailable() {
		return 0, fmt.Errorf(ioReportUnavailable)
	}
	return 0, fmt.Errorf("IOReport ANE not implemented via dlsym")
}
