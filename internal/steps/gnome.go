package steps

import (
	"fmt"
	"os"
	"strings"

	"devastation/internal/step"
	"devastation/internal/system"
)

// Gnome disables desktop animations / visual effects and applies a few
// developer-friendly GNOME tweaks. gsettings must run as the *logged-in user*
// against their session bus, even when the tool itself runs as root.
type Gnome struct{}

func (Gnome) ID() string    { return "gnome" }
func (Gnome) Title() string { return "GNOME: desabilitar animações + tweaks" }

// NeedsRoot is false: gsettings must run as the logged-in user, not root.
func (Gnome) NeedsRoot() bool { return false }

// gnomeSettings is the desired dconf state. Order does not matter.
var gnomeSettings = []struct{ schema, key, value string }{
	// Núcleo do pedido: matar animações e efeitos visuais.
	{"org.gnome.desktop.interface", "enable-animations", "false"},
	{"org.gnome.desktop.interface", "enable-hot-corners", "false"},
	// Menos distração / mais performance no desktop.
	{"org.gnome.desktop.interface", "gtk-enable-primary-paste", "false"},
	{"org.gnome.desktop.sound", "event-sounds", "false"},
	{"org.gnome.desktop.privacy", "remember-recent-files", "false"},
	// QoL de dev.
	{"org.gnome.desktop.interface", "clock-show-seconds", "true"},
	{"org.gnome.desktop.interface", "clock-show-weekday", "true"},
	{"org.gnome.desktop.interface", "show-battery-percentage", "true"},
	{"org.gnome.mutter", "attach-modal-dialogs", "false"},
	{"org.gnome.desktop.wm.preferences", "focus-mode", "'click'"},
}

// dbusEnv returns the env needed to reach the user's session bus from root.
func dbusEnv(u system.UserInfo) []string {
	return []string{
		fmt.Sprintf("DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/%d/bus", u.UID),
		"HOME=" + u.Home,
	}
}

// gsettingsGet reads a key as the target user (read-only; ignores dry-run).
func gsettingsGet(u system.UserInfo, schema, key string) (string, bool) {
	var out string
	var err error
	if os.Geteuid() == 0 && u.Name != "root" {
		args := []string{"-u", u.Name, "--", "env"}
		args = append(args, dbusEnv(u)...)
		args = append(args, "gsettings", "get", schema, key)
		out, err = system.Output("sudo", args...)
	} else {
		out, err = system.Output("gsettings", "get", schema, key)
	}
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(out), true
}

func (Gnome) Check(c *step.Context) (bool, error) {
	// If there is no session bus, we can't (and needn't) apply now.
	if _, err := os.Stat(fmt.Sprintf("/run/user/%d/bus", c.Target.UID)); err != nil {
		c.UI.Warn("sem session bus do usuário — pulo o GNOME (rode dentro da sessão gráfica)")
		return true, nil
	}
	for _, s := range gnomeSettings {
		cur, ok := gsettingsGet(c.Target, s.schema, s.key)
		if !ok {
			return false, nil // schema/key ausente → precisa aplicar (ou schema não existe)
		}
		if !equalSetting(cur, s.value) {
			return false, nil
		}
	}
	return true, nil
}

func (Gnome) Apply(c *step.Context) error {
	if _, err := os.Stat(fmt.Sprintf("/run/user/%d/bus", c.Target.UID)); err != nil {
		c.UI.Warn("sem session bus — rode este passo logado na sessão GNOME")
		return nil
	}
	var applied int
	for _, s := range gnomeSettings {
		// Só tenta se o schema existir, p/ não falhar em GNOME parcial.
		if cur, ok := gsettingsGet(c.Target, s.schema, s.key); !ok {
			_ = cur
			c.UI.Warn("schema %s ausente — pulando %s", s.schema, s.key)
			continue
		}
		if err := c.AsUser(c.Target, dbusEnv(c.Target), "gsettings", "set", s.schema, s.key, s.value); err != nil {
			c.UI.Warn("gsettings %s %s: %v", s.schema, s.key, err)
			continue
		}
		applied++
	}
	c.UI.Info("%d configurações do GNOME aplicadas (animações OFF)", applied)
	return nil
}

// equalSetting compares gsettings values, tolerating quote style differences.
func equalSetting(a, b string) bool {
	norm := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.Trim(s, "'\"")
		return s
	}
	return norm(a) == norm(b)
}
