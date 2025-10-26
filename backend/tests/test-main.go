package tests

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	// run
	code := m.Run()

	// clear db and close connection
	os.Exit(code)
}
