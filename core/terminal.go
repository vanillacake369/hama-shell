package core

import (
	"github.com/creack/pty"
	"golang.org/x/term"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func ConnectTerminalEmulator() {
	// PTY 쌍 생성
	ptyMaster, ptySlave, err := pty.Open()
	if err != nil {
		return
	}
	defer func(ptyMaster *os.File) {
		_ = ptyMaster.Close()
	}(ptyMaster)
	// PTY 크기를 현재 터미널 크기로 설정
	winSize, _ := pty.GetsizeFull(os.Stdin)
	_ = pty.Setsize(ptyMaster, winSize)

	// 사용자의 입력을 PTY 에 전달하고자 부모의 stdin 에 대해 RAW MODE 활성화
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// 자식 프로세스 생성
	// sid 를 지정하여 현재 세션으로부터 분리
	childProcess := exec.Command(os.Getenv("SHELL"))
	childProcess.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	childProcess.Stdin = ptySlave
	childProcess.Stdout = ptySlave
	childProcess.Stderr = ptySlave
	err = childProcess.Start()
	if err != nil {
		panic(err)
	}

	// 부모 프로세스는 PTY Slave 를 닫는다
	_ = ptySlave.Close()

	// PTY Master -> 현재 세션의 stdout (화면에 표시)
	// 현재 세션의 stdin -> PTY Master (입력을 PTY 에 전달하여 가상의 shell 에 전달)
	go func() {
		_, _ = io.Copy(os.Stdout, ptyMaster)
	}()
	go func() {
		_, _ = io.Copy(ptyMaster, os.Stdin)
	}()

	// TODO : 명령어를 전달 (테스트용)
	// Note: goroutine이 시작된 후 명령어를 전송해야 출력이 올바르게 표시됨
	// Shell 초기화를 위한 짧은 대기 시간
	commands := []string{
		"ls",
		"whoami",
		"neofetch",
	}
	go func() {
		// Shell 이 완전히 초기화될 때까지 대기
		// 대기하지 않으면 명령어 전달을 초기화 이전에 전달하게 되어
		// 부모에게 전달하게 될 수 있음
		time.Sleep(500 * time.Millisecond)

		for _, command := range commands {
			_, _ = io.Copy(ptyMaster, strings.NewReader(command+"\n"))
		}
	}()

	// 시그널을 통해 PTY 윈도우 크기 변경
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGWINCH)
	defer func() {
		signal.Stop(signalChan)
		close(signalChan)
	}()
	go func() {
		for sig := range signalChan {
			if sig == syscall.SIGWINCH {
				if winSize, err := pty.GetsizeFull(os.Stdin); err == nil {
					_ = pty.Setsize(ptyMaster, winSize)
				}
			}
		}
	}()

	// 자식 프로세스가 종료될 때까지 블로킹하며 대기
	err = childProcess.Wait()
}
