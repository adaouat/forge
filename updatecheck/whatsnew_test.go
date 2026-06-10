package updatecheck

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssemble(t *testing.T) {
	tests := []struct {
		name string
		in   []release
		want string
	}{
		{
			name: "single release",
			in:   []release{{Tag: "v1.3.0", Body: "## What changed\n- thing"}},
			want: "# v1.3.0\n\n## What changed\n- thing\n",
		},
		{
			name: "span keeps newest-first order",
			in:   []release{{Tag: "v1.4.0", Body: "b4"}, {Tag: "v1.3.0", Body: "b3"}},
			want: "# v1.4.0\n\nb4\n\n# v1.3.0\n\nb3\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, assemble(tc.in))
		})
	}
}

func TestRender(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, render(&buf, "# Hello\n\nworld"))
	out := buf.String()
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "Hello")
	assert.Contains(t, out, "world")
}
