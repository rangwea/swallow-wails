package backend

import (
	"testing"
)

func TestKillProcessByName(t *testing.T) {
	KillProcessByName("hugo")
}
