package updatecheck

import (
	"fmt"
	"io"
	"strings"

	"charm.land/glamour/v2"
)

// assemble renders the release list to a single markdown document, newest first.
func assemble(rels []release) string {
	var b strings.Builder
	for i, r := range rels {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "# %s\n\n%s\n", r.Tag, strings.TrimSpace(r.Body))
	}
	return b.String()
}

// render writes md to w through glamour, falling back to the raw markdown if glamour fails —
// the styled render is best-effort, but the content must always reach the user. See ADR-0012.
func render(w io.Writer, md string) error {
	out, err := glamour.Render(md, "auto")
	if err != nil {
		out = md
	}
	if _, err := io.WriteString(w, out); err != nil {
		return fmt.Errorf("writing changelog: %w", err)
	}
	return nil
}
