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
	// ...
	ChannelChangeType_openning ChannelChangeType = 0
	// Opened is used when this channel was just created, and haven't been in
	// local network before.
	ChannelChangeType_opened ChannelChangeType = 1
	// ...
	ChannelChangeType_closing ChannelChangeType = 2
	// Closed is used when number locked funds / balances of both channel
	// participant equal to zero.
	ChannelChangeType_closed ChannelChangeType = 3
	// ...
	ChannelChangeType_udpating ChannelChangeType = 4
	// Udpated is used when one of the participants decides to update its
	// channel balance.
	ChannelChangeType_udpated ChannelChangeType = 5
)

var ChannelChangeType_name = map[int32]string{
	0: "openning",
	1: "opened",
	2: "closing",
	3: "closed",
	4: "udpating",
	5: "udpated",
}
var ChannelChangeType_value = map[string]int32{
	"openning": 0,
	"opened":   1,
	"closing":  2,
	"closed":   3,
	"udpating": 4,
	"udpated":  5,
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
	IsPending     bool   `protobuf:"varint,5,opt,name=is_pending,json=isPending" json:"is_pending,omitempty"`
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

func (m *Channel) GetIsPending() bool {
	if m != nil {
		return m.IsPending
	}
	return false
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
	return ChannelChangeType_openning
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
	// 581 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0xd1, 0x6e, 0xd3, 0x3e,
	0x14, 0xc6, 0xeb, 0x26, 0x4d, 0xbb, 0xd3, 0xb5, 0xcb, 0xfc, 0xd7, 0x7f, 0x04, 0xd0, 0x50, 0xa9,
	0x84, 0x98, 0x36, 0x6d, 0x17, 0x70, 0xcf, 0xc5, 0x86, 0xc4, 0x26, 0x71, 0x31, 0x79, 0x70, 0x1d,
	0x79, 0xf1, 0x69, 0xb0, 0x94, 0x39, 0x91, 0xed, 0x4c, 0xec, 0x5d, 0x78, 0x02, 0x5e, 0x81, 0x47,
	0xe0, 0x8e, 0x27, 0x42, 0x76, 0x9c, 0xd2, 0x89, 0x09, 0xb8, 0xaa, 0xcf, 0x77, 0x7e, 0xa7, 0xfd,
	0x8e, 0x3f, 0xab, 0xb0, 0x55, 0xd5, 0xe5, 0x49, 0xa3, 0x6b, 0x5b, 0xd3, 0xa4, 0xaa, 0xcb, 0x12,
	0xf5, 0xf2, 0x1b, 0x81, 0xe8, 0x7d, 0x5d, 0x52, 0x0a, 0xb1, 0x95, 0x37, 0x98, 0x91, 0x05, 0x39,
	0x88, 0x98, 0x3f, 0xd3, 0x23, 0x18, 0x19, 0xcb, 0x2d, 0x66, 0xc3, 0x05, 0x39, 0x98, 0xbe, 0xfa,
	0xef, 0xa4, 0x9b, 0x39, 0x61, 0x75, 0x6b, 0x51, 0x5f, 0xb9, 0xd6, 0xf9, 0x80, 0x75, 0x0c, 0x3d,
	0x82, 0x71, 0xc3, 0xef, 0x6e, 0x50, 0xd9, 0x2c, 0xf2, 0xf8, 0x4e, 0x8f, 0x5f, 0x76, 0xf2, 0xf9,
	0x80, 0xf5, 0x04, 0x7d, 0x03, 0xf3, 0xe2, 0x13, 0x57, 0x0a, 0xab, 0xdc, 0x7d, 0x96, 0x98, 0xc5,
	0x7e, 0xe6, 0xff, 0x7e, 0xe6, 0xac, 0xeb, 0x9e, 0xf9, 0xe6, 0xf9, 0x80, 0xcd, 0x8a, 0x4d, 0xe1,
	0x34, 0x81, 0x58, 0x70, 0xcb, 0x97, 0xdf, 0x09, 0x4c, 0x37, 0xdc, 0xd0, 0x23, 0x98, 0x04, 0xd0,
	0x64, 0x64, 0x11, 0x6d, 0xba, 0x08, 0xdf, 0xc8, 0xd6, 0x00, 0x7d, 0x0e, 0xdb, 0x2b, 0x8d, 0x98,
	0x5f, 0xf3, 0x8a, 0xab, 0xa2, 0xdb, 0x32, 0x66, 0x53, 0xa7, 0x9d, 0x76, 0x12, 0x7d, 0x09, 0x3b,
	0x0d, 0x2a, 0x21, 0x55, 0xb9, 0xa6, 0x22, 0x4f, 0xcd, 0x83, 0xdc, 0x83, 0x67, 0xf0, 0x8c, 0xdf,
	0xa2, 0xe6, 0x25, 0x86, 0x85, 0xf2, 0xb6, 0x11, 0xdc, 0x62, 0x2e, 0x5a, 0xcd, 0xad, 0xac, 0x95,
	0x5f, 0x30, 0x66, 0x4f, 0x03, 0xd5, 0xed, 0xf1, 0xd1, 0x33, 0x6f, 0x03, 0xb2, 0xfc, 0x4a, 0x60,
	0x1c, 0x6c, 0xd2, 0x47, 0x30, 0x6e, 0x0d, 0xea, 0x5c, 0x0a, 0x1f, 0x49, 0xcc, 0x12, 0x57, 0x5e,
	0x08, 0xba, 0x0f, 0xd0, 0x5f, 0x9d, 0x14, 0xc1, 0xf3, 0x56, 0x50, 0x2e, 0x84, 0x5b, 0xca, 0xcf,
	0xdd, 0xb7, 0x3b, 0x75, 0x5a, 0xef, 0xf5, 0x05, 0xcc, 0xb5, 0xbf, 0xb3, 0x35, 0xd4, 0x79, 0x9b,
	0x75, 0x6a, 0x8f, 0xed, 0x03, 0x48, 0x93, 0x87, 0x3d, 0xb3, 0xd1, 0x82, 0x1c, 0x4c, 0xd8, 0x96,
	0x34, 0x97, 0x9d, 0xb0, 0xfc, 0x42, 0x60, 0x1c, 0x92, 0xa5, 0xc7, 0x90, 0xb8, 0x47, 0xd0, 0x1a,
	0xef, 0x75, 0xfe, 0x2b, 0xc6, 0x00, 0x5c, 0xf9, 0x26, 0x0b, 0x10, 0xdd, 0x83, 0xc4, 0xa0, 0x12,
	0xa8, 0x83, 0xfd, 0x50, 0xd1, 0x27, 0x30, 0xd1, 0x58, 0xa0, 0xbc, 0x45, 0x1d, 0x7c, 0xaf, 0x6b,
	0x37, 0xc3, 0x6f, 0xea, 0x56, 0x59, 0xef, 0x24, 0x66, 0xa1, 0x72, 0x3a, 0x72, 0xad, 0x50, 0x64,
	0x89, 0x7f, 0xb9, 0xa1, 0x5a, 0xfe, 0x20, 0x30, 0xbb, 0xf7, 0x88, 0xe8, 0x31, 0xc4, 0xf6, 0xae,
	0xc1, 0x60, 0xf1, 0xf1, 0x83, 0x2f, 0xed, 0xc3, 0x5d, 0x83, 0xcc, 0x63, 0x9b, 0x01, 0x0c, 0xff,
	0x10, 0x40, 0xf4, 0xb7, 0x00, 0xe2, 0x7f, 0x09, 0x60, 0xf4, 0x50, 0x00, 0x29, 0x44, 0x2b, 0x44,
	0xbf, 0x57, 0xcc, 0xdc, 0xf1, 0xf0, 0x1d, 0xcc, 0xee, 0xdd, 0x28, 0x9d, 0xc2, 0xd8, 0xb4, 0x45,
	0x81, 0xc6, 0xa4, 0x03, 0xba, 0x07, 0xb4, 0x55, 0xa6, 0x5d, 0xad, 0x64, 0x21, 0x51, 0xd9, 0x7c,
	0xd5, 0x2a, 0x61, 0x52, 0x42, 0x77, 0x61, 0x86, 0x9f, 0x2d, 0x6a, 0xc5, 0xab, 0x7c, 0xc5, 0x65,
	0x95, 0x0e, 0x0f, 0x39, 0xec, 0xfe, 0xb6, 0x37, 0xdd, 0x86, 0x49, 0xdd, 0xa0, 0x52, 0x52, 0x95,
	0xe9, 0x80, 0x02, 0x24, 0xae, 0x42, 0x91, 0x12, 0xf7, 0x33, 0x45, 0x55, 0x1b, 0xd7, 0x18, 0xba,
	0x86, 0x2b, 0x50, 0xa4, 0x91, 0x1b, 0x69, 0x45, 0xc3, 0xad, 0xeb, 0xc4, 0x0e, 0xf3, 0x15, 0x8a,
	0x74, 0x74, 0x9d, 0xf8, 0xff, 0x99, 0xd7, 0x3f, 0x03, 0x00, 0x00, 0xff, 0xff, 0xb3, 0xb7, 0xbe,
	0x2b, 0x74, 0x04, 0x00, 0x00,
}
