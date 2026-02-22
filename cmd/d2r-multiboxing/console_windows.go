package main

import (
	"golang.org/x/sys/windows"
)

var (
	kernel32            = windows.NewLazySystemDLL("kernel32.dll")
	procSetConsoleCP    = kernel32.NewProc("SetConsoleCP")
	procSetConsoleOutCP = kernel32.NewProc("SetConsoleOutputCP")
)

const cpUTF8 = 65001

func init() {
	// 設定控制台輸入輸出為 UTF-8，避免中文在 Windows 控制台顯示亂碼
	procSetConsoleCP.Call(cpUTF8)
	procSetConsoleOutCP.Call(cpUTF8)
}
