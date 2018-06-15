package libunlynxsmc

import (
	"github.com/dedis/onet/log"
	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/poly"
	"github.com/henrycg/prio/share"
	"github.com/henrycg/prio/triple"
	"github.com/henrycg/prio/utils"
	"math/big"
	"github.com/henrycg/prio/config"
	"math/rand"
)

//This should be run at client for proof start ClientRequest is sent to each server from one client, for all clients

//Request represent the Shares of the circuit and of the Beaver MPC triples
type Request struct {
	RequestID   []byte
	Hint        *share.PRGHints
	TripleShare *triple.Share
}

//JS: bool_min = 1 when we are executing the min operation
var bool_min = 0
//JS: bool_linreg = 1 when we are executing the linear regression operation
var bool_linreg = 0

//ClientRequest creates proof submission for one client
func ClientRequest(datas []*config.Field, ns int, leaderForReq int) []*Request {
	//utils.PrintTime("Initialize")

	prg := share.NewGenPRG(ns, leaderForReq)

	pub := make([]byte, 32)
	rand.Read(pub)

	out := make([]*Request, ns)
	for s := 0; s < ns; s++ {
		out[s] = new(Request)
		out[s].RequestID = pub
	}

	//log.Lvl1("Inputs are")
	inputs := make([]*big.Int, 0)
	for f := 0; f < len(datas); f++ {
		field := datas[f]
		switch field.Type {
		default:
			panic("Unexpected type!")
		case config.TypeInt:
			//log.LLvl1("INT")
			inputs = append(inputs, intNewRandom(int(field.IntBits))...)
		case config.TypeIntPow:
			//log.LLvl1("POW")
			log.LLvl1(int(field.IntPow))
			inputs = append(inputs, intPowNewRandom(int(field.IntBits), int(field.IntPow))...)
		case config.TypeIntUnsafe:
			//log.LLvl1("UNSAFE")
			inputs = append(inputs, intUnsafeNewRandom(int(field.IntBits))...)
		case config.TypeBoolOr:
			//log.LLvl1("OR")
			inputs = append(inputs, boolNewRandom()...)
		case config.TypeBoolAnd:
			//log.LLvl1("AND")
			inputs = append(inputs, boolNewRandom()...)
		case config.TypeCountMin:
			//log.LLvl1("MIN")
			bool_min = 1
			inputs = append(inputs, countMinNewRandom(int(field.CountMinHashes), int(field.CountMinBuckets))...)
		case config.TypeLinReg:
			//log.LLvl1("LIN_REG")
			inputs = append(inputs, linRegNewRandom(field)...)
			bool_linreg = 1
		}
		/*
			for f := 0; f < len(dataShared); f++ {
			//log.Lvl1(dataShared[f])
			inputs = append(inputs, toArrayBit(dataShared[f])...)
		*/
	}

	// Evaluate the Valid() circuit
	ckt := ConfigToCircuit(datas)
	//log.Lvl1("When evaluate request mod is ", ckt.Modulus())
	//can only evaluate on bit values,
	ckt.Eval(inputs)
	//log.LLvl1(ckt.Outputs())

	//JS: if we execute the min operation, print the corresponding min value proposed by every DP
	if bool_min == 1 {
		//var min_candidate= big.NewInt(0)
		for i := 0; i < len(ckt.Outputs()); i++ {
			if ckt.Outputs()[i].WireValue.Int64() == int64(1) {
				//min_candidate = big.NewInt(int64(i))
				break
			}
		}
		//log.Lvl1("value is ", min_candidate)
	} else if (bool_linreg == 1){
		//JS: if we execute the lin_reg operation, print the corresponding X and Y values proposed by every DP
		//log.Lvl1("X value is ", ckt.Outputs()[0].WireValue)
		//log.Lvl1("Y value is ", ckt.Outputs()[1].WireValue)
	} //else {log.Lvl1("value is ", ckt.Outputs()[0].WireValue)}

	// Generate sharings of the input wires and the multiplication gate wires
	ckt.ShareWires(prg)

	// Construct polynomials f, g, and h and share evaluations of h
	sharePolynomials(ckt, prg)

	//generate share of beaver Triple
	triples := triple.NewTriple(share.IntModulus, ns)
	for s := 0; s < ns; s++ {
		out[s].Hint = prg.Hints(s)
		out[s].TripleShare = triples[s]
	}

	return out
}

