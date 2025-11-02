package controller

import "hama-shell/core"

func Attach(commands []string) {
	core.ConnectTerminalEmulator(commands)
}
