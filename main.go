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
	modadvapi32 = syscall.NewLazyDLL("advapi32.dll")

	procRegOpenKeyEx    = modadvapi32.NewProc("RegOpenKeyExW")
	procRegQueryInfoKey = modadvapi32.NewProc("RegQueryInfoKeyW")
)

func RegOpenKeyEx(hKey syscall.Handle, subKey string, options, desiredAccess uint32) syscall.Handle {
	var result syscall.Handle
	subKeyPtr, _ := syscall.UTF16PtrFromString(subKey)
	procRegOpenKeyEx.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(subKeyPtr)),
		uintptr(options),
		uintptr(desiredAccess),
		uintptr(unsafe.Pointer(&result)))

	return result
}

func RegKeyHasContent(hKey syscall.Handle) bool {
	var subKeys, values uint32
	ret, _, _ := procRegQueryInfoKey.Call(
		uintptr(hKey),
		0, 0, 0,
		uintptr(unsafe.Pointer(&subKeys)),
		0, 0,
		uintptr(unsafe.Pointer(&values)),
		0, 0, 0, 0)
	return ret == 0 && (subKeys > 0 || values > 0)
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <hive> <registry_key>", os.Args[0])
	}

	hiveArg := os.Args[1]
	subKey := os.Args[2]

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
			currentStatus = "[*] No registry key found..."
		} else {
			currentStatus = "[*] Registry key found..."
			syscall.RegCloseKey(hKey) // Make sure to close the handle
		}

		if currentStatus != lastStatus {
			fmt.Println(currentStatus)
			lastStatus = currentStatus
		}

		time.Sleep(5 * time.Second) // Wait for 5 seconds before checking again
	}
}
