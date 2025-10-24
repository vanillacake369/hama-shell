package controller

import "hama-shell/core"

func Attach() {
	commands := []string{
		"ssh limjihoon@127.0.0.1",
		"ls",
		"whoami",
		"neofetch",
	}
	core.ConnectTerminalEmulator(commands)
}
