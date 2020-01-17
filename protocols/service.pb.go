// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protocols/service.proto

package protocols

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

func init() { proto.RegisterFile("protocols/service.proto", fileDescriptor_763b5d403f427316) }

var fileDescriptor_763b5d403f427316 = []byte{
	// 140 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2f, 0x28, 0xca, 0x2f,
	0xc9, 0x4f, 0xce, 0xcf, 0x29, 0xd6, 0x2f, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0xd5, 0x03, 0x8b,
	0x08, 0x71, 0xc2, 0x25, 0xa4, 0xc4, 0x10, 0x6a, 0x92, 0x12, 0x93, 0xb3, 0x4b, 0x0b, 0x20, 0x4a,
	0x8c, 0x3c, 0xb8, 0xd8, 0x43, 0x12, 0xf3, 0x9d, 0x72, 0xf2, 0xd3, 0x85, 0x6c, 0xb9, 0xd8, 0x9c,
	0xc0, 0x52, 0x42, 0x12, 0x7a, 0x70, 0xd5, 0x7a, 0x10, 0xa1, 0xa0, 0xd4, 0xc2, 0xd2, 0xd4, 0xe2,
	0x12, 0x29, 0x49, 0x2c, 0x32, 0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9, 0x4e, 0x2a, 0x51, 0x4a, 0xe9,
	0x99, 0x25, 0x19, 0xa5, 0x49, 0x7a, 0xc9, 0xf9, 0xb9, 0xfa, 0xb9, 0xf9, 0x65, 0xc5, 0x49, 0xfa,
	0x25, 0x89, 0xf9, 0x49, 0x39, 0xf9, 0xe9, 0xfa, 0x70, 0x4d, 0x49, 0x6c, 0x60, 0xa6, 0x31, 0x20,
	0x00, 0x00, 0xff, 0xff, 0x93, 0xc9, 0xf4, 0x83, 0xb4, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TaoBlogClient is the client API for TaoBlog service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TaoBlogClient interface {
	Backup(ctx context.Context, in *BackupRequest, opts ...grpc.CallOption) (*BackupResponse, error)
}

type taoBlogClient struct {
	cc *grpc.ClientConn
}

func NewTaoBlogClient(cc *grpc.ClientConn) TaoBlogClient {
	return &taoBlogClient{cc}
}

func (c *taoBlogClient) Backup(ctx context.Context, in *BackupRequest, opts ...grpc.CallOption) (*BackupResponse, error) {
	out := new(BackupResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/Backup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TaoBlogServer is the server API for TaoBlog service.
type TaoBlogServer interface {
	Backup(context.Context, *BackupRequest) (*BackupResponse, error)
}

// UnimplementedTaoBlogServer can be embedded to have forward compatible implementations.
type UnimplementedTaoBlogServer struct {
}

func (*UnimplementedTaoBlogServer) Backup(ctx context.Context, req *BackupRequest) (*BackupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Backup not implemented")
}

func RegisterTaoBlogServer(s *grpc.Server, srv TaoBlogServer) {
	s.RegisterService(&_TaoBlog_serviceDesc, srv)
}

func _TaoBlog_Backup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BackupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).Backup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/Backup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).Backup(ctx, req.(*BackupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TaoBlog_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protocols.TaoBlog",
	HandlerType: (*TaoBlogServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Backup",
			Handler:    _TaoBlog_Backup_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protocols/service.proto",
}
