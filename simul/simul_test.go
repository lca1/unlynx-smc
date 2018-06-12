package main_test

import (
	"github.com/dedis/onet/simul"
	"testing"
)

func TestSimulation(t *testing.T) {
	simul.Start("runfiles/verification.toml")
	//simul.Start("runfiles/aggregation.toml")
}