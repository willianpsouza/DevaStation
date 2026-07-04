# DevaStation — Handoff / Log da sessão de setup

> Registro detalhado de tudo que foi feito montando esta estação de dev.
> Serve de contexto pra qualquer sessão futura do Claude Code nesta máquina.
> Última atualização: 2026-07-04.

## A máquina

- **SO:** Ubuntu **26.04 LTS** ("Resolute Raccoon", codename `resolute`) — bleeding edge.
- **Arch:** x86_64 · **RAM:** ~14.8 GiB (16GB) · **Sessão:** Wayland / GNOME.
- **Usuário:** `willian_pires` (uid 1000, grupos incluem `sudo`, `docker`, `lxd`).
- **`sudo` PEDE SENHA** (não é NOPASSWD). Para rodar coisas privilegiadas, o usuário
  digita a senha — não peça/armazene a senha em arquivos.
- Começou praticamente zerada (só Chrome, `curl`, `wget`, `gnome-shell`, `gsettings`, `sudo`).

## O projeto: DevaStation

- **Local:** `~/projects/devastation` · módulo/binário Go `devastation` · marca "DevaStation".
- **GitHub:** https://github.com/willianpsouza/DevaStation (público). Conta `gh`: `willianpsouza` (autenticada).
- **O que é:** CLI em Go puro (stdlib) que transforma um Ubuntu 24.04+ pelado numa
  estação de dev Go completa. Idempotente, com `--dry-run`, `--only`, `--skip`, `--list`.
- **Arquitetura:**
  - `main.go` — flags, preflight (checa Ubuntu 24.04+), seleção de módulos, re-exec com sudo.
  - `internal/ui/` — logging colorido.
  - `internal/system/` — exec de comandos (respeita dry-run/verbose), detecção de usuário-alvo
    (via `SUDO_USER`), apt, escrita de arquivos com `chown` de volta pro usuário.
  - `internal/step/` — interface `Step` (`ID/Title/Check/Apply`) + runner idempotente.
    `NeedsRoot()` (default true) permite steps user-level.
  - `internal/steps/` — um arquivo por módulo; registrados em `registry.go`.
- **Privilégio inteligente:** só re-executa com `sudo` se algum step selecionado precisa de root.
  Steps user-level (`gnome`, `ssh-config`, `claude-code`) rodam como o usuário; `gnome`/ghostty
  usam o session bus (`DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus`).

### Como buildar e rodar
```bash
cd ~/projects/devastation
go build -o devastation .           # precisa do Go no PATH (veja abaixo)
sudo ./devastation                  # roda tudo (pede senha)
sudo ./devastation --dry-run        # mostra o plano, não altera nada
sudo ./devastation --only ghostty   # só um módulo
./devastation --list                # lista módulos (sem root)
```

## Módulos (21) — registrados em `internal/steps/registry.go`

| Módulo | O que faz | Estado nesta máquina |
|---|---|---|
| `system-update` | apt update + dist-upgrade + garante kernel + autoremove | aplicado |
| `apt-base` | build-essential, curl, wget, gnupg, jq, etc. | aplicado |
| `swap` | **remove swap se RAM > 8GB** (swapoff, comenta fstab c/ backup, rm /swap.img) | aplicado (swap removida, ~4GB recuperados) |
| `golang` | baixa/instala Go estável em `/usr/local/go` (+ /etc/profile.d/go.sh) | Go **1.26.4** |
| `nodejs` | Node LTS via NodeSource (apt) | Node **24.18** + npm 11.16 |
| `docker` | Docker CE + compose (repo oficial) + adiciona ao grupo docker | **29.6.1**, rodando |
| `gh` | GitHub CLI (repo apt oficial) | **2.96** (autenticado) |
| `wireguard` | wireguard + wireguard-tools | instalado |
| `fish` | fish shell (+ `--fish-default` torna padrão) | **4.2.1**, é o shell padrão |
| `starship` | prompt starship (bash + fish) | **1.26** |
| `ghostty` | Ghostty (PPA mkasberg) + config + terminal padrão + Ctrl+Alt+T | ver seção abaixo |
| `vim` | vim + ~/.vimrc pro Go | aplicado |
| `vscode` | VS Code de .deb pinado + extensão Go | **1.127** + golang.go |
| `claude` | claude-desktop (GUI) de .deb pinado | 1.18286.0 |
| `claude-code` | CLI Claude Code (installer nativo, user-local) | **2.1.201** em ~/.local/bin |
| `modern-cli` | ripgrep, fd, bat, eza, fzf, **htop** (+symlinks fd/bat) | aplicado |
| `nettools` | nmap, mtr, fping, hping3 | aplicado |
| `ssh-config` | keep-alive agressivo prependado em ~/.ssh/config (700/600) | aplicado |
| `git` | git + config global (nome/email/aliases) | Willian Pires / willianpsouza@gmail.com |
| `tmux` | tmux + ~/.tmux.conf | aplicado |
| `gnome` | desabilita animações + tweaks (via gsettings, como usuário) | aplicado |

