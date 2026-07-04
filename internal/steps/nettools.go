package steps

import (
	"devstation/internal/step"
	"devstation/internal/system"
)

// NetTools installs handy network diagnostics: nmap, mtr, fping, hping3.
type NetTools struct{}

func (NetTools) ID() string    { return "nettools" }
func (NetTools) Title() string { return "ferramentas de rede (nmap, mtr, fping, hping3)" }

// netToolsPkgs → binários: nmap, mtr, fping, hping3.
var netToolsPkgs = []string{"nmap", "mtr-tiny", "fping", "hping3"}

func (NetTools) Check(c *step.Context) (bool, error) {
	for _, cmd := range []string{"nmap", "mtr", "fping", "hping3"} {
		if !system.HasCommand(cmd) {
			return false, nil
		}
	}
	return true, nil
}

func (NetTools) Apply(c *step.Context) error {
	if err := c.AptInstall(netToolsPkgs...); err != nil {
		return err
	}
	c.UI.Info("ferramentas de rede instaladas: nmap, mtr, fping, hping3")
	return nil
}
