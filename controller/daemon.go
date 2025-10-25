package controller

import "fmt"

/*
Daemonize 는 hama-shell 데몬 프로세스로써 수행되며 아래와 같은 책임을 수행한다.
- config 변경점을 detect 하여 (elm architecture 기반)
- profile 에 대한 pty 상태를 정적 파일로 관리 (~/.hama-shell/profile.json 으로 관리)
- profile 에 대해 여러 서버 간 동일한 pty 를 공유하도록 ipc 처리
*/
func Daemonize() {
	// TODO : config 변경점을 어떻게 elm arch 방식으로 detection 하지 ?
	fmt.Println("Daemonize() 함수 진입!!")
}
