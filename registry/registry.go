package registry

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	KEY_QUERY_VALUE = 0x0001
	KEY_READ        = 0x20019
	REG_SZ          = 1
)

var (
	modadvapi32         = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyEx    = modadvapi32.NewProc("RegOpenKeyExW")
	procRegQueryInfoKey = modadvapi32.NewProc("RegQueryInfoKeyW")
	procRegEnumValue    = modadvapi32.NewProc("RegEnumValueW")
	procRegEnumKeyEx    = modadvapi32.NewProc("RegEnumKeyExW")
)

func RegEnumSubKey(hKey syscall.Handle, index uint32) (string, error) {
	var name [256]uint16
	var nameLen uint32 = 256

	ret, _, err := procRegEnumKeyEx.Call(
		uintptr(hKey),
		uintptr(index),
		uintptr(unsafe.Pointer(&name[0])),
		uintptr(unsafe.Pointer(&nameLen)),
		0,
		0,
		0,
		0,
	)
	if ret != 0 {
		return "", err
	}

	subKeyName := syscall.UTF16ToString(name[:nameLen])
	return subKeyName, nil
}

func RegOpenKeyEx(hKey syscall.Handle, subKey string, options, desiredAccess uint32) (syscall.Handle, error) {
	var result syscall.Handle
	subKeyPtr, _ := syscall.UTF16PtrFromString(subKey)
	ret, _, err := procRegOpenKeyEx.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(subKeyPtr)),
		uintptr(options),
		uintptr(desiredAccess),
		uintptr(unsafe.Pointer(&result)),
	)
	if ret != 0 {
		return 0, err
	}
	return result, nil
}

func RegKeyHasContent(hKey syscall.Handle) (uint32, uint32, error) {
	var subKeys, values uint32
	ret, _, err := procRegQueryInfoKey.Call(
		uintptr(hKey),
		0, 0, 0,
		uintptr(unsafe.Pointer(&subKeys)),
		0, 0,
		uintptr(unsafe.Pointer(&values)),
		0, 0, 0, 0,
	)
	if ret != 0 {
		return 0, 0, err
	}
	return subKeys, values, nil
}

func RegEnumValue(hKey syscall.Handle, index uint32) (string, string, error) {
	var name [256]uint16
	var nameLen uint32 = 256
	var dataType uint32
	var data [1024]byte
	var dataLen uint32 = 1024

	ret, _, err := procRegEnumValue.Call(
		uintptr(hKey),
		uintptr(index),
		uintptr(unsafe.Pointer(&name[0])),
		uintptr(unsafe.Pointer(&nameLen)),
		0,
		uintptr(unsafe.Pointer(&dataType)),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(unsafe.Pointer(&dataLen)),
	)
	if ret != 0 {
		return "", "", err
	}

	valueName := syscall.UTF16ToString(name[:nameLen])
	var valueData string
	if dataType == REG_SZ {
		valueData = syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(&data))[:dataLen/2])
	} else {
		valueData = fmt.Sprintf("%v", data[:dataLen])
	}

	return valueName, valueData, nil
}
