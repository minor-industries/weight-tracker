package _import

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestImport(t *testing.T) {
	content, err := os.ReadFile(os.ExpandEnv("$HOME/Downloads/Weight - Sheet1.csv"))
	require.NoError(t, err)

	reader := csv.NewReader(bytes.NewBuffer(content))
	records, err := reader.ReadAll()
	require.NoError(t, err)

	fmt.Println(len(records))

	for _, r := range records[1:] {
		fmt.Println(r)
	}
}
