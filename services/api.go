package serviceunlynxsmc

/**
Implementation of a data provider. All functions are used with the client name to specify the communication way but this client is in fact a data provider !
*/
import (
	"github.com/henrycg/prio/config"
	"github.com/henrycg/prio/share"
	"github.com/henrycg/prio/utils"
	"github.com/lca1/unlynx-smc/lib"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/log"
	"math/big"
)

//API is the data provider represented as its secret value ID and modulus
type API struct {
	*onet.Client
	ClientID    string
	secretValue []*config.Field
	modulus     *big.Int
}

//ConfigByte is just used because protobuff cannot handle int...
type ConfigByte struct {
	Name            string
	Type            int64
	IntBits         int64
	IntPow          int64
	CountMinHashes  int64
	CountMinBuckets int64
	LinRegBits      []int64
}

//NewUnLynxSMCClient constructor of a DataProvider
func NewUnLynxSMCClient(clientID string) *API {

	newClient := &API{
		Client:      onet.NewClient(ServiceName),
		ClientID:    clientID,
		secretValue: []*config.Field{&config.Field{Name: "Simul", Type: config.FieldType(byte(0)), IntBits: 2}},
		modulus:     share.IntModulus,
	}
	return newClient
}

//SendRequest split and then send each request for 1 client submission to each server.
//ProtoBuf do not support big.Int, we need to transform to []byte and transfer like this, reconstruction done at
//server/
func (c *API) SendRequest(entities *onet.Roster) (string, error) {
	numServer := len(entities.List)
	//dataSplited := prio_utils.Share(c.modulus,numServer,c.secretValue)

	//For the moment for almost all type, they are chosen randomly in function of the number of bits passed
	requests := libunlynxsmc.ClientRequest(c.secretValue, numServer, 0)

	//Conversion of field as protoBuf do not take int only int64
	circuitConfig := make([]ConfigByte, len(c.secretValue))
	for i := 0; i < len(c.secretValue); i++ {
		field := c.secretValue[i]
		linReg := make([]int64, 0)
		for j := 0; j < len(field.LinRegBits); j++ {
			linReg = append(linReg, int64(field.LinRegBits[j]))
		}
		circuitConfig[i] = ConfigByte{Name: field.Name, IntBits: int64(field.IntBits), Type: int64(field.Type), LinRegBits: linReg, IntPow: int64(field.IntPow), CountMinBuckets: int64(field.CountMinBuckets), CountMinHashes: int64(field.CountMinHashes)}
	}

	// The list is ordered first == root
	servList := entities.List

	resp := ServiceResult{}
	randomPoint := utils.RandInt(c.modulus).Bytes()

	for i := 0; i < len(servList); i++ {

		req := requests[i]
		shareA := req.TripleShare.ShareA.Bytes()
		shareB := req.TripleShare.ShareB.Bytes()
		shareC := req.TripleShare.ShareC.Bytes()
		hint := make([][]byte, 0)
		for _, v := range req.Hint.Delta {
			hint = append(hint, v.Bytes())
		}

		dsc := DataSentClient{
			Leader:        servList[0],
			Roster:        entities,
			CircuitConfig: circuitConfig,
			RandomPoint:   randomPoint,
			ShareA:        shareA,
			ShareB:        shareB,
			ShareC:        shareC,
			Hint:          hint,
			Key:           req.Hint.Key,
			RequestID:     req.RequestID,
		}

		err := c.SendProtobuf(servList[i], &dsc, &resp)
		log.Lvl1(err)
		if err != nil {
			return resp.Results, err
		}

	}
	//return the id of the request in the concurrent map of service if successful
	return resp.Results, nil
}

//ExecuteRequest executes the client submission verification
func (c *API) ExecuteRequest(entities *onet.Roster, id string) error {
	result := RequestResult{}

	for _, v := range entities.List {
		err := c.SendProtobuf(v, &ExecRequest{id}, &result)

		if err != nil {
			return err
		}
	}
	return nil
}

//Aggregate is used to aggregate the datas.
//For now DPs send request to aggregate even if not wanted, was to simplify simulation. However, servers do not aggregate
//if there are less than 2 data points.
func (c *API) Aggregate(entities *onet.Roster, id string) (*big.Int, error) {

	result := AggResult{}
	err := c.SendProtobuf(entities.List[0], &ExecAgg{id}, &result)

	if err != nil {
		return nil, err
	}

	return big.NewInt(0).SetBytes(result.Result), nil
}
