# Registry Key Monitor

This Go script continuously monitors specified registry keys on a Windows system and reports any changes to those keys in real-time.

## Features

- Monitor specific registry keys or folders for changes.
- Real-time update of the registry key status with detailed output.
- Can detect registry key addition or deletion.

The script takes two command-line arguments: the registry hive (HKLM or HKCU) and the registry key or folder path to monitor. 

```sh
go run main.go <HIVE> "<Registry_Path>"
```

## Examples

#### Monitoring a Registry Folder:
```sh
go run main.go HKCU "Software\Microsoft\Windows\CurrentVersion\App Paths"
```

#### output
```css
[*] Registry Key is Found!
  \ SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths
      [-] SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\python3.11.exe
      [-] SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\wt.exe
      [-] SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\winget.exe
```

## Monitoring a Specific Registry Key:
```go
go run main.go HKCU "Software\Microsoft\Windows\CurrentVersion\App Paths\control.exe"
```

#### output
```css
[*] Registry Key is Found!
  [-] SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths
[ ] No Registry Key Found
```

#### Detecting Modifications:
```css
[^] Registry Key Deleted
  [ ] SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\python3.11.exe

[^] Registry Key Added
  [+] SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\python3.11.exe
```
