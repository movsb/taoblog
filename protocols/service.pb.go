// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.21.9
// source: protocols/service.proto

package protocols

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PingRequest) Reset() {
	*x = PingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingRequest) ProtoMessage() {}

func (x *PingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingRequest.ProtoReflect.Descriptor instead.
func (*PingRequest) Descriptor() ([]byte, []int) {
	return file_protocols_service_proto_rawDescGZIP(), []int{0}
}

type PingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pong string `protobuf:"bytes,1,opt,name=pong,proto3" json:"pong,omitempty"`
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingResponse) ProtoMessage() {}

func (x *PingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingResponse.ProtoReflect.Descriptor instead.
func (*PingResponse) Descriptor() ([]byte, []int) {
	return file_protocols_service_proto_rawDescGZIP(), []int{1}
}

func (x *PingResponse) GetPong() string {
	if x != nil {
		return x.Pong
	}
	return ""
}

var File_protocols_service_proto protoreflect.FileDescriptor

var file_protocols_service_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x73, 0x1a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2f,
	0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x14, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2f, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2f, 0x63, 0x6f,
	0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2f, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e,
	0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2c,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x73, 0x77, 0x61, 0x67, 0x67,
	0x65, 0x72, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x0d, 0x0a, 0x0b,
	0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x22, 0x0a, 0x0c, 0x50,
	0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70,
	0x6f, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x6e, 0x67, 0x32,
	0xcb, 0x02, 0x0a, 0x0a, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x76,
	0x0a, 0x06, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x12, 0x18, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x42,
	0x61, 0x63, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x35, 0x92,
	0x41, 0x20, 0x1a, 0x1e, 0xe5, 0xaf, 0xbc, 0xe5, 0x87, 0xba, 0xe6, 0x95, 0xb0, 0xe6, 0x8d, 0xae,
	0xe5, 0xba, 0x93, 0xef, 0xbc, 0x88, 0xe6, 0x8b, 0x96, 0xe5, 0xba, 0x93, 0xef, 0xbc, 0x89, 0xe3,
	0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0c, 0x12, 0x0a, 0x2f, 0x76, 0x33, 0x2f, 0x62, 0x61,
	0x63, 0x6b, 0x75, 0x70, 0x30, 0x01, 0x12, 0x66, 0x0a, 0x0b, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70,
	0x46, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x73, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73,
	0x2e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x14, 0x92, 0x41, 0x11, 0x1a, 0x0f, 0xe5, 0xaf, 0xbc, 0xe5, 0x87,
	0xba, 0xe6, 0x96, 0x87, 0xe4, 0xbb, 0xb6, 0xe3, 0x80, 0x82, 0x28, 0x01, 0x30, 0x01, 0x12, 0x5d,
	0x0a, 0x0b, 0x53, 0x65, 0x74, 0x52, 0x65, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74, 0x12, 0x1d, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x53, 0x65, 0x74, 0x52, 0x65, 0x64,
	0x69, 0x72, 0x65, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45,
	0x6d, 0x70, 0x74, 0x79, 0x22, 0x17, 0x92, 0x41, 0x14, 0x1a, 0x12, 0xe8, 0xae, 0xbe, 0xe7, 0xbd,
	0xae, 0xe9, 0x87, 0x8d, 0xe5, 0xae, 0x9a, 0xe5, 0x90, 0x91, 0xe3, 0x80, 0x82, 0x32, 0x84, 0x01,
	0x0a, 0x06, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x12, 0x7a, 0x0a, 0x0b, 0x53, 0x65, 0x61, 0x72,
	0x63, 0x68, 0x50, 0x6f, 0x73, 0x74, 0x73, 0x12, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x73, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x50, 0x6f, 0x73, 0x74, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x73, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x50, 0x6f, 0x73, 0x74, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2c, 0x92, 0x41, 0x11, 0x1a, 0x0f, 0xe6, 0x96, 0x87,
	0xe7, 0xab, 0xa0, 0xe6, 0x90, 0x9c, 0xe7, 0xb4, 0xa2, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x12, 0x12, 0x10, 0x2f, 0x76, 0x33, 0x2f, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x2f, 0x70,
	0x6f, 0x73, 0x74, 0x73, 0x32, 0xaf, 0x0f, 0x0a, 0x07, 0x54, 0x61, 0x6f, 0x42, 0x6c, 0x6f, 0x67,
	0x12, 0x5f, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x16, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x50, 0x69, 0x6e,
	0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x26, 0x92, 0x41, 0x13, 0x1a, 0x11,
	0x50, 0x69, 0x6e, 0x67, 0x20, 0xe6, 0x9c, 0x8d, 0xe5, 0x8a, 0xa1, 0xe5, 0x99, 0xa8, 0xe3, 0x80,
	0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0a, 0x12, 0x08, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x69, 0x6e,
	0x67, 0x12, 0x5b, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x50, 0x6f, 0x73, 0x74, 0x12,
	0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x50, 0x6f, 0x73, 0x74,
	0x1a, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x50, 0x6f, 0x73,
	0x74, 0x22, 0x2b, 0x92, 0x41, 0x14, 0x1a, 0x12, 0xe5, 0x88, 0x9b, 0xe5, 0xbb, 0xba, 0xe6, 0x96,
	0xb0, 0xe6, 0x96, 0x87, 0xe7, 0xab, 0xa0, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0e,
	0x3a, 0x01, 0x2a, 0x22, 0x09, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6f, 0x73, 0x74, 0x73, 0x12, 0x69,
	0x0a, 0x07, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73,
	0x2e, 0x50, 0x6f, 0x73, 0x74, 0x22, 0x32, 0x92, 0x41, 0x17, 0x1a, 0x15, 0xe8, 0x8e, 0xb7, 0xe5,
	0x8f, 0x96, 0xe6, 0x9f, 0x90, 0xe7, 0xaf, 0x87, 0xe6, 0x96, 0x87, 0xe7, 0xab, 0xa0, 0xe3, 0x80,
	0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x12, 0x12, 0x10, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6f, 0x73,
	0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x3d, 0x2a, 0x7d, 0x12, 0x71, 0x0a, 0x0a, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x73, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x73, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x22, 0x34, 0x92, 0x41, 0x11, 0x1a, 0x0f, 0xe6, 0x9b, 0xb4,
	0xe6, 0x96, 0xb0, 0xe6, 0x96, 0x87, 0xe7, 0xab, 0xa0, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x1a, 0x3a, 0x01, 0x2a, 0x32, 0x15, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6f, 0x73, 0x74, 0x73,
	0x2f, 0x7b, 0x70, 0x6f, 0x73, 0x74, 0x2e, 0x69, 0x64, 0x3d, 0x2a, 0x7d, 0x12, 0x9e, 0x01, 0x0a,
	0x0a, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x1c, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x50, 0x6f,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x22, 0x5a, 0x92, 0x41, 0x41, 0x1a, 0x3f, 0xe5, 0x88, 0xa0, 0xe9, 0x99, 0xa4, 0xe6, 0x96,
	0x87, 0xe7, 0xab, 0xa0, 0xe5, 0x8f, 0x8a, 0xe5, 0x85, 0xb6, 0xe6, 0x89, 0x80, 0xe6, 0x9c, 0x89,
	0xe7, 0x9b, 0xb8, 0xe5, 0x85, 0xb3, 0xe8, 0xb5, 0x84, 0xe6, 0xba, 0x90, 0xef, 0xbc, 0x88, 0xe8,
	0xaf, 0x84, 0xe8, 0xae, 0xba, 0xe3, 0x80, 0x81, 0xe6, 0xa0, 0x87, 0xe7, 0xad, 0xbe, 0xe7, 0xad,
	0x89, 0xef, 0xbc, 0x89, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x10, 0x2a, 0x0e, 0x2f,
	0x76, 0x33, 0x2f, 0x70, 0x6f, 0x73, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x93, 0x01,
	0x0a, 0x0d, 0x53, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x1f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x53, 0x65, 0x74, 0x50,
	0x6f, 0x73, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x20, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x53, 0x65, 0x74,
	0x50, 0x6f, 0x73, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x3f, 0x92, 0x41, 0x17, 0x1a, 0x15, 0xe6, 0x98, 0xaf, 0xe5, 0x90, 0xa6, 0xe5,
	0x85, 0xac, 0xe5, 0xbc, 0x80, 0xe6, 0x96, 0x87, 0xe7, 0xab, 0xa0, 0xe3, 0x80, 0x82, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x1f, 0x3a, 0x01, 0x2a, 0x22, 0x1a, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6f, 0x73,
	0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x3d, 0x2a, 0x7d, 0x3a, 0x73, 0x65, 0x74, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x7c, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x6d,
	0x6d, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73,
	0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x1a, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x22, 0x43, 0x92, 0x41,
	0x17, 0x1a, 0x15, 0xe5, 0x88, 0x9b, 0xe5, 0xbb, 0xba, 0xe4, 0xb8, 0x80, 0xe6, 0x9d, 0xa1, 0xe8,
	0xaf, 0x84, 0xe8, 0xae, 0xba, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x23, 0x3a, 0x01,
	0x2a, 0x22, 0x1e, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6f, 0x73, 0x74, 0x73, 0x2f, 0x7b, 0x70, 0x6f,
	0x73, 0x74, 0x5f, 0x69, 0x64, 0x3d, 0x2a, 0x7d, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74,
	0x73, 0x12, 0x7e, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x12,
	0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x43,
	0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e,
	0x74, 0x22, 0x3e, 0x92, 0x41, 0x20, 0x1a, 0x1e, 0xe8, 0x8e, 0xb7, 0xe5, 0x8f, 0x96, 0xe6, 0x8c,
	0x87, 0xe5, 0xae, 0x9a, 0xe7, 0xbc, 0x96, 0xe5, 0x8f, 0xb7, 0xe7, 0x9a, 0x84, 0xe8, 0xaf, 0x84,
	0xe8, 0xae, 0xba, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x15, 0x12, 0x13, 0x2f, 0x76,
	0x33, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x3d, 0x2a,
	0x7d, 0x12, 0x8f, 0x01, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x6d, 0x6d,
	0x65, 0x6e, 0x74, 0x12, 0x1f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73,
	0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x22, 0x49, 0x92, 0x41, 0x20, 0x1a, 0x1e, 0xe6,
	0x9b, 0xb4, 0xe6, 0x96, 0xb0, 0xe6, 0x8c, 0x87, 0xe5, 0xae, 0x9a, 0xe7, 0xbc, 0x96, 0xe5, 0x8f,
	0xb7, 0xe7, 0x9a, 0x84, 0xe8, 0xaf, 0x84, 0xe8, 0xae, 0xba, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x20, 0x3a, 0x01, 0x2a, 0x32, 0x1b, 0x2f, 0x76, 0x33, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x65, 0x6e, 0x74, 0x73, 0x2f, 0x7b, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x69, 0x64,
	0x3d, 0x2a, 0x7d, 0x12, 0x89, 0x01, 0x0a, 0x0d, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x43, 0x6f,
	0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x1f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x73, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x35, 0x92, 0x41, 0x17, 0x1a, 0x15, 0xe5,
	0x88, 0xa0, 0xe9, 0x99, 0xa4, 0xe6, 0x9f, 0x90, 0xe6, 0x9d, 0xa1, 0xe8, 0xaf, 0x84, 0xe8, 0xae,
	0xba, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x15, 0x2a, 0x13, 0x2f, 0x76, 0x33, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x3d, 0x2a, 0x7d, 0x12,
	0xa6, 0x01, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73,
	0x12, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x4c, 0x69, 0x73,
	0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x55, 0x92, 0x41, 0x2c, 0x1a, 0x2a, 0xe8, 0x8e, 0xb7, 0xe5, 0x8f, 0x96, 0xef, 0xbc,
	0x88, 0xe6, 0x9f, 0x90, 0xe7, 0xaf, 0x87, 0xe6, 0x96, 0x87, 0xe7, 0xab, 0xa0, 0xe7, 0x9a, 0x84,
	0xef, 0xbc, 0x89, 0xe8, 0xaf, 0x84, 0xe8, 0xae, 0xba, 0xe5, 0x88, 0x97, 0xe8, 0xa1, 0xa8, 0xe3,
	0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x20, 0x12, 0x1e, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6f,
	0x73, 0x74, 0x73, 0x2f, 0x7b, 0x70, 0x6f, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x3d, 0x2a, 0x7d, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0xc9, 0x01, 0x0a, 0x10, 0x53, 0x65, 0x74,
	0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x49, 0x44, 0x12, 0x22, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6d,
	0x6d, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x49, 0x44, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x23, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x53, 0x65,
	0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x49, 0x44, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x6c, 0x92, 0x41, 0x41, 0x1a, 0x3f, 0xe8, 0xbd, 0xac,
	0xe7, 0xa7, 0xbb, 0xe6, 0x9f, 0x90, 0xe9, 0xa1, 0xb6, 0xe7, 0xba, 0xa7, 0xe8, 0xaf, 0x84, 0xe8,
	0xae, 0xba, 0xef, 0xbc, 0x88, 0xe8, 0xbf, 0x9e, 0xe5, 0x90, 0x8c, 0xe5, 0xad, 0x90, 0xe8, 0xaf,
	0x84, 0xe8, 0xae, 0xba, 0xef, 0xbc, 0x89, 0xe5, 0x88, 0xb0, 0xe5, 0x8f, 0xa6, 0xe4, 0xb8, 0x80,
	0xe7, 0xaf, 0x87, 0xe6, 0x96, 0x87, 0xe7, 0xab, 0xa0, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x22, 0x3a, 0x01, 0x2a, 0x22, 0x1d, 0x2f, 0x76, 0x33, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x65,
	0x6e, 0x74, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x3d, 0x2a, 0x7d, 0x3a, 0x73, 0x65, 0x74, 0x50, 0x6f,
	0x73, 0x74, 0x49, 0x44, 0x12, 0xb2, 0x01, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74,
	0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x26, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73,
	0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x73, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74,
	0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x49,
	0x92, 0x41, 0x1a, 0x1a, 0x18, 0xe8, 0x8e, 0xb7, 0xe5, 0x8f, 0x96, 0xe6, 0x96, 0x87, 0xe7, 0xab,
	0xa0, 0xe8, 0xaf, 0x84, 0xe8, 0xae, 0xba, 0xe6, 0x95, 0xb0, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x26, 0x12, 0x24, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6f, 0x73, 0x74, 0x73, 0x2f, 0x7b,
	0x70, 0x6f, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x3d, 0x2a, 0x7d, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x65,
	0x6e, 0x74, 0x73, 0x3a, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x8a, 0x01, 0x0a, 0x0e, 0x50, 0x72,
	0x65, 0x76, 0x69, 0x65, 0x77, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x20, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x65, 0x76, 0x69, 0x65, 0x77,
	0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x50, 0x72, 0x65, 0x76, 0x69,
	0x65, 0x77, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x33, 0x92, 0x41, 0x11, 0x1a, 0x0f, 0xe8, 0xaf, 0x84, 0xe8, 0xae, 0xba, 0xe9, 0xa2,
	0x84, 0xe8, 0xa7, 0x88, 0xe3, 0x80, 0x82, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x19, 0x3a, 0x01, 0x2a,
	0x22, 0x14, 0x2f, 0x76, 0x33, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x3a, 0x70,
	0x72, 0x65, 0x76, 0x69, 0x65, 0x77, 0x42, 0x24, 0x5a, 0x22, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6f, 0x76, 0x73, 0x62, 0x2f, 0x74, 0x61, 0x6f, 0x62, 0x6c,
	0x6f, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protocols_service_proto_rawDescOnce sync.Once
	file_protocols_service_proto_rawDescData = file_protocols_service_proto_rawDesc
)

