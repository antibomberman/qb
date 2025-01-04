package tests

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestMigrateCreate(t *testing.T) {

	cmd := exec.Command("go", "run", "cli/main.go", "migrate", "create", "post")
	cmd.Dir = ".."
	fmt.Println(cmd.String())
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to execute command: %s", err)
	}

}
