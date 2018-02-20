package protocolsUnLynxSMC

import (
	"errors"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"

	"math/big"
)

/**
This is a simple protocol that collects and aggregates by notifying the tree structure until the
leafs are reached. Then they locally aggregate the shares they have and send them to the parent.
The root collects all the data and publishes the final aggregation
*/

//AggregationProtocolName is the name of Prio's aggregation protocol.
const AggregationProtocolName = "PrioAggregation"

/*_________________________________________________________________________________________________________________
 */

//ReplySumCipherBytes is the reply from the children
type ReplySumCipherBytes struct {
	Bytes []byte
	Index int64
}

//AnnounceAggregation is the structure to announce start of protocol
type AnnounceAggregation struct{}

/*
_________________________________________________________________________________________________________________________
*/

//StructReply is the structure containing reply of node
type StructReply struct {
	*onet.TreeNode
	ReplySumCipherBytes
}

//StructAnnounceAggregation is the structure containing announce of node
type StructAnnounceAggregation struct {
	*onet.TreeNode
	AnnounceAggregation
}

//
//AggregationProtocol is the structure representing the protocol, the Feedback channel contains the
//result of the aggregation
type AggregationProtocol struct {
	*onet.TreeNodeInstance

	//the feedback final
	Feedback chan []*big.Int

	//Channel for up and down communication respectively
	ChildDataChannel chan []StructReply
	AnnounceChannel  chan StructAnnounceAggregation

	//The data of the protocol : shares from server, local sum and Modulus
	Shares  [][]*big.Int
	Sum     []*big.Int
	Modulus *big.Int
}

//Tell which message are gonna be used in the protocol
func init() {
	network.RegisterMessage(AnnounceAggregation{})
	network.RegisterMessage(ReplySumCipherBytes{})
	onet.GlobalProtocolRegister(AggregationProtocolName, NewAggregationProtocol)
}

//NewAggregationProtocol creates a new protocol instance
func NewAggregationProtocol(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	//initialize the local sum to 0 and channel
	st := &AggregationProtocol{
		TreeNodeInstance: n,
		Feedback:         make(chan []*big.Int),
		Sum:              make([]*big.Int, 0),
	}

	//register the channel for announce
	err := st.RegisterChannel(&st.AnnounceChannel)
	if err != nil {
		return nil, errors.New("couldn't register Announce data channel: " + err.Error())
	}

	//register the channel for child response
	err = st.RegisterChannel(&st.ChildDataChannel)
	if err != nil {
		return nil, errors.New("couldn't register Child Response channel" + err.Error())
	}

	return st, nil
}

//Start called at the root
func (p *AggregationProtocol) Start() error {
	// log.Lvl1(p.ServerIdentity(), " started a Sum Cipher Protocol (", len(p.Request), " different shares)")

	//The root announce to its children that we start the protocol
	p.SendToChildren(&AnnounceAggregation{})

	//start := time.Now()
	//log.Lvl1("time to send mesage to children of root ", time.Since(start))
	return nil
}

//Dispatch is called on the node and handle incoming messages
func (p *AggregationProtocol) Dispatch() error {

	//send if you're not the root (done in start), and only if you have children
	if !p.IsRoot() {
		if !p.IsLeaf() {
			p.SendToChildren(&AnnounceAggregation{})
		}
	}
	//log.Lvl1("Server p" ,p.Index() ," wait on ")
	//p.waitOnSignal()

	//Ascending aggreg
	//log.Lvl1(" Server p ",p.Index() , "start Aggreg")
	sum := p.ascendingAggregationPhase()
	//log.Lvl1(p.ServerIdentity(), " completed aggregation phase (", sum, " is the sum ) in ", time.Since(start))

	//report result
	if p.IsRoot() {
		p.Feedback <- sum
	}
	return nil
}

// Results pushing up the tree containing aggregation results.
func (p *AggregationProtocol) ascendingAggregationPhase() []*big.Int {
	p.Sum = make([]*big.Int, len(p.Shares[0]))

	for j := 0; j < len(p.Sum); j++ {
		p.Sum[j] = big.NewInt(0)
	}

	if !p.IsLeaf() {
		//wait on the channel for child to complete and add sum
		//take time only at the root
		for i := 0; i < len(p.Sum); i++ {

			for _, v := range <-p.ChildDataChannel {

				//get the bytes and turn them back in big.Int
				var sum big.Int
				sum.SetBytes(v.Bytes)

				index := int(v.Index)

				p.Sum[index].Add(p.Sum[index], &sum)
				p.Sum[index].Mod(p.Sum[index], p.Modulus)
			}
		}
	}

	//do the sum of ciphers
	for i := 0; i < len(p.Shares); i++ {
		for j := 0; j < len(p.Sum); j++ {

			p.Sum[j].Add(p.Sum[j], p.Shares[i][j])
			p.Sum[j].Mod(p.Sum[j], p.Modulus)
		}
	}

	//send to parent the sum to deblock channel wait
	if !p.IsRoot() {
		//send the big.Int in bytes
		for j := 0; j < len(p.Sum); j++ {
			p.SendToParent(&ReplySumCipherBytes{p.Sum[j].Bytes(), int64(j)})
			p.Sum[j] = big.NewInt(0)
		}

	}

	//finish by returning the sum of the root
	for j := 0; j < len(p.Sum); j++ {
		p.Sum[j].Mod(p.Sum[j], p.Modulus)
	}

	return p.Sum

}
