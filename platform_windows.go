//go:build windows

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

const executableName = "p2s"

const platformSystemPrompt = `Convert the user's natural-language request into one Windows PowerShell command.
Return JSON with exactly one string field named "command".
The command must work from the current directory unless the user asks otherwise.
Use PowerShell syntax and built-in cmdlets, preserve paths and constraints precisely, and do not request elevation unless explicitly asked.
Return only the command itself in the field: no Markdown, explanation, prompt prefix, or newline.
If the request is ambiguous, choose the safest non-destructive interpretation.`

func configPaths() ([]string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return nil, errors.New("APPDATA is not set")
	}
	programData := os.Getenv("ProgramData")
	if programData == "" {
		programData = `C:\ProgramData`
	}
	return []string{
		filepath.Join(appData, "prompt2shell", "prompt2shell.conf"),
		filepath.Join(programData, "prompt2shell", "prompt2shell.conf"),
	}, nil
}

const (
	enableEchoInput = 0x0004
	enableLineInput = 0x0002
)

var (
	kernel32       = syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode = kernel32.NewProc("GetConsoleMode")
	setConsoleMode = kernel32.NewProc("SetConsoleMode")
)

func enableSingleKeyInput(reader io.Reader) (func(), error) {
	file, ok := reader.(*os.File)
	if !ok {
		return func() {}, nil
	}
	handle := syscall.Handle(file.Fd())
	var original uint32
	result, _, callErr := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&original)))
	if result == 0 {
		if callErr == syscall.Errno(6) {
			return func() {}, nil
		}
		return nil, fmt.Errorf("GetConsoleMode: %w", callErr)
	}
	modified := original &^ (enableEchoInput | enableLineInput)
	result, _, callErr = setConsoleMode.Call(uintptr(handle), uintptr(modified))
	if result == 0 {
		return nil, fmt.Errorf("SetConsoleMode: %w", callErr)
	}
	return func() { _, _, _ = setConsoleMode.Call(uintptr(handle), uintptr(original)) }, nil
}

func executeCommand(command string, stdin io.Reader, stdout, stderr io.Writer) error {
	shell, err := exec.LookPath("pwsh.exe")
	if err != nil {
		shell, err = exec.LookPath("powershell.exe")
		if err != nil {
			return errors.New("PowerShell not found (tried pwsh.exe and powershell.exe)")
		}
	}
	child := exec.Command(shell, "-NoLogo", "-NoProfile", "-Command", command)
	child.Stdin, child.Stdout, child.Stderr = stdin, stdout, stderr
	return child.Run()
}