func toArrayBit(int *big.Int) []*big.Int {
	out := make([]*big.Int, int.BitLen())
	for i := 0; i < int.BitLen(); i++ {
		out[i] = big.NewInt(int64(int.Bit(i)))
	}
	return out
}

//ConfigToCircuit configures a circuit from input data.
func ConfigToCircuit(datas []*config.Field) *circuit.Circuit {

	nf := len(datas)
	ckts := make([]*circuit.Circuit, nf)
	for f := 0; f < nf; f++ {
		field := datas[f]
		switch field.Type {
		default:
			panic("Unexpected type!")
		case config.TypeInt:
			ckts[f] = intCircuit(field.Name, int(field.IntBits))
		case config.TypeIntPow:
			ckts[f] = intPowCircuit(field.Name, int(field.IntBits), int(field.IntPow))
		case config.TypeIntUnsafe:
			ckts[f] = intUnsafeCircuit(field.Name)
		case config.TypeBoolOr:
			ckts[f] = boolCircuit(field.Name)
		case config.TypeBoolAnd:
			ckts[f] = boolCircuit(field.Name)
		case config.TypeCountMin:
			ckts[f] = countMinCircuit(field.Name, int(field.CountMinHashes), int(field.CountMinBuckets))
		case config.TypeLinReg:
			ckts[f] = linRegCircuit(field)
		}
	}

	ckt := circuit.AndCircuits(ckts)
	return ckt
}

func sharePolynomials(ckt *circuit.Circuit, prg *share.GenPRG) {
	mulGates := ckt.MulGates()
	mod := ckt.Modulus()

	// Little n the number of points on the polynomials.
	// The constant term is randomized, so it's (mulGates + 1).
	n := len(mulGates) + 1
	//log.Lvl1("Mulgates: ", n)

	// Big N is n rounded up to a power of two
	N := utils.NextPowerOfTwo(n)

	// Get the n2-th roots of unity
	pointsF := make([]*big.Int, N)
	pointsG := make([]*big.Int, N)
	zeros := make([]*big.Int, N)
	for i := 0; i < N; i++ {
		zeros[i] = utils.Zero
	}

	// Compute f(X) and g(X)
	pointsF[0] = prg.ShareRand(mod)
	pointsG[0] = prg.ShareRand(mod)

	// Send a sharing of h(0) = f(0)*g(0).
	h0 := new(big.Int)
	h0.Mul(pointsF[0], pointsG[0])
	h0.Mod(h0, mod)
	prg.Share(mod, h0)

	for i := 1; i < n; i++ {
		pointsF[i] = mulGates[i-1].ParentL.WireValue
		pointsG[i] = mulGates[i-1].ParentR.WireValue
	}

	// Zero pad the upper coefficients of f(X) and g(X)
	for i := n; i < N; i++ {
		pointsF[i] = utils.Zero
		pointsG[i] = utils.Zero
	}

	// Interpolate through the Nth roots of unity
	polyF := poly.InverseFFT(pointsF)
	polyG := poly.InverseFFT(pointsG)
	paddedF := append(polyF, zeros...)
	paddedG := append(polyG, zeros...)

	// Evaluate at all 2N-th roots of unity
	evalsF := poly.FFT(paddedF)
	evalsG := poly.FFT(paddedG)

	// We need to send to the servers the evaluations of
	//   f(r) * g(r)
	// for all 2N-th roots of unity r that are not also
	// N-th roots of unity.
	hint := new(big.Int)
	for i := 1; i < 2*N-1; i += 2 {
		hint.Mul(evalsF[i], evalsG[i])
		hint.Mod(hint, mod)
		prg.Share(mod, hint)
	}
}
