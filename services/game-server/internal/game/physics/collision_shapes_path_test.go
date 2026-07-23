package physics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindSharedCollisionShapesPathUsesExecutableDirectory(t *testing.T) {
	packageRoot := t.TempDir()
	expected := filepath.Join(packageRoot, filepath.FromSlash(collisionShapesRelativePath))
	if err := os.MkdirAll(filepath.Dir(expected), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(expected, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	otherDirectory := t.TempDir()
	originalExecutablePath := executablePath
	originalWorkingDirectory := workingDirectory
	executablePath = func() (string, error) {
		return filepath.Join(packageRoot, "space-rocks-server.exe"), nil
	}
	workingDirectory = func() (string, error) {
		return otherDirectory, nil
	}
	t.Cleanup(func() {
		executablePath = originalExecutablePath
		workingDirectory = originalWorkingDirectory
	})

	actual, err := findSharedCollisionShapesPath()
	if err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("collision shapes path = %q, want %q", actual, expected)
	}
}

func TestFindCollisionShapesFromRootWalksUpward(t *testing.T) {
	root := t.TempDir()
	expected := filepath.Join(root, filepath.FromSlash(collisionShapesRelativePath))
	if err := os.MkdirAll(filepath.Dir(expected), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(expected, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	actual, found := findCollisionShapesFromRoot(filepath.Join(root, "nested", "deeper"))
	if !found {
		t.Fatal("collision shapes path was not found")
	}
	if actual != expected {
		t.Fatalf("collision shapes path = %q, want %q", actual, expected)
	}
}
