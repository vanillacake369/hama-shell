package executor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
	testCommands := []string{
		"cd ~/dev/tonys-nix/",
	}

	// WHEN
	// Test the RunSequence method
	err := exec.RunSequence(testKey, testCommands)
	if err != nil {
		t.Fatalf("Expected RunSequence to succeed, got error: %v", err)
	}

	// THEN
	// Verify the process is running
	status := exec.GetStatus()
	for k, v := range status {
		fmt.Println("Key:", k, "Value:", v)
	}
	assert.Len(t, status, 1)
}

func TestExecutor_MultipleRun_Success(t *testing.T) {
	// GIVEN
	// Create a new executor instance
	exec := New()
	if exec == nil {
		t.Fatal("Expected executor to be created, got nil")
	}
	testKey := "test.service"
	testCommands := []string{
		"cd ~/dev/tonys-nix/",
		"pwd",
		"ls -al",
		// "make help",
	}

	// WHEN
	// Test the RunSequence method
	err := exec.RunSequence(testKey, testCommands)
	if err != nil {
		t.Fatalf("Expected RunSequence to succeed, got error: %v", err)
	}

	// THEN
	// Verify the process is running
	status := exec.GetStatus()
	for k, v := range status {
		fmt.Println("Key:", k, "Value:", v)
	}
	assert.Len(t, status, 1)
}
