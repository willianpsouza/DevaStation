package steps

import (
	"devastation/internal/step"
	"devastation/internal/system"
)

// Ghostty installs the Ghostty GPU terminal (via the mkasberg PPA, which
// publishes for current Ubuntu incl. 26.04) and writes a user config:
// font-size 16, 20% transparency (opacity 0.8) and a generous scrollback.
type Ghostty struct{}

func (Ghostty) ID() string    { return "ghostty" }
func (Ghostty) Title() string { return "Ghostty (terminal GPU) + config" }

func (Ghostty) Check(c *step.Context) (bool, error) {
	if !system.HasCommand("ghostty") {
		return false, nil
	}
	return fileHasMarker(c.Target.Home+"/.config/ghostty/config", managedMarker), nil
}

func (Ghostty) Apply(c *step.Context) error {
	if !system.HasCommand("ghostty") {
		repo := `set -euo pipefail
add-apt-repository -y ppa:mkasberg/ghostty-ubuntu`
		if err := c.Sh(repo); err != nil {
			return err
		}
		if err := c.RunApt("update"); err != nil {
			return err
		}
		if err := c.AptInstall("ghostty"); err != nil {
			return err
		}
	}

	// scrollback-limit is in BYTES (not lines). ~20MB comfortably holds well
	// over 20.000 lines even with heavy styling.
	cfg := managedMarker + `
# ~/.config/ghostty/config

font-size = 16

# 20% transparente (0.8 opaco). Requer compositor com suporte (GNOME/Wayland ok).
background-opacity = 0.8

# "20 mil linhas" de scroll — Ghostty conta em BYTES; 20MB segura isso com folga.
scrollback-limit = 20000000
`
	if err := c.WriteUserFile(c.Target, c.Target.Home+"/.config/ghostty/config", cfg, 0o644); err != nil {
		return err
	}
	c.UI.Info("Ghostty pronto — config: font 16, opacidade 0.8, scrollback 20MB")
	return nil
}
