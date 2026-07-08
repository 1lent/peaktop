package apple

/*
#include <sys/socket.h>
#include <net/if.h>
#include <net/if_var.h>
#include <ifaddrs.h>
#include <string.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
)

type NetIface struct {
	Name     string
	RxBytes  uint64
	TxBytes  uint64
	IsActive bool
}

func GetNetworkStats() ([]NetIface, error) {
	var ifap *C.struct_ifaddrs
	result := C.getifaddrs(&ifap)
	if result != 0 {
		return nil, fmt.Errorf("getifaddrs failed: %d", result)
	}
	defer C.freeifaddrs(ifap)

	seen := make(map[string]bool)
	var interfaces []NetIface

	for ifa := ifap; ifa != nil; ifa = ifa.ifa_next {
		name := C.GoString(ifa.ifa_name)
		if seen[name] {
			continue
		}
		if ifa.ifa_addr == nil || ifa.ifa_data == nil {
			continue
		}
		if ifa.ifa_addr.sa_family != C.AF_LINK {
			continue
		}

		isUp := (ifa.ifa_flags & C.IFF_UP) != 0
		isLoopback := (ifa.ifa_flags & C.IFF_LOOPBACK) != 0
		if !isUp || isLoopback {
			continue
		}

		dl := (*C.struct_if_data)(ifa.ifa_data)

		seen[name] = true
		interfaces = append(interfaces, NetIface{
			Name:     name,
			RxBytes:  uint64(dl.ifi_ibytes),
			TxBytes:  uint64(dl.ifi_obytes),
			IsActive: true,
		})
	}

	return interfaces, nil
}
