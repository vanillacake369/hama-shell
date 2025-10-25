package core

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

func ConnectTerminalEmulator(commands []string) {
	ptyMaster, ptySlave := initializePTY()
	defer func(ptyMaster *os.File) {
		_ = ptyMaster.Close()
	}(ptyMaster)

	oldState := enableRawMode()
	defer func() {
		if oldState != nil {
			_ = term.Restore(int(os.Stdin.Fd()), oldState)
		}
	}()

	childProcess := startShellProcess(ptySlave)
	_ = ptySlave.Close()

	executeCommands(ptyMaster, commands)
	setupIOStreaming(ptyMaster)
	setupWindowResizeHandler(ptyMaster)

	_ = childProcess.Wait()
}

// initializePTY PTY 쌍을 생성하고 설정
func initializePTY() (*os.File, *os.File) {
	ptyMaster, ptySlave, err := pty.Open()
	if err != nil {
		panic(err)
	}

	winSize, _ := pty.GetsizeFull(os.Stdin)
	_ = pty.Setsize(ptyMaster, winSize)

	return ptyMaster, ptySlave
}

// enableRawMode 사용자의 입력을 PTY 에 직접 전달하기 위해 stdin 을 RAW MODE 로 활성화
func enableRawMode() *term.State {
	// stdin 이 터미널인지 확인
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// stdin 이 터미널이 아니면 nil 반환 (raw mode 불필요)
		return nil
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	return oldState
}

// startShellProcess PTY 에 연결된 자식 쉘 프로세스를 생성하고 시작
func startShellProcess(ptySlave *os.File) *exec.Cmd {
	childProcess := exec.Command(os.Getenv("SHELL"))
	childProcess.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	childProcess.Stdin = ptySlave
	childProcess.Stdout = ptySlave
	childProcess.Stderr = ptySlave

	err := childProcess.Start()
	if err != nil {
		panic(err)
	}

	return childProcess
}

// executeCommands 초기화 지연 후 쉘에 명령어를 전송
func executeCommands(ptyMaster *os.File, commands []string) {
	go func() {
		// Shell 이 완전히 초기화될 때까지 대기
		// 대기하지 않으면 명령어 전달을 초기화 이전에 전달하게 되어
		// 부모에게 전달하게 될 수 있음
		time.Sleep(500 * time.Millisecond)

		for _, command := range commands {
			_, _ = io.Copy(ptyMaster, strings.NewReader(command+"\n"))
			// TODO
			//		명령어가 PTY 에 너무 빨리 보내져서 부모가 출력해버리는 경우가 발생함
			//		이에 따라 처리한 명령어가 완전히 성공해야 다음 명령어가 처리되도록 해야함
			//		ptyMaster 의 stdout 이 idle 하면 성공했다고 판단하면 되지 않을까?
		}
	}()
}

// setupIOStreaming PTY 와 현재 세션 간의 양방향 I/O 설정
func setupIOStreaming(ptyMaster *os.File) {
	// PTY Master -> 현재 세션의 stdout (화면에 표시)
	go func() {
		_, _ = io.Copy(os.Stdout, ptyMaster)
	}()

	// 현재 세션의 stdin -> PTY Master (입력을 PTY 에 전달하여 가상의 shell 에 전달)
	go func() {
		_, _ = io.Copy(ptyMaster, os.Stdin)
	}()
}

// setupWindowResizeHandler 터미널 윈도우 크기 변경 시그널 처리
func setupWindowResizeHandler(ptyMaster *os.File) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGWINCH)

	go func() {
		defer func() {
			signal.Stop(signalChan)
			close(signalChan)
		}()

		for sig := range signalChan {
			if sig == syscall.SIGWINCH {
				if winSize, err := pty.GetsizeFull(os.Stdin); err == nil {
					_ = pty.Setsize(ptyMaster, winSize)
				}
			}
		}
	}()
}
