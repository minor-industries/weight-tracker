package graphs

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGraph(t *testing.T) {
	err := Graph()
	require.NoError(t, err)
}
