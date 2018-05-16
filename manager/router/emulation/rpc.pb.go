// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rpc.proto

/*
Package emulation is a generated protocol buffer package.

It is generated from these files:
	rpc.proto

It has these top-level messages:
	SendPaymentRequest
	SendPaymentResponse
	OpenChannelRequest
	OpenChannelResponse
	CloseChannelRequest
	CloseChannelResponse
	SetBlockGenDurationRequest
	SetBlockGenDurationResponse
	SetBlockchainFeeRequest
	SetBlockchainFeeResponse
*/
package emulation

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SendPaymentRequest struct {
	// Sender is the id of user which sends the payment. If equals zero than we
	// believe that router is the sender.
	Sender string `protobuf:"bytes,1,opt,name=sender" json:"sender,omitempty"`
	// Receiver is the id of user who receive the payment. If equals zero than we
	// believe that router is the receiver.
	Receiver string `protobuf:"bytes,2,opt,name=receiver" json:"receiver,omitempty"`
	Amount   int64  `protobuf:"varint,4,opt,name=amount" json:"amount,omitempty"`
}

func (m *SendPaymentRequest) Reset()                    { *m = SendPaymentRequest{} }
func (m *SendPaymentRequest) String() string            { return proto.CompactTextString(m) }
func (*SendPaymentRequest) ProtoMessage()               {}
func (*SendPaymentRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *SendPaymentRequest) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *SendPaymentRequest) GetReceiver() string {
	if m != nil {
		return m.Receiver
	}
	return ""
}

func (m *SendPaymentRequest) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

type SendPaymentResponse struct {
}

