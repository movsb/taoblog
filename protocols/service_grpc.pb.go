// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: protocols/service.proto

package protocols

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ManagementClient is the client API for Management service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ManagementClient interface {
	Backup(ctx context.Context, in *BackupRequest, opts ...grpc.CallOption) (Management_BackupClient, error)
	BackupFiles(ctx context.Context, opts ...grpc.CallOption) (Management_BackupFilesClient, error)
	SetRedirect(ctx context.Context, in *SetRedirectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	FileSystem(ctx context.Context, opts ...grpc.CallOption) (Management_FileSystemClient, error)
}

type managementClient struct {
	cc grpc.ClientConnInterface
}

func NewManagementClient(cc grpc.ClientConnInterface) ManagementClient {
	return &managementClient{cc}
}

func (c *managementClient) Backup(ctx context.Context, in *BackupRequest, opts ...grpc.CallOption) (Management_BackupClient, error) {
	stream, err := c.cc.NewStream(ctx, &Management_ServiceDesc.Streams[0], "/protocols.Management/Backup", opts...)
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
	stream, err := c.cc.NewStream(ctx, &Management_ServiceDesc.Streams[1], "/protocols.Management/BackupFiles", opts...)
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

func (c *managementClient) SetRedirect(ctx context.Context, in *SetRedirectRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/protocols.Management/SetRedirect", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *managementClient) FileSystem(ctx context.Context, opts ...grpc.CallOption) (Management_FileSystemClient, error) {
	stream, err := c.cc.NewStream(ctx, &Management_ServiceDesc.Streams[2], "/protocols.Management/FileSystem", opts...)
	if err != nil {
		return nil, err
	}
	x := &managementFileSystemClient{stream}
	return x, nil
}

type Management_FileSystemClient interface {
	Send(*FileSystemRequest) error
	Recv() (*FileSystemResponse, error)
	grpc.ClientStream
}

type managementFileSystemClient struct {
	grpc.ClientStream
}

func (x *managementFileSystemClient) Send(m *FileSystemRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *managementFileSystemClient) Recv() (*FileSystemResponse, error) {
	m := new(FileSystemResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ManagementServer is the server API for Management service.
// All implementations must embed UnimplementedManagementServer
// for forward compatibility
type ManagementServer interface {
	Backup(*BackupRequest, Management_BackupServer) error
	BackupFiles(Management_BackupFilesServer) error
	SetRedirect(context.Context, *SetRedirectRequest) (*emptypb.Empty, error)
	FileSystem(Management_FileSystemServer) error
	mustEmbedUnimplementedManagementServer()
}

// UnimplementedManagementServer must be embedded to have forward compatible implementations.
type UnimplementedManagementServer struct {
}

func (UnimplementedManagementServer) Backup(*BackupRequest, Management_BackupServer) error {
	return status.Errorf(codes.Unimplemented, "method Backup not implemented")
}
func (UnimplementedManagementServer) BackupFiles(Management_BackupFilesServer) error {
	return status.Errorf(codes.Unimplemented, "method BackupFiles not implemented")
}
func (UnimplementedManagementServer) SetRedirect(context.Context, *SetRedirectRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetRedirect not implemented")
}
func (UnimplementedManagementServer) FileSystem(Management_FileSystemServer) error {
	return status.Errorf(codes.Unimplemented, "method FileSystem not implemented")
}
func (UnimplementedManagementServer) mustEmbedUnimplementedManagementServer() {}

// UnsafeManagementServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ManagementServer will
// result in compilation errors.
type UnsafeManagementServer interface {
	mustEmbedUnimplementedManagementServer()
}

func RegisterManagementServer(s grpc.ServiceRegistrar, srv ManagementServer) {
	s.RegisterService(&Management_ServiceDesc, srv)
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

func _Management_SetRedirect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetRedirectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ManagementServer).SetRedirect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.Management/SetRedirect",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ManagementServer).SetRedirect(ctx, req.(*SetRedirectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Management_FileSystem_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ManagementServer).FileSystem(&managementFileSystemServer{stream})
}

type Management_FileSystemServer interface {
	Send(*FileSystemResponse) error
	Recv() (*FileSystemRequest, error)
	grpc.ServerStream
}

type managementFileSystemServer struct {
	grpc.ServerStream
}

func (x *managementFileSystemServer) Send(m *FileSystemResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *managementFileSystemServer) Recv() (*FileSystemRequest, error) {
	m := new(FileSystemRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Management_ServiceDesc is the grpc.ServiceDesc for Management service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Management_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protocols.Management",
	HandlerType: (*ManagementServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetRedirect",
			Handler:    _Management_SetRedirect_Handler,
		},
	},
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
		{
			StreamName:    "FileSystem",
			Handler:       _Management_FileSystem_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protocols/service.proto",
}

// SearchClient is the client API for Search service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SearchClient interface {
	SearchPosts(ctx context.Context, in *SearchPostsRequest, opts ...grpc.CallOption) (*SearchPostsResponse, error)
}

type searchClient struct {
	cc grpc.ClientConnInterface
}

func NewSearchClient(cc grpc.ClientConnInterface) SearchClient {
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
// All implementations must embed UnimplementedSearchServer
// for forward compatibility
type SearchServer interface {
	SearchPosts(context.Context, *SearchPostsRequest) (*SearchPostsResponse, error)
	mustEmbedUnimplementedSearchServer()
}

// UnimplementedSearchServer must be embedded to have forward compatible implementations.
type UnimplementedSearchServer struct {
}

func (UnimplementedSearchServer) SearchPosts(context.Context, *SearchPostsRequest) (*SearchPostsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchPosts not implemented")
}
func (UnimplementedSearchServer) mustEmbedUnimplementedSearchServer() {}

// UnsafeSearchServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SearchServer will
// result in compilation errors.
type UnsafeSearchServer interface {
	mustEmbedUnimplementedSearchServer()
}

func RegisterSearchServer(s grpc.ServiceRegistrar, srv SearchServer) {
	s.RegisterService(&Search_ServiceDesc, srv)
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

// Search_ServiceDesc is the grpc.ServiceDesc for Search service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Search_ServiceDesc = grpc.ServiceDesc{
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
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TaoBlogClient interface {
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	CreatePost(ctx context.Context, in *Post, opts ...grpc.CallOption) (*Post, error)
	GetPost(ctx context.Context, in *GetPostRequest, opts ...grpc.CallOption) (*Post, error)
	UpdatePost(ctx context.Context, in *UpdatePostRequest, opts ...grpc.CallOption) (*Post, error)
	DeletePost(ctx context.Context, in *DeletePostRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SetPostStatus(ctx context.Context, in *SetPostStatusRequest, opts ...grpc.CallOption) (*SetPostStatusResponse, error)
	GetPostComments(ctx context.Context, in *GetPostCommentsRequest, opts ...grpc.CallOption) (*GetPostCommentsResponse, error)
	GetPostsByTags(ctx context.Context, in *GetPostsByTagsRequest, opts ...grpc.CallOption) (*GetPostsByTagsResponse, error)
	CreateComment(ctx context.Context, in *Comment, opts ...grpc.CallOption) (*Comment, error)
	GetComment(ctx context.Context, in *GetCommentRequest, opts ...grpc.CallOption) (*Comment, error)
	UpdateComment(ctx context.Context, in *UpdateCommentRequest, opts ...grpc.CallOption) (*Comment, error)
	DeleteComment(ctx context.Context, in *DeleteCommentRequest, opts ...grpc.CallOption) (*DeleteCommentResponse, error)
	ListComments(ctx context.Context, in *ListCommentsRequest, opts ...grpc.CallOption) (*ListCommentsResponse, error)
	SetCommentPostID(ctx context.Context, in *SetCommentPostIDRequest, opts ...grpc.CallOption) (*SetCommentPostIDResponse, error)
	GetPostCommentsCount(ctx context.Context, in *GetPostCommentsCountRequest, opts ...grpc.CallOption) (*GetPostCommentsCountResponse, error)
	PreviewComment(ctx context.Context, in *PreviewCommentRequest, opts ...grpc.CallOption) (*PreviewCommentResponse, error)
}

type taoBlogClient struct {
	cc grpc.ClientConnInterface
}

func NewTaoBlogClient(cc grpc.ClientConnInterface) TaoBlogClient {
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

func (c *taoBlogClient) GetPostComments(ctx context.Context, in *GetPostCommentsRequest, opts ...grpc.CallOption) (*GetPostCommentsResponse, error) {
	out := new(GetPostCommentsResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/GetPostComments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *taoBlogClient) GetPostsByTags(ctx context.Context, in *GetPostsByTagsRequest, opts ...grpc.CallOption) (*GetPostsByTagsResponse, error) {
	out := new(GetPostsByTagsResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/GetPostsByTags", in, out, opts...)
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

func (c *taoBlogClient) PreviewComment(ctx context.Context, in *PreviewCommentRequest, opts ...grpc.CallOption) (*PreviewCommentResponse, error) {
	out := new(PreviewCommentResponse)
	err := c.cc.Invoke(ctx, "/protocols.TaoBlog/PreviewComment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TaoBlogServer is the server API for TaoBlog service.
// All implementations must embed UnimplementedTaoBlogServer
// for forward compatibility
type TaoBlogServer interface {
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	CreatePost(context.Context, *Post) (*Post, error)
	GetPost(context.Context, *GetPostRequest) (*Post, error)
	UpdatePost(context.Context, *UpdatePostRequest) (*Post, error)
	DeletePost(context.Context, *DeletePostRequest) (*emptypb.Empty, error)
	SetPostStatus(context.Context, *SetPostStatusRequest) (*SetPostStatusResponse, error)
	GetPostComments(context.Context, *GetPostCommentsRequest) (*GetPostCommentsResponse, error)
	GetPostsByTags(context.Context, *GetPostsByTagsRequest) (*GetPostsByTagsResponse, error)
	CreateComment(context.Context, *Comment) (*Comment, error)
	GetComment(context.Context, *GetCommentRequest) (*Comment, error)
	UpdateComment(context.Context, *UpdateCommentRequest) (*Comment, error)
	DeleteComment(context.Context, *DeleteCommentRequest) (*DeleteCommentResponse, error)
	ListComments(context.Context, *ListCommentsRequest) (*ListCommentsResponse, error)
	SetCommentPostID(context.Context, *SetCommentPostIDRequest) (*SetCommentPostIDResponse, error)
	GetPostCommentsCount(context.Context, *GetPostCommentsCountRequest) (*GetPostCommentsCountResponse, error)
	PreviewComment(context.Context, *PreviewCommentRequest) (*PreviewCommentResponse, error)
	mustEmbedUnimplementedTaoBlogServer()
}

// UnimplementedTaoBlogServer must be embedded to have forward compatible implementations.
type UnimplementedTaoBlogServer struct {
}

func (UnimplementedTaoBlogServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedTaoBlogServer) CreatePost(context.Context, *Post) (*Post, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePost not implemented")
}
func (UnimplementedTaoBlogServer) GetPost(context.Context, *GetPostRequest) (*Post, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPost not implemented")
}
func (UnimplementedTaoBlogServer) UpdatePost(context.Context, *UpdatePostRequest) (*Post, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePost not implemented")
}
func (UnimplementedTaoBlogServer) DeletePost(context.Context, *DeletePostRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeletePost not implemented")
}
func (UnimplementedTaoBlogServer) SetPostStatus(context.Context, *SetPostStatusRequest) (*SetPostStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetPostStatus not implemented")
}
func (UnimplementedTaoBlogServer) GetPostComments(context.Context, *GetPostCommentsRequest) (*GetPostCommentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPostComments not implemented")
}
func (UnimplementedTaoBlogServer) GetPostsByTags(context.Context, *GetPostsByTagsRequest) (*GetPostsByTagsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPostsByTags not implemented")
}
func (UnimplementedTaoBlogServer) CreateComment(context.Context, *Comment) (*Comment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateComment not implemented")
}
func (UnimplementedTaoBlogServer) GetComment(context.Context, *GetCommentRequest) (*Comment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetComment not implemented")
}
func (UnimplementedTaoBlogServer) UpdateComment(context.Context, *UpdateCommentRequest) (*Comment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateComment not implemented")
}
func (UnimplementedTaoBlogServer) DeleteComment(context.Context, *DeleteCommentRequest) (*DeleteCommentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteComment not implemented")
}
func (UnimplementedTaoBlogServer) ListComments(context.Context, *ListCommentsRequest) (*ListCommentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListComments not implemented")
}
func (UnimplementedTaoBlogServer) SetCommentPostID(context.Context, *SetCommentPostIDRequest) (*SetCommentPostIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetCommentPostID not implemented")
}
func (UnimplementedTaoBlogServer) GetPostCommentsCount(context.Context, *GetPostCommentsCountRequest) (*GetPostCommentsCountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPostCommentsCount not implemented")
}
func (UnimplementedTaoBlogServer) PreviewComment(context.Context, *PreviewCommentRequest) (*PreviewCommentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PreviewComment not implemented")
}
func (UnimplementedTaoBlogServer) mustEmbedUnimplementedTaoBlogServer() {}

// UnsafeTaoBlogServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TaoBlogServer will
// result in compilation errors.
type UnsafeTaoBlogServer interface {
	mustEmbedUnimplementedTaoBlogServer()
}

func RegisterTaoBlogServer(s grpc.ServiceRegistrar, srv TaoBlogServer) {
	s.RegisterService(&TaoBlog_ServiceDesc, srv)
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

func _TaoBlog_GetPostComments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPostCommentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).GetPostComments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/GetPostComments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).GetPostComments(ctx, req.(*GetPostCommentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TaoBlog_GetPostsByTags_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPostsByTagsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).GetPostsByTags(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/GetPostsByTags",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).GetPostsByTags(ctx, req.(*GetPostsByTagsRequest))
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

func _TaoBlog_PreviewComment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PreviewCommentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaoBlogServer).PreviewComment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocols.TaoBlog/PreviewComment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaoBlogServer).PreviewComment(ctx, req.(*PreviewCommentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TaoBlog_ServiceDesc is the grpc.ServiceDesc for TaoBlog service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TaoBlog_ServiceDesc = grpc.ServiceDesc{
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
			MethodName: "GetPostComments",
			Handler:    _TaoBlog_GetPostComments_Handler,
		},
		{
			MethodName: "GetPostsByTags",
			Handler:    _TaoBlog_GetPostsByTags_Handler,
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
		{
			MethodName: "PreviewComment",
			Handler:    _TaoBlog_PreviewComment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protocols/service.proto",
}
