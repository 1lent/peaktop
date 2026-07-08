package apple

/*
#cgo LDFLAGS: -lproc

#include <libproc.h>
#include <sys/proc_info.h>
#include <sys/proc.h>
#include <mach/mach.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	procNameMax = 256
)

type ProcInfo struct {
	PID         int32
	Name        string
	CPUPercent  float64
	MemoryBytes uint64
	UserTicks   uint64
	SystemTicks uint64
	TotalTicks  uint64
}

func ListPIDs() ([]int32, error) {
	bufSize := C.proc_listallpids(nil, 0)
	if bufSize <= 0 {
		return nil, fmt.Errorf("proc_listallpids size query failed: %d", bufSize)
	}

	pids := make([]C.int, bufSize/C.sizeof_int)
	numBytes := C.proc_listallpids(unsafe.Pointer(&pids[0]), bufSize)
	if numBytes <= 0 {
		return nil, fmt.Errorf("proc_listallpids failed: %d", numBytes)
	}

	numPIDs := int(numBytes) / int(C.sizeof_int)
	result := make([]int32, numPIDs)
	for i := 0; i < numPIDs; i++ {
		result[i] = int32(pids[i])
	}
	return result, nil
}

func GetProcInfo(pid int32) (*ProcInfo, error) {
	var taskInfo C.struct_proc_taskinfo
	taskInfoSize := C.int(unsafe.Sizeof(taskInfo))

	result := C.proc_pidinfo(
		C.int(pid),
		C.PROC_PIDTASKINFO,
		0,
		unsafe.Pointer(&taskInfo),
		taskInfoSize,
	)
	if result <= 0 {
		return nil, fmt.Errorf("proc_pidinfo TASKINFO pid %d: %d", pid, result)
	}

	name := getProcessName(pid)

	return &ProcInfo{
		PID:         pid,
		Name:        name,
		CPUPercent:  0,
		MemoryBytes: uint64(taskInfo.pti_resident_size),
		UserTicks:   uint64(taskInfo.pti_total_user),
		SystemTicks: uint64(taskInfo.pti_total_system),
		TotalTicks:  uint64(taskInfo.pti_total_user) + uint64(taskInfo.pti_total_system),
	}, nil
}

func getProcessName(pid int32) string {
	nameBuf := make([]C.char, procNameMax)
	result := C.proc_name(C.int(pid), unsafe.Pointer(&nameBuf[0]), C.uint32_t(len(nameBuf)))
	if result <= 0 {
		return fmt.Sprintf("pid:%d", pid)
	}
	return C.GoString(&nameBuf[0])
}
