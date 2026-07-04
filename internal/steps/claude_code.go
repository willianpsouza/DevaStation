package steps

import (
	"os"

	"devastation/internal/step"
	"devastation/internal/system"
)

// ClaudeCode installs the Claude Code CLI via the official native installer.
// It is user-local (~/.local/bin/claude) and needs no root — "não existe dev
// sem Claude". :)
type ClaudeCode struct{}

func (ClaudeCode) ID() string    { return "claude-code" }
func (ClaudeCode) Title() string { return "Claude Code CLI" }

// NeedsRoot is false: the native installer writes to the user's ~/.local.
func (ClaudeCode) NeedsRoot() bool { return false }

// claudeInstalled reports whether the claude binary is present.
func (ClaudeCode) claudeInstalled(c *step.Context) bool {
	if _, err := os.Stat(c.Target.Home + "/.local/bin/claude"); err == nil {
		return true
	}
	return system.HasCommand("claude")
}

// Check is true only when the binary exists AND fish already sees ~/.local/bin.
// The fish PATH line lives in Apply, so gating Check on the binary alone left
// machines where claude was pre-installed without the fish wiring.
func (c ClaudeCode) Check(ctx *step.Context) (bool, error) {
	fishCfg := ctx.Target.Home + "/.config/fish/config.fish"
	return c.claudeInstalled(ctx) && fileHasMarker(fishCfg, "claude-code path"), nil
}

func (c ClaudeCode) Apply(ctx *step.Context) error {
	if !c.claudeInstalled(ctx) {
		env := []string{"HOME=" + ctx.Target.Home}
		install := "curl -fsSL https://claude.ai/install.sh | bash"
		if err := ctx.AsUser(ctx.Target, env, "bash", "-c", install); err != nil {
			return err
		}
		ctx.UI.Info("Claude Code instalado em ~/.local/bin/claude")
	}
	// The installer wires PATH for bash/zsh; make sure fish sees ~/.local/bin too.
	fishCfg := ctx.Target.Home + "/.config/fish/config.fish"
	if !fileHasMarker(fishCfg, "claude-code path") {
		line := "# managed-by: devastation (claude-code path)\nfish_add_path -g $HOME/.local/bin\n"
		if err := appendUser(ctx, fishCfg, line); err != nil {
			ctx.UI.Warn("não consegui ajustar o PATH do fish: %v", err)
		}
	}
	return nil
}
