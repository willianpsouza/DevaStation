package steps

import (
	"devstation/internal/step"
	"devstation/internal/system"
)

// aptBasePkgs are the build essentials and repo-management tools every other
// step relies on.
var aptBasePkgs = []string{
	"build-essential",
	"ca-certificates",
	"curl",
	"wget",
	"gnupg",
	"lsb-release",
	"apt-transport-https",
	"software-properties-common",
	"unzip",
	"jq",
}

// AptBase installs foundational packages.
type AptBase struct{}

func (AptBase) ID() string    { return "apt-base" }
func (AptBase) Title() string { return "pacotes base de build" }

func (AptBase) Check(c *step.Context) (bool, error) {
	for _, p := range aptBasePkgs {
		if !system.PackageInstalled(p) {
			return false, nil
		}
	}
	return true, nil
}

func (AptBase) Apply(c *step.Context) error {
	if err := c.RunApt("update"); err != nil {
		return err
	}
	return c.AptInstall(aptBasePkgs...)
}
