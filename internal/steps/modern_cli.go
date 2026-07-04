package steps

import (
	"devastation/internal/step"
	"devastation/internal/system"
)

// ModernCLI installs modern terminal tooling: ripgrep, fd, bat, fzf and eza.
// On Ubuntu fd/bat ship as fdfind/batcat, so we add friendly symlinks.
type ModernCLI struct{}

func (ModernCLI) ID() string    { return "modern-cli" }
func (ModernCLI) Title() string { return "CLIs modernas (rg, fd, bat, eza, fzf)" }

var modernPkgs = []string{"ripgrep", "fd-find", "bat", "fzf"}

func (ModernCLI) Check(c *step.Context) (bool, error) {
	for _, p := range modernPkgs {
		if !system.PackageInstalled(p) {
			return false, nil
		}
	}
	if !system.HasCommand("eza") {
		return false, nil
	}
	return true, nil
}

func (ModernCLI) Apply(c *step.Context) error {
	if err := c.AptInstall(modernPkgs...); err != nil {
		return err
	}
	// Friendly names for Ubuntu's renamed binaries.
	link := `set -e
[ -x /usr/bin/fdfind ] && ln -sf /usr/bin/fdfind /usr/local/bin/fd || true
[ -x /usr/bin/batcat ] && ln -sf /usr/bin/batcat /usr/local/bin/bat || true`
	if err := c.Sh(link); err != nil {
		c.UI.Warn("symlinks fd/bat: %v", err)
	}

	// eza: try the distro repo first, then Gierens' official apt repo.
	if !system.HasCommand("eza") {
		if err := c.AptInstall("eza"); err != nil || !system.HasCommand("eza") {
			c.UI.Info("eza ausente no repo distro — adicionando repo oficial do eza")
			ezaRepo := `set -euo pipefail
install -m 0755 -d /etc/apt/keyrings
if [ ! -f /etc/apt/keyrings/gierens.gpg ]; then
  curl -fsSL https://raw.githubusercontent.com/eza-community/eza/main/deb.asc \
    | gpg --dearmor -o /etc/apt/keyrings/gierens.gpg
  chmod a+r /etc/apt/keyrings/gierens.gpg
fi
echo "deb [signed-by=/etc/apt/keyrings/gierens.gpg] http://deb.gierens.de stable main" \
  > /etc/apt/sources.list.d/gierens.list`
			if err := c.Sh(ezaRepo); err != nil {
				c.UI.Warn("não consegui configurar o repo do eza: %v", err)
				return nil
			}
			if err := c.RunApt("update"); err != nil {
				return err
			}
			if err := c.AptInstall("eza"); err != nil {
				c.UI.Warn("eza não instalado: %v", err)
			}
		}
	}
	return nil
}
