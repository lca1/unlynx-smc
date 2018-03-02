package libunlynxsmc

import "math/big"

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
	case "sum":
		result = append(result, IntNew(lenR, x)...)
	}
	// ADD OTHER OPERATIONS HERE

	return result
}
