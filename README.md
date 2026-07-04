# devstation

Transforma uma instalação **default de Ubuntu 24.04+** numa **estação completa de
desenvolvimento Go** com um único binário — idempotente, seguro e sem depender de
nada além da stdlib do Go.

```
sudo ./devstation                 # roda tudo
sudo ./devstation --dry-run       # mostra o que faria, sem mudar nada
sudo ./devstation --only fish,vim # roda só esses módulos
sudo ./devstation --skip docker   # roda tudo menos o docker
./devstation --list               # lista os módulos (não precisa root)
./devstation --only gnome         # tweaks do desktop (roda como você, sem sudo)
```

## O que ele faz

| Módulo         | Descrição |
|----------------|-----------|
| `system-update`| `apt update` + `dist-upgrade`, garante o **último kernel**, autoremove/clean, avisa se precisa reboot |
| `apt-base`     | build-essential, curl, wget, gnupg, ca-certificates, jq, … |
| `golang`       | baixa e instala a **última** toolchain Go em `/usr/local/go` (compara versão antes) |
| `docker`       | Docker CE + Compose (repo oficial) e adiciona você ao grupo `docker` |
| `fish`         | shell fish (opcional: `--fish-default` p/ torná-lo padrão) |
| `starship`     | prompt starship, integrado no bash **e** no fish |
| `vim`          | vim + `~/.vimrc` sensato p/ Go (gofmt-on-save, tabs, undo persistente) |
| `modern-cli`   | ripgrep, fd, bat, eza, fzf (com symlinks `fd`/`bat`) |
| `git`          | git + config global (nome, email, aliases, defaults) |
| `tmux`         | tmux + `~/.tmux.conf` (mouse, vi-keys, splits) |
| `gnome`        | **desabilita animações/efeitos** + tweaks (roda como usuário via session bus) |

## Design

- **Idempotente**: cada módulo tem `Check()` (já está feito?) → só aplica se preciso.
- **Privilégio inteligente**: se re-executa com `sudo` só quando algum módulo
  selecionado precisa de root. O módulo `gnome` roda como você (via
  `DBUS_SESSION_BUS_ADDRESS`), nunca como root.
- **`--dry-run`**: imprime cada comando que rodaria, sem tocar no sistema.
- **Sem dependências**: Go puro (stdlib). O binário é autocontido.

## Flags

```
--dry-run          mostra o que faria, sem alterar nada
--verbose          exibe a saída dos comandos
--list             lista os módulos e sai
--only a,b,c       roda apenas esses módulos
--skip a,b,c       pula esses módulos
--fish-default     define fish como shell padrão
--git-name  NOME   git user.name  (default: Willian Pires)
--git-email EMAIL  git user.email (default: willianpsouza@gmail.com)
```

## Build

```
go build -o devstation .
```

## Estrutura

```
main.go                    flags, preflight, seleção de módulos, re-exec sudo
internal/ui/               logging colorido
internal/system/           exec de comandos, detecção OS/usuário, apt, arquivos
internal/step/             interface Step + runner idempotente
internal/steps/            um arquivo por módulo
```

Adicionar um módulo novo = implementar a interface `Step` (`ID/Title/Check/Apply`)
e registrar em `internal/steps/registry.go`.
