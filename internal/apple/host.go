package apple

/*
#include <mach/mach.h>
#include <mach/mach_host.h>
#include <mach/host_info.h>
#include <mach/processor_info.h>
#include <sys/sysctl.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

const cpuStateMax = 4

type CPUTick struct {
	User uint32
	Sys  uint32
	Idle uint32
	Nice uint32
}

type PerfLevel struct {
	Name  string
	Count int
}

type HostCPUInfo struct {
	PerCoreTicks []CPUTick
	CoreCount    int
	PerfLevels   []PerfLevel
	CoreLabels   []string
	FrequencyHz  uint64
}

func GetHostCPUInfo() (*HostCPUInfo, error) {
	var processorInfo *C.processor_cpu_load_info_t
	var processorMsgCount C.mach_msg_type_number_t
	var numCPUs C.natural_t

	result := C.host_processor_info(
		C.mach_host_self(),
		C.PROCESSOR_CPU_LOAD_INFO,
		&numCPUs,
		(*C.processor_info_array_t)(unsafe.Pointer(&processorInfo)),
		&processorMsgCount,
	)
	if result != C.KERN_SUCCESS {
		return nil, fmt.Errorf("host_processor_info failed: %d", result)
	}
	defer func() {
		C.vm_deallocate(
			C.mach_task_self_,
			C.vm_address_t(uintptr(unsafe.Pointer(processorInfo))),
			C.vm_size_t(processorMsgCount),
		)
	}()

	totalElements := int(numCPUs) * cpuStateMax
	cpuValues := unsafe.Slice((*uint32)(unsafe.Pointer(processorInfo)), totalElements)

	cores := make([]CPUTick, numCPUs)
	for i := 0; i < int(numCPUs); i++ {
		offset := i * cpuStateMax
		cores[i] = CPUTick{
			User: cpuValues[offset],
			Sys:  cpuValues[offset+1],
			Idle: cpuValues[offset+2],
			Nice: cpuValues[offset+3],
		}
	}

	perfLevels := detectPerfLevels()
	coreLabels := generateCoreLabels(int(numCPUs), perfLevels)
	freq := getCPUFrequency()

	return &HostCPUInfo{
		PerCoreTicks: cores,
		CoreCount:    int(numCPUs),
		PerfLevels:   perfLevels,
		CoreLabels:   coreLabels,
		FrequencyHz:  freq,
	}, nil
}

func detectPerfLevels() []PerfLevel {
	numLevels := readSysctlInt("hw.nperflevels")
	if numLevels <= 0 {
		numLevels = 2
	}

	levels := make([]PerfLevel, numLevels)
	for i := 0; i < numLevels; i++ {
		count := readSysctlInt(fmt.Sprintf("hw.perflevel%d.logicalcpu", i))
		name := readPerfLevelName(i)
		levels[i] = PerfLevel{Name: name, Count: count}
	}

	return levels
}

func readPerfLevelName(index int) string {
	name, err := ReadSysctlString(fmt.Sprintf("hw.perflevel%d.name", index))
	if err != nil {
		name = fmt.Sprintf("Cluster%d", index)
	}
	return strings.TrimSpace(name)
}

func generateCoreLabels(totalCores int, perfLevels []PerfLevel) []string {
	labels := make([]string, totalCores)

	coreIndex := 0
	for _, level := range perfLevels {
		prefix := clusterPrefix(level.Name)
		for i := 0; i < level.Count && coreIndex < totalCores; i++ {
			labels[coreIndex] = fmt.Sprintf("%s%d", prefix, i)
			coreIndex++
		}
	}

	for coreIndex < totalCores {
		labels[coreIndex] = fmt.Sprintf("C%d", coreIndex)
		coreIndex++
	}

	return labels
}

func clusterPrefix(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "super") {
		return "S"
	}
	if strings.Contains(lower, "performance") || strings.Contains(lower, "p-cluster") || strings.Contains(lower, "pcluster") {
		return "P"
	}
	if strings.Contains(lower, "efficiency") || strings.Contains(lower, "e-cluster") || strings.Contains(lower, "ecluster") {
		return "E"
	}

	return "C"
}

func getCPUFrequency() uint64 {
	freq := readSysctlInt("hw.cpufrequency")
	if freq < 0 {
		return 0
	}
	return uint64(freq)
}

func readSysctlInt(name string) int {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var val C.int64_t
	size := C.size_t(unsafe.Sizeof(val))
	result := C.sysctlbyname(cName, unsafe.Pointer(&val), &size, nil, 0)
	if result != 0 {
		return -1
	}
	return int(val)
}

func ReadSysctlString(name string) (string, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var size C.size_t
	if result := C.sysctlbyname(cName, nil, &size, nil, 0); result != 0 {
		return "", fmt.Errorf("sysctlbyname size: %s: %d", name, result)
	}
	if size == 0 {
		return "", fmt.Errorf("sysctlbyname: %s: zero size", name)
	}

	buf := make([]byte, size)
	if result := C.sysctlbyname(cName, unsafe.Pointer(&buf[0]), &size, nil, 0); result != 0 {
		return "", fmt.Errorf("sysctlbyname: %s: %d", name, result)
	}

	return string(buf[:len(buf)-1]), nil
}

func ReadSysctlBytes(name string) ([]byte, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var size C.size_t
	if result := C.sysctlbyname(cName, nil, &size, nil, 0); result != 0 {
		return nil, fmt.Errorf("sysctlbyname size: %s: %d", name, result)
	}
	if size == 0 {
		return nil, fmt.Errorf("sysctlbyname: %s: zero size", name)
	}

	buf := make([]byte, size)
	if result := C.sysctlbyname(cName, unsafe.Pointer(&buf[0]), &size, nil, 0); result != 0 {
		return nil, fmt.Errorf("sysctlbyname: %s: %d", name, result)
	}

	return buf[:size], nil
}
