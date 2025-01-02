package tests

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestMigrateCreate(t *testing.T) {
	startTime := time.Now()

	cmd := exec.Command("go", "run", "cmd/dbl/main.go", "migrate", "create", "post")
	cmd.Dir = ".."
	fmt.Println(cmd.String())
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to execute command: %s", err)
	}

	// Look for migration file created after startTime
	files, err := os.ReadDir("migrations")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %s", err)
	}

	found := false
	for _, f := range files {
		if strings.Contains(f.Name(), "test_migration.sql") {
			info, _ := f.Info()
			if info.ModTime().After(startTime) {
				found = true
				break
			}
		}
	}

	if !found {
		t.Fatal("Migration file not found")
	}
}
