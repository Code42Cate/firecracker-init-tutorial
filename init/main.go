package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	mount("/dev", "devtmpfs")
	mount("/proc", "proc")
	mount("/sys", "sysfs")
	mount("/tmp", "tmpfs")
	mount("/run", "tmpfs")

	go reapZombies()

	if err := startServices("/etc/services", "/var/log"); err != nil {
		fmt.Printf("Failed to start processes: %v\n", err)
	}

	cmd := exec.Command("/bin/busybox", "sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = append(os.Environ(), "PS1=[e2b-init-tutorial]\\$ ")
	cmd.Run()

	<-signals
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func reapZombies() {
	for {
		var ws syscall.WaitStatus
		_, err := syscall.Wait4(-1, &ws, syscall.WNOHANG, nil)
		if err != nil {
			continue
		}
	}
}

func mount(target string, fstype string) {
	os.MkdirAll(target, 0755)
	syscall.Mount("none", target, fstype, 0, "")
}

func startServices(binaryDir string, logDir string) error {
	files, err := os.ReadDir(binaryDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(binaryDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		if info.Mode().Perm()&0111 != 0 {
			go startAndLogProcess(filePath, filepath.Join(logDir, fmt.Sprintf("%s.log", file.Name())))
		}
	}

	return nil
}

func startAndLogProcess(binaryPath string, logFilePath string) {

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil || logFile == nil {
		return
	}
	defer logFile.Close()

	cmd := exec.Command(binaryPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		return
	}

	cmd.Wait()
}
