package controller

import (
	"fmt"
	"hama-shell/core"
)

/*
Daemonize 는
데몬 프로세스로써 수행되며
아래와 같은 책임을 수행한다.
- profile.yaml 변경 감시를 통해 profile.go 와 동기화
- 여러 서버 간 동일한 pty 를 공유하도록 ipc 처리
- TODO 1) 어떻게 하면 hama-shell 명령 시에 데몬프로세스와 명령어 처리 프로세스 두 개를 동시에 띄우지 ?
- TODO 2) 위 두 가지 책임 이외에 데몬 프로세스가 책임져야 할 부분이 있을까?
*/
func Daemonize() {
	// profile.yaml 변경에 따른 DTO 동기화
	core.SyncProfile()

	// TODO : config 변경점을 어떻게 elm arch 방식으로 detection 하지 ?
	fmt.Println("Daemonize() 함수 진입!!")
}
