package protocolsunlynxsmc

import (
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"testing"
	"time"

	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/share"
	"github.com/henrycg/prio/utils"
	"github.com/lca1/unlynx-smc/lib"
	"github.com/lca1/unlynx/lib"
	"github.com/stretchr/testify/assert"
)

// the field cardinality must be superior to nbclient*2^b where b is the maximum number of bit a client need to encode its value

var field = share.IntModulus

var nbServ = 2

//3 random number to test
//var serv1Secret = big.NewInt(int64(55))

//the share of them
//var serv1Share = prio_utils.Share(field,nbServ,serv1Secret)

//var req = prio_utils.ClientRequest(serv1Share, 0)
//var datas = []*config.Field{&config.Field{Name:"test",Type:config.FieldType(byte(5)),LinRegBits:[]int{14,7,1,2,7,8,1,3,8,1,8,4,4,1}},&config.Field{Name:"Test2",Type:config.FieldType(byte(5)),LinRegBits:[]int{1,2,5,2,7,3,8,1,8,1,8,3,6,12}}}
var datas = []*config.Field{&config.Field{Name: "Int1", Type: config.FieldType(byte(0)), IntBits: 2}}
var req = libunlynxsmc.ClientRequest(datas, nbServ, 0)
var randomPoint = utils.RandInt(share.IntModulus)

func TestVerificationProtocol(t *testing.T) {

	local := onet.NewLocalTest(libunlynx.SuiTe)

	// You must register this protocol before creating the servers
	onet.GlobalProtocolRegister("VerificationTest", NewVerificationTest)
	_, _, tree := local.GenTree(nbServ, true)
	defer local.CloseAll()

	p, err := local.CreateProtocol("VerificationTest", tree)
	if err != nil {
		t.Fatal("Couldn't start protocol:", err)
	}

	protocol := p.(*VerificationProtocol)

	start := time.Now()
	protocol.Start()

	//timeout := network.WaitRetry * time.Duration(network.MaxRetryConnect*5*2) * time.Millisecond
	if protocol.IsRoot() {
		Result := <-protocol.AggregateData
		log.Lvl1("time elapsed is ", time.Since(start))
		assert.NotZero(t, len(Result))
	}

}

//inject Test data
func NewVerificationTest(tni *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	pi, err := NewVerificationProtocol(tni)
	protocol := pi.(*VerificationProtocol)

	//set circuit
	ckt := libunlynxsmc.ConfigToCircuit(datas)

	//set request, checker and preChecker
	protocol.Request = new(libunlynxsmc.Request)
	protocol.Request = req[tni.Index()]

	protocol.Checker = new(libunlynxsmc.Checker)
	protocol.Checker = libunlynxsmc.NewChecker(ckt, protocol.Index(), 0)

	protocol.Pre = new(libunlynxsmc.CheckerPrecomp)

	protocol.Pre = libunlynxsmc.NewCheckerPrecomp(ckt)
	protocol.Pre.SetCheckerPrecomp(randomPoint)

	return protocol, err
}
