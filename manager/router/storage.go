package router

// InfoStorage represents storage of the hub public information,
// and is used as cached between subsystem fetching the info and retrieving
// the info.
type InfoStorage interface {
	// StorePayment saves the payment.
	StorePayment(payment *DbPayment) error

	// UpdatePeers updates information about set of online and peers
	// connected to the hub.
	UpdatePeers(peers []*DbPeer) error

	// UpdateInfo updates information about the hub lightning network node.
	UpdateInfo(info *DbInfo) error

	// Payments returns the payments happening inside the hub local network,
	Payments() ([]*DbPayment, error)

	// Peers return the peers active and connected to the hub.
	Peers() ([]*DbPeer, error)

	// Info returns hub lighting network node information.
	Info() (*DbInfo, error)
}

type DbNodeInfo struct {
	Alias          string
	Host           string
	Port           string
	IdentityPubKey string
}

type DbNeutrinoInfo struct {
	Host string
	Port string
}

type DbInfo struct {
	Version      string
	Network      string
	BlockHeight  uint32
	BlockHash    string
	NodeInfo     *DbNodeInfo
	NeutrinoInfo *DbNeutrinoInfo
}

type DbPeer struct {
	Alias        string
	LockedByPeer int64
	LockedByHub  int64
}

type DbPayment struct {
	FromPeer string
	ToPeer   string
	Amount   int64
	Status   string
	Type     string
	Time     string
}

func NewInMemoryHubStorage() *InMemoryInfoStorage {
	return &InMemoryInfoStorage{}
}

// InMemoryInfoStorage represent in-memory storage for hub public information.
type InMemoryInfoStorage struct {
	info     *DbInfo
	payments []*DbPayment
	peers    []*DbPeer
}

func (s *InMemoryInfoStorage) StorePayment(payment *DbPayment) error {
	s.payments = append(s.payments, payment)
	return nil
}

func (s *InMemoryInfoStorage) UpdatePeers(peers []*DbPeer) error {
	s.peers = peers
	return nil
}

func (s *InMemoryInfoStorage) UpdateInfo(info *DbInfo) error {
	s.info = info
	return nil
}

func (s *InMemoryInfoStorage) Payments() ([]*DbPayment, error) {
	if s.payments == nil {
		return []*DbPayment{}, nil
	}

	return s.payments, nil
}
func (s *InMemoryInfoStorage) Peers() ([]*DbPeer, error) {
	if s.peers == nil {
		return []*DbPeer{}, nil
	}

	return s.peers, nil
}

func (s *InMemoryInfoStorage) Info() (*DbInfo, error) {
	if s.info == nil {
		return &DbInfo{
			NodeInfo:     &DbNodeInfo{},
			NeutrinoInfo: &DbNeutrinoInfo{},
		}, nil
	}

	return s.info, nil
}
