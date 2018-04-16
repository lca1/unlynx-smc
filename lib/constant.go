package libunlynxsmc

//JS: this is a file that holds all the constant parameters that will be used throughout the whole code base
const int_power = 2
const nHashes = 8
const nBuckets = 32
const nbHost = 5

//LinRegBits_temp: the 0th entry is the number of bits in the y value. The rest of the entries represent the number of bits in each x_i.
//JS: to be seen later, what should be the number of bits of the x and y values?
var LinRegBits_temp = []int{2, 2}