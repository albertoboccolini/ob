package git

import (
	"testing"
)

func TestIssueCommand(t *testing.T) {
	output, err := IssueCommand("echo", []string{"Hello\nWorld"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expected := []string{"Hello", "World"}
	if len(output) != len(expected) {
		t.Fatalf("Expected output length %d, got %d", len(expected), len(output))
	}
	for i, line := range expected {
		if output[i] != line {
			t.Errorf("Expected line %d to be %q, got %q", i, line, output[i])
		}
	}
}

func TestPushChanges_NonexistentPath(t *testing.T) {
	err := PushChanges("/nonexistent/path")
	t.Logf("Error returned: %v", err)
	t.Logf("Error is nil: %v", err == nil)

	if err == nil {
		t.Error("Expected error for nonexistent path, got nil")
	}
}
