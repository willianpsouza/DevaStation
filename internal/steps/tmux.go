package steps

import (
	"devastation/internal/step"
	"devastation/internal/system"
)

// Tmux installs tmux and writes a modest, comfortable config.
type Tmux struct{}

func (Tmux) ID() string    { return "tmux" }
func (Tmux) Title() string { return "tmux + config básica" }

func (Tmux) Check(c *step.Context) (bool, error) {
	if !system.HasCommand("tmux") {
		return false, nil
	}
	return fileHasMarker(c.Target.Home+"/.tmux.conf", managedMarker), nil
}

func (Tmux) Apply(c *step.Context) error {
	if err := c.AptInstall("tmux"); err != nil {
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
`
	if err := c.WriteUserFile(c.Target, c.Target.Home+"/.tmux.conf", conf, 0o644); err != nil {
		return err
	}
	c.UI.Info("~/.tmux.conf gerado")
	return nil
}
