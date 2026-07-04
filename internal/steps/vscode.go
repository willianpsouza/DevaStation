package steps

import (
	"devstation/internal/step"
	"devstation/internal/system"
)

// vscodeDebURL is the pinned VS Code .deb chosen for this station.
// (Installing this .deb also configures Microsoft's apt repo, so future
// updates flow via `apt upgrade`.)
const vscodeDebURL = "https://vscode.download.prss.microsoft.com/dbazure/download/stable/4fe60c8b1cdac1c4c174f2fb180d0d758272d713/code_1.127.0-1782814776_amd64.deb"

// VSCode installs Visual Studio Code from the pinned .deb and adds the Go
// extension to the target user's profile.
type VSCode struct{}

func (VSCode) ID() string    { return "vscode" }
func (VSCode) Title() string { return "VS Code + extensão Go" }

func (VSCode) Check(c *step.Context) (bool, error) {
	return system.HasCommand("code"), nil
}

func (VSCode) Apply(c *step.Context) error {
	if !system.HasCommand("code") {
		if err := installDeb(c, vscodeDebURL); err != nil {
			return err
		}
	}
	// Go extension goes into the target user's profile, never root's.
	env := []string{"HOME=" + c.Target.Home}
	if err := c.AsUser(c.Target, env, "code", "--install-extension", "golang.go", "--force"); err != nil {
		c.UI.Warn("não consegui instalar a extensão golang.go: %v", err)
	} else {
		c.UI.Info("VS Code + extensão Go prontos")
	}
	return nil
}
