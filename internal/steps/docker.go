package steps

import (
	"devastation/internal/step"
	"devastation/internal/system"
)

// Docker installs Docker CE from Docker's official apt repository, including
// the compose plugin, and adds the target user to the docker group so they can
// run docker without sudo.
type Docker struct{}

func (Docker) ID() string    { return "docker" }
func (Docker) Title() string { return "Docker Engine + Compose" }

var dockerPkgs = []string{
	"docker-ce",
	"docker-ce-cli",
	"containerd.io",
	"docker-buildx-plugin",
	"docker-compose-plugin",
}

func (Docker) Check(c *step.Context) (bool, error) {
	if !system.HasCommand("docker") {
		return false, nil
	}
	for _, p := range dockerPkgs {
		if !system.PackageInstalled(p) {
			return false, nil
		}
	}
	// Also require the user to be in the docker group.
	if !system.InGroup(c.Target, "docker") {
		return false, nil
	}
	return true, nil
}

func (Docker) Apply(c *step.Context) error {
	// 1. Add Docker's GPG key and apt repo (idempotent shell).
	setup := `set -euo pipefail
install -m 0755 -d /etc/apt/keyrings
if [ ! -f /etc/apt/keyrings/docker.asc ]; then
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  chmod a+r /etc/apt/keyrings/docker.asc
fi
codename="$(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")"
# Bleeding-edge Ubuntu may not have a matching Docker repo yet; fall back to
# the newest LTS Docker actually publishes for.
if ! curl -fsSL "https://download.docker.com/linux/ubuntu/dists/$codename/Release" >/dev/null 2>&1; then
  echo "devastation: repo Docker sem '$codename', usando 'noble' (24.04 LTS)" >&2
  codename="noble"
fi
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $codename stable" \
  > /etc/apt/sources.list.d/docker.list`
	if err := c.Sh(setup); err != nil {
		return err
	}
	if err := c.RunApt("update"); err != nil {
		return err
	}
	if err := c.AptInstall(dockerPkgs...); err != nil {
		return err
	}
	// 2. Enable + start the daemon.
	if err := c.Run("systemctl", "enable", "--now", "docker"); err != nil {
		c.UI.Warn("não consegui habilitar o serviço docker: %v", err)
	}
	// 3. Add the user to the docker group.
	if !system.InGroup(c.Target, "docker") {
		if err := c.Run("usermod", "-aG", "docker", c.Target.Name); err != nil {
			return err
		}
		c.UI.Info("usuário %s adicionado ao grupo docker (faça logout/login p/ valer)", c.Target.Name)
	}
	return nil
}
