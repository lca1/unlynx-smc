package servicesunlynxsmc_test

import (
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"github.com/lca1/unlynx-smc/services"
	"github.com/lca1/unlynx/lib"
	"testing"
)

var nbHost = 5
var nbServers = 5
var operation = "sum"

func TestServiceUnLynxSMC(t *testing.T) {
	//log.SetDebugVisible(3)
	local := onet.NewLocalTest(libunlynx.SuiTe)

	// generate 5 hosts, they don't connect, they process messages, and they
	// don't register the tree or entity list
	_, el, _ := local.GenTree(nbServers, false)
	defer local.CloseAll()

	dataPro := make([]*servicesunlynxsmc.API, nbHost)

	//init
	for i, _ := range dataPro {
		dataPro[i] = servicesunlynxsmc.NewUnLynxSMCClient("DP" + string(i))
	}

	//log.Lvl1("Secret value is ", (client.secretValue[0].IntBits) ,"bits")

	for i, v := range dataPro {
		res, _ := v.SendRequest(el)
		v.ExecuteRequest(el, res)
		if i == len(dataPro)-1 {
			final, _ := dataPro[i].Aggregate(el, res, operation)
			log.Lvl1(final)
		}
	}

}
