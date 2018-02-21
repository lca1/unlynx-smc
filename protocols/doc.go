package protocolsUnLynxSMC

/**
This Prio Implementation has two protocols. The SNIPs verification and the aggregation of shares.
In order to make the protocols work, you need to modify the following imports.
Each is described and well commented so that you understand them.

----------------------------------------------------------------------------------------------------------
Changes needed to be done to github files:

Replace the function in github references : dedis/paper_17_dfinity/pbc/group.go :

func (c *common) Read(r io.Reader, objs ...interface{}) error {
	//panic("Not implemented")
	print("Read")
	return nil

}

func (c *common) Write(w io.Writer, objs ...interface{}) error {
	//panic("not implemented")
	print("Write")
	return nil
}

func (c *common) New(t reflect.Type) interface{} {
	//panic("not implemented")
	print("New")
	return nil
}
-----------------------------------------------------------------------------------------------------------

Things to be changed in github dedis/onet.v1/network/encoding.go

var Suite = ed25519.NewAES128SHA256Ed25519(false)
-----------------------------------------------------------------------------------------------------------
Things to be changed in github dedis/crypto.v0/ed25519/suite.go

COMMENT ALL AND ADD

func NewAES128SHA256Ed25519(fullGroup bool) abstract.Suite {
	suite := pbc.NewPairingFp254BNb()
	return suite.G2()
}

*/
