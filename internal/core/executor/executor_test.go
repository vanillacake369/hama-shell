package executor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestExecutor_Run_Success(t *testing.T) {
	// GIVEN
	// Create a new executor instance
	exec := New()
	if exec == nil {
		t.Fatal("Expected executor to be created, got nil")
	}
	testKey := "test.service"
	var testCommand string
	if runtime.GOOS == "windows" {
		testCommand = "echo test"
	} else {
		testCommand = "echo test"
	}

	// WHEN
	// Test the Run method
	err := exec.Run(testKey, testCommand)
	if err != nil {
		t.Fatalf("Expected Run to succeed, got error: %v", err)
	}

	// THEN
	// Verify the process is running
	status := exec.GetStatus()
	for k, v := range status {
		fmt.Println("Key:", k, "Value:", v)
	}
	assert.Len(t, status, 1)
}
