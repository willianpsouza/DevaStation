package steps

import (
	"os"

	"devstation/internal/step"
	"devstation/internal/system"
)

// ClaudeCode installs the Claude Code CLI via the official native installer.
// It is user-local (~/.local/bin/claude) and needs no root — "não existe dev
// sem Claude". :)
type ClaudeCode struct{}

func (ClaudeCode) ID() string    { return "claude-code" }
func (ClaudeCode) Title() string { return "Claude Code CLI" }

// NeedsRoot is false: the native installer writes to the user's ~/.local.
func (ClaudeCode) NeedsRoot() bool { return false }

func (ClaudeCode) Check(c *step.Context) (bool, error) {
	if _, err := os.Stat(c.Target.Home + "/.local/bin/claude"); err == nil {
		return true, nil
	}
	return system.HasCommand("claude"), nil
}

func (ClaudeCode) Apply(c *step.Context) error {
	env := []string{"HOME=" + c.Target.Home}
	install := "curl -fsSL https://claude.ai/install.sh | bash"
	if err := c.AsUser(c.Target, env, "bash", "-c", install); err != nil {
		return err
	}
	// The installer wires PATH for bash/zsh; make sure fish sees ~/.local/bin too.
	fishCfg := c.Target.Home + "/.config/fish/config.fish"
	if !fileHasMarker(fishCfg, "claude-code path") {
		line := "# managed-by: devstation (claude-code path)\nfish_add_path -g $HOME/.local/bin\n"
		if err := appendUser(c, fishCfg, line); err != nil {
			c.UI.Warn("não consegui ajustar o PATH do fish: %v", err)
		}
	}
	c.UI.Info("Claude Code instalado em ~/.local/bin/claude")
	return nil
}
