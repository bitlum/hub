package lightning

type NodeInfo struct {
	Alias          string
	Host           string
	Port           string
	IdentityPubKey string
}

type NeutrinoInfo struct {
	Host string
	Port string
}

type Info struct {
	Version      string
	Network      string
	BlockHeight  uint32
	BlockHash    string
	NodeInfo     *NodeInfo
	NeutrinoInfo *NeutrinoInfo
}
