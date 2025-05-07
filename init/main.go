package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	// Handle signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	// Mount necessary filesystems
	// This isnt strictly necessary, but it's good practice to do it.
	// These filesystems are known as "virtual" or "pseudo" filesystems because they don't exist as physical files on the host.
	// They are created by the kernel and are used to provide a way to interact with the kernel and the system.
	mount("/dev", "devtmpfs")
	mount("/proc", "proc")
	mount("/sys", "sysfs")
	mount("/tmp", "tmpfs")
	mount("/run", "tmpfs")

	// Start reaping zombies
	go reapZombies()

	// Start the services
	if err := startServices("/etc/services", "/var/log"); err != nil {
		fmt.Printf("Failed to start processes: %v\n", err)
	}

	// Start a shell for us to interact with
	cmd := exec.Command("/bin/busybox", "sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// purely aesthetic, just to make it look nicer:)
	cmd.Env = append(os.Environ(), "PS1=[e2b-init-tutorial]\\$ ")
	cmd.Run()

	// Handle shutdown gracefully
	<-signals
	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

// reapZombies continuously waits for exited child processes (zombies) and reaps them.
// This prevents the accumulation of zombie processes, which can occur if the parent
// does not collect the exit status of its children. The inner loop calls Wait4 with
// WNOHANG to avoid blocking, and it reaps all available zombies in one pass.
// The outer loop runs indefinitely with a short sleep to avoid high CPU usage when no children exit.
func reapZombies() {
	for {
		var ws syscall.WaitStatus
		for {
			pid, err := syscall.Wait4(-1, &ws, syscall.WNOHANG, nil)
			if pid <= 0 || err != nil {
				break
			}
		}
		time.Sleep(time.Second)
	}
}

// mount mounts a filesystem on a target path.
func mount(target string, fstype string) {
	os.MkdirAll(target, 0755)
	syscall.Mount("none", target, fstype, 0, "")
}

// startServices starts all the services in the given directory.
// It reads all the files in the directory and starts them if they are executable.
// It also logs the output of the services to the given log directory.
func startServices(binaryDir string, logDir string) error {
	files, err := os.ReadDir(binaryDir)
	if err != nil {
		return err
	}

	// Iterate over all the files in the directory
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Get the full path of the file
		filePath := filepath.Join(binaryDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		// If the file is executable, start it and log its output
		if info.Mode().Perm()&0111 != 0 {
			go startAndLogProcess(filePath, filepath.Join(logDir, fmt.Sprintf("%s.log", file.Name())))
		}
	}

	return nil
}

// startAndLogProcess starts a process and logs its output to a file.
func startAndLogProcess(binaryPath string, logFilePath string) {
	// Open the log file for writing. If it doesn't exist, create it.
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil || logFile == nil {
		return
	}
	defer logFile.Close()

	// Create a new command to run the binary
	cmd := exec.Command(binaryPath)

	// Set the output of the command to the log file
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Start the process
	if err := cmd.Start(); err != nil {
		return
	}

	// Wait for the process to finish
	cmd.Wait()
}
