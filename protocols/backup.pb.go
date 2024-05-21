// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.21.9
// source: protocols/backup.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type BackupRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 是否需要压缩数据（zlib）
	Compress bool `protobuf:"varint,1,opt,name=compress,proto3" json:"compress,omitempty"`
}

func (x *BackupRequest) Reset() {
	*x = BackupRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupRequest) ProtoMessage() {}

func (x *BackupRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupRequest.ProtoReflect.Descriptor instead.
func (*BackupRequest) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{0}
}

func (x *BackupRequest) GetCompress() bool {
	if x != nil {
		return x.Compress
	}
	return false
}

type BackupResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to BackupResponseMessage:
	//
	//	*BackupResponse_Preparing_
	//	*BackupResponse_Transfering_
	BackupResponseMessage isBackupResponse_BackupResponseMessage `protobuf_oneof:"BackupResponseMessage"`
}

func (x *BackupResponse) Reset() {
	*x = BackupResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupResponse) ProtoMessage() {}

func (x *BackupResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupResponse.ProtoReflect.Descriptor instead.
func (*BackupResponse) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{1}
}

func (m *BackupResponse) GetBackupResponseMessage() isBackupResponse_BackupResponseMessage {
	if m != nil {
		return m.BackupResponseMessage
	}
	return nil
}

func (x *BackupResponse) GetPreparing() *BackupResponse_Preparing {
	if x, ok := x.GetBackupResponseMessage().(*BackupResponse_Preparing_); ok {
		return x.Preparing
	}
	return nil
}

func (x *BackupResponse) GetTransfering() *BackupResponse_Transfering {
	if x, ok := x.GetBackupResponseMessage().(*BackupResponse_Transfering_); ok {
		return x.Transfering
	}
	return nil
}

type isBackupResponse_BackupResponseMessage interface {
	isBackupResponse_BackupResponseMessage()
}

type BackupResponse_Preparing_ struct {
	Preparing *BackupResponse_Preparing `protobuf:"bytes,1,opt,name=preparing,proto3,oneof"`
}

type BackupResponse_Transfering_ struct {
	Transfering *BackupResponse_Transfering `protobuf:"bytes,2,opt,name=transfering,proto3,oneof"`
}

func (*BackupResponse_Preparing_) isBackupResponse_BackupResponseMessage() {}

func (*BackupResponse_Transfering_) isBackupResponse_BackupResponseMessage() {}

type BackupFileSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Mode uint32 `protobuf:"varint,2,opt,name=mode,proto3" json:"mode,omitempty"`
	Size uint32 `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	Time uint32 `protobuf:"varint,4,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *BackupFileSpec) Reset() {
	*x = BackupFileSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupFileSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupFileSpec) ProtoMessage() {}

func (x *BackupFileSpec) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupFileSpec.ProtoReflect.Descriptor instead.
func (*BackupFileSpec) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{2}
}

func (x *BackupFileSpec) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *BackupFileSpec) GetMode() uint32 {
	if x != nil {
		return x.Mode
	}
	return 0
}

func (x *BackupFileSpec) GetSize() uint32 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *BackupFileSpec) GetTime() uint32 {
	if x != nil {
		return x.Time
	}
	return 0
}

type BackupFilesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to BackupFilesMessage:
	//
	//	*BackupFilesRequest_ListFiles
	//	*BackupFilesRequest_SendFile
	BackupFilesMessage isBackupFilesRequest_BackupFilesMessage `protobuf_oneof:"BackupFilesMessage"`
}

func (x *BackupFilesRequest) Reset() {
	*x = BackupFilesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupFilesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupFilesRequest) ProtoMessage() {}

func (x *BackupFilesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupFilesRequest.ProtoReflect.Descriptor instead.
func (*BackupFilesRequest) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{3}
}

func (m *BackupFilesRequest) GetBackupFilesMessage() isBackupFilesRequest_BackupFilesMessage {
	if m != nil {
		return m.BackupFilesMessage
	}
	return nil
}

func (x *BackupFilesRequest) GetListFiles() *BackupFilesRequest_ListFilesRequest {
	if x, ok := x.GetBackupFilesMessage().(*BackupFilesRequest_ListFiles); ok {
		return x.ListFiles
	}
	return nil
}

func (x *BackupFilesRequest) GetSendFile() *BackupFilesRequest_SendFileRequest {
	if x, ok := x.GetBackupFilesMessage().(*BackupFilesRequest_SendFile); ok {
		return x.SendFile
	}
	return nil
}

type isBackupFilesRequest_BackupFilesMessage interface {
	isBackupFilesRequest_BackupFilesMessage()
}

type BackupFilesRequest_ListFiles struct {
	ListFiles *BackupFilesRequest_ListFilesRequest `protobuf:"bytes,1,opt,name=list_files,json=listFiles,proto3,oneof"`
}

