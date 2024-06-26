package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"weight-tracker/db"
)

func Test_writeWeightToDB(t *testing.T) {
	conn, err := db.Get("localhost")
	require.NoError(t, err)

	id, err := uuid.NewUUID()
	require.NoError(t, err)

	_, err = writeWeightToDB(conn, "78.1", "kg", id.String())
	require.NoError(t, err)
}
