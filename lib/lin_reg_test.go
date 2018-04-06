package libunlynxsmc

//File originally in Prio repository.
//Added by JS

import (
	"math/big"
	"testing"
	"github.com/henrycg/prio/config"
)

func getField() *config.Field {
	f := new(config.Field)
	f.Type = config.TypeLinReg
	f.LinRegBits = []int{3, 4, 5}
	return f
}

func TestLinRegGood(t *testing.T) {
	f := getField()
	ckt := linRegCircuit(f)
	vals := linRegNewRandom(f)

	if !ckt.Eval(vals) {
		t.Fail()
	}
}

func TestLinRegBad(t *testing.T) {
	f := getField()
	ckt := linRegCircuit(f)
	vals := linRegNewRandom(f)

	vals[0] = big.NewInt(123123123123)

	if ckt.Eval(vals) {
		t.Fail()
	}
}