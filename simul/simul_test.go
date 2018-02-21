package main_test

import (
	"gopkg.in/dedis/onet.v1/simul"
	"testing"
)

func TestSimulation(t *testing.T) {
	simul.Start("runfiles/verification.toml", "runfiles/aggregation.toml")
}
