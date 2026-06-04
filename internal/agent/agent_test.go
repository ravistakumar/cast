package agent

import "testing"

func TestNewRunnerUnknown(t *testing.T) {
	if _, err := NewRunner("nope"); err == nil {
		t.Fatal("unknown agent should error")
	}
}

func TestCommandArgsPutPromptLast(t *testing.T) {
	args, err := commandFor("claude", "do the thing")
	if err != nil {
		t.Fatal(err)
	}
	if len(args) == 0 || args[len(args)-1] != "do the thing" {
		t.Fatalf("prompt must be the final argument: %v", args)
	}
	if args[0] != "claude" {
		t.Fatalf("expected claude binary first: %v", args)
	}
}
