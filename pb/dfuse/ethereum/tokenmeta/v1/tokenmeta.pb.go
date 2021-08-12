// Code generated by protoc-gen-go. DO NOT EDIT.
// source: dfuse/ethereum/tokenmeta/v1/tokenmeta.proto

package pbtokenmeta

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type GetTokenRequest struct {
	Address              string   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetTokenRequest) Reset()         { *m = GetTokenRequest{} }
func (m *GetTokenRequest) String() string { return proto.CompactTextString(m) }
func (*GetTokenRequest) ProtoMessage()    {}
func (*GetTokenRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_75ae7669958e267d, []int{0}
}

func (m *GetTokenRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetTokenRequest.Unmarshal(m, b)
}
func (m *GetTokenRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetTokenRequest.Marshal(b, m, deterministic)
}
func (m *GetTokenRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetTokenRequest.Merge(m, src)
}
func (m *GetTokenRequest) XXX_Size() int {
	return xxx_messageInfo_GetTokenRequest.Size(m)
}
func (m *GetTokenRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetTokenRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetTokenRequest proto.InternalMessageInfo

func (m *GetTokenRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type StreamTokenRequest struct {
	Addresses            []string `protobuf:"bytes,1,rep,name=addresses,proto3" json:"addresses,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StreamTokenRequest) Reset()         { *m = StreamTokenRequest{} }
func (m *StreamTokenRequest) String() string { return proto.CompactTextString(m) }
func (*StreamTokenRequest) ProtoMessage()    {}
func (*StreamTokenRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_75ae7669958e267d, []int{1}
}

func (m *StreamTokenRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StreamTokenRequest.Unmarshal(m, b)
}
func (m *StreamTokenRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StreamTokenRequest.Marshal(b, m, deterministic)
}
func (m *StreamTokenRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StreamTokenRequest.Merge(m, src)
}
func (m *StreamTokenRequest) XXX_Size() int {
	return xxx_messageInfo_StreamTokenRequest.Size(m)
}
func (m *StreamTokenRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StreamTokenRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StreamTokenRequest proto.InternalMessageInfo

func (m *StreamTokenRequest) GetAddresses() []string {
	if m != nil {
		return m.Addresses
	}
	return nil
}

type Token struct {
	Address              []byte   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Name                 string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Symbol               string   `protobuf:"bytes,3,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Decimals             uint32   `protobuf:"varint,4,opt,name=decimals,proto3" json:"decimals,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Token) Reset()         { *m = Token{} }
func (m *Token) String() string { return proto.CompactTextString(m) }
func (*Token) ProtoMessage()    {}
func (*Token) Descriptor() ([]byte, []int) {
	return fileDescriptor_75ae7669958e267d, []int{2}
}

func (m *Token) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Token.Unmarshal(m, b)
}
func (m *Token) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Token.Marshal(b, m, deterministic)
}
func (m *Token) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Token.Merge(m, src)
}
func (m *Token) XXX_Size() int {
	return xxx_messageInfo_Token.Size(m)
}
func (m *Token) XXX_DiscardUnknown() {
	xxx_messageInfo_Token.DiscardUnknown(m)
}

var xxx_messageInfo_Token proto.InternalMessageInfo

func (m *Token) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *Token) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Token) GetSymbol() string {
	if m != nil {
		return m.Symbol
	}
	return ""
}

func (m *Token) GetDecimals() uint32 {
	if m != nil {
		return m.Decimals
	}
	return 0
}

func init() {
	proto.RegisterType((*GetTokenRequest)(nil), "dfuse.ethereum.tokenmeta.v1.GetTokenRequest")
	proto.RegisterType((*StreamTokenRequest)(nil), "dfuse.ethereum.tokenmeta.v1.StreamTokenRequest")
	proto.RegisterType((*Token)(nil), "dfuse.ethereum.tokenmeta.v1.Token")
}

func init() {
	proto.RegisterFile("dfuse/ethereum/tokenmeta/v1/tokenmeta.proto", fileDescriptor_75ae7669958e267d)
}

