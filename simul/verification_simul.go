package main

import (
	"crypto/rand"
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"github.com/henrycg/prio/circuit"
	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/share"
	"github.com/henrycg/prio/utils"
	"github.com/lca1/unlynx-smc/lib"
	"github.com/lca1/unlynx-smc/protocols"
	"github.com/lca1/unlynx/lib"
	"math/big"
	"time"
	"sync"
)

//VerificationSimulation holds the state of a simulation.
type VerificationSimulation struct {
	onet.SimulationBFTree
	OperationInt 	int
	NbrParallel 	int
	NbrSequential 	int
	NbrDPsTotal 	int
	Proofs      	bool
	Timeout 		time.Duration
	Sleep 			time.Duration

}
//var req []*libunlynxsmc.Request
//var ckt []*circuit.Circuit
//var mod = share.IntModulus
//var randomPoint = utils.RandInt(mod)

func init() {
	onet.SimulationRegister("Verification", NewVerificationSimulation)
}

//NewVerificationSimulation creates a new Verification Simulation
func NewVerificationSimulation(config string) (onet.Simulation, error) {
	sim := &VerificationSimulation{}
	_, err := toml.Decode(config, sim)
	if err != nil {
		return nil, err
	}

	return sim, nil
}

//Setup create the local roster for simulation
func (sim *VerificationSimulation) Setup(dir string, hosts []string) (*onet.SimulationConfig, error) {
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
func (sim *VerificationSimulation) Node(config *onet.SimulationConfig) error {
	//start := time.Now()
	config.Server.ProtocolRegister("VerificationSimul",
		func(tni *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
			return NewVerificationProtocolSimul(tni, sim)
		})
	//time := time.Since(start)
	//sum += time.Seconds()

	return sim.SimulationBFTree.Node(config)
}

//Run starts the simulation.
func (sim *VerificationSimulation) Run(config *onet.SimulationConfig) error {
	for round := 0; round < sim.Rounds; round++ {
		log.Lvl1("Starting round", round)
		//req, ckt = createCipherSet(sim.Hosts, sim.OperationInt)
		roundTime := libunlynx.StartTimer("Verification(Simulation")
		//new variable for nbValidation
		wg := libunlynx.StartParallelize(sim.NbrDPsTotal)
		start := time.Now()
		mu := sync.Mutex{}
		counter := 0

		for i := 0; i < sim.NbrParallel; i++ {
			go func(i int) {
				for j := 0; j <  sim.NbrSequential; j++ {
					rooti, err := config.Overlay.CreateProtocol("VerificationSimul", config.Tree, onet.NilServiceID)
					if err != nil {
						log.Fatal("Error while creating Protocol:", err)
					}
					root := rooti.(*protocolsunlynxsmc.VerificationProtocol)

					timer := time.Now()
					go root.Start()

					mutex := sync.Mutex{}
					finish := false
					if j > 0 {
						go func (startWait time.Time) {
							for time.Since(startWait) < time.Millisecond * sim.Timeout {
								time.Sleep(time.Millisecond*sim.Sleep)
								mutex.Lock()
								if finish {
									return
								}
								mutex.Unlock()
							}
							log.LLvl1("Go routine blocked. Unblocking...")
							root.AggregateData <-nil
						}(time.Now())
					}

					<-root.AggregateData
					log.Lvl2("It took:", time.Since(timer))
					mutex.Lock()
					finish = true
					mutex.Unlock()


					mu.Lock()
					log.LLvl1("Finished", counter, sim.NbrDPsTotal)
					counter = counter + 1
					if counter <= sim.NbrDPsTotal{
						wg.Done()
					}
					mu.Unlock()
				}
			}(i)
		}


		libunlynx.EndParallelize(wg)
		time := time.Since(start)
		libunlynx.EndTimer(roundTime)

		log.LLvl1("Verification simulation took:", time)

		/*filename := "../../../../lca1/unlynx-smc/simul/time"
		f, err := os.Create(filename, os.O_APPEND|os.O_WRONLY, 0600)
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


//NewVerificationProtocolSimul is the function called on each node to send data
func NewVerificationProtocolSimul(tni *onet.TreeNodeInstance, sim *VerificationSimulation) (onet.ProtocolInstance, error) {
	mod := share.IntModulus
	randomPoint := utils.RandInt(mod)
	protocol, err := protocolsunlynxsmc.NewVerificationProtocol(tni)
	pap := protocol.(*protocolsunlynxsmc.VerificationProtocol)

	pap.Request = new(libunlynxsmc.Request)
	pap.Checker = new(libunlynxsmc.Checker)
	pap.Pre = new(libunlynxsmc.CheckerPrecomp)

	//simulate sending of client to protocol, !! each server must have a different circuit which has the same value for
	//each client submission

	req, ckt := createCipherSet(tni.Tree().Size(), sim.OperationInt)
	pap.Request = req[tni.Index()]
	pap.Checker = libunlynxsmc.NewChecker(ckt[tni.Index()], pap.Index(), 0)
	pap.Pre = libunlynxsmc.NewCheckerPrecomp(ckt[tni.Index()])
	pap.Pre.SetCheckerPrecomp(randomPoint)

	return protocol, err
}

//create cipher text for test from a config file in UnLynxSMC
func createCipherSet(numberServer, operationInt int) ([]*libunlynxsmc.Request, []*circuit.Circuit) {

	circuit := make([]*circuit.Circuit, 0)
	result := make([]*libunlynxsmc.Request, numberServer)

	//secret := config.LoadFile("../../../../henrycg/prio/eval/cell-geneva.conf")
	//fields := make([]*config.Field, 0)
	/*for j := 0; j < len(secret.Fields); j++ {
		fields = append(fields, &(secret.Fields[j]))
	}*/
	fields := []*config.Field{&config.Field{Name: "Int1", Type: config.FieldType(byte(operationInt)), IntBits: libunlynxsmc.IntBits,
		IntPow: libunlynxsmc.IntPower, CountMinHashes: libunlynxsmc.NHashes, CountMinBuckets: libunlynxsmc.NBuckets, LinRegBits: libunlynxsmc.LinRegBits}}
	result = libunlynxsmc.ClientRequest(fields, numberServer, 0)

	for j := 0; j < numberServer; j++ {

		test := libunlynxsmc.ConfigToCircuit(fields)
		circuit = append(circuit, test)
	}

	return result, circuit
}

//function to generate a random big int between 0 and low^expo
func randomBig(low, expo *big.Int) (int *big.Int) {
	max := new(big.Int)
	max.Exp(low, expo, nil).Sub(max, big.NewInt(1))

	//Generate cryptographically strong pseudo-random between 0 - max
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Lvl2(errors.New("Could not create random Big int "))
	}
	return n
}
