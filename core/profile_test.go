package core

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestSomething(t *testing.T) {
	// GIVEN
	pwd, _ := os.Getwd()
	file := pwd + "/../profile.yaml"
	f, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	// WHEN
	var config ProfileConfig
	if err = yaml.Unmarshal(f, &config); err != nil {
		log.Fatal(err)
	}

	// THEN
	assert.NotEmpty(t, config.Profiles, "프로파일 목록이 비어있으면 안됨")
	if len(config.Profiles) > 0 {
		fmt.Println(config.ToString())
	}
}
