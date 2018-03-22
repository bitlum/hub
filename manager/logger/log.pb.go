// Code generated by protoc-gen-go. DO NOT EDIT.
// source: log.proto

/*
Package logger is a generated protocol buffer package.

It is generated from these files:
	log.proto

It has these top-level messages:
	Log
	RouterState
	Channel
	Payment
	ChannelChange
*/
package logger

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type PaymentStatus int32

const (
	PaymentStatus_success PaymentStatus = 0
	// UnsufficientFunds means that router haven't posses/locked enough funds
	// with receiver user to route thouth the payment.
	PaymentStatus_unsufficient_funds PaymentStatus = 1
	// ExternalFail means that receiver failed to receive payment because of
	// the unknown to us reason.
	PaymentStatus_external_fail PaymentStatus = 2
)

var PaymentStatus_name = map[int32]string{
	0: "success",
	1: "unsufficient_funds",
	2: "external_fail",
}
var PaymentStatus_value = map[string]int32{
	"success":            0,
	"unsufficient_funds": 1,
	"external_fail":      2,
}

func (x PaymentStatus) String() string {
	return proto.EnumName(PaymentStatus_name, int32(x))
}
func (PaymentStatus) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// ChannelChangeType represent the type of action which were attempted to
// apply to the channel.
type ChannelChangeType int32

const (
	// Open is used when this channel was just created, and haven't been in
	// local network before
	ChannelChangeType_open ChannelChangeType = 0
	// Close is used when number locked funds / balances of both channel
	// participant equal to zero.
	ChannelChangeType_close ChannelChangeType = 1
	// Udpate is used when one of the participants decides to update its
	// channel balance.
	ChannelChangeType_udpate ChannelChangeType = 2
)

var ChannelChangeType_name = map[int32]string{
	0: "open",
	1: "close",
	2: "udpate",
}
var ChannelChangeType_value = map[string]int32{
	"open":   0,
	"close":  1,
	"udpate": 2,
}

func (x ChannelChangeType) String() string {
	return proto.EnumName(ChannelChangeType_name, int32(x))
}
func (ChannelChangeType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// Log is the main object in the file which represent the log entry in the
// log file.
type Log struct {
	Time int64 `protobuf:"varint,1,opt,name=time" json:"time,omitempty"`
	// Types that are valid to be assigned to Data:
	//	*Log_State
	//	*Log_Payment
	//	*Log_ChannelChange
	Data isLog_Data `protobuf_oneof:"data"`
}

func (m *Log) Reset()                    { *m = Log{} }
func (m *Log) String() string            { return proto.CompactTextString(m) }
func (*Log) ProtoMessage()               {}
func (*Log) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type isLog_Data interface {
	isLog_Data()
}

type Log_State struct {
	State *RouterState `protobuf:"bytes,2,opt,name=state,oneof"`
}
type Log_Payment struct {
	Payment *Payment `protobuf:"bytes,3,opt,name=payment,oneof"`
}
type Log_ChannelChange struct {
	ChannelChange *ChannelChange `protobuf:"bytes,4,opt,name=channel_change,json=channelChange,oneof"`
}

func (*Log_State) isLog_Data()         {}
func (*Log_Payment) isLog_Data()       {}
func (*Log_ChannelChange) isLog_Data() {}

func (m *Log) GetData() isLog_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Log) GetTime() int64 {
	if m != nil {
		return m.Time
	}
	return 0
}

func (m *Log) GetState() *RouterState {
	if x, ok := m.GetData().(*Log_State); ok {
		return x.State
	}
	return nil
}

func (m *Log) GetPayment() *Payment {
	if x, ok := m.GetData().(*Log_Payment); ok {
		return x.Payment
	}
	return nil
}

func (m *Log) GetChannelChange() *ChannelChange {
	if x, ok := m.GetData().(*Log_ChannelChange); ok {
		return x.ChannelChange
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Log) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Log_OneofMarshaler, _Log_OneofUnmarshaler, _Log_OneofSizer, []interface{}{
		(*Log_State)(nil),
		(*Log_Payment)(nil),
		(*Log_ChannelChange)(nil),
	}
}