func file_protocols_service_proto_rawDescGZIP() []byte {
	file_protocols_service_proto_rawDescOnce.Do(func() {
		file_protocols_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_protocols_service_proto_rawDescData)
	})
	return file_protocols_service_proto_rawDescData
}

var file_protocols_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protocols_service_proto_goTypes = []interface{}{
	(*PingRequest)(nil),                  // 0: protocols.PingRequest
	(*PingResponse)(nil),                 // 1: protocols.PingResponse
	(*BackupRequest)(nil),                // 2: protocols.BackupRequest
	(*BackupFilesRequest)(nil),           // 3: protocols.BackupFilesRequest
	(*SetRedirectRequest)(nil),           // 4: protocols.SetRedirectRequest
	(*SearchPostsRequest)(nil),           // 5: protocols.SearchPostsRequest
	(*Post)(nil),                         // 6: protocols.Post
	(*GetPostRequest)(nil),               // 7: protocols.GetPostRequest
	(*UpdatePostRequest)(nil),            // 8: protocols.UpdatePostRequest
	(*DeletePostRequest)(nil),            // 9: protocols.DeletePostRequest
	(*SetPostStatusRequest)(nil),         // 10: protocols.SetPostStatusRequest
	(*Comment)(nil),                      // 11: protocols.Comment
	(*GetCommentRequest)(nil),            // 12: protocols.GetCommentRequest
	(*UpdateCommentRequest)(nil),         // 13: protocols.UpdateCommentRequest
	(*DeleteCommentRequest)(nil),         // 14: protocols.DeleteCommentRequest
	(*ListCommentsRequest)(nil),          // 15: protocols.ListCommentsRequest
	(*SetCommentPostIDRequest)(nil),      // 16: protocols.SetCommentPostIDRequest
	(*GetPostCommentsCountRequest)(nil),  // 17: protocols.GetPostCommentsCountRequest
	(*PreviewCommentRequest)(nil),        // 18: protocols.PreviewCommentRequest
	(*BackupResponse)(nil),               // 19: protocols.BackupResponse
	(*BackupFilesResponse)(nil),          // 20: protocols.BackupFilesResponse
	(*emptypb.Empty)(nil),                // 21: google.protobuf.Empty
	(*SearchPostsResponse)(nil),          // 22: protocols.SearchPostsResponse
	(*SetPostStatusResponse)(nil),        // 23: protocols.SetPostStatusResponse
	(*DeleteCommentResponse)(nil),        // 24: protocols.DeleteCommentResponse
	(*ListCommentsResponse)(nil),         // 25: protocols.ListCommentsResponse
	(*SetCommentPostIDResponse)(nil),     // 26: protocols.SetCommentPostIDResponse
	(*GetPostCommentsCountResponse)(nil), // 27: protocols.GetPostCommentsCountResponse
	(*PreviewCommentResponse)(nil),       // 28: protocols.PreviewCommentResponse
}
var file_protocols_service_proto_depIdxs = []int32{
	2,  // 0: protocols.Management.Backup:input_type -> protocols.BackupRequest
	3,  // 1: protocols.Management.BackupFiles:input_type -> protocols.BackupFilesRequest
	4,  // 2: protocols.Management.SetRedirect:input_type -> protocols.SetRedirectRequest
	5,  // 3: protocols.Search.SearchPosts:input_type -> protocols.SearchPostsRequest
	0,  // 4: protocols.TaoBlog.Ping:input_type -> protocols.PingRequest
	6,  // 5: protocols.TaoBlog.CreatePost:input_type -> protocols.Post
	7,  // 6: protocols.TaoBlog.GetPost:input_type -> protocols.GetPostRequest
	8,  // 7: protocols.TaoBlog.UpdatePost:input_type -> protocols.UpdatePostRequest
	9,  // 8: protocols.TaoBlog.DeletePost:input_type -> protocols.DeletePostRequest
	10, // 9: protocols.TaoBlog.SetPostStatus:input_type -> protocols.SetPostStatusRequest
	11, // 10: protocols.TaoBlog.CreateComment:input_type -> protocols.Comment
	12, // 11: protocols.TaoBlog.GetComment:input_type -> protocols.GetCommentRequest
	13, // 12: protocols.TaoBlog.UpdateComment:input_type -> protocols.UpdateCommentRequest
	14, // 13: protocols.TaoBlog.DeleteComment:input_type -> protocols.DeleteCommentRequest
	15, // 14: protocols.TaoBlog.ListComments:input_type -> protocols.ListCommentsRequest
	16, // 15: protocols.TaoBlog.SetCommentPostID:input_type -> protocols.SetCommentPostIDRequest
	17, // 16: protocols.TaoBlog.GetPostCommentsCount:input_type -> protocols.GetPostCommentsCountRequest
	18, // 17: protocols.TaoBlog.PreviewComment:input_type -> protocols.PreviewCommentRequest
	19, // 18: protocols.Management.Backup:output_type -> protocols.BackupResponse
	20, // 19: protocols.Management.BackupFiles:output_type -> protocols.BackupFilesResponse
	21, // 20: protocols.Management.SetRedirect:output_type -> google.protobuf.Empty
	22, // 21: protocols.Search.SearchPosts:output_type -> protocols.SearchPostsResponse
	1,  // 22: protocols.TaoBlog.Ping:output_type -> protocols.PingResponse
	6,  // 23: protocols.TaoBlog.CreatePost:output_type -> protocols.Post
	6,  // 24: protocols.TaoBlog.GetPost:output_type -> protocols.Post
	6,  // 25: protocols.TaoBlog.UpdatePost:output_type -> protocols.Post
	21, // 26: protocols.TaoBlog.DeletePost:output_type -> google.protobuf.Empty
	23, // 27: protocols.TaoBlog.SetPostStatus:output_type -> protocols.SetPostStatusResponse
	11, // 28: protocols.TaoBlog.CreateComment:output_type -> protocols.Comment
	11, // 29: protocols.TaoBlog.GetComment:output_type -> protocols.Comment
	11, // 30: protocols.TaoBlog.UpdateComment:output_type -> protocols.Comment
	24, // 31: protocols.TaoBlog.DeleteComment:output_type -> protocols.DeleteCommentResponse
	25, // 32: protocols.TaoBlog.ListComments:output_type -> protocols.ListCommentsResponse
	26, // 33: protocols.TaoBlog.SetCommentPostID:output_type -> protocols.SetCommentPostIDResponse
	27, // 34: protocols.TaoBlog.GetPostCommentsCount:output_type -> protocols.GetPostCommentsCountResponse
	28, // 35: protocols.TaoBlog.PreviewComment:output_type -> protocols.PreviewCommentResponse
	18, // [18:36] is the sub-list for method output_type
	0,  // [0:18] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_protocols_service_proto_init() }
func file_protocols_service_proto_init() {
	if File_protocols_service_proto != nil {
		return
	}
	file_protocols_backup_proto_init()
	file_protocols_post_proto_init()
	file_protocols_comment_proto_init()
	file_protocols_search_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_protocols_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_protocols_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protocols_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   3,
		},
		GoTypes:           file_protocols_service_proto_goTypes,
		DependencyIndexes: file_protocols_service_proto_depIdxs,
		MessageInfos:      file_protocols_service_proto_msgTypes,
	}.Build()
	File_protocols_service_proto = out.File
	file_protocols_service_proto_rawDesc = nil
	file_protocols_service_proto_goTypes = nil
	file_protocols_service_proto_depIdxs = nil
}