type BackupFilesRequest_SendFile struct {
	SendFile *BackupFilesRequest_SendFileRequest `protobuf:"bytes,2,opt,name=send_file,json=sendFile,proto3,oneof"`
}

func (*BackupFilesRequest_ListFiles) isBackupFilesRequest_BackupFilesMessage() {}

func (*BackupFilesRequest_SendFile) isBackupFilesRequest_BackupFilesMessage() {}

type BackupFilesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to BackupFilesMessage:
	//
	//	*BackupFilesResponse_ListFiles
	//	*BackupFilesResponse_SendFile
	BackupFilesMessage isBackupFilesResponse_BackupFilesMessage `protobuf_oneof:"BackupFilesMessage"`
}

func (x *BackupFilesResponse) Reset() {
	*x = BackupFilesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupFilesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupFilesResponse) ProtoMessage() {}

func (x *BackupFilesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupFilesResponse.ProtoReflect.Descriptor instead.
func (*BackupFilesResponse) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{4}
}

func (m *BackupFilesResponse) GetBackupFilesMessage() isBackupFilesResponse_BackupFilesMessage {
	if m != nil {
		return m.BackupFilesMessage
	}
	return nil
}

func (x *BackupFilesResponse) GetListFiles() *BackupFilesResponse_ListFilesResponse {
	if x, ok := x.GetBackupFilesMessage().(*BackupFilesResponse_ListFiles); ok {
		return x.ListFiles
	}
	return nil
}

func (x *BackupFilesResponse) GetSendFile() *BackupFilesResponse_SendFileResponse {
	if x, ok := x.GetBackupFilesMessage().(*BackupFilesResponse_SendFile); ok {
		return x.SendFile
	}
	return nil
}

type isBackupFilesResponse_BackupFilesMessage interface {
	isBackupFilesResponse_BackupFilesMessage()
}

type BackupFilesResponse_ListFiles struct {
	ListFiles *BackupFilesResponse_ListFilesResponse `protobuf:"bytes,1,opt,name=list_files,json=listFiles,proto3,oneof"`
}

type BackupFilesResponse_SendFile struct {
	SendFile *BackupFilesResponse_SendFileResponse `protobuf:"bytes,2,opt,name=send_file,json=sendFile,proto3,oneof"`
}

func (*BackupFilesResponse_ListFiles) isBackupFilesResponse_BackupFilesMessage() {}

func (*BackupFilesResponse_SendFile) isBackupFilesResponse_BackupFilesMessage() {}

type BackupResponse_Preparing struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Progress float32 `protobuf:"fixed32,1,opt,name=progress,proto3" json:"progress,omitempty"`
}

func (x *BackupResponse_Preparing) Reset() {
	*x = BackupResponse_Preparing{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupResponse_Preparing) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupResponse_Preparing) ProtoMessage() {}

func (x *BackupResponse_Preparing) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupResponse_Preparing.ProtoReflect.Descriptor instead.
func (*BackupResponse_Preparing) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{1, 0}
}

func (x *BackupResponse_Preparing) GetProgress() float32 {
	if x != nil {
		return x.Progress
	}
	return 0
}

