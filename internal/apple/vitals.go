package apple

/*
#include <mach/mach.h>
#include <mach/mach_host.h>
#include <mach/host_info.h>
#include <sys/sysctl.h>
#include <stdlib.h>
#include <unistd.h>
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

type VMStats struct {
	FreeBytes       uint64
	ActiveBytes     uint64
	InactiveBytes   uint64
	WiredBytes      uint64
	CompressedBytes uint64
	TotalBytes      uint64
}

type SwapInfo struct {
	TotalBytes uint64
	UsedBytes  uint64
}

var cachedTotal uint64
var totalCached bool

func GetVMStats() (VMStats, error) {
	var count C.mach_msg_type_number_t = C.HOST_VM_INFO64_COUNT
	var raw C.vm_statistics64_data_t

	result := C.host_statistics64(
		C.mach_host_self(),
		C.HOST_VM_INFO64,
		(C.host_info64_t)(unsafe.Pointer(&raw)),
		&count,
	)
	if result != C.KERN_SUCCESS {
		return VMStats{}, fmt.Errorf("host_statistics64 failed: %d", result)
	}

	pageSize := uint64(C.getpagesize())
	total := getTotalMemoryCache()

	return VMStats{
		FreeBytes:       uint64(raw.free_count) * pageSize,
		ActiveBytes:     uint64(raw.active_count) * pageSize,
		InactiveBytes:   uint64(raw.inactive_count) * pageSize,
		WiredBytes:      uint64(raw.wire_count) * pageSize,
		CompressedBytes: uint64(raw.compressor_page_count) * pageSize,
		TotalBytes:      total,
	}, nil
}

func getTotalMemoryCache() uint64 {
	if totalCached {
		return cachedTotal
	}
	val := readSysctlInt("hw.memsize")
	if val < 0 {
		return 0
	}
	cachedTotal = uint64(val)
	totalCached = true
	return cachedTotal
}

func GetVMPressure() (int, error) {
	pressure := readSysctlInt("kern.memorystatus_vm_pressure_level")
	if pressure < 0 {
		return 0, fmt.Errorf("memory pressure sysctl unavailable")
	}
	return pressure, nil
}

func GetSwapUsage() (SwapInfo, error) {
	data, err := ReadSysctlBytes("vm.swapusage")
	if err != nil {
		return SwapInfo{}, nil
	}

	if len(data) > 0 && data[0] >= 'a' && data[0] <= 'z' {
		total, used := parseSwapUsage(string(data))
		return SwapInfo{TotalBytes: total, UsedBytes: used}, nil
	}

	total, used := parseSwapBinary(data)
	return SwapInfo{TotalBytes: total, UsedBytes: used}, nil
}

func parseSwapBinary(data []byte) (uint64, uint64) {
	if len(data) < 16 {
		return 0, 0
	}
	total := binary.LittleEndian.Uint64(data[0:8])
	avail := binary.LittleEndian.Uint64(data[8:16])
	used := total - avail
	return total, used
}

func parseSwapUsage(raw string) (uint64, uint64) {
	total, used := uint64(0), uint64(0)

	fields := strings.Fields(raw)
	for i, field := range fields {
		if field == "total" && i+2 < len(fields) {
			total = parseSwapBytes(fields[i+2])
		}
		if field == "used" && i+2 < len(fields) {
			used = parseSwapBytes(fields[i+2])
		}
	}
	return total, used
}

func parseSwapBytes(val string) uint64 {
	val = strings.TrimRight(val, "M")
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return uint64(f * 1024 * 1024)
}

func GetMemoryPressureString(level int) string {
	switch level {
	case 0:
		return "Normal"
	case 1:
		return "Warning"
	case 2:
		return "Critical"
	case 3:
		return "Urgent"
	default:
		return "Unknown"
	}
}