func (m *SendPaymentResponse) Reset()                    { *m = SendPaymentResponse{} }
func (m *SendPaymentResponse) String() string            { return proto.CompactTextString(m) }
func (*SendPaymentResponse) ProtoMessage()               {}
func (*SendPaymentResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type OpenChannelRequest struct {
	UserId       string `protobuf:"bytes,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	LockedByUser int64  `protobuf:"varint,2,opt,name=locked_by_user,json=lockedByUser" json:"locked_by_user,omitempty"`
}

func (m *OpenChannelRequest) Reset()                    { *m = OpenChannelRequest{} }
func (m *OpenChannelRequest) String() string            { return proto.CompactTextString(m) }
func (*OpenChannelRequest) ProtoMessage()               {}
func (*OpenChannelRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *OpenChannelRequest) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *OpenChannelRequest) GetLockedByUser() int64 {
	if m != nil {
		return m.LockedByUser
	}
	return 0
}

type OpenChannelResponse struct {
	ChannelId string `protobuf:"bytes,3,opt,name=channel_id,json=channelId" json:"channel_id,omitempty"`
}

func (m *OpenChannelResponse) Reset()                    { *m = OpenChannelResponse{} }
func (m *OpenChannelResponse) String() string            { return proto.CompactTextString(m) }
func (*OpenChannelResponse) ProtoMessage()               {}
func (*OpenChannelResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *OpenChannelResponse) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

type CloseChannelRequest struct {
	ChannelId string `protobuf:"bytes,3,opt,name=channel_id,json=channelId" json:"channel_id,omitempty"`
}

func (m *CloseChannelRequest) Reset()                    { *m = CloseChannelRequest{} }
func (m *CloseChannelRequest) String() string            { return proto.CompactTextString(m) }
func (*CloseChannelRequest) ProtoMessage()               {}
func (*CloseChannelRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *CloseChannelRequest) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

type CloseChannelResponse struct {
}

func (m *CloseChannelResponse) Reset()                    { *m = CloseChannelResponse{} }
func (m *CloseChannelResponse) String() string            { return proto.CompactTextString(m) }
func (*CloseChannelResponse) ProtoMessage()               {}
func (*CloseChannelResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type SetBlockGenDurationRequest struct {
	// Duration is the number of millisecond channel has to be in the locked
	// state after it openes or updates.
	Duration int64 `protobuf:"varint,3,opt,name=duration" json:"duration,omitempty"`
}

func (m *SetBlockGenDurationRequest) Reset()                    { *m = SetBlockGenDurationRequest{} }
func (m *SetBlockGenDurationRequest) String() string            { return proto.CompactTextString(m) }
func (*SetBlockGenDurationRequest) ProtoMessage()               {}
func (*SetBlockGenDurationRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *SetBlockGenDurationRequest) GetDuration() int64 {
	if m != nil {
		return m.Duration
	}
	return 0
}

type SetBlockGenDurationResponse struct {
}

func (m *SetBlockGenDurationResponse) Reset()                    { *m = SetBlockGenDurationResponse{} }
func (m *SetBlockGenDurationResponse) String() string            { return proto.CompactTextString(m) }
func (*SetBlockGenDurationResponse) ProtoMessage()               {}
func (*SetBlockGenDurationResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type SetBlockchainFeeRequest struct {
	// fee is the number expressed in satoshis, which blockchain takes for
	// making an computation, transaction creation, i.e. channel updates.
	// TODO(andrew.shvv) Make it real world friendly - calculate fee
	// depending on tx size and fee per kilobyte.
	Fee int64 `protobuf:"varint,3,opt,name=fee" json:"fee,omitempty"`
}

func (m *SetBlockchainFeeRequest) Reset()                    { *m = SetBlockchainFeeRequest{} }
func (m *SetBlockchainFeeRequest) String() string            { return proto.CompactTextString(m) }
func (*SetBlockchainFeeRequest) ProtoMessage()               {}
func (*SetBlockchainFeeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *SetBlockchainFeeRequest) GetFee() int64 {
	if m != nil {
		return m.Fee
	}
	return 0
}

type SetBlockchainFeeResponse struct {
}

func (m *SetBlockchainFeeResponse) Reset()                    { *m = SetBlockchainFeeResponse{} }
func (m *SetBlockchainFeeResponse) String() string            { return proto.CompactTextString(m) }
func (*SetBlockchainFeeResponse) ProtoMessage()               {}
func (*SetBlockchainFeeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func init() {
	proto.RegisterType((*SendPaymentRequest)(nil), "emulation.SendPaymentRequest")
	proto.RegisterType((*SendPaymentResponse)(nil), "emulation.SendPaymentResponse")
	proto.RegisterType((*OpenChannelRequest)(nil), "emulation.OpenChannelRequest")
	proto.RegisterType((*OpenChannelResponse)(nil), "emulation.OpenChannelResponse")
	proto.RegisterType((*CloseChannelRequest)(nil), "emulation.CloseChannelRequest")
	proto.RegisterType((*CloseChannelResponse)(nil), "emulation.CloseChannelResponse")
	proto.RegisterType((*SetBlockGenDurationRequest)(nil), "emulation.SetBlockGenDurationRequest")
	proto.RegisterType((*SetBlockGenDurationResponse)(nil), "emulation.SetBlockGenDurationResponse")
	proto.RegisterType((*SetBlockchainFeeRequest)(nil), "emulation.SetBlockchainFeeRequest")
	proto.RegisterType((*SetBlockchainFeeResponse)(nil), "emulation.SetBlockchainFeeResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Emulator service

type EmulatorClient interface {
	//
	// SendPayment is used to emulate the activity of one user sending payment
	// to another within the local router network.
	SendPayment(ctx context.Context, in *SendPaymentRequest, opts ...grpc.CallOption) (*SendPaymentResponse, error)
	//
	// OpenChannel is used to emulate that user has opened the channel with the
	// router.
	OpenChannel(ctx context.Context, in *OpenChannelRequest, opts ...grpc.CallOption) (*OpenChannelResponse, error)
	//
	// CloseChannel is used to emulate that user has closed the channel with the
	// router.
	CloseChannel(ctx context.Context, in *CloseChannelRequest, opts ...grpc.CallOption) (*CloseChannelResponse, error)
	//
	// SetBlockGenDuration is used to set the time which is needed for blokc
	// to be generatedtime. This would impact channel creation, channel
	// update and channel close.
	SetBlockGenDuration(ctx context.Context, in *SetBlockGenDurationRequest, opts ...grpc.CallOption) (*SetBlockGenDurationResponse, error)
	//
	// SetBlockchainFee is used to set the fee which blockchain takes for
	// making an computation, transaction creation, i.e. channel updates.
	SetBlockchainFee(ctx context.Context, in *SetBlockchainFeeRequest, opts ...grpc.CallOption) (*SetBlockchainFeeResponse, error)
}

type emulatorClient struct {
	cc *grpc.ClientConn
}

func NewEmulatorClient(cc *grpc.ClientConn) EmulatorClient {
	return &emulatorClient{cc}
}

func (c *emulatorClient) SendPayment(ctx context.Context, in *SendPaymentRequest, opts ...grpc.CallOption) (*SendPaymentResponse, error) {
	out := new(SendPaymentResponse)
	err := grpc.Invoke(ctx, "/emulation.Emulator/SendPayment", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *emulatorClient) OpenChannel(ctx context.Context, in *OpenChannelRequest, opts ...grpc.CallOption) (*OpenChannelResponse, error) {
	out := new(OpenChannelResponse)
	err := grpc.Invoke(ctx, "/emulation.Emulator/OpenChannel", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *emulatorClient) CloseChannel(ctx context.Context, in *CloseChannelRequest, opts ...grpc.CallOption) (*CloseChannelResponse, error) {
	out := new(CloseChannelResponse)
	err := grpc.Invoke(ctx, "/emulation.Emulator/CloseChannel", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *emulatorClient) SetBlockGenDuration(ctx context.Context, in *SetBlockGenDurationRequest, opts ...grpc.CallOption) (*SetBlockGenDurationResponse, error) {
	out := new(SetBlockGenDurationResponse)
	err := grpc.Invoke(ctx, "/emulation.Emulator/SetBlockGenDuration", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *emulatorClient) SetBlockchainFee(ctx context.Context, in *SetBlockchainFeeRequest, opts ...grpc.CallOption) (*SetBlockchainFeeResponse, error) {
	out := new(SetBlockchainFeeResponse)
	err := grpc.Invoke(ctx, "/emulation.Emulator/SetBlockchainFee", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Emulator service

type EmulatorServer interface {
	//
	// SendPayment is used to emulate the activity of one user sending payment
	// to another within the local router network.
	SendPayment(context.Context, *SendPaymentRequest) (*SendPaymentResponse, error)
	//
	// OpenChannel is used to emulate that user has opened the channel with the
	// router.
	OpenChannel(context.Context, *OpenChannelRequest) (*OpenChannelResponse, error)
	//
	// CloseChannel is used to emulate that user has closed the channel with the
	// router.
	CloseChannel(context.Context, *CloseChannelRequest) (*CloseChannelResponse, error)
	//
	// SetBlockGenDuration is used to set the time which is needed for blokc
	// to be generatedtime. This would impact channel creation, channel
	// update and channel close.
	SetBlockGenDuration(context.Context, *SetBlockGenDurationRequest) (*SetBlockGenDurationResponse, error)
	//
	// SetBlockchainFee is used to set the fee which blockchain takes for
	// making an computation, transaction creation, i.e. channel updates.
	SetBlockchainFee(context.Context, *SetBlockchainFeeRequest) (*SetBlockchainFeeResponse, error)
}

func RegisterEmulatorServer(s *grpc.Server, srv EmulatorServer) {
	s.RegisterService(&_Emulator_serviceDesc, srv)
}

func _Emulator_SendPayment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendPaymentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmulatorServer).SendPayment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/emulation.Emulator/SendPayment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmulatorServer).SendPayment(ctx, req.(*SendPaymentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Emulator_OpenChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OpenChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmulatorServer).OpenChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/emulation.Emulator/OpenChannel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmulatorServer).OpenChannel(ctx, req.(*OpenChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Emulator_CloseChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmulatorServer).CloseChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/emulation.Emulator/CloseChannel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmulatorServer).CloseChannel(ctx, req.(*CloseChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Emulator_SetBlockGenDuration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetBlockGenDurationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmulatorServer).SetBlockGenDuration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/emulation.Emulator/SetBlockGenDuration",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmulatorServer).SetBlockGenDuration(ctx, req.(*SetBlockGenDurationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Emulator_SetBlockchainFee_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetBlockchainFeeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmulatorServer).SetBlockchainFee(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/emulation.Emulator/SetBlockchainFee",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmulatorServer).SetBlockchainFee(ctx, req.(*SetBlockchainFeeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Emulator_serviceDesc = grpc.ServiceDesc{
	ServiceName: "emulation.Emulator",
	HandlerType: (*EmulatorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendPayment",
			Handler:    _Emulator_SendPayment_Handler,
		},
		{
			MethodName: "OpenChannel",
			Handler:    _Emulator_OpenChannel_Handler,
		},
		{
			MethodName: "CloseChannel",
			Handler:    _Emulator_CloseChannel_Handler,
		},
		{
			MethodName: "SetBlockGenDuration",
			Handler:    _Emulator_SetBlockGenDuration_Handler,
		},
		{
			MethodName: "SetBlockchainFee",
			Handler:    _Emulator_SetBlockchainFee_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc.proto",
}

func init() { proto.RegisterFile("rpc.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 399 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x93, 0x5b, 0x8b, 0xda, 0x40,
	0x14, 0xc7, 0xb1, 0x29, 0xd6, 0x1c, 0xa5, 0xc8, 0xd8, 0x6a, 0x98, 0x62, 0x2b, 0xe9, 0x05, 0xa1,
	0xe0, 0x43, 0xdb, 0x87, 0x3e, 0xeb, 0x5e, 0x10, 0x04, 0x97, 0xc8, 0x3e, 0xed, 0x83, 0x1b, 0x33,
	0x67, 0x31, 0x6c, 0x9c, 0xc9, 0x4e, 0x92, 0x05, 0xbf, 0xda, 0x7e, 0xba, 0x65, 0x92, 0x89, 0x24,
	0x26, 0xae, 0x6f, 0x9e, 0xdb, 0xef, 0x9f, 0x39, 0xff, 0x23, 0x98, 0x32, 0xf4, 0x26, 0xa1, 0x14,
	0xb1, 0x20, 0x26, 0xee, 0x92, 0xc0, 0x8d, 0x7d, 0xc1, 0xed, 0x7b, 0x20, 0x2b, 0xe4, 0xec, 0xc6,
	0xdd, 0xef, 0x90, 0xc7, 0x0e, 0x3e, 0x25, 0x18, 0xc5, 0xa4, 0x0f, 0xcd, 0x08, 0x39, 0x43, 0x69,
	0x35, 0x46, 0x8d, 0xb1, 0xe9, 0xe8, 0x88, 0x50, 0x68, 0x49, 0xf4, 0xd0, 0x7f, 0x46, 0x69, 0xbd,
	0x4b, 0x2b, 0x87, 0x58, 0xcd, 0xb8, 0x3b, 0x91, 0xf0, 0xd8, 0x7a, 0x3f, 0x6a, 0x8c, 0x0d, 0x47,
	0x47, 0xf6, 0x67, 0xe8, 0x95, 0x14, 0xa2, 0x50, 0xf0, 0x08, 0xed, 0x15, 0x90, 0x65, 0x88, 0x7c,
	0xb6, 0x75, 0x39, 0xc7, 0x20, 0x17, 0x1e, 0xc0, 0x87, 0x24, 0x42, 0xb9, 0xf6, 0x59, 0xae, 0xac,
	0xc2, 0x39, 0x23, 0x3f, 0xe0, 0x63, 0x20, 0xbc, 0x47, 0x64, 0xeb, 0xcd, 0x7e, 0xad, 0x72, 0xa9,
	0xbe, 0xe1, 0x74, 0xb2, 0xec, 0x74, 0x7f, 0x1b, 0xa1, 0xb4, 0xff, 0x41, 0xaf, 0x04, 0xcd, 0xb4,
	0xc8, 0x10, 0xc0, 0xcb, 0x52, 0x0a, 0x6c, 0xa4, 0x60, 0x53, 0x67, 0xe6, 0x4c, 0x4d, 0xcd, 0x02,
	0x11, 0xe1, 0xd1, 0xb7, 0x9c, 0x99, 0xea, 0xc3, 0xa7, 0xf2, 0x94, 0x7e, 0xd8, 0x7f, 0xa0, 0x2b,
	0x8c, 0xa7, 0xea, 0xbb, 0xae, 0x91, 0x5f, 0x24, 0x32, 0x5d, 0x74, 0x0e, 0xa5, 0xd0, 0x62, 0x3a,
	0x95, 0x22, 0x0d, 0xe7, 0x10, 0xdb, 0x43, 0xf8, 0x52, 0x3b, 0xa9, 0xc1, 0xbf, 0x61, 0x90, 0x97,
	0xbd, 0xad, 0xeb, 0xf3, 0x2b, 0xc4, 0x9c, 0xda, 0x05, 0xe3, 0x01, 0x51, 0x03, 0xd5, 0x4f, 0x9b,
	0x82, 0x55, 0x6d, 0xce, 0x40, 0x7f, 0x5e, 0x0c, 0x68, 0x5d, 0xa6, 0x17, 0x20, 0x24, 0x59, 0x40,
	0xbb, 0x60, 0x0f, 0x19, 0x4e, 0x0e, 0xb7, 0x31, 0xa9, 0x1e, 0x06, 0xfd, 0x7a, 0xaa, 0xac, 0x37,
	0xbd, 0x80, 0x76, 0xc1, 0x80, 0x12, 0xad, 0xea, 0x76, 0x89, 0x56, 0xe7, 0xdb, 0x12, 0x3a, 0xc5,
	0x15, 0x93, 0x62, 0x7f, 0x8d, 0x63, 0xf4, 0xdb, 0xc9, 0xba, 0x06, 0x32, 0x75, 0x8b, 0x95, 0x0d,
	0x93, 0x9f, 0xa5, 0x57, 0x9d, 0xf2, 0x8e, 0xfe, 0x3a, 0xd7, 0xa6, 0x55, 0xee, 0xa0, 0x7b, 0xbc,
	0x7b, 0x62, 0xd7, 0xcc, 0x1e, 0xb9, 0x48, 0xbf, 0xbf, 0xd9, 0x93, 0xc1, 0x37, 0xcd, 0xf4, 0x2f,
	0xfc, 0xf7, 0x35, 0x00, 0x00, 0xff, 0xff, 0xb5, 0xa1, 0xed, 0x37, 0xcf, 0x03, 0x00, 0x00,
}
