package servicesunlynxsmc_test

import (
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"github.com/lca1/unlynx-smc/services"
	"github.com/lca1/unlynx/lib"
	"testing"
	lib "github.com/lca1/unlynx-smc/lib"
)

func TestServiceUnLynxSMC(t *testing.T) {
	//JS: print the chosen operation
	println("Operation:", lib.Operation)

	//JS: Set the appropriate operationInt, depending on operation
	switch lib.Operation {
	case "variance":
		lib.OperationInt = 1
		break
	case "bool_OR":
		lib.OperationInt = 2
		break
	case "bool_AND":
		lib.OperationInt = 3
		break
	case "min":
		lib.OperationInt = 4
		break
	case "lin_reg":
		lib.OperationInt = 5
		break
	}

	//log.SetDebugVisible(3)
	local := onet.NewLocalTest(libunlynx.SuiTe)

	// generate 5 hosts, they don't connect, they process messages, and they
	// don't register the tree or entity list
	_, el, _ := local.GenTree(lib.NbServers, false)
	defer local.CloseAll()

	dataPro := make([]*servicesunlynxsmc.API, lib.NbHost)

	//init
	for i, _ := range dataPro {
		dataPro[i] = servicesunlynxsmc.NewUnLynxSMCClient("DP" + string(i), lib.OperationInt)
	}

	//log.Lvl1("Secret value is ", (client.secretValue[0].IntBits) ,"bits")

	for i, v := range dataPro {
		res, _ := v.SendRequest(el)
		v.ExecuteRequest(el, res)
		if i == len(dataPro)-1 {
			final, _ := dataPro[i].Aggregate(el, res, lib.Operation)
			log.Lvl1(final)
		}
	}

}