package main

import (
	"BlueEyE/registry"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 3 {
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
	var lastValues map[string]string

	for {
		hKey, err := registry.RegOpenKeyEx(syscall.Handle(hive), subKey, 0, registry.KEY_READ)
		var currentStatus string
		var currentValues map[string]string = make(map[string]string)

		if err != nil {
			currentStatus = "[ ] NO Registry Key Found"
		} else {
			subKeys, values, err := registry.RegKeyHasContent(hKey)
			if err != nil || (subKeys == 0 && values == 0) {
				currentStatus = "[ ] NO Registry Key Found"
			} else {
				currentStatus = fmt.Sprintf("[*] Registry Key Found\n \\ %s:%s", hiveArg, subKey)
				if values > 0 {
					// This is a registry key with values
					for i := uint32(0); i < values; i++ {
						valueName, valueData, err := registry.RegEnumValue(hKey, i)
						if err != nil {
							fmt.Printf("Error enumerating value: %s\n", err.Error())
						} else {
							currentValues[valueName] = valueData
							currentStatus += fmt.Sprintf("\n   %s: %s", valueName, valueData)
						}
					}
				} else {
					// This is a folder, enumerate subkeys
					for i := uint32(0); i < subKeys; i++ {
						subKeyName, err := registry.RegEnumSubKey(hKey, i)
						if err != nil {
							fmt.Printf("Error enumerating subkey: %s\n", err.Error())
						} else {
							currentStatus += fmt.Sprintf("\n   \\ %s:%s\\%s", hiveArg, subKey, subKeyName)
						}
					}
				}
			}
			syscall.RegCloseKey(hKey) // Make sure to close the handle
		}

		if currentStatus != lastStatus {
			fmt.Println(currentStatus)
			if lastValues != nil {
				for key, value := range lastValues {
					if currentValues[key] != value {
						fmt.Printf("   - Modified: %s = %s (was %s)\n", key, currentValues[key], value)
					}
				}
				for key := range currentValues {
					if lastValues[key] == "" {
						fmt.Printf("   - Added: %s = %s\n", key, currentValues[key])
					}
				}
				for key := range lastValues {
					if currentValues[key] == "" {
						fmt.Printf("   - Removed: %s\n", key)
					}
				}
			}
			lastStatus = currentStatus
			lastValues = currentValues
		}

		time.Sleep(3 * time.Second)
	}
}
