// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protocols/post.proto

package protocols

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	field_mask "google.golang.org/genproto/protobuf/field_mask"
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

type PostType int32

const (
	PostType_PostType_Unspecified PostType = 0
	PostType_PostType_Post        PostType = 1
	PostType_PostType_Page        PostType = 2
)

var PostType_name = map[int32]string{
	0: "PostType_Unspecified",
	1: "PostType_Post",
	2: "PostType_Page",
}

var PostType_value = map[string]int32{
	"PostType_Unspecified": 0,
	"PostType_Post":        1,
	"PostType_Page":        2,
}

func (x PostType) String() string {
	return proto.EnumName(PostType_name, int32(x))
}

func (PostType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_3b419c78abee5f34, []int{0}
}

type Post struct {
	Id                   int64                `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Date                 *timestamp.Timestamp `protobuf:"bytes,2,opt,name=date,proto3" json:"date,omitempty"`
	Modified             *timestamp.Timestamp `protobuf:"bytes,3,opt,name=modified,proto3" json:"modified,omitempty"`
	Title                string               `protobuf:"bytes,4,opt,name=title,proto3" json:"title,omitempty"`
	Content              string               `protobuf:"bytes,5,opt,name=content,proto3" json:"content,omitempty"`
	Slug                 string               `protobuf:"bytes,6,opt,name=slug,proto3" json:"slug,omitempty"`
	Type                 PostType             `protobuf:"varint,7,opt,name=type,proto3,enum=protocols.PostType" json:"type,omitempty"`
	Category             int64                `protobuf:"varint,8,opt,name=category,proto3" json:"category,omitempty"`
	Status               string               `protobuf:"bytes,9,opt,name=status,proto3" json:"status,omitempty"`
	PageView             int64                `protobuf:"varint,10,opt,name=page_view,json=pageView,proto3" json:"page_view,omitempty"`
	CommentStatus        bool                 `protobuf:"varint,11,opt,name=comment_status,json=commentStatus,proto3" json:"comment_status,omitempty"`
	Comments             int64                `protobuf:"varint,12,opt,name=comments,proto3" json:"comments,omitempty"`
	Metas                string               `protobuf:"bytes,13,opt,name=metas,proto3" json:"metas,omitempty"`
	Source               string               `protobuf:"bytes,14,opt,name=source,proto3" json:"source,omitempty"`
	SourceType           string               `protobuf:"bytes,15,opt,name=source_type,json=sourceType,proto3" json:"source_type,omitempty"`
	Tags                 []string             `protobuf:"bytes,16,rep,name=tags,proto3" json:"tags,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Post) Reset()         { *m = Post{} }
func (m *Post) String() string { return proto.CompactTextString(m) }
func (*Post) ProtoMessage()    {}
func (*Post) Descriptor() ([]byte, []int) {
	return fileDescriptor_3b419c78abee5f34, []int{0}
}

func (m *Post) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Post.Unmarshal(m, b)
}
func (m *Post) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Post.Marshal(b, m, deterministic)
}
func (m *Post) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Post.Merge(m, src)
}
func (m *Post) XXX_Size() int {
	return xxx_messageInfo_Post.Size(m)
}
func (m *Post) XXX_DiscardUnknown() {
	xxx_messageInfo_Post.DiscardUnknown(m)
}

var xxx_messageInfo_Post proto.InternalMessageInfo

func (m *Post) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Post) GetDate() *timestamp.Timestamp {
	if m != nil {
		return m.Date
	}
	return nil
}

func (m *Post) GetModified() *timestamp.Timestamp {
	if m != nil {
		return m.Modified
	}
	return nil
}

func (m *Post) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *Post) GetContent() string {
	if m != nil {
		return m.Content
	}
	return ""
}

func (m *Post) GetSlug() string {
	if m != nil {
		return m.Slug
	}
	return ""
}

func (m *Post) GetType() PostType {
	if m != nil {
		return m.Type
	}
	return PostType_PostType_Unspecified
}

func (m *Post) GetCategory() int64 {
	if m != nil {
		return m.Category
	}
	return 0
}

func (m *Post) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *Post) GetPageView() int64 {
	if m != nil {
		return m.PageView
	}
	return 0
}

func (m *Post) GetCommentStatus() bool {
	if m != nil {
		return m.CommentStatus
	}
	return false
}

