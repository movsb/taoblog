// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protocols/service.proto

package protocols

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
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

type PingRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingRequest) Reset()         { *m = PingRequest{} }
func (m *PingRequest) String() string { return proto.CompactTextString(m) }
func (*PingRequest) ProtoMessage()    {}
func (*PingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_763b5d403f427316, []int{0}
}

func (m *PingRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingRequest.Unmarshal(m, b)
}
func (m *PingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingRequest.Marshal(b, m, deterministic)
}
func (m *PingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingRequest.Merge(m, src)
}
func (m *PingRequest) XXX_Size() int {
	return xxx_messageInfo_PingRequest.Size(m)
}
func (m *PingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PingRequest proto.InternalMessageInfo

type PingResponse struct {
	Pong                 string   `protobuf:"bytes,1,opt,name=pong,proto3" json:"pong,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingResponse) Reset()         { *m = PingResponse{} }
func (m *PingResponse) String() string { return proto.CompactTextString(m) }
func (*PingResponse) ProtoMessage()    {}
func (*PingResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_763b5d403f427316, []int{1}
}

func (m *PingResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingResponse.Unmarshal(m, b)
}
func (m *PingResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingResponse.Marshal(b, m, deterministic)
}
func (m *PingResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingResponse.Merge(m, src)
}
func (m *PingResponse) XXX_Size() int {
	return xxx_messageInfo_PingResponse.Size(m)
}
func (m *PingResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PingResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PingResponse proto.InternalMessageInfo

func (m *PingResponse) GetPong() string {
	if m != nil {
		return m.Pong
	}
	return ""
}

func init() {
	proto.RegisterType((*PingRequest)(nil), "protocols.PingRequest")
	proto.RegisterType((*PingResponse)(nil), "protocols.PingResponse")
}

func init() { proto.RegisterFile("protocols/service.proto", fileDescriptor_763b5d403f427316) }

var fileDescriptor_763b5d403f427316 = []byte{
	// 1135 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x55, 0x5b, 0x6f, 0x1b, 0x45,
	0x14, 0x66, 0x51, 0x95, 0x92, 0x49, 0xdc, 0xa4, 0xd3, 0xdc, 0xd8, 0xe6, 0x62, 0x6d, 0xab, 0x50,
	0x85, 0xd4, 0xae, 0x5c, 0x78, 0xa9, 0x04, 0x95, 0x93, 0x42, 0x15, 0x09, 0xa4, 0xaa, 0xa1, 0x2f,
	0xe5, 0xa1, 0x5a, 0xbb, 0xd3, 0xad, 0x85, 0xe3, 0x71, 0xbc, 0xeb, 0x20, 0x2e, 0x45, 0x26, 0x4a,
	0x64, 0xbb, 0xb4, 0xc4, 0x72, 0x14, 0x43, 0x93, 0xd0, 0x24, 0x08, 0xa5, 0x48, 0x34, 0x17, 0x41,
	0x49, 0x23, 0xc7, 0xc0, 0x33, 0xcf, 0xfc, 0x84, 0x64, 0x76, 0x5d, 0x09, 0x29, 0x7f, 0x01, 0xed,
	0x5c, 0x9c, 0x5d, 0xdb, 0x69, 0x78, 0x9b, 0x3d, 0xe7, 0xe8, 0x7c, 0xdf, 0x77, 0x6e, 0x0b, 0x3a,
	0xe3, 0x09, 0x6c, 0xe0, 0x30, 0x8e, 0xea, 0x7e, 0x1d, 0x25, 0x26, 0x22, 0x61, 0xe4, 0xa3, 0x16,
	0xd8, 0x58, 0x71, 0xc8, 0x1d, 0x07, 0x31, 0x21, 0x35, 0xfc, 0x49, 0x32, 0xce, 0x42, 0xe4, 0xb6,
	0x03, 0x7b, 0x1c, 0xeb, 0x06, 0xb7, 0x3a, 0x32, 0x86, 0xf1, 0xd8, 0x18, 0x8a, 0x09, 0x47, 0x87,
	0x13, 0x4a, 0x4d, 0x84, 0xef, 0x72, 0xfb, 0x69, 0x0d, 0x63, 0x2d, 0x8a, 0xfc, 0xf4, 0x2b, 0x94,
	0xbc, 0xe3, 0x47, 0x63, 0x71, 0xe3, 0x33, 0xee, 0xec, 0xe6, 0x4e, 0x35, 0x1e, 0xf1, 0xab, 0xb1,
	0x18, 0x36, 0x54, 0x23, 0x82, 0x63, 0x3a, 0xf7, 0x0e, 0xb2, 0x94, 0xe7, 0x35, 0x14, 0x3b, 0xaf,
	0x7f, 0xaa, 0x6a, 0x1a, 0x4a, 0xf8, 0x71, 0x9c, 0x46, 0xd4, 0x46, 0x2b, 0x1e, 0xd0, 0x74, 0x2d,
	0x12, 0xd3, 0xae, 0xa3, 0xf1, 0x24, 0xd2, 0x0d, 0x45, 0x01, 0xcd, 0xec, 0x53, 0x8f, 0xe3, 0x98,
	0x8e, 0x20, 0x04, 0xc7, 0xe2, 0x38, 0xa6, 0x75, 0x49, 0x5e, 0xe9, 0x5c, 0xe3, 0x75, 0xfa, 0x0e,
	0xfc, 0x2b, 0x01, 0xf0, 0xa1, 0x1a, 0x53, 0x35, 0x64, 0x0b, 0x81, 0x13, 0xa0, 0x61, 0x88, 0x56,
	0x00, 0x76, 0xf9, 0x2a, 0x6a, 0x7c, 0xcc, 0xc4, 0xd3, 0xca, 0xaf, 0xd7, 0xf1, 0x30, 0x04, 0xe5,
	0xed, 0x7c, 0xd0, 0x2b, 0xf7, 0x92, 0x8d, 0x12, 0x49, 0x17, 0xcd, 0xb9, 0x4d, 0xf3, 0xe1, 0x3a,
	0x29, 0xce, 0xee, 0x97, 0x32, 0xe6, 0x37, 0x05, 0xfa, 0xc8, 0xee, 0xa6, 0x26, 0x27, 0xff, 0xd9,
	0x9b, 0x7d, 0xb5, 0x19, 0x02, 0xff, 0xc4, 0x45, 0x5e, 0xed, 0x0b, 0x12, 0xbc, 0x03, 0x9a, 0x58,
	0xaa, 0xf7, 0x23, 0x51, 0xa4, 0xc3, 0x9e, 0x1a, 0x08, 0x6a, 0x17, 0x0c, 0x7a, 0x0f, 0x73, 0x73,
	0x1a, 0x6d, 0xf9, 0xe0, 0x49, 0xb9, 0x85, 0xd3, 0x28, 0xa4, 0xf7, 0x76, 0xb6, 0x76, 0x53, 0x93,
	0xe7, 0xa4, 0x0b, 0x52, 0x60, 0x4a, 0x02, 0x0d, 0xa3, 0xb4, 0x37, 0xf0, 0x73, 0xd0, 0xc4, 0x5e,
	0xd7, 0xb0, 0x6e, 0xb8, 0x21, 0x1d, 0xf6, 0x7a, 0x90, 0x2e, 0x37, 0x87, 0x1c, 0xa4, 0x90, 0x66,
	0x21, 0x6d, 0xfd, 0xb2, 0x6c, 0xe6, 0x16, 0xad, 0x67, 0x3f, 0x0a, 0xa9, 0x10, 0xb6, 0xda, 0x52,
	0xd9, 0x44, 0xd0, 0x39, 0xd2, 0x03, 0x73, 0x2d, 0xe0, 0xf8, 0x47, 0x2a, 0x1e, 0x8a, 0x62, 0x0d,
	0xde, 0x02, 0xc7, 0xec, 0x2e, 0xc1, 0x0e, 0x07, 0x82, 0xa3, 0x8b, 0x72, 0x67, 0x8d, 0x9d, 0x43,
	0xf6, 0xe7, 0x83, 0xa7, 0xe4, 0x93, 0xb6, 0xc9, 0x6b, 0x2e, 0x3e, 0x24, 0xf7, 0x57, 0xc8, 0xa3,
	0x9f, 0x05, 0x28, 0x80, 0xaf, 0xd9, 0xa0, 0x71, 0x3b, 0xf1, 0xc7, 0x00, 0x0c, 0x27, 0x90, 0x6a,
	0x20, 0x9b, 0x31, 0x6c, 0x71, 0xa6, 0xc3, 0xba, 0x21, 0x57, 0x1b, 0x94, 0x37, 0xf3, 0xc1, 0x36,
	0x19, 0x92, 0xcc, 0x02, 0xd9, 0x29, 0x9a, 0x85, 0x4d, 0xa6, 0x49, 0x24, 0x3e, 0xa1, 0x34, 0xd2,
	0xc4, 0xb6, 0x8c, 0x4b, 0xd2, 0x00, 0x8c, 0x80, 0xe3, 0x57, 0x91, 0x41, 0x33, 0x3b, 0xe7, 0x82,
	0xdb, 0x84, 0x86, 0x1a, 0x8c, 0x40, 0x3e, 0xd8, 0x29, 0xb7, 0x97, 0xbf, 0x7d, 0x4e, 0x66, 0x0a,
	0xe6, 0x52, 0xce, 0xda, 0x48, 0xbb, 0x61, 0x78, 0xd1, 0x28, 0x8c, 0xff, 0x8b, 0xc8, 0xed, 0x77,
	0x06, 0xee, 0xc1, 0x71, 0x00, 0x6e, 0xc4, 0x6f, 0x0b, 0x1d, 0xdd, 0x8e, 0x94, 0x07, 0xe6, 0x43,
	0x01, 0xdf, 0x62, 0xfd, 0x59, 0x78, 0x56, 0xa3, 0x48, 0x0e, 0xb4, 0x3b, 0xa0, 0xe8, 0x9e, 0x53,
	0x3c, 0x5b, 0xdd, 0x63, 0x09, 0x80, 0x2b, 0x28, 0x8a, 0xea, 0x60, 0x1e, 0x98, 0x05, 0x66, 0x87,
	0x8f, 0xad, 0xb2, 0x4f, 0xec, 0xb9, 0xef, 0x3d, 0x7b, 0xcf, 0x95, 0x9b, 0xf9, 0x60, 0x50, 0xbe,
	0x4c, 0x32, 0xcb, 0x2f, 0x1e, 0xfd, 0xc4, 0xa0, 0xc9, 0xcc, 0x7d, 0x32, 0xbd, 0x65, 0x66, 0x53,
	0xe6, 0x62, 0xd6, 0x5a, 0xd8, 0x26, 0xd3, 0xbf, 0x95, 0x7f, 0x9f, 0x32, 0x8b, 0xb9, 0xfd, 0x52,
	0xa6, 0xbc, 0x31, 0x55, 0x5e, 0x2f, 0xee, 0xa6, 0xbe, 0x36, 0x97, 0xd3, 0xd6, 0xda, 0x9f, 0xd6,
	0x5a, 0xd6, 0xb1, 0x35, 0xad, 0x03, 0x27, 0x5c, 0x55, 0xb9, 0x07, 0x67, 0x25, 0xe0, 0x19, 0x65,
	0xb5, 0x1e, 0x35, 0x54, 0x23, 0xa9, 0xc3, 0x3e, 0xd7, 0xa0, 0x3a, 0x3c, 0x82, 0xa6, 0xf7, 0xf0,
	0x00, 0x3e, 0x58, 0x97, 0x69, 0x73, 0xcc, 0xef, 0x37, 0x48, 0x6e, 0x95, 0x4c, 0x3f, 0x25, 0xa5,
	0x94, 0xbb, 0x62, 0x7d, 0x8a, 0x5c, 0xdd, 0x9c, 0x4b, 0x3a, 0xe2, 0x59, 0xec, 0xb2, 0x8d, 0x03,
	0x0f, 0x1f, 0x80, 0x51, 0x9c, 0x4c, 0x84, 0x91, 0x8b, 0x94, 0xcb, 0x53, 0x8f, 0x54, 0x55, 0x00,
	0x27, 0xd5, 0x43, 0xa7, 0x92, 0x4f, 0x0c, 0x5b, 0xb3, 0x62, 0x4e, 0x30, 0x7a, 0x05, 0x7e, 0x09,
	0x3c, 0x6c, 0xc8, 0x87, 0xd9, 0x49, 0x86, 0xd0, 0x91, 0x91, 0xdb, 0xe4, 0x3a, 0x36, 0x65, 0x98,
	0x8a, 0x65, 0xd3, 0xbe, 0xb7, 0x9d, 0x32, 0x7f, 0x58, 0x11, 0x2d, 0x60, 0xa9, 0xcf, 0x28, 0xbd,
	0x55, 0xe3, 0x71, 0x8b, 0x2a, 0x16, 0xb7, 0x9f, 0x0a, 0xfe, 0x0a, 0x80, 0xab, 0xc8, 0x10, 0xd0,
	0xdd, 0x6e, 0x31, 0xdc, 0x2c, 0xa4, 0xd6, 0x23, 0xf1, 0x2e, 0xbd, 0x9b, 0x5c, 0xdc, 0x83, 0x34,
	0x59, 0x9f, 0xb7, 0x4a, 0x05, 0x32, 0xf3, 0xdc, 0x9a, 0x9f, 0x72, 0xb3, 0x69, 0x87, 0xa7, 0x6c,
	0x36, 0x02, 0x5b, 0xac, 0xc6, 0x8c, 0x04, 0x3c, 0x6c, 0x09, 0x04, 0x87, 0xbe, 0x9a, 0xf5, 0xf8,
	0x1f, 0x34, 0x46, 0x28, 0x0d, 0xbe, 0x24, 0x2f, 0xa3, 0xe1, 0x0d, 0x9c, 0x76, 0xd3, 0x10, 0x3f,
	0xc2, 0xca, 0xe6, 0x64, 0x25, 0xe0, 0x61, 0x2b, 0x52, 0x8f, 0x91, 0xcb, 0x53, 0x6f, 0x06, 0xaa,
	0x02, 0x0e, 0x7e, 0x2f, 0xb4, 0x57, 0x74, 0x93, 0x96, 0x72, 0x35, 0xbd, 0x6a, 0x1f, 0xa8, 0x5b,
	0x9d, 0x55, 0x09, 0x34, 0x7f, 0x10, 0xd1, 0x45, 0x23, 0x74, 0xe8, 0x3c, 0xe6, 0x4e, 0x87, 0x60,
	0xd2, 0x77, 0xa8, 0x9f, 0x13, 0xb9, 0x91, 0x0f, 0x0e, 0xca, 0x03, 0xac, 0x5f, 0xf6, 0xef, 0xcd,
	0x71, 0xc1, 0xac, 0xf9, 0xa9, 0xfd, 0x52, 0x96, 0xf1, 0x22, 0x99, 0xef, 0xca, 0x2b, 0x95, 0x9b,
	0xec, 0x85, 0x47, 0x4c, 0x12, 0xfc, 0x5b, 0x02, 0xad, 0xa3, 0x95, 0x81, 0xb1, 0x97, 0x60, 0xe4,
	0x0a, 0x54, 0xdc, 0xfb, 0xea, 0x72, 0x0a, 0xc2, 0x67, 0x5e, 0x1a, 0xc3, 0x49, 0x47, 0xe9, 0x1d,
	0x2a, 0xff, 0xf1, 0xd4, 0x7a, 0xb2, 0x63, 0x2e, 0xe5, 0x5e, 0xac, 0x6c, 0x59, 0xc5, 0x27, 0x8c,
	0xa8, 0x7d, 0x78, 0xfe, 0x7a, 0x4c, 0x72, 0x0f, 0xc8, 0x5a, 0x4e, 0x58, 0xb2, 0x24, 0xb3, 0x49,
	0x66, 0x56, 0xf7, 0xb6, 0x53, 0x35, 0xd7, 0x59, 0x51, 0x7a, 0xea, 0xd4, 0xd9, 0xbe, 0x01, 0x0c,
	0xd2, 0x1e, 0x80, 0x5f, 0x25, 0xd0, 0xc6, 0x37, 0x59, 0x94, 0x6f, 0x18, 0x27, 0x63, 0x06, 0xec,
	0xaf, 0x5d, 0x75, 0x57, 0x80, 0xd0, 0xf4, 0xc6, 0x91, 0x71, 0x5c, 0xd7, 0x48, 0x3e, 0x28, 0xcb,
	0x5d, 0xce, 0xcb, 0xc0, 0x24, 0x98, 0x73, 0x9b, 0x82, 0x70, 0x3f, 0x3c, 0x7b, 0xc4, 0x12, 0x87,
	0xed, 0x94, 0x43, 0x67, 0x6f, 0x2a, 0x5a, 0xc4, 0xb8, 0x9b, 0x0c, 0xf9, 0xc2, 0x78, 0xcc, 0x3f,
	0x86, 0x27, 0xf4, 0x90, 0xdf, 0x50, 0x71, 0x28, 0x8a, 0x35, 0x7f, 0x85, 0x4d, 0xa8, 0x81, 0x3e,
	0x2f, 0xfe, 0x17, 0x00, 0x00, 0xff, 0xff, 0x1a, 0x9a, 0xcf, 0xee, 0x5b, 0x0a, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ManagementClient is the client API for Management service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ManagementClient interface {
	Backup(ctx context.Context, in *BackupRequest, opts ...grpc.CallOption) (Management_BackupClient, error)
	BackupFiles(ctx context.Context, opts ...grpc.CallOption) (Management_BackupFilesClient, error)
}

type managementClient struct {
	cc *grpc.ClientConn
}

func NewManagementClient(cc *grpc.ClientConn) ManagementClient {
	return &managementClient{cc}
}

func (c *managementClient) Backup(ctx context.Context, in *BackupRequest, opts ...grpc.CallOption) (Management_BackupClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Management_serviceDesc.Streams[0], "/protocols.Management/Backup", opts...)
	if err != nil {
		return nil, err
	}
	x := &managementBackupClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Management_BackupClient interface {
	Recv() (*BackupResponse, error)
	grpc.ClientStream
}

type managementBackupClient struct {
	grpc.ClientStream
}

func (x *managementBackupClient) Recv() (*BackupResponse, error) {
	m := new(BackupResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *managementClient) BackupFiles(ctx context.Context, opts ...grpc.CallOption) (Management_BackupFilesClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Management_serviceDesc.Streams[1], "/protocols.Management/BackupFiles", opts...)
	if err != nil {
		return nil, err
	}
	x := &managementBackupFilesClient{stream}
	return x, nil
}

type Management_BackupFilesClient interface {
	Send(*BackupFilesRequest) error
	Recv() (*BackupFilesResponse, error)
	grpc.ClientStream
}

type managementBackupFilesClient struct {
	grpc.ClientStream
}

func (x *managementBackupFilesClient) Send(m *BackupFilesRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *managementBackupFilesClient) Recv() (*BackupFilesResponse, error) {
	m := new(BackupFilesResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ManagementServer is the server API for Management service.
type ManagementServer interface {
	Backup(*BackupRequest, Management_BackupServer) error
	BackupFiles(Management_BackupFilesServer) error
}

// UnimplementedManagementServer can be embedded to have forward compatible implementations.
type UnimplementedManagementServer struct {
}

func (*UnimplementedManagementServer) Backup(req *BackupRequest, srv Management_BackupServer) error {
	return status.Errorf(codes.Unimplemented, "method Backup not implemented")
}
func (*UnimplementedManagementServer) BackupFiles(srv Management_BackupFilesServer) error {
	return status.Errorf(codes.Unimplemented, "method BackupFiles not implemented")
}

func RegisterManagementServer(s *grpc.Server, srv ManagementServer) {
	s.RegisterService(&_Management_serviceDesc, srv)
}

func _Management_Backup_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BackupRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ManagementServer).Backup(m, &managementBackupServer{stream})
}

type Management_BackupServer interface {
	Send(*BackupResponse) error
	grpc.ServerStream
}

type managementBackupServer struct {
	grpc.ServerStream
}

func (x *managementBackupServer) Send(m *BackupResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Management_BackupFiles_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ManagementServer).BackupFiles(&managementBackupFilesServer{stream})
}

type Management_BackupFilesServer interface {
	Send(*BackupFilesResponse) error
	Recv() (*BackupFilesRequest, error)
	grpc.ServerStream
}

type managementBackupFilesServer struct {
	grpc.ServerStream
}

func (x *managementBackupFilesServer) Send(m *BackupFilesResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *managementBackupFilesServer) Recv() (*BackupFilesRequest, error) {
	m := new(BackupFilesRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Management_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protocols.Management",
	HandlerType: (*ManagementServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Backup",
			Handler:       _Management_Backup_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "BackupFiles",
			Handler:       _Management_BackupFiles_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protocols/service.proto",
}

// SearchClient is the client API for Search service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SearchClient interface {
	SearchPosts(ctx context.Context, in *SearchPostsRequest, opts ...grpc.CallOption) (*SearchPostsResponse, error)
}

type searchClient struct {
	cc *grpc.ClientConn
}

func NewSearchClient(cc *grpc.ClientConn) SearchClient {
	return &searchClient{cc}
}

func (c *searchClient) SearchPosts(ctx context.Context, in *SearchPostsRequest, opts ...grpc.CallOption) (*SearchPostsResponse, error) {
	out := new(SearchPostsResponse)
	err := c.cc.Invoke(ctx, "/protocols.Search/SearchPosts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SearchServer is the server API for Search service.
type SearchServer interface {
	SearchPosts(context.Context, *SearchPostsRequest) (*SearchPostsResponse, error)
}

// UnimplementedSearchServer can be embedded to have forward compatible implementations.
type UnimplementedSearchServer struct {
}

func (*UnimplementedSearchServer) SearchPosts(ctx context.Context, req *SearchPostsRequest) (*SearchPostsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchPosts not implemented")
}

func RegisterSearchServer(s *grpc.Server, srv SearchServer) {
	s.RegisterService(&_Search_serviceDesc, srv)
}

func _Search_SearchPosts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchPostsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchServer).SearchPosts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.Search/SearchPosts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SearchServer).SearchPosts(ctx, req.(*SearchPostsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Search_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protocols.Search",
	HandlerType: (*SearchServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SearchPosts",
			Handler:    _Search_SearchPosts_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protocols/service.proto",
}

// TaoBlogClient is the client API for TaoBlog service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TaoBlogClient interface {
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	CreatePost(ctx context.Context, in *Post, opts ...grpc.CallOption) (*Post, error)
	GetPost(ctx context.Context, in *GetPostRequest, opts ...grpc.CallOption) (*Post, error)
	UpdatePost(ctx context.Context, in *UpdatePostRequest, opts ...grpc.CallOption) (*Post, error)
	DeletePost(ctx context.Context, in *DeletePostRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SetPostStatus(ctx context.Context, in *SetPostStatusRequest, opts ...grpc.CallOption) (*SetPostStatusResponse, error)
	GetPostSource(ctx context.Context, in *GetPostSourceRequest, opts ...grpc.CallOption) (*GetPostSourceResponse, error)
	CreateComment(ctx context.Context, in *Comment, opts ...grpc.CallOption) (*Comment, error)
	GetComment(ctx context.Context, in *GetCommentRequest, opts ...grpc.CallOption) (*Comment, error)
	UpdateComment(ctx context.Context, in *UpdateCommentRequest, opts ...grpc.CallOption) (*Comment, error)
	DeleteComment(ctx context.Context, in *DeleteCommentRequest, opts ...grpc.CallOption) (*DeleteCommentResponse, error)
	ListComments(ctx context.Context, in *ListCommentsRequest, opts ...grpc.CallOption) (*ListCommentsResponse, error)
	SetCommentPostID(ctx context.Context, in *SetCommentPostIDRequest, opts ...grpc.CallOption) (*SetCommentPostIDResponse, error)
	GetPostCommentsCount(ctx context.Context, in *GetPostCommentsCountRequest, opts ...grpc.CallOption) (*GetPostCommentsCountResponse, error)
}

type taoBlogClient struct {
	cc *grpc.ClientConn
}

func NewTaoBlogClient(cc *grpc.ClientConn) TaoBlogClient {
	return &taoBlogClient{cc}
}

func (c *taoBlogClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) CreatePost(ctx context.Context, in *Post, opts ...grpc.CallOption) (*Post, error) {
	out := new(Post)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/CreatePost", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) GetPost(ctx context.Context, in *GetPostRequest, opts ...grpc.CallOption) (*Post, error) {
	out := new(Post)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/GetPost", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) UpdatePost(ctx context.Context, in *UpdatePostRequest, opts ...grpc.CallOption) (*Post, error) {
	out := new(Post)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/UpdatePost", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) DeletePost(ctx context.Context, in *DeletePostRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/DeletePost", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) SetPostStatus(ctx context.Context, in *SetPostStatusRequest, opts ...grpc.CallOption) (*SetPostStatusResponse, error) {
	out := new(SetPostStatusResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/SetPostStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) GetPostSource(ctx context.Context, in *GetPostSourceRequest, opts ...grpc.CallOption) (*GetPostSourceResponse, error) {
	out := new(GetPostSourceResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/GetPostSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) CreateComment(ctx context.Context, in *Comment, opts ...grpc.CallOption) (*Comment, error) {
	out := new(Comment)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/CreateComment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) GetComment(ctx context.Context, in *GetCommentRequest, opts ...grpc.CallOption) (*Comment, error) {
	out := new(Comment)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/GetComment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) UpdateComment(ctx context.Context, in *UpdateCommentRequest, opts ...grpc.CallOption) (*Comment, error) {
	out := new(Comment)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/UpdateComment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) DeleteComment(ctx context.Context, in *DeleteCommentRequest, opts ...grpc.CallOption) (*DeleteCommentResponse, error) {
	out := new(DeleteCommentResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/DeleteComment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) ListComments(ctx context.Context, in *ListCommentsRequest, opts ...grpc.CallOption) (*ListCommentsResponse, error) {
	out := new(ListCommentsResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/ListComments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) SetCommentPostID(ctx context.Context, in *SetCommentPostIDRequest, opts ...grpc.CallOption) (*SetCommentPostIDResponse, error) {
	out := new(SetCommentPostIDResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/SetCommentPostID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) GetPostCommentsCount(ctx context.Context, in *GetPostCommentsCountRequest, opts ...grpc.CallOption) (*GetPostCommentsCountResponse, error) {
	out := new(GetPostCommentsCountResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/GetPostCommentsCount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TaoBlogServer is the server API for TaoBlog service.
type TaoBlogServer interface {
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	CreatePost(context.Context, *Post) (*Post, error)
	GetPost(context.Context, *GetPostRequest) (*Post, error)
	UpdatePost(context.Context, *UpdatePostRequest) (*Post, error)
	DeletePost(context.Context, *DeletePostRequest) (*emptypb.Empty, error)
	SetPostStatus(context.Context, *SetPostStatusRequest) (*SetPostStatusResponse, error)
	GetPostSource(context.Context, *GetPostSourceRequest) (*GetPostSourceResponse, error)
	CreateComment(context.Context, *Comment) (*Comment, error)
	GetComment(context.Context, *GetCommentRequest) (*Comment, error)
	UpdateComment(context.Context, *UpdateCommentRequest) (*Comment, error)
	DeleteComment(context.Context, *DeleteCommentRequest) (*DeleteCommentResponse, error)
	ListComments(context.Context, *ListCommentsRequest) (*ListCommentsResponse, error)
	SetCommentPostID(context.Context, *SetCommentPostIDRequest) (*SetCommentPostIDResponse, error)
	GetPostCommentsCount(context.Context, *GetPostCommentsCountRequest) (*GetPostCommentsCountResponse, error)
}

// UnimplementedTaoBlogServer can be embedded to have forward compatible implementations.
type UnimplementedTaoBlogServer struct {
}

func (*UnimplementedTaoBlogServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedTaoBlogServer) CreatePost(ctx context.Context, req *Post) (*Post, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePost not implemented")
}
func (*UnimplementedTaoBlogServer) GetPost(ctx context.Context, req *GetPostRequest) (*Post, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPost not implemented")
}
func (*UnimplementedTaoBlogServer) UpdatePost(ctx context.Context, req *UpdatePostRequest) (*Post, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePost not implemented")
}
func (*UnimplementedTaoBlogServer) DeletePost(ctx context.Context, req *DeletePostRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeletePost not implemented")
}
func (*UnimplementedTaoBlogServer) SetPostStatus(ctx context.Context, req *SetPostStatusRequest) (*SetPostStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetPostStatus not implemented")
}
func (*UnimplementedTaoBlogServer) GetPostSource(ctx context.Context, req *GetPostSourceRequest) (*GetPostSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPostSource not implemented")
}
func (*UnimplementedTaoBlogServer) CreateComment(ctx context.Context, req *Comment) (*Comment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateComment not implemented")
}
func (*UnimplementedTaoBlogServer) GetComment(ctx context.Context, req *GetCommentRequest) (*Comment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetComment not implemented")
}
func (*UnimplementedTaoBlogServer) UpdateComment(ctx context.Context, req *UpdateCommentRequest) (*Comment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateComment not implemented")
}
func (*UnimplementedTaoBlogServer) DeleteComment(ctx context.Context, req *DeleteCommentRequest) (*DeleteCommentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteComment not implemented")
}
func (*UnimplementedTaoBlogServer) ListComments(ctx context.Context, req *ListCommentsRequest) (*ListCommentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListComments not implemented")
}
func (*UnimplementedTaoBlogServer) SetCommentPostID(ctx context.Context, req *SetCommentPostIDRequest) (*SetCommentPostIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetCommentPostID not implemented")
}
func (*UnimplementedTaoBlogServer) GetPostCommentsCount(ctx context.Context, req *GetPostCommentsCountRequest) (*GetPostCommentsCountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPostCommentsCount not implemented")
}

func RegisterTaoBlogServer(s *grpc.Server, srv TaoBlogServer) {
	s.RegisterService(&_TaoBlog_serviceDesc, srv)
}

func _TaoBlog_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_CreatePost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Post)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).CreatePost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/CreatePost",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).CreatePost(ctx, req.(*Post))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_GetPost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPostRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).GetPost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/GetPost",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).GetPost(ctx, req.(*GetPostRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_UpdatePost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdatePostRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).UpdatePost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/UpdatePost",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).UpdatePost(ctx, req.(*UpdatePostRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_DeletePost_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeletePostRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).DeletePost(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/DeletePost",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).DeletePost(ctx, req.(*DeletePostRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_SetPostStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetPostStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).SetPostStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/SetPostStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).SetPostStatus(ctx, req.(*SetPostStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_GetPostSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPostSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).GetPostSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/GetPostSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).GetPostSource(ctx, req.(*GetPostSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_CreateComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Comment)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).CreateComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/CreateComment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).CreateComment(ctx, req.(*Comment))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_GetComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).GetComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/GetComment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).GetComment(ctx, req.(*GetCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_UpdateComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).UpdateComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/UpdateComment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).UpdateComment(ctx, req.(*UpdateCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_DeleteComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).DeleteComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/DeleteComment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).DeleteComment(ctx, req.(*DeleteCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_ListComments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListCommentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).ListComments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/ListComments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).ListComments(ctx, req.(*ListCommentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_SetCommentPostID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetCommentPostIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).SetCommentPostID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/SetCommentPostID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).SetCommentPostID(ctx, req.(*SetCommentPostIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_GetPostCommentsCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPostCommentsCountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).GetPostCommentsCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/GetPostCommentsCount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).GetPostCommentsCount(ctx, req.(*GetPostCommentsCountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TaoBlog_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protocols.TaoBlog",
	HandlerType: (*TaoBlogServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _TaoBlog_Ping_Handler,
		},
		{
			MethodName: "CreatePost",
			Handler:    _TaoBlog_CreatePost_Handler,
		},
		{
			MethodName: "GetPost",
			Handler:    _TaoBlog_GetPost_Handler,
		},
		{
			MethodName: "UpdatePost",
			Handler:    _TaoBlog_UpdatePost_Handler,
		},
		{
			MethodName: "DeletePost",
			Handler:    _TaoBlog_DeletePost_Handler,
		},
		{
			MethodName: "SetPostStatus",
			Handler:    _TaoBlog_SetPostStatus_Handler,
		},
		{
			MethodName: "GetPostSource",
			Handler:    _TaoBlog_GetPostSource_Handler,
		},
		{
			MethodName: "CreateComment",
			Handler:    _TaoBlog_CreateComment_Handler,
		},
		{
			MethodName: "GetComment",
			Handler:    _TaoBlog_GetComment_Handler,
		},
		{
			MethodName: "UpdateComment",
			Handler:    _TaoBlog_UpdateComment_Handler,
		},
		{
			MethodName: "DeleteComment",
			Handler:    _TaoBlog_DeleteComment_Handler,
		},
		{
			MethodName: "ListComments",
			Handler:    _TaoBlog_ListComments_Handler,
		},
		{
			MethodName: "SetCommentPostID",
			Handler:    _TaoBlog_SetCommentPostID_Handler,
		},
		{
			MethodName: "GetPostCommentsCount",
			Handler:    _TaoBlog_GetPostCommentsCount_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protocols/service.proto",
}
