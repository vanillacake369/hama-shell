package controller

import (
	"hama-shell/core"
)

/*
StartDaemonSelf 는
데몬 프로세스로써 수행되며
아래와 같은 책임을 수행한다.
- profile.yaml 변경 감시를 통해 profile.go 와 동기화
- 여러 서버 간 동일한 pty 를 공유하도록 ipc 처리
*/
func StartDaemonSelf() error {
	// TODO : 어떻게 하면 fork 를 떠서 데몬 프로세스를 처리하지 ?

	// profile.yaml 변경에 따른 DTO 동기화
	core.SyncProfile()

	return nil
}
