package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
)

func ssh() (err error) {
	cmd := exec.Command("ssh", "limjihoon@127.0.0.1")
	stdin, _ := cmd.StdinPipe()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, "1026")
		fmt.Fprintln(stdin, "cd ~/")
		fmt.Fprintln(stdin, "pwd")
		fmt.Fprintln(stdin, "ls")
		// ToDo : exit 없이도 처리될 수 있도록 할 수는 없을까?
		//   1) 아마도 segment 끝에 무조건 exit 이 오도록 처리 ??
		//   2) delay 를 사용 ,,, ?
		fmt.Fprintln(stdin, "exit")
	}()

	return cmd.Run()
}

func ssh_with_pty() (err error) {
	// Force interactive bash shell with simpler prompt
	cmd := exec.Command("ssh", "-tt", "limjihoon@127.0.0.1", "bash", "--norc", "-i")

	// Start with a pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer ptmx.Close()

	// Buffer to accumulate output
	outputBuffer := make([]byte, 4096)
	passwordSent := false
	loginSuccess := false

	// Commands to execute in order
	commands := []string{
		"whoami",
		"pwd",
		"ls -la | head -5",
		"cd ~/dev/",
		"pwd",
		"ls | head -5",
		"echo 'All commands completed'",
	}
	commandIndex := 0

	// Handle I/O in a goroutine
	go func() {
		for {
			n, err := ptmx.Read(outputBuffer)
			if err != nil {
				return
			}

			output := string(outputBuffer[:n])

			// Always echo to console to see what's happening
			os.Stdout.Write(outputBuffer[:n])
			os.Stdout.Sync() // Force flush

			// Check for password prompt
			if !passwordSent && strings.Contains(strings.ToLower(output), "password:") {
				fmt.Println("\n[DEBUG] Password prompt detected, sending password...")
				time.Sleep(500 * time.Millisecond)
				ptmx.Write([]byte("1026\n"))
				passwordSent = true
			}

			// Check for successful login - look for bash prompt
			if passwordSent && !loginSuccess {
				// Bash prompt indicators (bash-3.2$ is the prompt we see)
				if strings.Contains(output, "bash-3.2$") {
					fmt.Println("\n[DEBUG] Bash shell ready!")
					loginSuccess = true

					// Send first command
					if commandIndex < len(commands) {
						fmt.Printf("[DEBUG] Sending command %d: %s\n", commandIndex+1, commands[commandIndex])
						ptmx.Write([]byte(commands[commandIndex] + "\n"))
						commandIndex++
					}
				}
			}

			// After login, wait for prompts to send next commands
			if loginSuccess && strings.Contains(output, "bash-3.2$") {
				if commandIndex < len(commands) {
					// Wait a bit to ensure prompt is fully ready
					time.Sleep(200 * time.Millisecond)

					fmt.Printf("[DEBUG] Sending command %d: %s\n", commandIndex+1, commands[commandIndex])
					ptmx.Write([]byte(commands[commandIndex] + "\n"))
					commandIndex++
				} else if commandIndex == len(commands) {
					// All commands sent, wait to see final output then exit
					fmt.Println("\n[DEBUG] All commands executed successfully!")
					time.Sleep(1000 * time.Millisecond)
					ptmx.Write([]byte("exit\n"))
					time.Sleep(500 * time.Millisecond)
					return
				}
			}
		}
	}()

	// Create a channel to signal when we're done
	done := make(chan error, 1)

	go func() {
		done <- cmd.Wait()
	}()

	// Wait for either completion or timeout
	select {
	case err := <-done:
		// If killed by us, that's expected
		if err != nil && strings.Contains(err.Error(), "signal: killed") {
			fmt.Println("[DEBUG] SSH process terminated as expected")
			return nil
		}
		return err
	case <-time.After(15 * time.Second):
		// Timeout - force kill
		fmt.Println("[DEBUG] Timeout reached, force killing SSH...")
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			fmt.Println("[DEBUG] Force kill didn't work, returning anyway")
		}
		return nil
	}
}
