// Package system centralizes command execution, privilege handling,
// OS detection and user-owned file writing.
package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"devastation/internal/ui"
)

// Exec runs external commands honoring dry-run and verbose modes.
type Exec struct {
	DryRun  bool
	Verbose bool
	UI      *ui.UI
}

// Run executes a command that mutates the system. In dry-run it only prints.
func (e *Exec) Run(name string, args ...string) error {
	line := name + " " + strings.Join(args, " ")
	if e.DryRun {
		e.UI.Cmd("would run:", line)
		return nil
	}
	if e.Verbose {
		e.UI.Cmd("run:", line)
	}
	cmd := exec.Command(name, args...)
	if e.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("comando falhou: %s: %w", line, err)
	}
	return nil
}

// Sh runs a bash snippet (needed for pipes / redirects). Honors dry-run.
func (e *Exec) Sh(script string) error {
	if e.DryRun {
		e.UI.Cmd("would run:", "bash -c "+strconv.Quote(script))
		return nil
	}
	if e.Verbose {
		e.UI.Cmd("run:", "bash -c "+strconv.Quote(script))
	}
	cmd := exec.Command("bash", "-c", script)
	if e.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script falhou: %w", err)
	}
	return nil
}

// Output runs a read-only command and returns trimmed stdout. Always executes,
// even in dry-run, because callers use it for Check() probes.
func Output(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return strings.TrimSpace(string(out)), err
}

// AsUser runs a command as the target user, optionally injecting env vars.
// When we are root it drops privileges via `sudo -u`; otherwise runs directly.
func (e *Exec) AsUser(u UserInfo, env []string, name string, args ...string) error {
	if os.Geteuid() == 0 && u.Name != "root" {
		full := []string{"-u", u.Name, "--"}
		if len(env) > 0 {
			full = append(full, "env")
			full = append(full, env...)
		}
		full = append(full, name)
		full = append(full, args...)
		return e.Run("sudo", full...)
	}
	// Already the target user: run directly, applying env.
	if len(env) > 0 {
		full := append([]string{}, env...)
		full = append(full, name)
		full = append(full, args...)
		return e.Run("env", full...)
	}
	return e.Run(name, args...)
}

// UserInfo describes the human whose desktop/dotfiles we are configuring.
type UserInfo struct {
	Name string
	UID  int
	GID  int
	Home string
}

// Target resolves the real user even when invoked through sudo.
func Target() (UserInfo, error) {
	name := os.Getenv("SUDO_USER")
	if name == "" || name == "root" {
		name = os.Getenv("USER")
	}
	if name == "" || name == "root" {
		if u, err := user.Current(); err == nil {
			name = u.Username
		}
	}
	u, err := user.Lookup(name)
	if err != nil {
		return UserInfo{}, fmt.Errorf("não consegui resolver o usuário-alvo %q: %w", name, err)
	}
	uid, _ := strconv.Atoi(u.Uid)
	gid, _ := strconv.Atoi(u.Gid)
	return UserInfo{Name: u.Username, UID: uid, GID: gid, Home: u.HomeDir}, nil
}

// HasCommand reports whether a binary is on PATH.
func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// PackageInstalled reports whether a dpkg package is installed.
func PackageInstalled(pkg string) bool {
	out, err := Output("dpkg-query", "-W", "-f=${Status}", pkg)
	if err != nil {
		return false
	}
	return strings.Contains(out, "install ok installed")
}

// AptInstall installs packages non-interactively (idempotent; skips satisfied).
func (e *Exec) AptInstall(pkgs ...string) error {
	var missing []string
	for _, p := range pkgs {
		if !PackageInstalled(p) {
			missing = append(missing, p)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	args := append([]string{"-y", "install"}, missing...)
	return e.RunApt(args...)
}

// RunApt runs apt-get with DEBIAN_FRONTEND=noninteractive.
func (e *Exec) RunApt(args ...string) error {
	full := append([]string{"env", "DEBIAN_FRONTEND=noninteractive", "apt-get"}, args...)
	return e.Run(full[0], full[1:]...)
}

// WriteUserFile writes content to a path inside the user's home, creating
// parent dirs and chowning everything back to the target user.
func (e *Exec) WriteUserFile(u UserInfo, path, content string, mode os.FileMode) error {
	if e.DryRun {
		e.UI.Cmd("would write:", path)
		return nil
	}
	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		return err
	}
	// Chown the file and any newly created dirs up to (but not including) home.
	_ = os.Chown(path, u.UID, u.GID)
	for d := dir; strings.HasPrefix(d, u.Home) && d != u.Home; d = d[:strings.LastIndex(d, "/")] {
		_ = os.Chown(d, u.UID, u.GID)
	}
	return nil
}

// InGroup reports whether the user already belongs to a unix group.
func InGroup(u UserInfo, group string) bool {
	out, err := Output("id", "-nG", u.Name)
	if err != nil {
		return false
	}
	for _, g := range strings.Fields(out) {
		if g == group {
			return true
		}
	}
	return false
}

// RebootRequired reports whether the kernel/base packages want a reboot.
func RebootRequired() bool {
	_, err := os.Stat("/var/run/reboot-required")
	return err == nil
}

// OSRelease parses /etc/os-release into a map.
func OSRelease() map[string]string {
	m := map[string]string{}
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return m
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		m[k] = strings.Trim(v, `"`)
	}
	return m
}
