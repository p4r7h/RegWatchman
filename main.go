package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	KEY_QUERY_VALUE = 0x0001
	KEY_READ        = 0x20019
)

var (
	modadvapi32         = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyEx    = modadvapi32.NewProc("RegOpenKeyExW")
	procRegQueryInfoKey = modadvapi32.NewProc("RegQueryInfoKeyW")
	procRegQueryValueEx = modadvapi32.NewProc("RegQueryValueExW")
)

func RegOpenKeyEx(hKey syscall.Handle, subKey string, options, desiredAccess uint32) syscall.Handle {
	var result syscall.Handle
	subKeyPtr, _ := syscall.UTF16PtrFromString(subKey)
	procRegOpenKeyEx.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(subKeyPtr)),
		uintptr(options),
		uintptr(desiredAccess),
		uintptr(unsafe.Pointer(&result)),
	)
	return result
}

func RegKeyHasContent(hKey syscall.Handle) bool {
	var subKeys, values uint32
	procRegQueryInfoKey.Call(
		uintptr(hKey),
		0, 0, 0,
		uintptr(unsafe.Pointer(&subKeys)),
		0, 0,
		uintptr(unsafe.Pointer(&values)),
		0, 0, 0, 0,
	)
	return subKeys > 0 || values > 0
}

func RegQueryValueEx(hKey syscall.Handle, valueName string) (string, error) {
	valueNamePtr, _ := syscall.UTF16PtrFromString(valueName)
	var buf [1024]uint16
	var bufLen uint32 = 1024
	var valueType uint32
	ret, _, _ := procRegQueryValueEx.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valueNamePtr)),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		uintptr(unsafe.Pointer(&buf)),
		uintptr(unsafe.Pointer(&bufLen)),
	)
	if ret != 0 {
		return "", fmt.Errorf("error querying value: %d", ret)
	}
	value := syscall.UTF16ToString(buf[:bufLen/2])
	return value, nil
}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <hive> <registry_key> [value_name]", os.Args[0])
	}

	hiveArg := os.Args[1]
	subKey := os.Args[2]

	var valueName string
	if len(os.Args) == 4 {
		valueName = os.Args[3]
	}

	var hive uintptr
	switch hiveArg {
	case "HKLM":
		hive = uintptr(syscall.HKEY_LOCAL_MACHINE)
	case "HKCU":
		hive = uintptr(syscall.HKEY_CURRENT_USER)
	default:
		log.Fatalf("Invalid hive specified: %s", hiveArg)
	}

	var lastStatus string
	for {
		hKey := RegOpenKeyEx(syscall.Handle(hive), subKey, 0, KEY_READ)
		currentStatus := ""
		if hKey == 0 || !RegKeyHasContent(hKey) {
			currentStatus = "[ ] NO Registry Key Found"
		} else {
			if valueName != "" {
				value, err := RegQueryValueEx(hKey, valueName)
				if err != nil {
					currentStatus = fmt.Sprintf("[*] Registry Key Found\n    \\ %s:%s : Error querying registry value: %s", hiveArg, subKey, err.Error())
				} else {
					currentStatus = fmt.Sprintf("[*] Registry Key Found\n    \\ %s:%s : %s", hiveArg, subKey, value)
				}
			} else {
				currentStatus = fmt.Sprintf("[*] Registry Key Found\n    \\ %s:%s", hiveArg, subKey)
			}
			syscall.RegCloseKey(hKey) // Make sure to close the handle
		}

		if currentStatus != lastStatus {
			fmt.Println(currentStatus)
			lastStatus = currentStatus
		}

		time.Sleep(3 * time.Second)
	}
}
