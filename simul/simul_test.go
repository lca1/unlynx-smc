package main_test

import (
	"github.com/dedis/onet/simul"
	"testing"
)

func TestSimulation(t *testing.T) {
	simul.Start("runfiles/paper_verification.toml")
	//simul.Start("runfiles/aggregationM.toml")
}