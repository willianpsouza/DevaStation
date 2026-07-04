package steps

import (
	"devastation/internal/step"
	"devastation/internal/system"
)

// WireGuard installs the WireGuard VPN kernel tooling (wg, wg-quick).
type WireGuard struct{}

func (WireGuard) ID() string    { return "wireguard" }
func (WireGuard) Title() string { return "WireGuard VPN" }

func (WireGuard) Check(c *step.Context) (bool, error) {
	return system.HasCommand("wg") && system.HasCommand("wg-quick"), nil
}

func (WireGuard) Apply(c *step.Context) error {
	if err := c.AptInstall("wireguard", "wireguard-tools"); err != nil {
		return err
	}
	c.UI.Info("WireGuard instalado — configs em /etc/wireguard/*.conf, suba com: wg-quick up <iface>")
	return nil
}
