package libunlynxsmc

import (
	"fmt"
	"math/big"

	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/share"
	"github.com/henrycg/prio/utils"
)

//File originally in Prio repository.
//Copied here to show what can be done with each type.

//JS: to be seen later, what should be the number of bits of the x and y values?
var LinRegBits_temp = [2]int{2, 2}

func linRegCircuit(field *config.Field) *circuit.Circuit {

	//JS
	//nTerms := len(field.LinRegBits)
	nTerms := len(LinRegBits_temp)

	// Check x_i's
	xCkts := make([]*circuit.Circuit, nTerms)
	for t := 0; t < nTerms; t++ {
		name := fmt.Sprintf("%v-bits[%v]", field.Name, t)
		//JS
		//xCkts[t] = circuit.NBits(field.LinRegBits[t], name)
		xCkts[t] = circuit.NBits(LinRegBits_temp[t], name)
	}

	// Check x_i * x_j
	prodCkts := make([]*circuit.Circuit, 0)
	prodMulCkts := make([]*circuit.Circuit, 0)
	for i := 0; i < nTerms; i++ {
		for j := 0; j < nTerms; j++ {
			if i >= j {
				name := fmt.Sprintf("%v-prod[%v*%v]", field.Name, i, j)
				prod := circuit.UncheckedInput(name)

				xI := xCkts[i].Outputs()[0]
				xJ := xCkts[j].Outputs()[0]
				prodMulCkts = append(prodMulCkts, circuit.CheckMul(xI, xJ, prod.Outputs()[0]))
				prodCkts = append(prodCkts, prod)
			}
		}
	}

	ckts := make([]*circuit.Circuit, 0)
	ckts = append(ckts, xCkts...)
	ckts = append(ckts, prodCkts...)
	ckts = append(ckts, prodMulCkts...)

	return circuit.AndCircuits(ckts)
}

func linRegNewRandom(field *config.Field) []*big.Int {
	//JS
	//nTerms := len(field.LinRegBits)
	nTerms := len(LinRegBits_temp)
	max := new(big.Int)
	values := make([]*big.Int, nTerms)
	for t := 0; t < nTerms; t++ {
		max.SetUint64(1)
		//JS
		//max.Lsh(max, uint(field.LinRegBits[t]))
		max.Lsh(max, uint(LinRegBits_temp[t]))
		values[t] = utils.RandInt(max)
	}

	return linRegNew(field, values)
}

func linRegNew(field *config.Field, values []*big.Int) []*big.Int {
	//JS
	//nTerms := len(field.LinRegBits)
	nTerms := len(LinRegBits_temp)
	out := make([]*big.Int, 0)

	if len(values) != nTerms {
		panic("Invalid data input")
	}

	// Output x_i's in bits
	for t := 0; t < nTerms; t++ {
		//JS
		//out = append(out, bigToBits(field.LinRegBits[t], values[t])...)
		out = append(out, bigToBits(LinRegBits_temp[t], values[t])...)
	}

	// Compute  (x_i * x_j) for all (i,j)
	for i := 0; i < nTerms; i++ {
		for j := 0; j < nTerms; j++ {
			if i >= j {
				v := new(big.Int)
				v.Mul(values[i], values[j])
				v.Mod(v, share.IntModulus)
				out = append(out, v)
			}
		}
	}

	return out
}

//JS: Same as linRegNew, but without the field parameter
//LinRegBits: the 0th entry is the number of bits in the y value. The rest of the entries represent the number of bits in each x_i.
func linRegNew_updated(values []*big.Int, LinRegBits []int) []*big.Int {

	nTerms := len(LinRegBits)
	out := make([]*big.Int, 0)

	if len(values) != nTerms {
		panic("Invalid data input")
	}

	// Output x_i's in bits
	for t := 0; t < nTerms; t++ {
		out = append(out, bigToBits(LinRegBits[t], values[t])...)
	}

	// Compute  (x_i * x_j) for all (i,j)
	for i := 0; i < nTerms; i++ {
		for j := 0; j < nTerms; j++ {
			if i >= j {
				v := new(big.Int)
				v.Mul(values[i], values[j])
				v.Mod(v, share.IntModulus)
				out = append(out, v)
			}
		}
	}
	return out
}
