package monitor

import (
	"testing"
)

func TestParseDisplayConfig_SinglePrimary(t *testing.T) {
	// Simplified Mutter output with one primary monitor
	input := `(uint32 2, [('DP-0', 'Dell', 'U2720Q', '1234')], [(0, 0, 1.0, uint32 0, true, [('DP-0', 'Dell', 'U2720Q', '1234')], {})], {})`

	monitors, err := parseDisplayConfig(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}
	m := monitors[0]
	if m.Connector != "DP-0" {
		t.Errorf("expected connector DP-0, got %q", m.Connector)
	}
	if m.X != 0 || m.Y != 0 {
		t.Errorf("expected position (0, 0), got (%d, %d)", m.X, m.Y)
	}
	if !m.Primary {
		t.Error("expected primary=true")
	}
}

func TestParseDisplayConfig_MultiMonitor(t *testing.T) {
	// Three monitors: primary at origin, secondary to the right, third further right
	input := `(uint32 3, [('DP-0', 'Dell', 'U2720Q', '1'), ('DP-1', 'LG', 'UL850', '2'), ('DP-2', 'Samsung', 'C34', '3')], [(0, 0, 1.0, uint32 0, true, [('DP-0', 'Dell', 'U2720Q', '1')], {}), (2560, 0, 1.0, uint32 0, false, [('DP-1', 'LG', 'UL850', '2')], {}), (5120, 0, 1.0, uint32 0, false, [('DP-2', 'Samsung', 'C34', '3')], {})], {})`

	monitors, err := parseDisplayConfig(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 3 {
		t.Fatalf("expected 3 monitors, got %d", len(monitors))
	}

	tests := []struct {
		connector string
		x, y      int
		primary   bool
	}{
		{"DP-0", 0, 0, true},
		{"DP-1", 2560, 0, false},
		{"DP-2", 5120, 0, false},
	}
	for i, tt := range tests {
		m := monitors[i]
		if m.Connector != tt.connector {
			t.Errorf("[%d] expected connector %q, got %q", i, tt.connector, m.Connector)
		}
		if m.X != tt.x || m.Y != tt.y {
			t.Errorf("[%d] expected position (%d, %d), got (%d, %d)", i, tt.x, tt.y, m.X, m.Y)
		}
		if m.Primary != tt.primary {
			t.Errorf("[%d] expected primary=%v, got %v", i, tt.primary, m.Primary)
		}
	}
}

func TestParseDisplayConfig_VerticalStack(t *testing.T) {
	// Two monitors stacked vertically
	input := `(uint32 2, [('HDMI-1', 'Dell', 'P2419H', '1'), ('eDP-1', 'BOE', 'Panel', '2')], [(0, 0, 1.0, uint32 0, false, [('HDMI-1', 'Dell', 'P2419H', '1')], {}), (0, 1080, 1.0, uint32 0, true, [('eDP-1', 'BOE', 'Panel', '2')], {})], {})`

	monitors, err := parseDisplayConfig(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 2 {
		t.Fatalf("expected 2 monitors, got %d", len(monitors))
	}

	if monitors[0].Connector != "HDMI-1" || monitors[0].Y != 0 {
		t.Errorf("first monitor: expected HDMI-1 at y=0, got %q at y=%d", monitors[0].Connector, monitors[0].Y)
	}
	if monitors[1].Connector != "eDP-1" || monitors[1].Y != 1080 {
		t.Errorf("second monitor: expected eDP-1 at y=1080, got %q at y=%d", monitors[1].Connector, monitors[1].Y)
	}
	if monitors[0].Primary {
		t.Error("first monitor should not be primary")
	}
	if !monitors[1].Primary {
		t.Error("second monitor should be primary")
	}
}

func TestParseDisplayConfig_DoubleQuotes(t *testing.T) {
	// Some gdbus versions use double quotes instead of single
	input := `(uint32 1, [("DP-0", "Dell", "U2720Q", "1234")], [(0, 0, 1.0, uint32 0, true, [("DP-0", "Dell", "U2720Q", "1234")], {})], {})`

	monitors, err := parseDisplayConfig(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(monitors))
	}
	if monitors[0].Connector != "DP-0" {
		t.Errorf("expected connector DP-0, got %q", monitors[0].Connector)
	}
}

func TestParseDisplayConfig_Empty(t *testing.T) {
	_, err := parseDisplayConfig("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseDisplayConfig_NegativeCoordinates(t *testing.T) {
	// Monitor to the left of primary has negative X
	input := `(uint32 2, [('DP-0', 'Dell', 'A', '1'), ('DP-1', 'LG', 'B', '2')], [(-2560, 0, 1.0, uint32 0, false, [('DP-0', 'Dell', 'A', '1')], {}), (0, 0, 1.0, uint32 0, true, [('DP-1', 'LG', 'B', '2')], {})], {})`

	monitors, err := parseDisplayConfig(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(monitors) != 2 {
		t.Fatalf("expected 2 monitors, got %d", len(monitors))
	}
	if monitors[0].X != -2560 {
		t.Errorf("expected X=-2560, got %d", monitors[0].X)
	}
}
