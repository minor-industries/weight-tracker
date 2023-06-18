package _import

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"testing"
	"time"
	"weight-tracker/db"
)

func TestImport(t *testing.T) {
	content, err := os.ReadFile(os.ExpandEnv("$HOME/Downloads/Weight - Sheet1.csv"))
	require.NoError(t, err)

	reader := csv.NewReader(bytes.NewBuffer(content))
	records, err := reader.ReadAll()
	require.NoError(t, err)

	fmt.Println(len(records))

	// 2006-01-02 15:04:05.999999999 -0700 MST

	tokyo, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)

	la, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)

	dbmap, err := db.Get("127.0.0.1")
	require.NoError(t, err)

	for _, r := range records[1:] {
		fmt.Println(r)
		T, err := time.ParseInLocation("1/2/2006 15:04:05", r[0], tokyo)
		require.NoError(t, err)

		if T.After(time.Date(2013, 4, 1, 0, 0, 0, 0, tokyo)) {
			T = T.In(la)
		}

		weight, err := strconv.ParseFloat(r[1], 64)
		require.NoError(t, err)

		err = dbmap.Insert(&db.Weight{
			Id:       newID(),
			T:        T.UTC(),
			Location: T.Location().String(),
			Weight:   weight,
			Unit:     "kg",
		})
		require.NoError(t, err)
	}
}

func newID() []byte {
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}

	result, err := id.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return result
}
