package steps

import (
	"fmt"
	"os"

	"devastation/internal/step"
	"devastation/internal/system"
)

// Ghostty installs the Ghostty GPU terminal (via the mkasberg PPA), installs
// JetBrains Mono, writes a user config (font, transparency, scrollback) and
// makes Ghostty the system default terminal.
type Ghostty struct{}

func (Ghostty) ID() string    { return "ghostty" }
func (Ghostty) Title() string { return "Ghostty (terminal GPU) + config + padrão" }

const ghosttyBin = "/usr/bin/ghostty"

const ghosttyKeybindPath = "/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/ghostty/"
const ghosttyKeybindSchema = "org.gnome.settings-daemon.plugins.media-keys.custom-keybinding:" + ghosttyKeybindPath

func (Ghostty) Check(c *step.Context) (bool, error) {
	if !system.HasCommand("ghostty") {
		return false, nil
	}
	if !fileHasMarker(c.Target.Home+"/.config/ghostty/config", managedMarker) {
		return false, nil
	}
	// Is Ghostty the default x-terminal-emulator?
	if out, _ := system.Output("readlink", "-f", "/etc/alternatives/x-terminal-emulator"); out != ghosttyBin {
		return false, nil
	}
	// Ctrl+Alt+T → ghostty? (only checkable with a live session bus)
	if _, err := os.Stat(fmt.Sprintf("/run/user/%d/bus", c.Target.UID)); err == nil {
		if v, ok := gsettingsGet(c.Target, ghosttyKeybindSchema, "command"); !ok || !equalSetting(v, "ghostty") {
			return false, nil
		}
	}
	return true, nil
}

func (Ghostty) Apply(c *step.Context) error {
	if !system.HasCommand("ghostty") {
		if err := c.Sh("set -euo pipefail\nadd-apt-repository -y ppa:mkasberg/ghostty-ubuntu"); err != nil {
			return err
		}
		if err := c.RunApt("update"); err != nil {
			return err
		}
		if err := c.AptInstall("ghostty"); err != nil {
			return err
		}
	}
	// JetBrains Mono font.
	if err := c.AptInstall("fonts-jetbrains-mono"); err != nil {
		c.UI.Warn("não consegui instalar fonts-jetbrains-mono: %v", err)
	}

	// User config. scrollback-limit is in BYTES; ~20MB holds well over 20k lines.
	cfg := managedMarker + `
# ~/.config/ghostty/config

font-family = JetBrains Mono
font-size = 16

# 5% transparente (0.95 opaco). Requer compositor com suporte (GNOME/Wayland ok).
background-opacity = 0.95

# "20 mil linhas" de scroll — Ghostty conta em BYTES; 20MB segura com folga.
scrollback-limit = 20000000
`
	if err := c.WriteUserFile(c.Target, c.Target.Home+"/.config/ghostty/config", cfg, 0o644); err != nil {
		return err
	}

	// Default terminal via Debian alternatives (used por Nautilus e afins).
	if err := c.Run("update-alternatives", "--install", "/usr/bin/x-terminal-emulator",
		"x-terminal-emulator", ghosttyBin, "50"); err != nil {
		c.UI.Warn("update-alternatives --install: %v", err)
	}
	if err := c.Run("update-alternatives", "--set", "x-terminal-emulator", ghosttyBin); err != nil {
		c.UI.Warn("update-alternatives --set: %v", err)
	}

	// GNOME bits, best-effort, run as the logged-in user against their bus.
	if _, err := os.Stat(fmt.Sprintf("/run/user/%d/bus", c.Target.UID)); err == nil {
		setg := func(schema, key, val string) {
			_ = c.AsUser(c.Target, dbusEnv(c.Target), "gsettings", "set", schema, key, val)
		}
		// Legacy default-terminal (some apps honor it).
		if _, ok := gsettingsGet(c.Target, "org.gnome.desktop.default-applications.terminal", "exec"); ok {
			setg("org.gnome.desktop.default-applications.terminal", "exec", "ghostty")
			setg("org.gnome.desktop.default-applications.terminal", "exec-arg", "-e")
		}
		// Ctrl+Alt+T → Ghostty: disable the built-in binding, add an explicit one.
		sd := "org.gnome.settings-daemon.plugins.media-keys"
		setg(sd, "terminal", "[]")
		setg(sd, "custom-keybindings", "['"+ghosttyKeybindPath+"']")
		setg(ghosttyKeybindSchema, "name", "Ghostty")
		setg(ghosttyKeybindSchema, "command", "ghostty")
		setg(ghosttyKeybindSchema, "binding", "<Primary><Alt>t")
		c.UI.Info("Ctrl+Alt+T mapeado para o Ghostty")
	}

	c.UI.Info("Ghostty: JetBrains Mono 16, opacidade 0.95, terminal padrão")
	return nil
}
