package db

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var day = time.Hour * 24

func TestTime(t *testing.T) {
	location1, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)
	fmt.Println(location1)

	location2, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	fmt.Println(location2)

	t1 := time.Now().AddDate(0, 0, -365/2).In(location2)

	fmt.Println(t1)

	//time.Local = time.UTC
	unix := time.Now().Unix()
	fmt.Println(unix)

	t0 := time.Unix(unix, 0)

	fmt.Println(t0)
	fmt.Println(t0.In(time.UTC).Truncate(24 * time.Hour))
}

func TestRows(t *testing.T) {

}
