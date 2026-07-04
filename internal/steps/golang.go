package steps

import (
	"fmt"
	"strings"

	"devstation/internal/step"
	"devstation/internal/system"
)

// Golang installs the official Go toolchain system-wide under /usr/local/go
// and exposes it via /etc/profile.d, matching the upstream install docs.
type Golang struct{}

func (Golang) ID() string    { return "golang" }
func (Golang) Title() string { return "toolchain Go (última estável)" }

const goRoot = "/usr/local/go"

// latestGo returns the current stable version tag, e.g. "go1.26.4".
func latestGo() (string, error) {
	out, err := system.Output("bash", "-c", "curl -fsSL 'https://go.dev/VERSION?m=text' | head -1")
	if err != nil {
		return "", fmt.Errorf("não consegui consultar a última versão do Go: %w", err)
	}
	v := strings.TrimSpace(out)
	if !strings.HasPrefix(v, "go") {
		return "", fmt.Errorf("resposta inesperada de go.dev/VERSION: %q", v)
	}
	return v, nil
}

func (Golang) Check(c *step.Context) (bool, error) {
	out, err := system.Output(goRoot+"/bin/go", "version")
	if err != nil {
		return false, nil // not installed
	}
	latest, err := latestGo()
	if err != nil {
		// Can't compare; treat an existing install as good enough.
		return true, nil
	}
	// `go version` → "go version go1.26.4 linux/amd64"
	installed := ""
	if f := strings.Fields(out); len(f) >= 3 {
		installed = f[2]
	}
	if installed == latest {
		c.UI.Skip("Go %s já instalado", installed)
		return true, nil
	}
	c.UI.Info("Go %s instalado; última é %s — vou atualizar", installed, latest)
	return false, nil
}

func (Golang) Apply(c *step.Context) error {
	ver, err := latestGo()
	if err != nil {
		return err
	}
	tarball := fmt.Sprintf("%s.linux-amd64.tar.gz", ver)
	url := "https://go.dev/dl/" + tarball
	c.UI.Info("baixando %s", url)

	script := fmt.Sprintf(`set -euo pipefail
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
curl -fsSL -o "$tmp/go.tgz" %q
rm -rf %q
tar -C /usr/local -xzf "$tmp/go.tgz"`, url, goRoot)
	if err := c.Sh(script); err != nil {
		return err
	}

	// System-wide PATH for interactive login shells.
	profile := "# Go toolchain (devstation)\nexport PATH=$PATH:/usr/local/go/bin\n"
	if !c.DryRun {
		if err := writeRoot("/etc/profile.d/go.sh", profile, 0o644); err != nil {
			return err
		}
	} else {
		c.UI.Cmd("would write:", "/etc/profile.d/go.sh")
	}
	c.UI.Info("Go %s instalado em %s", ver, goRoot)
	return nil
}
