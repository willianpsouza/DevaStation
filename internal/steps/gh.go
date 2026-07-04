package steps

import (
	"devstation/internal/step"
	"devstation/internal/system"
)

// GH installs GitHub's official CLI (gh) from their apt repository.
type GH struct{}

func (GH) ID() string    { return "gh" }
func (GH) Title() string { return "GitHub CLI (gh)" }

func (GH) Check(c *step.Context) (bool, error) {
	return system.HasCommand("gh"), nil
}

func (GH) Apply(c *step.Context) error {
	setup := `set -euo pipefail
install -m 0755 -d /etc/apt/keyrings
if [ ! -f /etc/apt/keyrings/githubcli-archive-keyring.gpg ]; then
  curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
    -o /etc/apt/keyrings/githubcli-archive-keyring.gpg
  chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg
fi
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
  > /etc/apt/sources.list.d/github-cli.list`
	if err := c.Sh(setup); err != nil {
		return err
	}
	if err := c.RunApt("update"); err != nil {
		return err
	}
	if err := c.AptInstall("gh"); err != nil {
		return err
	}
	c.UI.Info("gh instalado — autentique depois com: gh auth login")
	return nil
}
