package steps

import (
	"strings"

	"devastation/internal/step"
	"devastation/internal/system"
)

// SystemUpdate refreshes apt metadata, upgrades all packages (including the
// kernel via dist-upgrade) and cleans up. This covers the "small Ubuntu
// things" like landing on the latest kernel.
type SystemUpdate struct{}

func (SystemUpdate) ID() string    { return "system-update" }
func (SystemUpdate) Title() string { return "atualizar Ubuntu + kernel" }

func (SystemUpdate) Check(c *step.Context) (bool, error) {
	// Consider done only when there is nothing to upgrade.
	out, err := system.Output("bash", "-c",
		"apt-get -s dist-upgrade 2>/dev/null | grep -c '^Inst'")
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(out) == "0", nil
}

func (SystemUpdate) Apply(c *step.Context) error {
	if err := c.RunApt("update"); err != nil {
		return err
	}
	if err := c.RunApt("-y", "dist-upgrade"); err != nil {
		return err
	}
	// Ensure the generic kernel metapackage is present so future kernels flow in.
	if err := c.AptInstall("linux-generic"); err != nil {
		return err
	}
	if err := c.RunApt("-y", "autoremove"); err != nil {
		return err
	}
	if err := c.RunApt("-y", "autoclean"); err != nil {
		return err
	}
	// Report kernel state for visibility.
	if running, err := system.Output("uname", "-r"); err == nil {
		c.UI.Info("kernel em execução: %s", running)
	}
	if system.RebootRequired() {
		c.UI.Warn("novo kernel/base instalado — reboot recomendado ao final")
	}
	return nil
}
