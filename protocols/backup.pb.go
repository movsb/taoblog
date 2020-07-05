// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protocols/backup.proto

package protocols

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type BackupRequest struct {
	// 是否需要压缩数据（zlib）
	Compress             bool     `protobuf:"varint,1,opt,name=compress,proto3" json:"compress,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BackupRequest) Reset()         { *m = BackupRequest{} }
func (m *BackupRequest) String() string { return proto.CompactTextString(m) }
func (*BackupRequest) ProtoMessage()    {}
func (*BackupRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_b079f803aa0e9563, []int{0}
}

func (m *BackupRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BackupRequest.Unmarshal(m, b)
}
func (m *BackupRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BackupRequest.Marshal(b, m, deterministic)
}
func (m *BackupRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BackupRequest.Merge(m, src)
}
func (m *BackupRequest) XXX_Size() int {
	return xxx_messageInfo_BackupRequest.Size(m)
}
func (m *BackupRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_BackupRequest.DiscardUnknown(m)
}

var xxx_messageInfo_BackupRequest proto.InternalMessageInfo

func (m *BackupRequest) GetCompress() bool {
	if m != nil {
		return m.Compress
	}
	return false
}

type BackupResponse struct {
	// 数据库语句
	Data                 []byte   `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BackupResponse) Reset()         { *m = BackupResponse{} }
func (m *BackupResponse) String() string { return proto.CompactTextString(m) }
func (*BackupResponse) ProtoMessage()    {}
func (*BackupResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_b079f803aa0e9563, []int{1}
}

func (m *BackupResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BackupResponse.Unmarshal(m, b)
}
func (m *BackupResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BackupResponse.Marshal(b, m, deterministic)
}
func (m *BackupResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BackupResponse.Merge(m, src)
}
func (m *BackupResponse) XXX_Size() int {
	return xxx_messageInfo_BackupResponse.Size(m)
}
func (m *BackupResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_BackupResponse.DiscardUnknown(m)
}

var xxx_messageInfo_BackupResponse proto.InternalMessageInfo

func (m *BackupResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*BackupRequest)(nil), "protocols.BackupRequest")
	proto.RegisterType((*BackupResponse)(nil), "protocols.BackupResponse")
}

func init() { proto.RegisterFile("protocols/backup.proto", fileDescriptor_b079f803aa0e9563) }

var fileDescriptor_b079f803aa0e9563 = []byte{
	// 145 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2b, 0x28, 0xca, 0x2f,
	0xc9, 0x4f, 0xce, 0xcf, 0x29, 0xd6, 0x4f, 0x4a, 0x4c, 0xce, 0x2e, 0x2d, 0xd0, 0x03, 0x0b, 0x08,
	0x71, 0xc2, 0xc5, 0x95, 0xb4, 0xb9, 0x78, 0x9d, 0xc0, 0x52, 0x41, 0xa9, 0x85, 0xa5, 0xa9, 0xc5,
	0x25, 0x42, 0x52, 0x5c, 0x1c, 0xc9, 0xf9, 0xb9, 0x05, 0x45, 0xa9, 0xc5, 0xc5, 0x12, 0x8c, 0x0a,
	0x8c, 0x1a, 0x1c, 0x41, 0x70, 0xbe, 0x92, 0x0a, 0x17, 0x1f, 0x4c, 0x71, 0x71, 0x41, 0x7e, 0x5e,
	0x71, 0xaa, 0x90, 0x10, 0x17, 0x4b, 0x4a, 0x62, 0x49, 0x22, 0x58, 0x25, 0x4f, 0x10, 0x98, 0xed,
	0xa4, 0x12, 0xa5, 0x94, 0x9e, 0x59, 0x92, 0x51, 0x9a, 0xa4, 0x97, 0x9c, 0x9f, 0xab, 0x9f, 0x9b,
	0x5f, 0x56, 0x9c, 0xa4, 0x5f, 0x92, 0x98, 0x9f, 0x94, 0x93, 0x9f, 0xae, 0x0f, 0xb7, 0x38, 0x89,
	0x0d, 0xcc, 0x34, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0x4c, 0xf2, 0x18, 0x60, 0xa4, 0x00, 0x00,
	0x00,
}
