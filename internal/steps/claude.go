package steps

import (
	"devastation/internal/step"
	"devastation/internal/system"
)

// claudeDebURL is the pinned Claude .deb chosen for this station.
const claudeDebURL = "https://downloads.claude.ai/releases/linux/x64/1.18286.0/Claude-259c3fc2a647e4222ca15e99bba9b2649f31f051.deb"

// Claude installs Anthropic's Claude package from the pinned .deb.
type Claude struct{}

func (Claude) ID() string    { return "claude" }
func (Claude) Title() string { return "Claude (.deb oficial)" }

func (Claude) Check(c *step.Context) (bool, error) {
	// Done when any dpkg package starting with "claude" is installed.
	out, _ := system.Output("bash", "-c",
		"dpkg-query -W -f='${Package}\n' 2>/dev/null | grep -qi '^claude' && echo yes")
	return out == "yes", nil
}

func (Claude) Apply(c *step.Context) error {
	if err := installDeb(c, claudeDebURL); err != nil {
		return err
	}
	c.UI.Info("pacote Claude instalado")
	return nil
}
