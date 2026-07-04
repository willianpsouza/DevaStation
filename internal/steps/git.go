package steps

import (
	"devastation/internal/step"
	"devastation/internal/system"
)

// Git installs git and applies a sensible global config for the target user.
type Git struct{}

func (Git) ID() string    { return "git" }
func (Git) Title() string { return "git + configuração global" }

func (Git) Check(c *step.Context) (bool, error) {
	if !system.HasCommand("git") {
		return false, nil
	}
	// Consider configured when user.email is already set. We must read the
	// TARGET user's ~/.gitconfig (via HOME), not root's, since we may run as root.
	out, _ := system.Output("env", "HOME="+c.Target.Home, "git",
		"config", "--global", "--get", "user.email")
	return out != "", nil
}

func (Git) Apply(c *step.Context) error {
	if err := c.AptInstall("git"); err != nil {
		return err
	}
	home := []string{"HOME=" + c.Target.Home}
	set := func(k, v string) {
		if v == "" {
			return
		}
		if err := c.AsUser(c.Target, home, "git", "config", "--global", k, v); err != nil {
			c.UI.Warn("git config %s: %v", k, err)
		}
	}
	set("user.name", c.Cfg.GitName)
	set("user.email", c.Cfg.GitEmail)
	set("init.defaultBranch", "main")
	set("pull.rebase", "false")
	set("core.editor", "vim")
	set("color.ui", "auto")
	set("push.autoSetupRemote", "true")
	set("alias.st", "status -sb")
	set("alias.lg", "log --oneline --graph --decorate --all")
	set("alias.co", "checkout")
	set("alias.br", "branch")
	c.UI.Info("git configurado p/ %s <%s>", c.Cfg.GitName, c.Cfg.GitEmail)
	return nil
}
