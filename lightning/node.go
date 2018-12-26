package lightning

// NodeID uniquely identifies the user in the local lightning network.
// A this moment it is public key of node.
type NodeID string

// Node...
type Node struct {
	// NodeID...
	NodeID NodeID

	// IsConnected is user connected to hub with tcp/ip connection and active
	// for communication.
	IsConnected bool
}
