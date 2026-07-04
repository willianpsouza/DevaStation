package steps

import (
	"os/exec"
	"strings"

	"devastation/internal/step"
	"devastation/internal/system"
)

// Fish installs the fish shell. When Cfg.FishDefault is set it also makes fish
// the target user's login shell.
type Fish struct{}

func (Fish) ID() string    { return "fish" }
func (Fish) Title() string { return "fish shell" }

func (Fish) Check(c *step.Context) (bool, error) {
	if !system.HasCommand("fish") {
		return false, nil
	}
	if c.Cfg.FishDefault && !usesFishShell(c.Target.Name) {
		return false, nil
	}
	return true, nil
}

func (Fish) Apply(c *step.Context) error {
	if err := c.AptInstall("fish"); err != nil {
		return err
	}
	if c.Cfg.FishDefault {
		fishPath, err := exec.LookPath("fish")
		if err != nil {
			fishPath = "/usr/bin/fish"
		}
		if err := c.Run("chsh", "-s", fishPath, c.Target.Name); err != nil {
			c.UI.Warn("não consegui definir fish como shell padrão: %v", err)
		} else {
			c.UI.Info("fish definido como shell padrão de %s", c.Target.Name)
		}
	} else {
		c.UI.Info("fish instalado (use --fish-default p/ torná-lo o shell padrão)")
	}
	return nil
}

// usesFishShell checks /etc/passwd for the user's login shell.
func usesFishShell(user string) bool {
	out, err := system.Output("getent", "passwd", user)
	if err != nil {
		return false
	}
	return strings.HasSuffix(strings.TrimSpace(out), "/fish")
}
