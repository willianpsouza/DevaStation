package steps

import (
	"os"

	"devastation/internal/step"
	"devastation/internal/system"
)

// Tmux installs tmux and writes a modest, comfortable config, then bootstraps
// the Tmux Plugin Manager (tpm) with resurrect + continuum for session
// save/restore.
type Tmux struct{}

func (Tmux) ID() string    { return "tmux" }
func (Tmux) Title() string { return "tmux + config + tpm (resurrect/continuum)" }

func (Tmux) Check(c *step.Context) (bool, error) {
	if !system.HasCommand("tmux") {
		return false, nil
	}
	if !fileHasMarker(c.Target.Home+"/.tmux.conf", managedMarker) {
		return false, nil
	}
	// tpm cloned? if not, we still need to bootstrap the plugins.
	if _, err := os.Stat(c.Target.Home + "/.tmux/plugins/tpm/tpm"); err != nil {
		return false, nil
	}
	return true, nil
}

func (Tmux) Apply(c *step.Context) error {
	// git is needed to clone tpm and the plugins.
	if err := c.AptInstall("tmux", "git"); err != nil {
		return err
	}
	conf := managedMarker + `
# ~/.tmux.conf

set -g default-terminal "tmux-256color"
set -ga terminal-overrides ",*256col*:Tc"

set -g mouse on
set -g history-limit 50000
set -g base-index 1
setw -g pane-base-index 1
set -g renumber-windows on
set -sg escape-time 10
set -g focus-events on
setw -g mode-keys vi

# Recarrega a config com prefix + r
bind r source-file ~/.tmux.conf \; display "config recarregada"

# Split mantendo o diretório atual
bind | split-window -h -c "#{pane_current_path}"
bind - split-window -v -c "#{pane_current_path}"

# Navegação entre panes com Alt+setas (sem prefix)
bind -n M-Left  select-pane -L
bind -n M-Right select-pane -R
bind -n M-Up    select-pane -U
bind -n M-Down  select-pane -D

# Status bar discreta
set -g status-style "bg=default fg=white"
set -g status-left "#[bold] #S "
set -g status-right "#[fg=cyan]%H:%M #[fg=white]%d-%b "

# ── Plugins (tpm) ────────────────────────────────────────────
set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tmux-resurrect'
set -g @plugin 'tmux-plugins/tmux-continuum'

# resurrect: também salva/restaura o conteúdo dos panes
set -g @resurrect-capture-pane-contents 'on'
# continuum: restaura a última sessão ao subir o servidor e
# auto-salva a cada 15 minutos
set -g @continuum-restore 'on'
set -g @continuum-save-interval '15'

# Inicializa o tpm — MANTENHA como última linha
run '~/.tmux/plugins/tpm/tpm'
`
	if err := c.WriteUserFile(c.Target, c.Target.Home+"/.tmux.conf", conf, 0o644); err != nil {
		return err
	}
	c.UI.Info("~/.tmux.conf gerado")

	// Clona o tpm (se faltar) e instala os plugins declarados acima, como o
	// usuário-alvo. HOME é injetado explicitamente porque `sudo -u` não o
	// redefine por padrão, e tanto o tpm quanto o clone dependem de $HOME.
	script := `set -euo pipefail
TPM="$HOME/.tmux/plugins/tpm"
if [ ! -d "$TPM" ]; then
  git clone --depth 1 https://github.com/tmux-plugins/tpm "$TPM"
fi
# sobe um servidor tmux headless que carrega ~/.tmux.conf (registra os @plugin),
# instala os plugins e encerra a sessão temporária
tmux start-server
tmux new-session -d -s __tpm_bootstrap 2>/dev/null || true
"$TPM/bin/install_plugins" || true
tmux kill-session -t __tpm_bootstrap 2>/dev/null || true`
	if err := c.AsUser(c.Target, []string{"HOME=" + c.Target.Home}, "bash", "-c", script); err != nil {
		return err
	}
	c.UI.Info("tpm + resurrect + continuum instalados")
	return nil
}