var fileDescriptor_75ae7669958e267d = []byte{
	// 281 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x4f, 0x4b, 0xc3, 0x40,
	0x10, 0xc5, 0x89, 0xfd, 0x63, 0x33, 0x2a, 0xc2, 0x1c, 0x24, 0x54, 0x0f, 0x25, 0xa7, 0x40, 0x75,
	0xd7, 0xd6, 0xa3, 0x37, 0x2f, 0x9e, 0x44, 0x48, 0x3d, 0x89, 0x97, 0x4d, 0x33, 0xda, 0x60, 0xb7,
	0x1b, 0xb3, 0x93, 0x82, 0x9f, 0xd1, 0x2f, 0x25, 0x5d, 0xf2, 0x47, 0x2b, 0x04, 0x6f, 0xf3, 0x26,
	0xbf, 0x79, 0xd9, 0x7d, 0x3b, 0x30, 0x4d, 0x5f, 0x4b, 0x4b, 0x92, 0x78, 0x45, 0x05, 0x95, 0x5a,
	0xb2, 0x79, 0xa7, 0x8d, 0x26, 0x56, 0x72, 0x3b, 0x6b, 0x85, 0xc8, 0x0b, 0xc3, 0x06, 0xcf, 0x1d,
	0x2c, 0x6a, 0x58, 0xb4, 0xdf, 0xb7, 0xb3, 0x70, 0x0a, 0xa7, 0xf7, 0xc4, 0x4f, 0xbb, 0x56, 0x4c,
	0x1f, 0x25, 0x59, 0xc6, 0x00, 0x0e, 0x55, 0x9a, 0x16, 0x64, 0x6d, 0xe0, 0x4d, 0xbc, 0xc8, 0x8f,
	0x6b, 0x19, 0xce, 0x01, 0x17, 0x5c, 0x90, 0xd2, 0xbf, 0xf8, 0x0b, 0xf0, 0x2b, 0x80, 0x76, 0x13,
	0xbd, 0xc8, 0x8f, 0xdb, 0x46, 0x98, 0xc1, 0xc0, 0xd1, 0xfb, 0xb6, 0xc7, 0x8d, 0x2d, 0x22, 0xf4,
	0x37, 0x4a, 0x53, 0x70, 0xe0, 0xfe, 0xe6, 0x6a, 0x3c, 0x83, 0xa1, 0xfd, 0xd4, 0x89, 0x59, 0x07,
	0x3d, 0xd7, 0xad, 0x14, 0x8e, 0x61, 0x94, 0xd2, 0x32, 0xd3, 0x6a, 0x6d, 0x83, 0xfe, 0xc4, 0x8b,
	0x4e, 0xe2, 0x46, 0xcf, 0xbf, 0x3c, 0x18, 0x2c, 0x58, 0x31, 0x61, 0x0a, 0x47, 0x3f, 0x0e, 0x8a,
	0x52, 0x74, 0x44, 0x20, 0xfe, 0x5e, 0x69, 0x1c, 0x76, 0x0e, 0x38, 0xf4, 0xda, 0xc3, 0x17, 0x18,
	0xd5, 0xd9, 0xe1, 0x65, 0xe7, 0xc4, 0x5e, 0xc4, 0xff, 0xf1, 0xbf, 0x7b, 0x7c, 0x7e, 0x78, 0xcb,
	0x78, 0x55, 0x26, 0x62, 0x69, 0xb4, 0x74, 0xfc, 0x55, 0x66, 0xaa, 0xa2, 0x79, 0xf9, 0x3c, 0x91,
	0x1d, 0xbb, 0x70, 0x9b, 0x27, 0x8d, 0x4c, 0x86, 0x6e, 0x1d, 0x6e, 0xbe, 0x03, 0x00, 0x00, 0xff,
	0xff, 0xdf, 0xeb, 0x38, 0xf8, 0x3d, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// StateClient is the client API for State service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type StateClient interface {
	StreamToken(ctx context.Context, in *StreamTokenRequest, opts ...grpc.CallOption) (State_StreamTokenClient, error)
	GetToken(ctx context.Context, in *GetTokenRequest, opts ...grpc.CallOption) (*Token, error)
}

type stateClient struct {
	cc *grpc.ClientConn
}

func NewStateClient(cc *grpc.ClientConn) StateClient {
	return &stateClient{cc}
}

func (c *stateClient) StreamToken(ctx context.Context, in *StreamTokenRequest, opts ...grpc.CallOption) (State_StreamTokenClient, error) {
	stream, err := c.cc.NewStream(ctx, &_State_serviceDesc.Streams[0], "/dfuse.ethereum.tokenmeta.v1.State/StreamToken", opts...)
	if err != nil {
		return nil, err
	}
	x := &stateStreamTokenClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type State_StreamTokenClient interface {
	Recv() (*Token, error)
	grpc.ClientStream
}

type stateStreamTokenClient struct {
	grpc.ClientStream
}

func (x *stateStreamTokenClient) Recv() (*Token, error) {
	m := new(Token)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *stateClient) GetToken(ctx context.Context, in *GetTokenRequest, opts ...grpc.CallOption) (*Token, error) {
	out := new(Token)
	err := c.cc.Invoke(ctx, "/dfuse.ethereum.tokenmeta.v1.State/GetToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StateServer is the server API for State service.
type StateServer interface {
	StreamToken(*StreamTokenRequest, State_StreamTokenServer) error
	GetToken(context.Context, *GetTokenRequest) (*Token, error)
}

// UnimplementedStateServer can be embedded to have forward compatible implementations.
type UnimplementedStateServer struct {
}

func (*UnimplementedStateServer) StreamToken(req *StreamTokenRequest, srv State_StreamTokenServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamToken not implemented")
}
func (*UnimplementedStateServer) GetToken(ctx context.Context, req *GetTokenRequest) (*Token, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetToken not implemented")
}

func RegisterStateServer(s *grpc.Server, srv StateServer) {
	s.RegisterService(&_State_serviceDesc, srv)
}

func _State_StreamToken_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StreamTokenRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StateServer).StreamToken(m, &stateStreamTokenServer{stream})
}

type State_StreamTokenServer interface {
	Send(*Token) error
	grpc.ServerStream
}

type stateStreamTokenServer struct {
	grpc.ServerStream
}

func (x *stateStreamTokenServer) Send(m *Token) error {
	return x.ServerStream.SendMsg(m)
}

func _State_GetToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StateServer).GetToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dfuse.ethereum.tokenmeta.v1.State/GetToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StateServer).GetToken(ctx, req.(*GetTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _State_serviceDesc = grpc.ServiceDesc{
	ServiceName: "dfuse.ethereum.tokenmeta.v1.State",
	HandlerType: (*StateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetToken",
			Handler:    _State_GetToken_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamToken",
			Handler:       _State_StreamToken_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "dfuse/ethereum/tokenmeta/v1/tokenmeta.proto",
}
