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
	enableEchoInput            = 0x0004
	enableLineInput            = 0x0002
	enableVirtualTerminalInput = 0x0200
	keyEvent                   = 0x0001
	virtualKeyReturn           = 0x000D
	virtualKeyEscape           = 0x001B
)

type inputRecord struct {
	eventType uint16
	_         uint16
	event     [16]byte
}

type keyEventRecord struct {
	keyDown         int32
	repeatCount     uint16
	virtualKeyCode  uint16
	virtualScanCode uint16
	unicodeChar     uint16
	controlKeyState uint32
}

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode   = kernel32.NewProc("GetConsoleMode")
	setConsoleMode   = kernel32.NewProc("SetConsoleMode")
	readConsoleInput = kernel32.NewProc("ReadConsoleInputW")
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
	// Virtual terminal input treats Esc as the start of an escape sequence,
	// which can leave a lone Esc waiting for more input. Disable it while
	// reading the confirmation key so Esc is delivered immediately.
	modified := original &^ (enableEchoInput | enableLineInput | enableVirtualTerminalInput)
	result, _, callErr = setConsoleMode.Call(uintptr(handle), uintptr(modified))
	if result == 0 {
		return nil, fmt.Errorf("SetConsoleMode: %w", callErr)
	}
	return func() { _, _, _ = setConsoleMode.Call(uintptr(handle), uintptr(original)) }, nil
}

func readConfirmation(reader io.Reader) (bool, error) {
	file, ok := reader.(*os.File)
	if !ok {
		return readByteConfirmation(reader)
	}
	handle := syscall.Handle(file.Fd())
	var mode uint32
	result, _, _ := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if result == 0 {
		return readByteConfirmation(reader)
	}
	for {
		var record inputRecord
		var recordsRead uint32
		result, _, callErr := readConsoleInput.Call(
			uintptr(handle),
			uintptr(unsafe.Pointer(&record)),
			1,
			uintptr(unsafe.Pointer(&recordsRead)),
		)
		if result == 0 {
			return false, fmt.Errorf("ReadConsoleInputW: %w", callErr)
		}
		if recordsRead == 0 || record.eventType != keyEvent {
			continue
		}
		key := (*keyEventRecord)(unsafe.Pointer(&record.event[0]))
		if key.keyDown == 0 {
			continue
		}
		switch key.virtualKeyCode {
		case virtualKeyReturn:
			return true, nil
		case virtualKeyEscape:
			return false, nil
		default:
			return false, nil
		}
	}
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
