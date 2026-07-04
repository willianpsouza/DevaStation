package steps

import (
	"os"
	"strings"

	"devstation/internal/step"
)

// writeRoot writes a root-owned system file, creating parent dirs.
func writeRoot(path, content string, mode os.FileMode) error {
	dir := path[:strings.LastIndex(path, "/")]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), mode)
}

// fileHasMarker reports whether a file exists and contains a marker string
// (used to detect config we previously generated, for idempotency).
func fileHasMarker(path, marker string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(b), marker)
}

// managedMarker tags every file this tool generates.
const managedMarker = "# managed-by: devstation"

// appendUser appends content to a user-owned file (creating it if needed),
// preserving prior content and restoring ownership to the target user.
func appendUser(c *step.Context, path, content string) error {
	existing := ""
	if b, err := os.ReadFile(path); err == nil {
		existing = string(b)
	}
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	return c.WriteUserFile(c.Target, path, existing+content, 0o644)
}
