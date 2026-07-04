// Package step defines the Step contract and the runner that drives them.
package step

import (
	"fmt"

	"devstation/internal/system"
	"devstation/internal/ui"
)

// Config holds user-tunable knobs passed down to steps.
type Config struct {
	GitName     string
	GitEmail    string
	FishDefault bool // set fish as the login shell for the target user
}

// Context is handed to every step; it bundles execution + config + target user.
type Context struct {
	*system.Exec
	UI     *ui.UI
	Target system.UserInfo
	Cfg    Config
}

// rootAware lets a step declare it does NOT require root (default is: it does).
type rootAware interface{ NeedsRoot() bool }

// NeedsRoot reports whether a step requires root privileges. Steps that only
// touch the user's session/dotfiles implement rootAware and return false.
func NeedsRoot(s Step) bool {
	if r, ok := s.(rootAware); ok {
		return r.NeedsRoot()
	}
	return true
}

// AnyNeedsRoot reports whether at least one selected step requires root.
func AnyNeedsRoot(steps []Step) bool {
	for _, s := range steps {
		if NeedsRoot(s) {
			return true
		}
	}
	return false
}

// Step is one idempotent unit of provisioning.
type Step interface {
	// ID is the stable, lowercase identifier used by --skip/--only.
	ID() string
	// Title is a short human description.
	Title() string
	// Check reports whether the desired state already holds (skip Apply if true).
	Check(*Context) (bool, error)
	// Apply performs the change.
	Apply(*Context) error
}

// Result records what happened to a single step.
type Result struct {
	ID      string
	Skipped bool
	Err     error
}

// Runner executes a selected list of steps in order.
type Runner struct {
	Ctx   *Context
	Steps []Step
}

// Run drives each step: probe with Check, then Apply when needed.
func (r *Runner) Run() []Result {
	var results []Result
	for _, s := range r.Steps {
		r.Ctx.UI.Step(s.ID(), s.Title())
		res := Result{ID: s.ID()}

		done, err := s.Check(r.Ctx)
		if err != nil {
			r.Ctx.UI.Warn("check falhou (%v) — tentando aplicar mesmo assim", err)
		}
		if err == nil && done {
			r.Ctx.UI.Skip("já configurado")
			res.Skipped = true
			results = append(results, res)
			continue
		}

		if err := s.Apply(r.Ctx); err != nil {
			r.Ctx.UI.Fail("%v", err)
			res.Err = err
		} else {
			r.Ctx.UI.OK("ok")
		}
		results = append(results, res)
	}
	return results
}

// Summarize prints a final tally and returns the number of failures.
func (r *Runner) Summarize(results []Result) int {
	var applied, skipped, failed int
	for _, res := range results {
		switch {
		case res.Err != nil:
			failed++
		case res.Skipped:
			skipped++
		default:
			applied++
		}
	}
	r.Ctx.UI.Section("Resumo")
	r.Ctx.UI.Info("%d aplicados · %d já ok · %d falharam", applied, skipped, failed)
	if failed > 0 {
		for _, res := range results {
			if res.Err != nil {
				r.Ctx.UI.Fail("%s: %v", res.ID, res.Err)
			}
		}
	}
	if system.RebootRequired() {
		r.Ctx.UI.Warn("um REBOOT é necessário (kernel/base atualizados) — rode: sudo reboot")
	}
	fmt.Println()
	return failed
}