type BackupResponse_Transfering struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Progress float32 `protobuf:"fixed32,1,opt,name=progress,proto3" json:"progress,omitempty"`
	Data     []byte  `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *BackupResponse_Transfering) Reset() {
	*x = BackupResponse_Transfering{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupResponse_Transfering) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupResponse_Transfering) ProtoMessage() {}

func (x *BackupResponse_Transfering) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupResponse_Transfering.ProtoReflect.Descriptor instead.
func (*BackupResponse_Transfering) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{1, 1}
}

func (x *BackupResponse_Transfering) GetProgress() float32 {
	if x != nil {
		return x.Progress
	}
	return 0
}

func (x *BackupResponse_Transfering) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type BackupFilesRequest_ListFilesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *BackupFilesRequest_ListFilesRequest) Reset() {
	*x = BackupFilesRequest_ListFilesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupFilesRequest_ListFilesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupFilesRequest_ListFilesRequest) ProtoMessage() {}

func (x *BackupFilesRequest_ListFilesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupFilesRequest_ListFilesRequest.ProtoReflect.Descriptor instead.
func (*BackupFilesRequest_ListFilesRequest) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{3, 0}
}

type BackupFilesRequest_SendFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
}

func (x *BackupFilesRequest_SendFileRequest) Reset() {
	*x = BackupFilesRequest_SendFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupFilesRequest_SendFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupFilesRequest_SendFileRequest) ProtoMessage() {}

func (x *BackupFilesRequest_SendFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupFilesRequest_SendFileRequest.ProtoReflect.Descriptor instead.
func (*BackupFilesRequest_SendFileRequest) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{3, 1}
}

func (x *BackupFilesRequest_SendFileRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type BackupFilesResponse_ListFilesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Files []*BackupFileSpec `protobuf:"bytes,1,rep,name=files,proto3" json:"files,omitempty"`
}

func (x *BackupFilesResponse_ListFilesResponse) Reset() {
	*x = BackupFilesResponse_ListFilesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupFilesResponse_ListFilesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupFilesResponse_ListFilesResponse) ProtoMessage() {}

func (x *BackupFilesResponse_ListFilesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupFilesResponse_ListFilesResponse.ProtoReflect.Descriptor instead.
func (*BackupFilesResponse_ListFilesResponse) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{4, 0}
}

func (x *BackupFilesResponse_ListFilesResponse) GetFiles() []*BackupFileSpec {
	if x != nil {
		return x.Files
	}
	return nil
}

type BackupFilesResponse_SendFileResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *BackupFilesResponse_SendFileResponse) Reset() {
	*x = BackupFilesResponse_SendFileResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_backup_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BackupFilesResponse_SendFileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BackupFilesResponse_SendFileResponse) ProtoMessage() {}

func (x *BackupFilesResponse_SendFileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_backup_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BackupFilesResponse_SendFileResponse.ProtoReflect.Descriptor instead.
func (*BackupFilesResponse_SendFileResponse) Descriptor() ([]byte, []int) {
	return file_protocols_backup_proto_rawDescGZIP(), []int{4, 1}
}

func (x *BackupFilesResponse_SendFileResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_protocols_backup_proto protoreflect.FileDescriptor

var file_protocols_backup_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2f, 0x62, 0x61, 0x63, 0x6b,
	0x75, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x73, 0x22, 0x2b, 0x0a, 0x0d, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73,
	0x22, 0xa1, 0x02, 0x0a, 0x0e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x43, 0x0a, 0x09, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x69, 0x6e, 0x67,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x73, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x2e, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x69, 0x6e, 0x67, 0x48, 0x00, 0x52, 0x09, 0x70,
	0x72, 0x65, 0x70, 0x61, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x49, 0x0a, 0x0b, 0x74, 0x72, 0x61, 0x6e,
	0x73, 0x66, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65,
	0x72, 0x69, 0x6e, 0x67, 0x48, 0x00, 0x52, 0x0b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72,
	0x69, 0x6e, 0x67, 0x1a, 0x27, 0x0a, 0x09, 0x50, 0x72, 0x65, 0x70, 0x61, 0x72, 0x69, 0x6e, 0x67,
	0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x02, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x1a, 0x3d, 0x0a, 0x0b,
	0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x70,
	0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x02, 0x52, 0x08, 0x70,
	0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42, 0x17, 0x0a, 0x15, 0x42,
	0x61, 0x63, 0x6b, 0x75, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x22, 0x60, 0x0a, 0x0e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x46, 0x69,
	0x6c, 0x65, 0x53, 0x70, 0x65, 0x63, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f,
	0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x73, 0x69,
	0x7a, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x22, 0x84, 0x02, 0x0a, 0x12, 0x42, 0x61, 0x63, 0x6b, 0x75,
	0x70, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x4f, 0x0a,
	0x0a, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x2e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x42, 0x61,
	0x63, 0x6b, 0x75, 0x70, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x48, 0x00, 0x52, 0x09, 0x6c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x4c,
	0x0a, 0x09, 0x73, 0x65, 0x6e, 0x64, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x2d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x42, 0x61,
	0x63, 0x6b, 0x75, 0x70, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x2e, 0x53, 0x65, 0x6e, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x48, 0x00, 0x52, 0x08, 0x73, 0x65, 0x6e, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x1a, 0x12, 0x0a, 0x10,
	0x4c, 0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x25, 0x0a, 0x0f, 0x53, 0x65, 0x6e, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x42, 0x14, 0x0a, 0x12, 0x42, 0x61, 0x63, 0x6b, 0x75,
	0x70, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0xbc, 0x02,
	0x0a, 0x13, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x51, 0x0a, 0x0a, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x66, 0x69,
	0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x46, 0x69, 0x6c, 0x65,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x46, 0x69,
	0x6c, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x09, 0x6c,
	0x69, 0x73, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x4e, 0x0a, 0x09, 0x73, 0x65, 0x6e, 0x64,
	0x5f, 0x66, 0x69, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x46, 0x69,
	0x6c, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x53, 0x65, 0x6e, 0x64,
	0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x08,
	0x73, 0x65, 0x6e, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x1a, 0x44, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74,
	0x46, 0x69, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2f, 0x0a,
	0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x46,
	0x69, 0x6c, 0x65, 0x53, 0x70, 0x65, 0x63, 0x52, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x1a, 0x26,
	0x0a, 0x10, 0x53, 0x65, 0x6e, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x42, 0x14, 0x0a, 0x12, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70,
	0x46, 0x69, 0x6c, 0x65, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x20, 0x5a, 0x1e,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6f, 0x76, 0x73, 0x62,
	0x2f, 0x74, 0x61, 0x6f, 0x62, 0x6c, 0x6f, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protocols_backup_proto_rawDescOnce sync.Once
	file_protocols_backup_proto_rawDescData = file_protocols_backup_proto_rawDesc
)

func file_protocols_backup_proto_rawDescGZIP() []byte {
	file_protocols_backup_proto_rawDescOnce.Do(func() {
		file_protocols_backup_proto_rawDescData = protoimpl.X.CompressGZIP(file_protocols_backup_proto_rawDescData)
	})
	return file_protocols_backup_proto_rawDescData
}

var file_protocols_backup_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_protocols_backup_proto_goTypes = []interface{}{
	(*BackupRequest)(nil),                         // 0: protocols.BackupRequest
	(*BackupResponse)(nil),                        // 1: protocols.BackupResponse
	(*BackupFileSpec)(nil),                        // 2: protocols.BackupFileSpec
	(*BackupFilesRequest)(nil),                    // 3: protocols.BackupFilesRequest
	(*BackupFilesResponse)(nil),                   // 4: protocols.BackupFilesResponse
	(*BackupResponse_Preparing)(nil),              // 5: protocols.BackupResponse.Preparing
	(*BackupResponse_Transfering)(nil),            // 6: protocols.BackupResponse.Transfering
	(*BackupFilesRequest_ListFilesRequest)(nil),   // 7: protocols.BackupFilesRequest.ListFilesRequest
	(*BackupFilesRequest_SendFileRequest)(nil),    // 8: protocols.BackupFilesRequest.SendFileRequest
	(*BackupFilesResponse_ListFilesResponse)(nil), // 9: protocols.BackupFilesResponse.ListFilesResponse
	(*BackupFilesResponse_SendFileResponse)(nil),  // 10: protocols.BackupFilesResponse.SendFileResponse
}
var file_protocols_backup_proto_depIdxs = []int32{
	5,  // 0: protocols.BackupResponse.preparing:type_name -> protocols.BackupResponse.Preparing
	6,  // 1: protocols.BackupResponse.transfering:type_name -> protocols.BackupResponse.Transfering
	7,  // 2: protocols.BackupFilesRequest.list_files:type_name -> protocols.BackupFilesRequest.ListFilesRequest
	8,  // 3: protocols.BackupFilesRequest.send_file:type_name -> protocols.BackupFilesRequest.SendFileRequest
	9,  // 4: protocols.BackupFilesResponse.list_files:type_name -> protocols.BackupFilesResponse.ListFilesResponse
	10, // 5: protocols.BackupFilesResponse.send_file:type_name -> protocols.BackupFilesResponse.SendFileResponse
	2,  // 6: protocols.BackupFilesResponse.ListFilesResponse.files:type_name -> protocols.BackupFileSpec
	7,  // [7:7] is the sub-list for method output_type
	7,  // [7:7] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_protocols_backup_proto_init() }
func file_protocols_backup_proto_init() {
	if File_protocols_backup_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protocols_backup_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupRequest); i {
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
		file_protocols_backup_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupResponse); i {
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
		file_protocols_backup_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupFileSpec); i {
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
		file_protocols_backup_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupFilesRequest); i {
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
		file_protocols_backup_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupFilesResponse); i {
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
		file_protocols_backup_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupResponse_Preparing); i {
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
		file_protocols_backup_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupResponse_Transfering); i {
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
		file_protocols_backup_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupFilesRequest_ListFilesRequest); i {
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
		file_protocols_backup_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupFilesRequest_SendFileRequest); i {
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
		file_protocols_backup_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupFilesResponse_ListFilesResponse); i {
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
		file_protocols_backup_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BackupFilesResponse_SendFileResponse); i {
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
	file_protocols_backup_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*BackupResponse_Preparing_)(nil),
		(*BackupResponse_Transfering_)(nil),
	}
	file_protocols_backup_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*BackupFilesRequest_ListFiles)(nil),
		(*BackupFilesRequest_SendFile)(nil),
	}
	file_protocols_backup_proto_msgTypes[4].OneofWrappers = []interface{}{
		(*BackupFilesResponse_ListFiles)(nil),
		(*BackupFilesResponse_SendFile)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protocols_backup_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protocols_backup_proto_goTypes,
		DependencyIndexes: file_protocols_backup_proto_depIdxs,
		MessageInfos:      file_protocols_backup_proto_msgTypes,
	}.Build()
	File_protocols_backup_proto = out.File
	file_protocols_backup_proto_rawDesc = nil
	file_protocols_backup_proto_goTypes = nil
	file_protocols_backup_proto_depIdxs = nil
}
