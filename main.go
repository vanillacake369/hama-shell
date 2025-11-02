package main

import (
	"hama-shell/controller"
	"hama-shell/global"
	"hama-shell/view/inputs"
	"log"
)

func main() {
	// 데몬 프로세스 존재 여부에 따라
	// 데몬 프로세스 생성
	var isDaemonAlive bool = global.IsDaemonAlive()
	if !isDaemonAlive {
		if err := controller.StartDaemonSelf(); err != nil {
			log.Printf("failed to start daemon: %v", err)
		}
	}

	// CLI 명령어 전달
	inputs.Execute()
}
