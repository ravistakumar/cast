// Package cli wires cast's cobra commands.
package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ravistakumar/cast/internal/agent"
	"github.com/ravistakumar/cast/internal/config"
	"github.com/ravistakumar/cast/internal/emit"
	"github.com/ravistakumar/cast/internal/install"
	"github.com/ravistakumar/cast/internal/optimize"
	"github.com/ravistakumar/cast/internal/profile"
	"github.com/ravistakumar/cast/internal/skill"
	"github.com/ravistakumar/cast/internal/warn"
)

// New builds the root cast command.
func New() *cobra.Command {
	root := &cobra.Command{Use: "cast", Short: "Compile one SKILL.md into harness-native skills"}
	root.AddCommand(buildCmd(), checkCmd(), targetsCmd())
	return root
}

func targetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "targets",
		Short: "List supported target harnesses",
		RunE: func(cmd *cobra.Command, _ []string) error {
			for _, n := range profile.Names() {
				fmt.Fprintln(cmd.OutOrStdout(), n)
			}
			return nil
		},
	}
}

func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <skill-dir>",
		Short: "Validate a canonical SKILL.md against the spec",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := skill.Parse(args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "ok")
			return nil
		},
	}
}

func buildCmd() *cobra.Command {
	var targets []string
	var outdir string
	var doInstall, doOptimize bool
	var agentName string

	cmd := &cobra.Command{
		Use:   "build <skill-dir>",
		Short: "Compile a skill to harness-native output",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if len(targets) == 0 {
				targets = cfg.Targets
			}
			if outdir == "" {
				outdir = cfg.Outdir
			}
			if !cmd.Flags().Changed("optimize") {
				doOptimize = cfg.Optimize
			}
			if agentName == "" {
				agentName = cfg.Agent
			}

			s, err := skill.Parse(args[0]) // strict gate
			if err != nil {
				return err
			}

			var allWarns []warn.Warning
			for _, t := range targets {
				p, ok := profile.Get(t)
				if !ok {
					return fmt.Errorf("unknown target %q (see `cast targets`)", t)
				}
				files, warns := emit.Emit(s, p)

				if doOptimize && len(warns) > 0 {
					files, warns = applyOptimize(s, p, files, warns, agentName)
				}

				root := filepath.Join(outdir, t)
				if err := install.Install(files, root); err != nil {
					return err
				}
				if doInstall {
					live, err := install.DefaultRoot(t)
					if err != nil {
						allWarns = append(allWarns, warn.Warning{Harness: t, Severity: warn.Warn, Code: "install-skipped", Message: err.Error()})
					} else if err := install.Install(files, live); err != nil {
						allWarns = append(allWarns, warn.Warning{Harness: t, Severity: warn.Warn, Code: "install-skipped", Message: err.Error()})
					}
				}
				allWarns = append(allWarns, warns...)
			}

			if report := warn.Render(allWarns); report != "" {
				fmt.Fprint(cmd.ErrOrStderr(), report)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "cast: compiled %s to %d target(s)\n", s.Name, len(targets))
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&targets, "target", nil, "target harness(es); default from config")
	cmd.Flags().StringVar(&outdir, "outdir", "", "output directory (default from config)")
	cmd.Flags().BoolVar(&doInstall, "install", false, "also install into live harness dirs")
	cmd.Flags().BoolVar(&doOptimize, "optimize", false, "run the optional LLM adaptation pass")
	cmd.Flags().StringVar(&agentName, "agent", "", "agent CLI for --optimize: auto|claude|codex")
	return cmd
}

// applyOptimize rewrites the body via the agent CLI and re-emits the SKILL.md.
func applyOptimize(s *skill.Skill, p profile.Profile, files []profile.File, warns []warn.Warning, agentName string) ([]profile.File, []warn.Warning) {
	name := agentName
	if name == "" || name == "auto" {
		detected, err := agent.Detect()
		if err != nil {
			return files, append(warns, warn.Warning{Harness: p.Name(), Severity: warn.Warn, Code: "optimize-failed", Message: err.Error()})
		}
		name = detected
	}
	r, err := agent.NewRunner(name)
	if err != nil {
		return files, append(warns, warn.Warning{Harness: p.Name(), Severity: warn.Warn, Code: "optimize-failed", Message: err.Error()})
	}
	newBody, optWarns := optimize.Optimize(s, p, warns, r)
	adapted := *s
	adapted.Body = newBody
	// Re-emit with the (possibly adapted) body and keep any residual structural
	// warnings — a partial optimization may leave untranslatable tools in place.
	files, residual := emit.Emit(&adapted, p)
	return files, append(residual, optWarns...)
}