func (m *Post) GetComments() int64 {
	if m != nil {
		return m.Comments
	}
	return 0
}

func (m *Post) GetMetas() string {
	if m != nil {
		return m.Metas
	}
	return ""
}

func (m *Post) GetSource() string {
	if m != nil {
		return m.Source
	}
	return ""
}

func (m *Post) GetSourceType() string {
	if m != nil {
		return m.SourceType
	}
	return ""
}

func (m *Post) GetTags() []string {
	if m != nil {
		return m.Tags
	}
	return nil
}

type UpdatePostRequest struct {
	Post                 *Post                 `protobuf:"bytes,1,opt,name=post,proto3" json:"post,omitempty"`
	UpdateMask           *field_mask.FieldMask `protobuf:"bytes,2,opt,name=update_mask,json=updateMask,proto3" json:"update_mask,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *UpdatePostRequest) Reset()         { *m = UpdatePostRequest{} }
func (m *UpdatePostRequest) String() string { return proto.CompactTextString(m) }
func (*UpdatePostRequest) ProtoMessage()    {}
func (*UpdatePostRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3b419c78abee5f34, []int{1}
}

func (m *UpdatePostRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdatePostRequest.Unmarshal(m, b)
}
func (m *UpdatePostRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdatePostRequest.Marshal(b, m, deterministic)
}
func (m *UpdatePostRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdatePostRequest.Merge(m, src)
}
func (m *UpdatePostRequest) XXX_Size() int {
	return xxx_messageInfo_UpdatePostRequest.Size(m)
}
func (m *UpdatePostRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdatePostRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpdatePostRequest proto.InternalMessageInfo

func (m *UpdatePostRequest) GetPost() *Post {
	if m != nil {
		return m.Post
	}
	return nil
}

func (m *UpdatePostRequest) GetUpdateMask() *field_mask.FieldMask {
	if m != nil {
		return m.UpdateMask
	}
	return nil
}

func init() {
	proto.RegisterEnum("protocols.PostType", PostType_name, PostType_value)
	proto.RegisterType((*Post)(nil), "protocols.Post")
	proto.RegisterType((*UpdatePostRequest)(nil), "protocols.UpdatePostRequest")
}

func init() { proto.RegisterFile("protocols/post.proto", fileDescriptor_3b419c78abee5f34) }

var fileDescriptor_3b419c78abee5f34 = []byte{
	// 472 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x52, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0xc5, 0x89, 0xdb, 0x26, 0x13, 0x92, 0xa6, 0x4b, 0x84, 0x56, 0xe1, 0x50, 0x2b, 0x80, 0x88,
	0x38, 0xd8, 0x52, 0x90, 0xb8, 0x70, 0xe3, 0xc0, 0x01, 0x09, 0x09, 0x99, 0x96, 0x03, 0x97, 0x68,
	0x63, 0x4f, 0xcc, 0xaa, 0x76, 0xd6, 0x64, 0xc7, 0xad, 0xf2, 0xbb, 0x7c, 0x09, 0xda, 0x59, 0xdb,
	0x12, 0x95, 0x10, 0xb7, 0xf7, 0xde, 0xbe, 0x99, 0xdd, 0x79, 0x3b, 0xb0, 0xa8, 0x8f, 0x86, 0x4c,
	0x66, 0x4a, 0x9b, 0xd4, 0xc6, 0x52, 0xcc, 0x54, 0x8c, 0x7b, 0x75, 0x79, 0x5d, 0x18, 0x53, 0x94,
	0x98, 0xb0, 0xb2, 0x6b, 0xf6, 0x09, 0xe9, 0x0a, 0x2d, 0xa9, 0xaa, 0xf6, 0xde, 0x65, 0xf4, 0xd8,
	0xb0, 0xd7, 0x58, 0xe6, 0xdb, 0x4a, 0xd9, 0x3b, 0xef, 0x58, 0xfd, 0x1e, 0x42, 0xf8, 0xd5, 0x58,
	0x12, 0x33, 0x18, 0xe8, 0x5c, 0x06, 0x51, 0xb0, 0x1e, 0xa6, 0x03, 0x9d, 0x8b, 0x18, 0xc2, 0x5c,
	0x11, 0xca, 0x41, 0x14, 0xac, 0x27, 0x9b, 0x65, 0xec, 0x3b, 0xc5, 0x5d, 0xa7, 0xf8, 0xa6, 0xbb,
	0x2a, 0x65, 0x9f, 0x78, 0x0f, 0xa3, 0xca, 0xe4, 0x7a, 0xaf, 0x31, 0x97, 0xc3, 0xff, 0xd6, 0xf4,
	0x5e, 0xb1, 0x80, 0x33, 0xd2, 0x54, 0xa2, 0x0c, 0xa3, 0x60, 0x3d, 0x4e, 0x3d, 0x11, 0x12, 0x2e,
	0x32, 0x73, 0x20, 0x3c, 0x90, 0x3c, 0x63, 0xbd, 0xa3, 0x42, 0x40, 0x68, 0xcb, 0xa6, 0x90, 0xe7,
	0x2c, 0x33, 0x16, 0x6f, 0x20, 0xa4, 0x53, 0x8d, 0xf2, 0x22, 0x0a, 0xd6, 0xb3, 0xcd, 0xb3, 0xb8,
	0x4f, 0x28, 0x76, 0xa3, 0xdd, 0x9c, 0x6a, 0x4c, 0xd9, 0x20, 0x96, 0x30, 0xca, 0x14, 0x61, 0x61,
	0x8e, 0x27, 0x39, 0xe2, 0x51, 0x7b, 0x2e, 0x9e, 0xc3, 0xb9, 0x25, 0x45, 0x8d, 0x95, 0x63, 0x6e,
	0xdd, 0x32, 0xf1, 0x02, 0xc6, 0xb5, 0x2a, 0x70, 0x7b, 0xaf, 0xf1, 0x41, 0x82, 0x2f, 0x72, 0xc2,
	0x77, 0x8d, 0x0f, 0xe2, 0x35, 0xcc, 0x32, 0x53, 0x55, 0x78, 0xa0, 0x6d, 0x5b, 0x3c, 0x89, 0x82,
	0xf5, 0x28, 0x9d, 0xb6, 0xea, 0x37, 0xdf, 0xc3, 0xdd, 0xeb, 0x05, 0x2b, 0x9f, 0xb6, 0xf7, 0xb6,
	0xdc, 0x05, 0x50, 0x21, 0x29, 0x2b, 0xa7, 0x3e, 0x00, 0x26, 0xfc, 0x1a, 0xd3, 0x1c, 0x33, 0x94,
	0xb3, 0xf6, 0x35, 0xcc, 0xc4, 0x35, 0x4c, 0x3c, 0xda, 0xf2, 0xc4, 0x97, 0x7c, 0x08, 0x5e, 0x72,
	0x83, 0xba, 0x7c, 0x48, 0x15, 0x56, 0xce, 0xa3, 0xa1, 0xcb, 0xc7, 0xe1, 0x55, 0x03, 0x57, 0xb7,
	0xb5, 0xfb, 0x25, 0x17, 0x47, 0x8a, 0xbf, 0x1a, 0xb4, 0x24, 0x5e, 0x42, 0xe8, 0xb6, 0x8a, 0xbf,
	0x7c, 0xb2, 0xb9, 0x7c, 0x14, 0x5a, 0xca, 0x87, 0xe2, 0x03, 0x4c, 0x1a, 0xae, 0xe4, 0x9d, 0xf9,
	0xe7, 0x32, 0x7c, 0x72, 0x6b, 0xf5, 0x45, 0xd9, 0xbb, 0x14, 0xbc, 0xdd, 0xe1, 0xb7, 0x9f, 0x61,
	0xd4, 0xe5, 0x2f, 0x24, 0x2c, 0x3a, 0xbc, 0xbd, 0x3d, 0xd8, 0x1a, 0x33, 0xfe, 0xfe, 0xf9, 0x13,
	0x71, 0x05, 0xd3, 0xfe, 0xc4, 0x81, 0x79, 0xf0, 0xb7, 0xa4, 0x0a, 0x9c, 0x0f, 0x3e, 0xbe, 0xfa,
	0xb1, 0x2a, 0x34, 0xfd, 0x6c, 0x76, 0x71, 0x66, 0xaa, 0xa4, 0x32, 0xf7, 0x76, 0x97, 0x90, 0x32,
	0xbb, 0xd2, 0x14, 0x49, 0xff, 0xf2, 0xdd, 0x39, 0xc3, 0x77, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff,
	0x89, 0x17, 0x40, 0x7c, 0x3a, 0x03, 0x00, 0x00,
}