func _Log_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Log)
	// data
	switch x := m.Data.(type) {
	case *Log_State:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.State); err != nil {
			return err
		}
	case *Log_Payment:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Payment); err != nil {
			return err
		}
	case *Log_ChannelChange:
		b.EncodeVarint(4<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.ChannelChange); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Log.Data has unexpected type %T", x)
	}
	return nil
}

func _Log_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Log)
	switch tag {
	case 2: // data.state
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(RouterState)
		err := b.DecodeMessage(msg)
		m.Data = &Log_State{msg}
		return true, err
	case 3: // data.payment
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Payment)
		err := b.DecodeMessage(msg)
		m.Data = &Log_Payment{msg}
		return true, err
	case 4: // data.channel_change
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(ChannelChange)
		err := b.DecodeMessage(msg)
		m.Data = &Log_ChannelChange{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Log_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Log)
	// data
	switch x := m.Data.(type) {
	case *Log_State:
		s := proto.Size(x.State)
		n += proto.SizeVarint(2<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Log_Payment:
		s := proto.Size(x.Payment)
		n += proto.SizeVarint(3<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Log_ChannelChange:
		s := proto.Size(x.ChannelChange)
		n += proto.SizeVarint(4<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// RouterState is a type of log entry which describes the state/view of the
// router local lightning network and number of free funds which exist under
// control router.
type RouterState struct {
	// Channels represent the local lightning network topology.
	Channels []*Channel `protobuf:"bytes,1,rep,name=channels" json:"channels,omitempty"`
	// FreeBalance it is free number of funds under router managment which
	// could be used to lock them in the channels.
	FreeBalance uint64 `protobuf:"varint,2,opt,name=free_balance,json=freeBalance" json:"free_balance,omitempty"`
	// PendingBalance is the amount of funds which in the process of
	// being accepted by blockchain.
	PendingBalance uint64 `protobuf:"varint,3,opt,name=pending_balance,json=pendingBalance" json:"pending_balance,omitempty"`
	// AverageChangeUpdateDuration is number of milliseconds which
	// is needed to change of state of chanel over blockchain.
	AverageChangeUpdateDuration uint64 `protobuf:"varint,4,opt,name=average_change_update_duration,json=averageChangeUpdateDuration" json:"average_change_update_duration,omitempty"`
}

func (m *RouterState) Reset()                    { *m = RouterState{} }
func (m *RouterState) String() string            { return proto.CompactTextString(m) }
func (*RouterState) ProtoMessage()               {}
func (*RouterState) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RouterState) GetChannels() []*Channel {
	if m != nil {
		return m.Channels
	}
	return nil
}

func (m *RouterState) GetFreeBalance() uint64 {
	if m != nil {
		return m.FreeBalance
	}
	return 0
}

func (m *RouterState) GetPendingBalance() uint64 {
	if m != nil {
		return m.PendingBalance
	}
	return 0
}

func (m *RouterState) GetAverageChangeUpdateDuration() uint64 {
	if m != nil {
		return m.AverageChangeUpdateDuration
	}
	return 0
}

// Channel is used as the building block in describing of the lightning
// network topology.
type Channel struct {
	UserId        uint64 `protobuf:"varint,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	ChannelId     uint64 `protobuf:"varint,2,opt,name=channel_id,json=channelId" json:"channel_id,omitempty"`
	UserBalance   uint64 `protobuf:"varint,3,opt,name=user_balance,json=userBalance" json:"user_balance,omitempty"`
	RouterBalance uint64 `protobuf:"varint,4,opt,name=router_balance,json=routerBalance" json:"router_balance,omitempty"`
}

func (m *Channel) Reset()                    { *m = Channel{} }
func (m *Channel) String() string            { return proto.CompactTextString(m) }
func (*Channel) ProtoMessage()               {}
func (*Channel) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Channel) GetUserId() uint64 {
	if m != nil {
		return m.UserId
	}
	return 0
}

func (m *Channel) GetChannelId() uint64 {
	if m != nil {
		return m.ChannelId
	}
	return 0
}

func (m *Channel) GetUserBalance() uint64 {
	if m != nil {
		return m.UserBalance
	}
	return 0
}

func (m *Channel) GetRouterBalance() uint64 {
	if m != nil {
		return m.RouterBalance
	}
	return 0
}

// Payment represent the attempt of peer in the local lightning network to
// send the payment to some another peer in the network.
type Payment struct {
	Status   PaymentStatus `protobuf:"varint,1,opt,name=status,enum=logger.PaymentStatus" json:"status,omitempty"`
	Sender   uint64        `protobuf:"varint,2,opt,name=sender" json:"sender,omitempty"`
	Receiver uint64        `protobuf:"varint,3,opt,name=receiver" json:"receiver,omitempty"`
	Amount   uint64        `protobuf:"varint,5,opt,name=amount" json:"amount,omitempty"`
	// Earned is the number of funds which router earned by making this payment.
	// In case of rebalncing router will pay the fee, for that reason this
	// number will be negative.
	Earned int64 `protobuf:"varint,6,opt,name=earned" json:"earned,omitempty"`
}

func (m *Payment) Reset()                    { *m = Payment{} }
func (m *Payment) String() string            { return proto.CompactTextString(m) }
func (*Payment) ProtoMessage()               {}
func (*Payment) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Payment) GetStatus() PaymentStatus {
	if m != nil {
		return m.Status
	}
	return PaymentStatus_success
}

func (m *Payment) GetSender() uint64 {
	if m != nil {
		return m.Sender
	}
	return 0
}

func (m *Payment) GetReceiver() uint64 {
	if m != nil {
		return m.Receiver
	}
	return 0
}

func (m *Payment) GetAmount() uint64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *Payment) GetEarned() int64 {
	if m != nil {
		return m.Earned
	}
	return 0
}

type ChannelChange struct {
	Type          ChannelChangeType `protobuf:"varint,1,opt,name=type,enum=logger.ChannelChangeType" json:"type,omitempty"`
	UserId        uint64            `protobuf:"varint,2,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	ChannelId     uint64            `protobuf:"varint,3,opt,name=channel_id,json=channelId" json:"channel_id,omitempty"`
	UserBalance   uint64            `protobuf:"varint,4,opt,name=user_balance,json=userBalance" json:"user_balance,omitempty"`
	RouterBalance uint64            `protobuf:"varint,5,opt,name=router_balance,json=routerBalance" json:"router_balance,omitempty"`
	// Fee which was taken by blockchain decentrilized computer / mainers or
	// some other form of smart contract manager.
	Fee uint64 `protobuf:"varint,6,opt,name=fee" json:"fee,omitempty"`
}

func (m *ChannelChange) Reset()                    { *m = ChannelChange{} }
func (m *ChannelChange) String() string            { return proto.CompactTextString(m) }
func (*ChannelChange) ProtoMessage()               {}
func (*ChannelChange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *ChannelChange) GetType() ChannelChangeType {
	if m != nil {
		return m.Type
	}
	return ChannelChangeType_open
}

func (m *ChannelChange) GetUserId() uint64 {
	if m != nil {
		return m.UserId
	}
	return 0
}

func (m *ChannelChange) GetChannelId() uint64 {
	if m != nil {
		return m.ChannelId
	}
	return 0
}

func (m *ChannelChange) GetUserBalance() uint64 {
	if m != nil {
		return m.UserBalance
	}
	return 0
}

func (m *ChannelChange) GetRouterBalance() uint64 {
	if m != nil {
		return m.RouterBalance
	}
	return 0
}

func (m *ChannelChange) GetFee() uint64 {
	if m != nil {
		return m.Fee
	}
	return 0
}

func init() {
	proto.RegisterType((*Log)(nil), "logger.Log")
	proto.RegisterType((*RouterState)(nil), "logger.RouterState")
	proto.RegisterType((*Channel)(nil), "logger.Channel")
	proto.RegisterType((*Payment)(nil), "logger.Payment")
	proto.RegisterType((*ChannelChange)(nil), "logger.ChannelChange")
	proto.RegisterEnum("logger.PaymentStatus", PaymentStatus_name, PaymentStatus_value)
	proto.RegisterEnum("logger.ChannelChangeType", ChannelChangeType_name, ChannelChangeType_value)
}

func init() { proto.RegisterFile("log.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 546 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0xc1, 0x6e, 0xd3, 0x4e,
	0x10, 0xc6, 0xe3, 0xd8, 0x71, 0x9a, 0xc9, 0x3f, 0xa9, 0x3b, 0x7f, 0x51, 0x0c, 0x08, 0x14, 0x22,
	0x21, 0xaa, 0x46, 0xcd, 0x01, 0x38, 0x73, 0x68, 0x90, 0x48, 0x25, 0x0e, 0x68, 0x0b, 0x67, 0x6b,
	0xeb, 0x1d, 0x1b, 0x4b, 0xce, 0xda, 0x5a, 0xaf, 0x2b, 0xf2, 0x0c, 0xbc, 0x02, 0x6f, 0xc2, 0x23,
	0x70, 0xe3, 0x89, 0x90, 0xd7, 0xeb, 0x90, 0x88, 0x0a, 0x38, 0x65, 0xe7, 0xdb, 0xdf, 0x44, 0xdf,
	0x7c, 0xb3, 0x32, 0x8c, 0xf2, 0x22, 0x5d, 0x96, 0xaa, 0xd0, 0x05, 0xfa, 0x79, 0x91, 0xa6, 0xa4,
	0xe6, 0xdf, 0x1c, 0x70, 0xdf, 0x15, 0x29, 0x22, 0x78, 0x3a, 0xdb, 0x50, 0xe8, 0xcc, 0x9c, 0x33,
	0x97, 0x99, 0x33, 0x2e, 0x60, 0x50, 0x69, 0xae, 0x29, 0xec, 0xcf, 0x9c, 0xb3, 0xf1, 0x8b, 0xff,
	0x97, 0x6d, 0xcf, 0x92, 0x15, 0xb5, 0x26, 0x75, 0xdd, 0x5c, 0xad, 0x7b, 0xac, 0x65, 0x70, 0x01,
	0xc3, 0x92, 0x6f, 0x37, 0x24, 0x75, 0xe8, 0x1a, 0xfc, 0xb8, 0xc3, 0xdf, 0xb7, 0xf2, 0xba, 0xc7,
	0x3a, 0x02, 0x5f, 0xc3, 0x34, 0xfe, 0xc4, 0xa5, 0xa4, 0x3c, 0x6a, 0x7e, 0x53, 0x0a, 0x3d, 0xd3,
	0x73, 0xaf, 0xeb, 0x59, 0xb5, 0xb7, 0x2b, 0x73, 0xb9, 0xee, 0xb1, 0x49, 0xbc, 0x2f, 0x5c, 0xfa,
	0xe0, 0x09, 0xae, 0xf9, 0xfc, 0xbb, 0x03, 0xe3, 0x3d, 0x37, 0xb8, 0x80, 0x23, 0x0b, 0x56, 0xa1,
	0x33, 0x73, 0xf7, 0x5d, 0xd8, 0x7f, 0x64, 0x3b, 0x00, 0x9f, 0xc2, 0x7f, 0x89, 0x22, 0x8a, 0x6e,
	0x78, 0xce, 0x65, 0xdc, 0x4e, 0xe9, 0xb1, 0x71, 0xa3, 0x5d, 0xb6, 0x12, 0x3e, 0x87, 0xe3, 0x92,
	0xa4, 0xc8, 0x64, 0xba, 0xa3, 0x5c, 0x43, 0x4d, 0xad, 0xdc, 0x81, 0x2b, 0x78, 0xc2, 0x6f, 0x49,
	0xf1, 0x94, 0xec, 0x40, 0x51, 0x5d, 0x0a, 0xae, 0x29, 0x12, 0xb5, 0xe2, 0x3a, 0x2b, 0xa4, 0x19,
	0xd0, 0x63, 0x8f, 0x2c, 0xd5, 0xce, 0xf1, 0xd1, 0x30, 0x6f, 0x2c, 0x32, 0xff, 0xe2, 0xc0, 0xd0,
	0xda, 0xc4, 0xfb, 0x30, 0xac, 0x2b, 0x52, 0x51, 0x26, 0xcc, 0x4a, 0x3c, 0xe6, 0x37, 0xe5, 0x95,
	0xc0, 0xc7, 0x00, 0x5d, 0x74, 0x99, 0xb0, 0x9e, 0x47, 0x56, 0xb9, 0x12, 0xcd, 0x50, 0xa6, 0xef,
	0xd0, 0xee, 0xb8, 0xd1, 0x3a, 0xaf, 0xcf, 0x60, 0xaa, 0x4c, 0x66, 0x3b, 0xa8, 0xf5, 0x36, 0x69,
	0x55, 0x8b, 0xcd, 0xbf, 0x3a, 0x30, 0xb4, 0xab, 0xc3, 0x0b, 0xf0, 0x9b, 0x2d, 0xd7, 0x95, 0x31,
	0x33, 0xfd, 0xb5, 0x27, 0x0b, 0x5c, 0x9b, 0x4b, 0x66, 0x21, 0x3c, 0x05, 0xbf, 0x22, 0x29, 0x48,
	0x59, 0x7f, 0xb6, 0xc2, 0x87, 0x70, 0xa4, 0x28, 0xa6, 0xec, 0x96, 0x94, 0x35, 0xb6, 0xab, 0x9b,
	0x1e, 0xbe, 0x29, 0x6a, 0xa9, 0xc3, 0x41, 0xdb, 0xd3, 0x56, 0x8d, 0x4e, 0x5c, 0x49, 0x12, 0xa1,
	0x6f, 0x9e, 0xa6, 0xad, 0xe6, 0x3f, 0x1c, 0x98, 0x1c, 0xbc, 0x12, 0xbc, 0x00, 0x4f, 0x6f, 0x4b,
	0xb2, 0x16, 0x1f, 0xdc, 0xf9, 0x94, 0x3e, 0x6c, 0x4b, 0x62, 0x06, 0xdb, 0x4f, 0xb8, 0xff, 0x87,
	0x84, 0xdd, 0xbf, 0x25, 0xec, 0xfd, 0x4b, 0xc2, 0x83, 0x3b, 0x12, 0xc6, 0x00, 0xdc, 0x84, 0xc8,
	0xcc, 0xe5, 0xb1, 0xe6, 0x78, 0xfe, 0x16, 0x26, 0x07, 0x89, 0xe2, 0x18, 0x86, 0x55, 0x1d, 0xc7,
	0x54, 0x55, 0x41, 0x0f, 0x4f, 0x01, 0x6b, 0x59, 0xd5, 0x49, 0x92, 0xc5, 0x19, 0x49, 0x1d, 0x25,
	0xb5, 0x14, 0x55, 0xe0, 0xe0, 0x09, 0x4c, 0xe8, 0xb3, 0x26, 0x25, 0x79, 0x1e, 0x25, 0x3c, 0xcb,
	0x83, 0xfe, 0xf9, 0x2b, 0x38, 0xf9, 0x6d, 0x6e, 0x3c, 0x02, 0xaf, 0x28, 0x49, 0x06, 0x3d, 0x1c,
	0xc1, 0x20, 0xce, 0x8b, 0x8a, 0x02, 0x07, 0x01, 0xfc, 0x5a, 0x94, 0x5c, 0x53, 0xd0, 0xbf, 0xf1,
	0xcd, 0xb7, 0xe1, 0xe5, 0xcf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x17, 0x57, 0x22, 0xd3, 0x28, 0x04,
	0x00, 0x00,
}