### Ghostty (detalhe)
- Instalado via PPA `ppa:mkasberg/ghostty-ubuntu` (tem build pra `resolute`/26.04). Versão **1.3.1**.
- Config em `~/.config/ghostty/config`: `font-family = JetBrains Mono`, `font-size = 16`,
  `background-opacity = 0.95` (5% transparente), `scrollback-limit = 20000000` (20MB;
  **Ghostty conta scrollback em BYTES, não linhas**).
- Fonte `fonts-jetbrains-mono` instalada.
- **Terminal padrão:** `update-alternatives` (`x-terminal-emulator → /usr/bin/ghostty`) +
  gsettings `default-applications.terminal` = `ghostty`.
- **Ctrl+Alt+T → Ghostty:** desabilitou o binding embutido do GNOME (`media-keys terminal = []`)
  e criou um custom-keybinding (`.../custom-keybindings/ghostty/`, command `ghostty`, `<Primary><Alt>t`).

## Ferramentas de IA / dev assistido (feitas via CLI, NÃO são módulos do devastation)

- **caveman** — plugin do Claude Code (`caveman@caveman`, marketplace `juliusbrussee/caveman`).
  Faz o Claude responder em "caveman-speak" pra cortar ~65% dos tokens. **ENABLED.**
- **rtk** (Rust Token Killer) — `~/.local/bin/rtk` v0.43. Tem um **hook `PreToolUse`** em
  `~/.claude/settings.json` (`rtk hook claude`) que reescreve comandos Bash pra economizar tokens.
  Backup do settings pré-rtk: `~/.claude/settings.json.bak`. Config: `~/.config/rtk/filters.toml`.
- **ruflo** (`ruvnet/ruflo`) — meta-harness de agentes. Marketplace `ruvnet/ruflo` adicionado,
  plugin `ruflo-core@ruflo` instalado, e **MCP server `ruflo`** registrado (escopo **user**:
  `claude mcp add -s user ruflo -- npx ruflo@latest mcp start`). Ambos os MCP conectam ✔.

> ⚠️ **caveman + rtk ficam ATIVOS em sessões novas do Claude Code** (carregam no início da sessão).
> É intencional. Pra desligar: `claude plugin disable caveman@caveman` e remover o bloco `hooks`
> do settings.json (restaurar do `.bak`). Ver memória `[[caveman-rtk-intentional]]`.

## Pegadinhas / notas

- **Go em dois lugares:** `~/.local/go` (instalado no início pra ambiente de build; está no
  `~/.bashrc`) **e** `/usr/local/go` (system-wide, via módulo `golang`, no /etc/profile.d).
  Ambos são 1.26.4 — sem conflito, mas bom saber.
- **Marcador renomeado:** dotfiles gerados antes do rename do projeto (`.vimrc`, `.tmux.conf`,
  `~/.ssh/config`, config.fish) têm o marcador antigo `# managed-by: devstation`. Os módulos
  `vim`/`tmux`/`ssh-config` vão **reaplicar uma vez** num próximo run completo pra atualizar o
  marcador pra `devastation` — inofensivo (só reescreve).
- **Docker sem sudo:** usuário está no grupo `docker`, mas só vale após **logout/login** (ou `newgrp docker`).
- **26.04 é novíssima:** repos do Docker e do Ghostty JÁ têm build pra `resolute`. O módulo docker
  tem fallback pra `noble` (24.04) caso um dia falte.
- **Idempotência:** todo módulo tem `Check()`; rodar de novo pula o que já está feito.

## O que faltava / ideias (não feitas)
- `Makefile` (`make build/install`), flag `--version`, GitHub Actions de build no push.
- Normalizar os marcadores antigos dos dotfiles (ou só deixar reaplicar no próximo run).

## Histórico git (9 commits)
Commit inicial → módulos vscode/claude → claude-code/gh/wireguard/nettools/ssh-config →
nodejs → **rename para DevaStation** → ghostty → ghostty tweaks (JetBrains/padrão/opacity) →
Ctrl+Alt+T + htop → **swap**.
