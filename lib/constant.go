package libunlynxsmc
//JS: this is a file that holds all the constant parameters that will be used throughout the whole code base

//JS: Field attributes
const IntPower = 2
const NHashes = 8
const NBuckets = 1000
const IntBits = 2
//LinRegBits: the 0th entry is the number of bits in the y value. The rest of the entries represent the number of bits in each x_i.
//JS: to be seen later, what should be the number of bits of the x and y values?
var LinRegBits = []int{64, 64, 64, 64}

//JS: other constants
const NbHost = 5
const NbServers = 5

//JS: choose the operation to execute
var OperationList = [8]string{"sum", "mean", "variance", "bool_AND", "bool_OR", "min", "lin_reg", "unsafe"}
var Operation = OperationList[7]
var OperationInt = -1