package libunlynxsmc

import (
	"math/big"
)

//Encode contains the encodings of different operations
func Encode(x *big.Int, operation string) []*big.Int {
	result := make([]*big.Int, 1)
	result[0] = x

	lenR := len(result)
	// if we use a 63-bit modulus
	if lenR > 0 && lenR <= 64 {
		lenR = 63
		// if we use a 87-bit modulus
	} else if lenR > 64 && lenR <= 88 {
		lenR = 87
		// if we use a 102-bit modulus
	} else if lenR > 88 && lenR <= 104 {
		lenR = 103
		// if we use a 265-bit modulus
	} else if lenR > 104 && lenR <= 266 {
		lenR = 255
	}

	switch operation {
		case "sum", "mean":
			result = append(result, IntNew(lenR, x) ...)
			break

		case "variance":
			result = append(result, IntPowNew(lenR, int_power, x) ...)
			break

		case "bool_AND", "bool_OR":
			//JS: Should this be done this way? (lamda zeros in prio!)
			if x == big.NewInt(1) {
				result = append(result, boolNew(true) ...)
			} else {result = append(result, boolNew(false) ...)}
			break

		case "min":
			total := nHashes * nBuckets
			values := make([]bool, total)
			//JS: set values[i] to true for i >= x and to false otherwise
			for i := 0; i < total; i++ {
				if int64(i) >= x.Int64() {values[i] = true} else {values[i] = false}
			}
			result = append(result, countMinNew(nHashes, nBuckets, values)...)
			break

		case "lin_reg":
			result = append(result, linRegNew_updated(result, LinRegBits_temp) ...)
			break
		}

	return result
}

func Decode(output []*big.Int, operation string) *big.Int {
	result := big.NewInt(int64(0))
	switch operation {
		case "sum", "mean":
			result = output[0]
			break

		case "variance":
			result = output[len(output) - 1]
			break

		case "bool_AND":
			res :=  output[0].Int64()
			if res == int64(NbHost) {result = big.NewInt(1)
			} else {result = big.NewInt(0)}
			break

		case "bool_OR":
			res :=  output[0].Int64()
			if res == int64(0) {result = big.NewInt(0)
			} else {result = big.NewInt(1)}
			break

		case "min":
			for i := 0; i < len(output); i++ {
				if output[i].Int64() == int64(1) {
						result = big.NewInt(int64(i))
					break
				}
			}
			break

		case "lin_reg":
			sum_x := output[0].Int64()
			sum_y := output[1].Int64()
			sum_x_squared := output[2].Int64()
			sum_x_y := output[3].Int64()
			//sum_y_squared := output[4].Int64()

			//JS: we need to return both (c0, c1) for linear regression
			//but since result is not an array, for now we should return one of them
			//JS: c1 and c0 below are int, but for more precise results, c1 and c0 need to be float
			nbHost_64 := int64(NbHost)
			c1 := (nbHost_64 * sum_x_y - sum_x*sum_y)/((nbHost_64*sum_x_squared) - sum_x*sum_x)
			c0 := (sum_y - sum_x*c1)/nbHost_64
			//c1 := float64(nbHost_64 * sum_x_y - sum_x*sum_y)/float64((nbHost_64*sum_x_squared) - sum_x*sum_x)
			//c0 := (float64(sum_y) - float64(sum_x)*c1)/float64(nbHost_64)
			result = big.NewInt(c0)
			break
		}

	return result
}