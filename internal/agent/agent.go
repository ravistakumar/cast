// Package agent invokes an already-authenticated coding-agent CLI headlessly.
package agent

import (
	"fmt"
	"os/exec"
	"strings"
)

// Runner runs a prompt through an agent and returns its stdout.
type Runner interface {
	Run(prompt string) (string, error)
}

// commandFor builds the headless argv for a known agent; prompt is always last.
func commandFor(name, prompt string) ([]string, error) {
	switch name {
	case "claude":
		return []string{"claude", "-p", prompt}, nil
	case "codex":
		return []string{"codex", "exec", prompt}, nil
	default:
		return nil, fmt.Errorf("agent: unknown agent %q", name)
	}
}

type execRunner struct{ name string }

func (r execRunner) Run(prompt string) (string, error) {
	args, err := commandFor(r.name, prompt)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(args[0], args[1:]...) // #nosec G204 — fixed binary, prompt is a single arg
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("agent: %s failed: %w", r.name, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// NewRunner returns a Runner for the named agent.
func NewRunner(name string) (Runner, error) {
	if _, err := commandFor(name, ""); err != nil {
		return nil, err
	}
	return execRunner{name: name}, nil
}

// Detect returns the first agent CLI found on PATH (claude, then codex).
func Detect() (string, error) {
	for _, name := range []string{"claude", "codex"} {
		if _, err := exec.LookPath(name); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("agent: no supported agent CLI found on PATH")
}
