package main

import (
	"hama-shell/controller"
	"hama-shell/view/inputs"
	"os"
)

func main() {
	// 데몬 모드로 실행되었다면
	// 데몬 함수 호출. CLI 명령어 전달 X
	for _, arg := range os.Args {
		if arg == "--daemonize" {
			controller.Daemonize()
			return
		}
	}

	// 데몬 모드가 아니라면
	// CLI 명령어 전달
	inputs.Execute()
}
