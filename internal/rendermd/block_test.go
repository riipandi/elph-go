package rendermd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPreprocessFootnotesRendersReadableNotes(t *testing.T) {
	in := "Lihat footnote[^1] di sini.\n\n[^1]: Catatan penting."
	got := preprocessFootnotes(in)
	require.Contains(t, got, "footnote(1)")
	require.Contains(t, got, "> [1] Catatan penting.")
	require.NotContains(t, got, "[^1]")
}

func TestPreprocessHTMLDetails(t *testing.T) {
	in := "<details><summary>Buka</summary>\n\nIsi tersembunyi.\n</details>"
	got := preprocessHTMLBlocks(in)
	require.Contains(t, got, "> **Buka**")
	require.Contains(t, got, "Isi tersembunyi.")
	require.NotContains(t, got, "<details>")
}

func TestNormalizeBlockquoteMarkdownClosesNestedDepth(t *testing.T) {
	in := "> Outer\n> > Nested\n> Still"
	got := NormalizeBlockquote(in)
	require.Contains(t, got, "> > Nested\n>\n> Still")
}
