//go:build linux

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

const executableName = "p2s"

const platformSystemPrompt = `Convert the user's natural-language request into one Linux shell command.
Return JSON with exactly one string field named "command".
The command must work from the current directory unless the user asks otherwise.
Prefer standard Linux utilities, preserve paths and constraints precisely, and do not add sudo unless explicitly requested.
Return only the command itself in the field: no Markdown, explanation, prompt prefix, or newline.
If the request is ambiguous, choose the safest non-destructive interpretation.`

func configPaths() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find home directory: %w", err)
	}
	return []string{filepath.Join(home, ".config", "prompt2shell.conf"), "/etc/prompt2shell.conf"}, nil
}

func enableSingleKeyInput(reader io.Reader) (func(), error) {
	file, ok := reader.(*os.File)
	if !ok {
		return func() {}, nil
	}
	fd := file.Fd()
	var original syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&original)), 0, 0, 0)
	if errno != 0 {
		if errno == syscall.ENOTTY {
			return func() {}, nil
		}
		return nil, errno
	}
	modified := original
	modified.Lflag &^= syscall.ICANON | syscall.ECHO
	modified.Cc[syscall.VMIN] = 1
	modified.Cc[syscall.VTIME] = 0
	_, _, errno = syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&modified)), 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	return func() {
		_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&original)), 0, 0, 0)
	}, nil
}

func readConfirmation(reader io.Reader) (bool, error) {
	return readByteConfirmation(reader)
}

func executeCommand(command string, stdin io.Reader, stdout, stderr io.Writer) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	child := exec.Command(shell, "-c", command)
	child.Stdin, child.Stdout, child.Stderr = stdin, stdout, stderr
	return child.Run()
}
