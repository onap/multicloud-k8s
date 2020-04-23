// Code generated by protoc-gen-go. DO NOT EDIT.
// source: contextupdate.proto

package contextupdate

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

type ContextUpdateRequest struct {
	AppContext           string   `protobuf:"bytes,1,opt,name=app_context,json=appContext,proto3" json:"app_context,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ContextUpdateRequest) Reset()         { *m = ContextUpdateRequest{} }
func (m *ContextUpdateRequest) String() string { return proto.CompactTextString(m) }
func (*ContextUpdateRequest) ProtoMessage()    {}
func (*ContextUpdateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8ebb57f5310873be, []int{0}
}

func (m *ContextUpdateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ContextUpdateRequest.Unmarshal(m, b)
}
func (m *ContextUpdateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ContextUpdateRequest.Marshal(b, m, deterministic)
}
func (m *ContextUpdateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContextUpdateRequest.Merge(m, src)
}
func (m *ContextUpdateRequest) XXX_Size() int {
	return xxx_messageInfo_ContextUpdateRequest.Size(m)
}
func (m *ContextUpdateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ContextUpdateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ContextUpdateRequest proto.InternalMessageInfo

func (m *ContextUpdateRequest) GetAppContext() string {
	if m != nil {
		return m.AppContext
	}
	return ""
}

type ContextUpdateResponse struct {
	AppContextUpdated       bool     `protobuf:"varint,1,opt,name=app_context_updated,json=appContextUpdated,proto3" json:"app_context_updated,omitempty"`
	AppContextUpdateMessage string   `protobuf:"bytes,2,opt,name=app_context_update_message,json=appContextUpdateMessage,proto3" json:"app_context_update_message,omitempty"`
	XXX_NoUnkeyedLiteral    struct{} `json:"-"`
	XXX_unrecognized        []byte   `json:"-"`
	XXX_sizecache           int32    `json:"-"`
}

func (m *ContextUpdateResponse) Reset()         { *m = ContextUpdateResponse{} }
func (m *ContextUpdateResponse) String() string { return proto.CompactTextString(m) }
func (*ContextUpdateResponse) ProtoMessage()    {}
func (*ContextUpdateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_8ebb57f5310873be, []int{1}
}

func (m *ContextUpdateResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ContextUpdateResponse.Unmarshal(m, b)
}
func (m *ContextUpdateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ContextUpdateResponse.Marshal(b, m, deterministic)
}
func (m *ContextUpdateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContextUpdateResponse.Merge(m, src)
}
func (m *ContextUpdateResponse) XXX_Size() int {
	return xxx_messageInfo_ContextUpdateResponse.Size(m)
}
func (m *ContextUpdateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ContextUpdateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ContextUpdateResponse proto.InternalMessageInfo

func (m *ContextUpdateResponse) GetAppContextUpdated() bool {
	if m != nil {
		return m.AppContextUpdated
	}
	return false
}

func (m *ContextUpdateResponse) GetAppContextUpdateMessage() string {
	if m != nil {
		return m.AppContextUpdateMessage
	}
	return ""
}

func init() {
	proto.RegisterType((*ContextUpdateRequest)(nil), "ContextUpdateRequest")
	proto.RegisterType((*ContextUpdateResponse)(nil), "ContextUpdateResponse")
}

func init() {
	proto.RegisterFile("contextupdate.proto", fileDescriptor_8ebb57f5310873be)
}

var fileDescriptor_8ebb57f5310873be = []byte{
	// 176 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4e, 0xce, 0xcf, 0x2b,
	0x49, 0xad, 0x28, 0x29, 0x2d, 0x48, 0x49, 0x2c, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x57,
	0x32, 0xe7, 0x12, 0x71, 0x86, 0x08, 0x87, 0x82, 0x85, 0x83, 0x52, 0x0b, 0x4b, 0x53, 0x8b, 0x4b,
	0x84, 0xe4, 0xb9, 0xb8, 0x13, 0x0b, 0x0a, 0xe2, 0xa1, 0x5a, 0x24, 0x18, 0x15, 0x18, 0x35, 0x38,
	0x83, 0xb8, 0x12, 0x0b, 0x0a, 0xa0, 0xaa, 0x95, 0x5a, 0x18, 0xb9, 0x44, 0xd1, 0x74, 0x16, 0x17,
	0xe4, 0xe7, 0x15, 0xa7, 0x0a, 0xe9, 0x71, 0x09, 0x23, 0x69, 0x8d, 0x87, 0x58, 0x97, 0x02, 0x36,
	0x82, 0x23, 0x48, 0x10, 0x61, 0x04, 0x44, 0x5b, 0x8a, 0x90, 0x35, 0x97, 0x14, 0xa6, 0xfa, 0xf8,
	0xdc, 0xd4, 0xe2, 0xe2, 0xc4, 0xf4, 0x54, 0x09, 0x26, 0xb0, 0xcd, 0xe2, 0xe8, 0xda, 0x7c, 0x21,
	0xd2, 0x46, 0x21, 0x5c, 0xbc, 0x28, 0xde, 0x12, 0x72, 0xe6, 0x12, 0x80, 0xa8, 0x70, 0x84, 0xeb,
	0x10, 0x12, 0xd5, 0xc3, 0xe6, 0x47, 0x29, 0x31, 0x3d, 0xac, 0x1e, 0x50, 0x62, 0x48, 0x62, 0x03,
	0x07, 0x8e, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0x75, 0x78, 0xc0, 0x4d, 0x33, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ContextupdateClient is the client API for Contextupdate service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ContextupdateClient interface {
	// Controllers
	UpdateAppContext(ctx context.Context, in *ContextUpdateRequest, opts ...grpc.CallOption) (*ContextUpdateResponse, error)
}

type contextupdateClient struct {
	cc grpc.ClientConnInterface
}

func NewContextupdateClient(cc grpc.ClientConnInterface) ContextupdateClient {
	return &contextupdateClient{cc}
}

func (c *contextupdateClient) UpdateAppContext(ctx context.Context, in *ContextUpdateRequest, opts ...grpc.CallOption) (*ContextUpdateResponse, error) {
	out := new(ContextUpdateResponse)
	err := c.cc.Invoke(ctx, "/contextupdate/UpdateAppContext", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ContextupdateServer is the server API for Contextupdate service.
type ContextupdateServer interface {
	// Controllers
	UpdateAppContext(context.Context, *ContextUpdateRequest) (*ContextUpdateResponse, error)
}

// UnimplementedContextupdateServer can be embedded to have forward compatible implementations.
type UnimplementedContextupdateServer struct {
}

func (*UnimplementedContextupdateServer) UpdateAppContext(ctx context.Context, req *ContextUpdateRequest) (*ContextUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAppContext not implemented")
}

func RegisterContextupdateServer(s *grpc.Server, srv ContextupdateServer) {
	s.RegisterService(&_Contextupdate_serviceDesc, srv)
}

func _Contextupdate_UpdateAppContext_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContextUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContextupdateServer).UpdateAppContext(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/contextupdate/UpdateAppContext",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContextupdateServer).UpdateAppContext(ctx, req.(*ContextUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Contextupdate_serviceDesc = grpc.ServiceDesc{
	ServiceName: "contextupdate",
	HandlerType: (*ContextupdateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpdateAppContext",
			Handler:    _Contextupdate_UpdateAppContext_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "contextupdate.proto",
}
