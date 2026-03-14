package state

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDirs(t *testing.T) (stateDir, configDir string) {
	t.Helper()
	tmp := t.TempDir()
	sd := filepath.Join(tmp, "state")
	cd := filepath.Join(tmp, "config", "ws-manager", "workspaces")
	os.MkdirAll(sd, 0755)
	os.MkdirAll(cd, 0755)
	t.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "state", ".."))
	// stateDir() uses XDG_STATE_HOME + "ws-manager", so set it one level up
	t.Setenv("XDG_STATE_HOME", tmp)
	// We need stateDir() to return sd — but stateDir appends "ws-manager"
	// So create the ws-manager subdir
	wsStateDir := filepath.Join(tmp, "ws-manager")
	os.MkdirAll(wsStateDir, 0755)
	t.Setenv("XDG_STATE_HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "config"))
	return wsStateDir, cd
}

func TestSaveAndLoad(t *testing.T) {
	setupTestDirs(t)

	ws := WorkspaceState{
		Name:     "test-ws",
		KittyPID: 12345,
		Active:   true,
		HomeX:    100,
		HomeY:    200,
	}

	if err := Save(ws); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load("test-ws")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Name != ws.Name {
		t.Errorf("expected name %q, got %q", ws.Name, loaded.Name)
	}
	if loaded.KittyPID != ws.KittyPID {
		t.Errorf("expected PID %d, got %d", ws.KittyPID, loaded.KittyPID)
	}
	if loaded.HomeX != 100 || loaded.HomeY != 200 {
		t.Errorf("expected home (100, 200), got (%d, %d)", loaded.HomeX, loaded.HomeY)
	}
	if !loaded.Active {
		t.Error("expected active=true")
	}
}

func TestLoadNonExistent(t *testing.T) {
	setupTestDirs(t)

	loaded, err := Load("does-not-exist")
	if err != nil {
		t.Fatalf("expected no error for missing state, got: %v", err)
	}
	if loaded.Name != "does-not-exist" {
		t.Errorf("expected name %q, got %q", "does-not-exist", loaded.Name)
	}
	if loaded.Active {
		t.Error("expected active=false for non-existent state")
	}
}

func TestSaveAndRemove(t *testing.T) {
	setupTestDirs(t)

	ws := WorkspaceState{Name: "removeme", Active: true}
	Save(ws)

	if err := Remove("removeme"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	loaded, err := Load("removeme")
	if err != nil {
		t.Fatalf("Load after remove failed: %v", err)
	}
	if loaded.Active {
		t.Error("expected inactive after removal")
	}
}

func TestHomeCapturedFlag(t *testing.T) {
	setupTestDirs(t)

	ws := WorkspaceState{
		Name:         "captured",
		Active:       true,
		HomeX:        0,
		HomeY:        0,
		HomeCaptured: true,
	}
	Save(ws)

	loaded, err := Load("captured")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !loaded.HomeCaptured {
		t.Error("expected HomeCaptured=true")
	}
	if loaded.HomeX != 0 || loaded.HomeY != 0 {
		t.Errorf("expected (0,0), got (%d,%d)", loaded.HomeX, loaded.HomeY)
	}
}

func TestFocusedState(t *testing.T) {
	setupTestDirs(t)

	if f := LoadFocused(); f != "" {
		t.Errorf("expected empty focused, got %q", f)
	}

	SaveFocused("my-ws")
	if f := LoadFocused(); f != "my-ws" {
		t.Errorf("expected focused %q, got %q", "my-ws", f)
	}

	SaveFocused("")
	if f := LoadFocused(); f != "" {
		t.Errorf("expected empty focused after clear, got %q", f)
	}
}

func TestRotateIndex(t *testing.T) {
	setupTestDirs(t)

	if idx := LoadRotateIndex(); idx != -1 {
		t.Errorf("expected -1 for missing index, got %d", idx)
	}

	SaveRotateIndex(3)
	if idx := LoadRotateIndex(); idx != 3 {
		t.Errorf("expected 3, got %d", idx)
	}

	SaveRotateIndex(0)
	if idx := LoadRotateIndex(); idx != 0 {
		t.Errorf("expected 0, got %d", idx)
	}
}

func TestListActive_CleansStaleState(t *testing.T) {
	sd, cd := setupTestDirs(t)

	// Create a workspace config for "real-ws" but not for "stale-ws"
	os.WriteFile(filepath.Join(cd, "real-ws.yaml"), []byte("name: real-ws\ndir: /tmp\n"), 0644)

	// Create state files for both
	realState := WorkspaceState{Name: "real-ws", Active: true, KittyPID: 1}
	staleState := WorkspaceState{Name: "stale-ws", Active: true, KittyPID: 2}
	Save(realState)
	Save(staleState)

	// Verify both state files exist
	if _, err := os.Stat(filepath.Join(sd, "stale-ws.yaml")); err != nil {
		t.Fatalf("stale state file should exist before ListActive")
	}

	active, err := ListActive()
	if err != nil {
		t.Fatalf("ListActive failed: %v", err)
	}

	// Only real-ws should be returned
	if len(active) != 1 {
		t.Fatalf("expected 1 active workspace, got %d", len(active))
	}
	if active[0].Name != "real-ws" {
		t.Errorf("expected active workspace %q, got %q", "real-ws", active[0].Name)
	}

	// Stale state file should have been cleaned up
	if _, err := os.Stat(filepath.Join(sd, "stale-ws.yaml")); !os.IsNotExist(err) {
		t.Error("stale state file should have been removed")
	}
}

func TestListActive_SkipsInactive(t *testing.T) {
	_, cd := setupTestDirs(t)

	os.WriteFile(filepath.Join(cd, "inactive-ws.yaml"), []byte("name: inactive-ws\ndir: /tmp\n"), 0644)

	ws := WorkspaceState{Name: "inactive-ws", Active: false}
	Save(ws)

	active, err := ListActive()
	if err != nil {
		t.Fatalf("ListActive failed: %v", err)
	}
	if len(active) != 0 {
		t.Errorf("expected 0 active workspaces, got %d", len(active))
	}
}

func TestListActive_EmptyStateDir(t *testing.T) {
	setupTestDirs(t)

	active, err := ListActive()
	if err != nil {
		t.Fatalf("ListActive failed: %v", err)
	}
	if len(active) != 0 {
		t.Errorf("expected 0, got %d", len(active))
	}
}
