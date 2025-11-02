package controller

import "hama-shell/core"

func Attach(commands []string) {
	// TODO
	// 	재접속 방지를 위해 접속 상태를 저장 및 관리해야함
	// 	1) 마지막 접속일자
	// 	2) 또 어떤 정보들을 ,,?
	core.ConnectTerminalEmulator(commands)
}
