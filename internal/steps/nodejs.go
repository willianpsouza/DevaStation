package steps

import (
	"devstation/internal/step"
	"devstation/internal/system"
)

// NodeJS installs the current Node.js LTS from NodeSource's apt repository
// (includes npm). Needed by JS-based tooling like ruflo.
type NodeJS struct{}

func (NodeJS) ID() string    { return "nodejs" }
func (NodeJS) Title() string { return "Node.js LTS (NodeSource)" }

func (NodeJS) Check(c *step.Context) (bool, error) {
	return system.HasCommand("node") && system.HasCommand("npm"), nil
}

func (NodeJS) Apply(c *step.Context) error {
	// NodeSource's setup script adds the repo + runs apt-get update (needs root).
	if err := c.Sh("curl -fsSL https://deb.nodesource.com/setup_lts.x | bash -"); err != nil {
		return err
	}
	if err := c.AptInstall("nodejs"); err != nil {
		return err
	}
	if v, err := system.Output("node", "--version"); err == nil {
		c.UI.Info("Node.js %s + npm instalados", v)
	}
	return nil
}
