package steps

import (
	"fmt"
	"strings"

	"devastation/internal/step"
	"devastation/internal/system"
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

// fishPathWired reports whether config.fish already exposes the Go paths.
func fishPathWired(c *step.Context) bool {
	return fileHasMarker(c.Target.Home+"/.config/fish/config.fish", "golang path")
}

func (Golang) Check(c *step.Context) (bool, error) {
	out, err := system.Output(goRoot+"/bin/go", "version")
	if err != nil {
		return false, nil // not installed
	}
	latest, err := latestGo()
	if err != nil {
		// Can't compare; treat an existing install as good enough, but still
		// require the fish PATH to be wired.
		return fishPathWired(c), nil
	}
	// `go version` → "go version go1.26.4 linux/amd64"
	installed := ""
	if f := strings.Fields(out); len(f) >= 3 {
		installed = f[2]
	}
	if installed == latest {
		if !fishPathWired(c) {
			c.UI.Info("Go %s já instalado; falta expor o PATH no fish", installed)
			return false, nil
		}
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
	// Skip the download when the toolchain is already current — this Apply may
	// be running only to wire the fish PATH on an already-provisioned machine.
	if out, err := system.Output(goRoot+"/bin/go", "version"); err == nil &&
		strings.Contains(out, " "+ver+" ") {
		c.UI.Skip("Go %s já é a última — pulando download", ver)
	} else {
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
	}

	// System-wide PATH for interactive login shells (bash/zsh).
	profile := "# Go toolchain (devastation)\nexport PATH=$PATH:/usr/local/go/bin:$HOME/go/bin\n"
	if !c.DryRun {
		if err := writeRoot("/etc/profile.d/go.sh", profile, 0o644); err != nil {
			return err
		}
	} else {
		c.UI.Cmd("would write:", "/etc/profile.d/go.sh")
	}

	// fish does not source /etc/profile.d, so wire the Go paths into config.fish
	// explicitly: the toolchain (/usr/local/go/bin) and `go install` targets
	// (~/go/bin). fish_add_path is idempotent, but guard with a marker too.
	fishCfg := c.Target.Home + "/.config/fish/config.fish"
	if !fileHasMarker(fishCfg, "golang path") {
		line := managedMarker + " (golang path)\nfish_add_path -g /usr/local/go/bin $HOME/go/bin\n"
		if err := appendUser(c, fishCfg, line); err != nil {
			c.UI.Warn("não consegui ajustar o PATH do fish: %v", err)
		}
	}

	c.UI.Info("Go %s instalado em %s", ver, goRoot)
	return nil
}
