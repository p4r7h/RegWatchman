// Registry Key Monitor
// Author: Parth Gol
// School: Computer Science and Engineering
// This script continuously monitors specified registry keys on a Windows system and reports any changes to those keys in real-time

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

	knownKeys := make(map[string]struct{})
	firstIteration := true
	lastOutput := ""

	for {
		hKey, err := registry.RegOpenKeyEx(syscall.Handle(hive), subKey, 0, registry.KEY_READ)
		currentOutput := ""
		if err != nil {
			currentOutput += fmt.Sprintf("[ ] No Registry Key Found\n  \\ %s\n", subKey)
		} else {
			subKeysCount, _, err := registry.RegKeyHasContent(hKey)
			if err != nil {
				log.Printf("Error querying registry key: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}

			currentKeys := make(map[string]struct{})
			for i := uint32(0); i < subKeysCount; i++ {
				subKeyName, err := registry.RegEnumSubKey(hKey, i)
				if err != nil {
					log.Printf("Error enumerating subkeys: %v", err)
					time.Sleep(3 * time.Second)
					continue
				}
				currentKeys[subKeyName] = struct{}{}
			}
			syscall.RegCloseKey(hKey)

			if firstIteration {
				if subKeysCount > 0 {
					currentOutput += fmt.Sprintf("[*] Registry Key is Found!\n  \\ %s\n", subKey)
				} else {
					currentOutput += fmt.Sprintf("[*] Registry Key is Found!\n  [-] %s\n", subKey)
				}
				for key := range currentKeys {
					knownKeys[key] = struct{}{}
					currentOutput += fmt.Sprintf("      [-] %s\\%s\n", subKey, key)
				}
				firstIteration = false
			} else {
				for key := range currentKeys {
					if _, found := knownKeys[key]; !found {
						currentOutput += fmt.Sprintf("[^] Registry Key Added\n  [+] %s\\%s\n", subKey, key)
						knownKeys[key] = struct{}{}
					}
				}

				for key := range knownKeys {
					if _, found := currentKeys[key]; !found {
						currentOutput += fmt.Sprintf("[^] Registry Key Deleted\n  [-] %s\\%s\n", subKey, key)
						delete(knownKeys, key)
					}
				}
			}
		}

		if currentOutput != lastOutput {
			fmt.Print(currentOutput)
			lastOutput = currentOutput
		}

		time.Sleep(3 * time.Second)
	}
}
