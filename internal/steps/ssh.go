package steps

import (
	"os"

	"devstation/internal/step"
)

// SSHConfig ensures ~/.ssh exists with correct perms and installs an aggressive
// keep-alive block at the TOP of ~/.ssh/config (ssh honors the first value seen,
// so prepending guarantees our keep-alive wins). It never clobbers an existing
// config — it validates and prepends our managed block only if absent.
type SSHConfig struct{}

func (SSHConfig) ID() string    { return "ssh-config" }
func (SSHConfig) Title() string { return "config SSH (keep-alive agressivo)" }

// NeedsRoot is false: only touches the user's ~/.ssh.
func (SSHConfig) NeedsRoot() bool { return false }

const sshKeepAlive = managedMarker + `
# keep-alive agressivo: derruba sessões mortas rápido e mantém as vivas de pé
Host *
    ServerAliveInterval 15
    ServerAliveCountMax 4
    TCPKeepAlive yes
    AddKeysToAgent yes
`

func (SSHConfig) Check(c *step.Context) (bool, error) {
	cfg := c.Target.Home + "/.ssh/config"
	return fileHasMarker(cfg, managedMarker), nil
}

func (SSHConfig) Apply(c *step.Context) error {
	sshDir := c.Target.Home + "/.ssh"
	cfg := sshDir + "/config"

	// Validate whether the user already has a local SSH config.
	existing := ""
	if b, err := os.ReadFile(cfg); err == nil {
		existing = string(b)
		c.UI.Info("config SSH já existe — vou preservar e só prepender o keep-alive")
	} else {
		c.UI.Info("nenhuma config SSH local — criando ~/.ssh/config")
	}

	if c.DryRun {
		c.UI.Cmd("would ensure:", sshDir+" (700)")
		c.UI.Cmd("would prepend keep-alive to:", cfg+" (600)")
		return nil
	}

	// ~/.ssh must be 0700 or ssh refuses to use it.
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return err
	}
	_ = os.Chown(sshDir, c.Target.UID, c.Target.GID)
	_ = os.Chmod(sshDir, 0o700)

	// Prepend our managed block; keep any pre-existing content intact below it.
	content := sshKeepAlive
	if existing != "" {
		content = sshKeepAlive + "\n" + existing
	}
	if err := os.WriteFile(cfg, []byte(content), 0o600); err != nil {
		return err
	}
	_ = os.Chown(cfg, c.Target.UID, c.Target.GID)
	_ = os.Chmod(cfg, 0o600)
	c.UI.Info("keep-alive SSH aplicado (ServerAliveInterval 15, CountMax 4)")
	return nil
}
