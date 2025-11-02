package global

import (
	"os"
	"path/filepath"
)

const (
	// ConfigDirName 설정 디렉토리 이름
	ConfigDirName = "hama-shell"

	// ProfileConfigFileName profile 설정 파일 이름
	ProfileConfigFileName = "profile.yaml"

	// DefaultStage 기본 스테이지 이름
	DefaultStage = "dev"

	// Flag names
	FlagProject      = "project"
	FlagProjectShort = "p"
	FlagStage        = "stage"
	FlagStageShort   = "s"
)

// GetConfigDir 설정 디렉토리 경로 반환 (~/.config/hama-shell/)
func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// fallback to current directory
		return "."
	}
	return filepath.Join(homeDir, ".config", ConfigDirName)
}

// GetProfileConfigPath profile.yaml 전체 경로 반환
func GetProfileConfigPath() string {
	return filepath.Join(GetConfigDir(), ProfileConfigFileName)
}
