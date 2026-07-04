package steps

import (
	"os"
	"strconv"
	"strings"

	"devastation/internal/step"
	"devastation/internal/system"
)

// Swap removes the (annoying) default swap on machines with plenty of RAM.
// Ubuntu/Debian ship a /swap.img swapfile; with >8GB RAM it mostly just adds
// latency, so we disable it, drop it from fstab and reclaim the file. On
// machines with <=8GB RAM we leave swap alone for safety.
type Swap struct{}

func (Swap) ID() string    { return "swap" }
func (Swap) Title() string { return "remove swap (se RAM > 8GB)" }

// > 8 GiB in kB.
const swapRAMThresholdKB = 8 * 1024 * 1024

func ramTotalKB() int64 {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			if f := strings.Fields(line); len(f) >= 2 {
				n, _ := strconv.ParseInt(f[1], 10, 64)
				return n
			}
		}
	}
	return 0
}

func fstabHasSwap() bool {
	b, err := os.ReadFile("/etc/fstab")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if f := strings.Fields(t); len(f) >= 3 && f[2] == "swap" {
			return true
		}
	}
	return false
}

func (Swap) Check(c *step.Context) (bool, error) {
	if ramTotalKB() <= swapRAMThresholdKB {
		return true, nil // pouca RAM: mantém swap, nada a fazer
	}
	active, _ := system.Output("swapon", "--show", "--noheadings")
	if strings.TrimSpace(active) != "" {
		return false, nil
	}
	return !fstabHasSwap(), nil
}

func (Swap) Apply(c *step.Context) error {
	gib := float64(ramTotalKB()) / 1048576.0
	if ramTotalKB() <= swapRAMThresholdKB {
		c.UI.Info("RAM %.1f GiB (<= 8GB) — mantendo a swap por segurança", gib)
		return nil
	}
	c.UI.Info("RAM %.1f GiB (> 8GB) — removendo a swap", gib)

	script := `set -euo pipefail
# arquivos de swap ativos (p/ remover depois de desligar)
files="$(swapon --show=NAME,TYPE --noheadings 2>/dev/null | awk '$2=="file"{print $1}')"
swapoff -a || true
# comenta linhas de swap no /etc/fstab (com backup)
if grep -qE '^[^#]*[[:space:]]swap[[:space:]]' /etc/fstab; then
  cp -n /etc/fstab /etc/fstab.bak || true
  awk '($1 ~ /^#/){print; next} ($3=="swap"){print "#"$0; next} {print}' /etc/fstab > /etc/fstab.new && mv /etc/fstab.new /etc/fstab
fi
# remove os arquivos de swap
for f in $files; do [ -f "$f" ] && rm -f "$f"; done
[ -f /swap.img ] && rm -f /swap.img || true`
	if err := c.Sh(script); err != nil {
		return err
	}
	c.UI.Info("swap desativada · /etc/fstab ajustado (backup em fstab.bak) · arquivo removido")
	return nil
}
