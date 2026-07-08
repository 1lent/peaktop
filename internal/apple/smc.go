package apple

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <IOKit/IOKitLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

typedef struct {
	uint32_t key;
	uint8_t  data[32];
	uint8_t  dataSize;
	uint8_t  result;
} SMCParamStruct;

static io_connect_t smcOpenConnection(void) {
	io_service_t service = IOServiceGetMatchingService(
		kIOMainPortDefault,
		IOServiceMatching("AppleSMC"));
	if (!service) return 0;

	io_connect_t conn;
	kern_return_t result = IOServiceOpen(service, mach_task_self(), 0, &conn);
	IOObjectRelease(service);

	if (result != KERN_SUCCESS) return 0;
	return conn;
}

static kern_return_t smcCall(io_connect_t conn, uint32_t selector,
                              SMCParamStruct *input, SMCParamStruct *output) {
	size_t inputSize = sizeof(SMCParamStruct);
	size_t outputSize = sizeof(SMCParamStruct);
	return IOConnectCallStructMethod(conn, selector, input, inputSize, output, &outputSize);
}
*/
import "C"

import (
	"fmt"
	"math"
	"unsafe"
)

const smcSelectorReadKey = 5

var smcConnection C.io_connect_t
var smcInitialized bool

func ensureSMCConnection() error {
	if smcInitialized {
		if smcConnection == 0 {
			return fmt.Errorf("SMC not available on this system")
		}
		return nil
	}
	smcInitialized = true

	conn := C.smcOpenConnection()
	if conn == 0 {
		return fmt.Errorf("failed to open SMC connection (requires com.apple.private.smc entitlement on Apple Silicon)")
	}
	smcConnection = conn
	return nil
}

func SMCClose() {
	if smcConnection != 0 {
		C.IOServiceClose(smcConnection)
		smcConnection = 0
	}
}

func SMCIsAvailable() bool {
	return ensureSMCConnection() == nil
}

func ReadSMCKey(key string) (float64, error) {
	if err := ensureSMCConnection(); err != nil {
		return 0, err
	}

	if len(key) != 4 {
		return 0, fmt.Errorf("SMC key must be exactly 4 characters: %s", key)
	}

	var input, output C.SMCParamStruct
	input.key = C.uint32_t(uint32(key[0])<<24 | uint32(key[1])<<16 | uint32(key[2])<<8 | uint32(key[3]))
	input.dataSize = 4

	result := C.smcCall(smcConnection, C.uint32_t(smcSelectorReadKey), &input, &output)
	if result != C.KERN_SUCCESS {
		return 0, fmt.Errorf("SMC call failed with code 0x%x", result)
	}
	if output.result != 0 {
		return 0, fmt.Errorf("SMC read failed with result 0x%x", output.result)
	}

	data := C.GoBytes(unsafe.Pointer(&output.data[0]), C.int(output.dataSize))
	value := decodeSMCFloat(data)
	return value, nil
}

func decodeSMCFloat(data []byte) float64 {
	if len(data) < 4 {
		return 0
	}

	sign := data[0] & 0x80
	exp := (int(data[0]&0x7F) << 1) | int(data[1]>>7)
	mant := int(data[3]) | (int(data[2]) << 8)

	if exp == 0 && mant == 0 {
		return 0
	}

	f := float64(mant) / 512.0
	f *= math.Pow(2, float64(exp-15))

	if sign != 0 {
		f = -f
	}
	return f
}

func GetCPUTemperature() (float64, error) {
	val, err := ReadSMCKey("TC0P")
	if err != nil {
		return 0, fmt.Errorf("CPU temperature unavailable: %w", err)
	}
	return val, nil
}

func GetGPUTemperature() (float64, error) {
	val, err := ReadSMCKey("TG0P")
	if err != nil {
		return 0, fmt.Errorf("GPU temperature unavailable: %w", err)
	}
	return val, nil
}

func GetFanRPMViaSMC(fanIndex int) (float64, error) {
	var key string
	switch fanIndex {
	case 0:
		key = "F0Ac"
	case 1:
		key = "F1Ac"
	default:
		return 0, fmt.Errorf("unsupported fan index: %d", fanIndex)
	}
	return ReadSMCKey(key)
}
