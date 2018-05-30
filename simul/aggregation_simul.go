package main

import (
	"github.com/BurntSushi/toml"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"github.com/henrycg/prio/share"
	"github.com/lca1/unlynx-smc/protocols"
	"github.com/lca1/unlynx/lib"
	"math/big"
	"time"
)

var aggData [][]*big.Int
var sumCipher *big.Int

func init() {
	onet.SimulationRegister("Aggregation", NewAggregationSimulation)
}

//AggregationSimulation holds the state of a simulation.
type AggregationSimulation struct {
	onet.SimulationBFTree

	NbrRequestByProto int
	Proofs            bool
}

//NewAggregationSimulation creates a new Aggregation simulation
func NewAggregationSimulation(config string) (onet.Simulation, error) {
	sim := &AggregationSimulation{}
	_, err := toml.Decode(config, sim)
	if err != nil {
		return nil, err
	}

	return sim, nil
}

//Setup create the local roster for simulation
func (sim *AggregationSimulation) Setup(dir string, hosts []string) (*onet.SimulationConfig, error) {
	sc := &onet.SimulationConfig{}
	sim.CreateRoster(sc, hosts, 2000)
	err := sim.CreateTree(sc)

	if err != nil {
		return nil, err
	}

	log.Lvl1("Setup done")

	return sc, nil
}

//Node creates the protocol at each nodes.
func (sim *AggregationSimulation) Node(config *onet.SimulationConfig) error {
	//start := time.Now()
	config.Server.ProtocolRegister("AggregationSimul",
		func(tni *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
			return NewAggregationProtocolSimul(tni, sim)
		})
	//time := time.Since(start)
	//sum += time.Seconds()

	return sim.SimulationBFTree.Node(config)
}

//Run starts the simulation.
func (sim *AggregationSimulation) Run(config *onet.SimulationConfig) error {
	for round := 0; round < sim.Rounds; round++ {
		log.Lvl1("Starting round", round)

		aggData = createAggData(sim.NbrRequestByProto, config.Tree.Size())

		roundTime := libunlynx.StartTimer("Aggregation(Simulation")
		//new variable for nbValidation
		//start := time.Now()

		rooti, err := config.Overlay.CreateProtocol("AggregationSimul", config.Tree, onet.NilServiceID)
		if err != nil {
			return nil
		}
		start := time.Now()
		root := rooti.(*protocolsunlynxsmc.AggregationProtocol)
		root.Start()
		result := <-root.Feedback
		log.Lvl1("res is ", result)
		log.Lvl1(sumCipher)
		//time := time.Since(start)
		time := time.Since(start)
		libunlynx.EndTimer(roundTime)

		log.LLvl1("Aggregation simulation took:", time)

		/*filename := "time"
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		if _, err = f.WriteString(time.String() + "\n"); err != nil {
			panic(err)
		}
		f.Close()*/
	}
	return nil
}

//NewAggregationProtocolSimul is called on each node to send data
func NewAggregationProtocolSimul(tni *onet.TreeNodeInstance, sim *AggregationSimulation) (onet.ProtocolInstance, error) {

	protocol, err := protocolsunlynxsmc.NewAggregationProtocol(tni)
	pap := protocol.(*protocolsunlynxsmc.AggregationProtocol)

	pap.Modulus = share.IntModulus
	pap.Shares = aggData

	return protocol, err
}

func createAggData(numberClient, numberServer int) [][]*big.Int {

	//secret value of clients
	sumCipher = big.NewInt(0)
	result := make([][]*big.Int, numberServer)
	secretValues := make([][]*big.Int, numberClient)
	for i := 0; i < numberClient; i++ {
		secretValues[i] = share.Share(share.IntModulus, numberServer, randomBig(big.NewInt(2), big.NewInt(64)))
		log.LLvl1(secretValues)
		for j := 0; j < len(secretValues[i]); j++ {
			sumCipher.Add(sumCipher, secretValues[i][j])
			sumCipher.Mod(sumCipher, share.IntModulus)
		}
	}
	for k := 0; k < numberServer; k++ {
		for l := 0; l < numberClient; l++ {
			result[k] = append(result[k], secretValues[l][k])
		}
	}
	sumCipher.Mod(sumCipher, share.IntModulus)
	return result
}
