// devastation transforms a fresh Ubuntu 24.04+ install into a complete Go
// development workstation: latest kernel, Go toolchain, Docker, fish + starship,
// vim, modern CLIs, git, tmux and GNOME tweaks (animations off).
//
// Usage:
//
//	sudo ./devastation                 # run everything
//	sudo ./devastation --dry-run       # show what would happen, change nothing
//	sudo ./devastation --only fish,vim # run just these modules
//	sudo ./devastation --skip docker   # run all but docker
//	./devastation --list               # list modules (no root needed)
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"devastation/internal/step"
	"devastation/internal/steps"
	"devastation/internal/system"
	"devastation/internal/ui"
)

const version = "v0.1.0"

func main() {
	var (
		dryRun      = flag.Bool("dry-run", false, "mostra o que faria, sem alterar nada")
		verbose     = flag.Bool("verbose", false, "exibe a saída dos comandos")
		list        = flag.Bool("list", false, "lista os módulos disponíveis e sai")
		only        = flag.String("only", "", "roda apenas estes módulos (csv), ex: fish,vim")
		skip        = flag.String("skip", "", "pula estes módulos (csv), ex: docker,gnome")
		fishDefault = flag.Bool("fish-default", false, "define fish como shell padrão do usuário")
		gitName     = flag.String("git-name", "Willian Pires", "git user.name")
		gitEmail    = flag.String("git-email", "willianpsouza@gmail.com", "git user.email")
	)
	flag.Parse()

	u := ui.New()
	all := steps.All()

	if *list {
		u.Banner(version)
		u.Section("Módulos")
		for _, s := range all {
			fmt.Printf("  %-14s %s\n", s.ID(), s.Title())
		}
		fmt.Println()
		return
	}

	u.Banner(version)

	// Resolve the target user before any privilege dance so SUDO_USER is intact.
	target, err := system.Target()
	if err != nil {
		u.Fail("%v", err)
		os.Exit(1)
	}

	// Preflight: OS sanity. Non-fatal warnings only, except hard blockers.
	if err := preflight(u); err != nil {
		u.Fail("%v", err)
		os.Exit(1)
	}

	// Select the steps to run.
	selected, err := selectSteps(all, *only, *skip)
	if err != nil {
		u.Fail("%v", err)
		os.Exit(1)
	}

	// We need root for apt/systemctl. Re-exec under sudo unless dry-running,
	// already root, or none of the selected steps actually need root (e.g.
	// running only --only gnome, which must stay as the real user).
	if !*dryRun && os.Geteuid() != 0 && step.AnyNeedsRoot(selected) {
		u.Info("privilégios de root necessários — reexecutando com sudo…")
		reexecWithSudo()
		return // unreachable on success
	}

	ctx := &step.Context{
		Exec: &system.Exec{
			DryRun:  *dryRun,
			Verbose: *verbose,
			UI:      u,
		},
		UI:     u,
		Target: target,
		Cfg: step.Config{
			GitName:     *gitName,
			GitEmail:    *gitEmail,
			FishDefault: *fishDefault,
		},
	}

	if *dryRun {
		u.Warn("modo DRY-RUN: nada será alterado")
	}
	u.Info("usuário-alvo: %s (home: %s)", target.Name, target.Home)

	runner := &step.Runner{Ctx: ctx, Steps: selected}
	results := runner.Run()
	failed := runner.Summarize(results)
	if failed > 0 {
		os.Exit(1)
	}
}

// preflight validates we're on a supported platform.
func preflight(u *ui.UI) error {
	rel := system.OSRelease()
	if rel["ID"] != "ubuntu" {
		return fmt.Errorf("este tool é para Ubuntu (encontrei ID=%q)", rel["ID"])
	}
	ver := rel["VERSION_ID"]
	if ver < "24.04" {
		u.Warn("Ubuntu %s é anterior ao alvo (24.04+) — pode haver imprevistos", ver)
	} else {
		u.Info("Ubuntu %s detectado", ver)
	}
	if !system.HasCommand("apt-get") {
		return fmt.Errorf("apt-get não encontrado — sistema não suportado")
	}
	return nil
}

// selectSteps applies --only / --skip filters, preserving execution order.
func selectSteps(all []step.Step, only, skip string) ([]step.Step, error) {
	onlySet := csvSet(only)
	skipSet := csvSet(skip)
	if len(onlySet) > 0 && len(skipSet) > 0 {
		return nil, fmt.Errorf("use --only OU --skip, não os dois")
	}
	valid := map[string]bool{}
	for _, s := range all {
		valid[s.ID()] = true
	}
	for id := range onlySet {
		if !valid[id] {
			return nil, fmt.Errorf("módulo desconhecido em --only: %q (veja --list)", id)
		}
	}
	for id := range skipSet {
		if !valid[id] {
			return nil, fmt.Errorf("módulo desconhecido em --skip: %q (veja --list)", id)
		}
	}
	var out []step.Step
	for _, s := range all {
		if len(onlySet) > 0 && !onlySet[s.ID()] {
			continue
		}
		if skipSet[s.ID()] {
			continue
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("nenhum módulo selecionado")
	}
	return out, nil
}

func csvSet(s string) map[string]bool {
	m := map[string]bool{}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			m[p] = true
		}
	}
	return m
}

// reexecWithSudo replaces the current process with `sudo <self> <args...>`,
// preserving arguments. sudo keeps SUDO_USER pointing at the real user.
func reexecWithSudo() {
	self, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "não consegui localizar o binário:", err)
		os.Exit(1)
	}
	sudoPath, err := lookSudo()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sudo não encontrado:", err)
		os.Exit(1)
	}
	argv := append([]string{"sudo", self}, os.Args[1:]...)
	if err := syscall.Exec(sudoPath, argv, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, "falha ao reexecutar com sudo:", err)
		os.Exit(1)
	}
}

func lookSudo() (string, error) {
	for _, p := range []string{"/usr/bin/sudo", "/bin/sudo"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("sudo ausente")
}
