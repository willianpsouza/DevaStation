package steps

import (
	"devstation/internal/step"
	"devstation/internal/system"
)

// Starship installs the starship prompt and wires it into both bash and fish
// for the target user.
type Starship struct{}

func (Starship) ID() string    { return "starship" }
func (Starship) Title() string { return "starship prompt" }

const (
	starshipBashInit = "# managed-by: devstation\neval \"$(starship init bash)\"\n"
	starshipFishInit = "# managed-by: devstation\nstarship init fish | source\n"
)

func (Starship) Check(c *step.Context) (bool, error) {
	if !system.HasCommand("starship") {
		return false, nil
	}
	bashrc := c.Target.Home + "/.bashrc"
	fishCfg := c.Target.Home + "/.config/fish/config.fish"
	if !fileHasMarker(bashrc, "starship init bash") {
		return false, nil
	}
	if !fileHasMarker(fishCfg, "starship init fish") {
		return false, nil
	}
	return true, nil
}

func (Starship) Apply(c *step.Context) error {
	// Install the prebuilt binary to /usr/local/bin via the official script.
	if !system.HasCommand("starship") {
		install := `set -euo pipefail
curl -fsSL https://starship.rs/install.sh | sh -s -- --yes --bin-dir /usr/local/bin`
		if err := c.Sh(install); err != nil {
			return err
		}
	}

	// Wire bash (append once).
	bashrc := c.Target.Home + "/.bashrc"
	if !fileHasMarker(bashrc, "starship init bash") {
		if err := appendUser(c, bashrc, "\n"+starshipBashInit); err != nil {
			return err
		}
	}
	// Wire fish (append once).
	fishCfg := c.Target.Home + "/.config/fish/config.fish"
	if !fileHasMarker(fishCfg, "starship init fish") {
		if err := appendUser(c, fishCfg, starshipFishInit); err != nil {
			return err
		}
	}
	c.UI.Info("starship ativado no bash e no fish")
	return nil
}
