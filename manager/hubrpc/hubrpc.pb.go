// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hubrpc.proto

/*
Package hubrpc is a generated protocol buffer package.

It is generated from these files:
	hubrpc.proto

It has these top-level messages:
	SetStateRequest
	SetStateResponse
	Channel
*/
package hubrpc

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

type SetStateRequest struct {
	// Time of the last sync with state of the router.
	Time uint64 `protobuf:"varint,1,opt,name=time" json:"time,omitempty"`
	// Channels is the set of channels of local router lightning network.
	Channels []*Channel `protobuf:"bytes,2,rep,name=channels" json:"channels,omitempty"`
}

func (m *SetStateRequest) Reset()                    { *m = SetStateRequest{} }
func (m *SetStateRequest) String() string            { return proto.CompactTextString(m) }
func (*SetStateRequest) ProtoMessage()               {}
func (*SetStateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *SetStateRequest) GetTime() uint64 {
	if m != nil {
		return m.Time
	}
	return 0
}

func (m *SetStateRequest) GetChannels() []*Channel {
	if m != nil {
		return m.Channels
	}
	return nil
}

type SetStateResponse struct {
}

func (m *SetStateResponse) Reset()                    { *m = SetStateResponse{} }
func (m *SetStateResponse) String() string            { return proto.CompactTextString(m) }
func (*SetStateResponse) ProtoMessage()               {}
func (*SetStateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type Channel struct {
	UserId        uint64 `protobuf:"varint,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	ChannelId     uint64 `protobuf:"varint,2,opt,name=channel_id,json=channelId" json:"channel_id,omitempty"`
	RouterBalance uint64 `protobuf:"varint,3,opt,name=router_balance,json=routerBalance" json:"router_balance,omitempty"`
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

func (m *Channel) GetRouterBalance() uint64 {
	if m != nil {
		return m.RouterBalance
	}
	return 0
}

func init() {
	proto.RegisterType((*SetStateRequest)(nil), "hubrpc.SetStateRequest")
	proto.RegisterType((*SetStateResponse)(nil), "hubrpc.SetStateResponse")
	proto.RegisterType((*Channel)(nil), "hubrpc.Channel")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Manager service

type ManagerClient interface {
	//
	// SetState is used to change the local state of the
	SetState(ctx context.Context, in *SetStateRequest, opts ...grpc.CallOption) (*SetStateResponse, error)
}

type managerClient struct {
	cc *grpc.ClientConn
}

func NewManagerClient(cc *grpc.ClientConn) ManagerClient {
	return &managerClient{cc}
}

func (c *managerClient) SetState(ctx context.Context, in *SetStateRequest, opts ...grpc.CallOption) (*SetStateResponse, error) {
	out := new(SetStateResponse)
	err := grpc.Invoke(ctx, "/hubrpc.Manager/SetState", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Manager service

type ManagerServer interface {
	//
	// SetState is used to change the local state of the
	SetState(context.Context, *SetStateRequest) (*SetStateResponse, error)
}

func RegisterManagerServer(s *grpc.Server, srv ManagerServer) {
	s.RegisterService(&_Manager_serviceDesc, srv)
}

func _Manager_SetState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetStateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ManagerServer).SetState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/hubrpc.Manager/SetState",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ManagerServer).SetState(ctx, req.(*SetStateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Manager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "hubrpc.Manager",
	HandlerType: (*ManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetState",
			Handler:    _Manager_SetState_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hubrpc.proto",
}

func init() { proto.RegisterFile("hubrpc.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 220 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0x4b, 0x4b, 0xc4, 0x30,
	0x14, 0x85, 0x99, 0x07, 0xed, 0x78, 0x7d, 0x8c, 0xdc, 0xcd, 0x04, 0x41, 0x18, 0x0a, 0x42, 0x41,
	0xe8, 0xa2, 0xae, 0xdd, 0xe8, 0xc6, 0x2e, 0xdc, 0xa4, 0x3f, 0xa0, 0xa4, 0xed, 0xc5, 0x16, 0x6a,
	0x52, 0xf3, 0xf8, 0xff, 0xd2, 0x24, 0x55, 0xd0, 0x5d, 0xce, 0x77, 0x0e, 0x27, 0x27, 0x81, 0xab,
	0xc1, 0xb5, 0x7a, 0xee, 0x8a, 0x59, 0x2b, 0xab, 0x30, 0x09, 0x2a, 0xe3, 0x70, 0xac, 0xc9, 0xd6,
	0x56, 0x58, 0xe2, 0xf4, 0xe5, 0xc8, 0x58, 0x44, 0xd8, 0xdb, 0xf1, 0x93, 0xd8, 0xe6, 0xbc, 0xc9,
	0xf7, 0xdc, 0x9f, 0xf1, 0x11, 0x0e, 0xdd, 0x20, 0xa4, 0xa4, 0xc9, 0xb0, 0xed, 0x79, 0x97, 0x5f,
	0x96, 0xc7, 0x22, 0xf6, 0xbd, 0x06, 0xce, 0x7f, 0x02, 0x19, 0xc2, 0xed, 0x6f, 0xa7, 0x99, 0x95,
	0x34, 0x94, 0x0d, 0x90, 0xc6, 0x20, 0x9e, 0x20, 0x75, 0x86, 0x74, 0x33, 0xf6, 0xf1, 0x8a, 0x64,
	0x91, 0x55, 0x8f, 0xf7, 0x00, 0xb1, 0x63, 0xf1, 0xb6, 0xde, 0xbb, 0x88, 0xa4, 0xea, 0xf1, 0x01,
	0x6e, 0xb4, 0x72, 0x96, 0x74, 0xd3, 0x8a, 0x49, 0xc8, 0x8e, 0xd8, 0xce, 0x47, 0xae, 0x03, 0x7d,
	0x09, 0xb0, 0x7c, 0x83, 0xf4, 0x5d, 0x48, 0xf1, 0x41, 0x1a, 0x9f, 0xe1, 0xb0, 0x0e, 0xc1, 0xd3,
	0xba, 0xf7, 0xcf, 0x73, 0xef, 0xd8, 0x7f, 0x23, 0x6c, 0x6e, 0x13, 0xff, 0x55, 0x4f, 0xdf, 0x01,
	0x00, 0x00, 0xff, 0xff, 0xf8, 0x12, 0x48, 0xba, 0x3a, 0x01, 0x00, 0x00,
}
