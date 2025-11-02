package global

func IsDaemonAlive() bool {
	// TODO
	// 		데몬이 살아있는지 없는지 체크를 어떻게 하지 ???
	// 		GPT 나 Gemini 나 /var/run 에 PID File Path 를 선언해두고 Lock 을 통해 데몬
	//		zellij 는 아래와 같이 처리함
	//		https://deepwiki.com/search/how-does-zellij-creates-a-daem_47913b9d-58e2-4c3f-a2d7-6e7b54b5b515?mode=fast
	return true
}
