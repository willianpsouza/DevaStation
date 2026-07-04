package steps

import "devstation/internal/step"

// All returns every step in execution order. System-level steps run first so
// that later user-level steps (dotfiles, gnome) have their tools available.
func All() []step.Step {
	return []step.Step{
		SystemUpdate{},
		AptBase{},
		Golang{},
		Docker{},
		GH{},
		WireGuard{},
		Fish{},
		Starship{},
		Vim{},
		VSCode{},
		Claude{},
		ClaudeCode{},
		ModernCLI{},
		NetTools{},
		SSHConfig{},
		Git{},
		Tmux{},
		Gnome{},
	}
}
