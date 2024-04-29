// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.21.9
// source: protocols/comment.proto

package protocols

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ListCommentsRequest_Mode int32

const (
	ListCommentsRequest_Unspecified ListCommentsRequest_Mode = 0
	// limit 控制顶级评论的数量，子评论全部返回。用于以树型结构展示。
	ListCommentsRequest_Tree ListCommentsRequest_Mode = 1
	// limit 控制全部评论的数量，不区分父子评论。用于平铺展示近期评论。
	ListCommentsRequest_Flat ListCommentsRequest_Mode = 2
)

// Enum value maps for ListCommentsRequest_Mode.
var (
	ListCommentsRequest_Mode_name = map[int32]string{
		0: "Unspecified",
		1: "Tree",
		2: "Flat",
	}
	ListCommentsRequest_Mode_value = map[string]int32{
		"Unspecified": 0,
		"Tree":        1,
		"Flat":        2,
	}
)

func (x ListCommentsRequest_Mode) Enum() *ListCommentsRequest_Mode {
	p := new(ListCommentsRequest_Mode)
	*p = x
	return p
}

func (x ListCommentsRequest_Mode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ListCommentsRequest_Mode) Descriptor() protoreflect.EnumDescriptor {
	return file_protocols_comment_proto_enumTypes[0].Descriptor()
}

func (ListCommentsRequest_Mode) Type() protoreflect.EnumType {
	return &file_protocols_comment_proto_enumTypes[0]
}

func (x ListCommentsRequest_Mode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ListCommentsRequest_Mode.Descriptor instead.
func (ListCommentsRequest_Mode) EnumDescriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{5, 0}
}

type Comment struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Parent      int64  `protobuf:"varint,2,opt,name=parent,proto3" json:"parent,omitempty"`
	Root        int64  `protobuf:"varint,3,opt,name=root,proto3" json:"root,omitempty"`
	PostId      int64  `protobuf:"varint,4,opt,name=post_id,json=postId,proto3" json:"post_id,omitempty"`
	Author      string `protobuf:"bytes,5,opt,name=author,proto3" json:"author,omitempty"`
	Email       string `protobuf:"bytes,6,opt,name=email,proto3" json:"email,omitempty"`
	Url         string `protobuf:"bytes,7,opt,name=url,proto3" json:"url,omitempty"`
	Ip          string `protobuf:"bytes,8,opt,name=ip,proto3" json:"ip,omitempty"`
	Date        int32  `protobuf:"varint,9,opt,name=date,proto3" json:"date,omitempty"`
	SourceType  string `protobuf:"bytes,10,opt,name=source_type,json=sourceType,proto3" json:"source_type,omitempty"`
	Source      string `protobuf:"bytes,11,opt,name=source,proto3" json:"source,omitempty"`
	Content     string `protobuf:"bytes,12,opt,name=content,proto3" json:"content,omitempty"`
	IsAdmin     bool   `protobuf:"varint,14,opt,name=is_admin,json=isAdmin,proto3" json:"is_admin,omitempty"`
	DateFuzzy   string `protobuf:"bytes,15,opt,name=date_fuzzy,json=dateFuzzy,proto3" json:"date_fuzzy,omitempty"`
	GeoLocation string `protobuf:"bytes,16,opt,name=geo_location,json=geoLocation,proto3" json:"geo_location,omitempty"`
	// 前端用户是否可以编辑此评论？
	// 仅在 list/create 接口中返回。
	CanEdit bool `protobuf:"varint,17,opt,name=can_edit,json=canEdit,proto3" json:"can_edit,omitempty"`
	// 头像内部临时ID
	Avatar   int32 `protobuf:"varint,18,opt,name=avatar,proto3" json:"avatar,omitempty"`
	Modified int32 `protobuf:"varint,19,opt,name=modified,proto3" json:"modified,omitempty"`
}

func (x *Comment) Reset() {
	*x = Comment{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Comment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Comment) ProtoMessage() {}

func (x *Comment) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Comment.ProtoReflect.Descriptor instead.
func (*Comment) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{0}
}

func (x *Comment) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Comment) GetParent() int64 {
	if x != nil {
		return x.Parent
	}
	return 0
}

func (x *Comment) GetRoot() int64 {
	if x != nil {
		return x.Root
	}
	return 0
}

func (x *Comment) GetPostId() int64 {
	if x != nil {
		return x.PostId
	}
	return 0
}

func (x *Comment) GetAuthor() string {
	if x != nil {
		return x.Author
	}
	return ""
}

func (x *Comment) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *Comment) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Comment) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *Comment) GetDate() int32 {
	if x != nil {
		return x.Date
	}
	return 0
}

func (x *Comment) GetSourceType() string {
	if x != nil {
		return x.SourceType
	}
	return ""
}

func (x *Comment) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *Comment) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *Comment) GetIsAdmin() bool {
	if x != nil {
		return x.IsAdmin
	}
	return false
}

func (x *Comment) GetDateFuzzy() string {
	if x != nil {
		return x.DateFuzzy
	}
	return ""
}

func (x *Comment) GetGeoLocation() string {
	if x != nil {
		return x.GeoLocation
	}
	return ""
}

func (x *Comment) GetCanEdit() bool {
	if x != nil {
		return x.CanEdit
	}
	return false
}

func (x *Comment) GetAvatar() int32 {
	if x != nil {
		return x.Avatar
	}
	return 0
}

func (x *Comment) GetModified() int32 {
	if x != nil {
		return x.Modified
	}
	return 0
}

type GetCommentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetCommentRequest) Reset() {
	*x = GetCommentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCommentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCommentRequest) ProtoMessage() {}

func (x *GetCommentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCommentRequest.ProtoReflect.Descriptor instead.
func (*GetCommentRequest) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{1}
}

func (x *GetCommentRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type UpdateCommentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Comment    *Comment               `protobuf:"bytes,1,opt,name=comment,proto3" json:"comment,omitempty"`
	UpdateMask *fieldmaskpb.FieldMask `protobuf:"bytes,2,opt,name=update_mask,json=updateMask,proto3" json:"update_mask,omitempty"`
}

func (x *UpdateCommentRequest) Reset() {
	*x = UpdateCommentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateCommentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateCommentRequest) ProtoMessage() {}

func (x *UpdateCommentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateCommentRequest.ProtoReflect.Descriptor instead.
func (*UpdateCommentRequest) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{2}
}

func (x *UpdateCommentRequest) GetComment() *Comment {
	if x != nil {
		return x.Comment
	}
	return nil
}

func (x *UpdateCommentRequest) GetUpdateMask() *fieldmaskpb.FieldMask {
	if x != nil {
		return x.UpdateMask
	}
	return nil
}

type DeleteCommentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id int32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeleteCommentRequest) Reset() {
	*x = DeleteCommentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteCommentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteCommentRequest) ProtoMessage() {}

func (x *DeleteCommentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteCommentRequest.ProtoReflect.Descriptor instead.
func (*DeleteCommentRequest) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{3}
}

func (x *DeleteCommentRequest) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type DeleteCommentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteCommentResponse) Reset() {
	*x = DeleteCommentResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteCommentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteCommentResponse) ProtoMessage() {}

func (x *DeleteCommentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteCommentResponse.ProtoReflect.Descriptor instead.
func (*DeleteCommentResponse) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{4}
}

type ListCommentsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Mode ListCommentsRequest_Mode `protobuf:"varint,1,opt,name=mode,proto3,enum=protocols.ListCommentsRequest_Mode" json:"mode,omitempty"`
	// 0 for all posts.
	PostId int64 `protobuf:"varint,2,opt,name=post_id,json=postId,proto3" json:"post_id,omitempty"`
	// defaults to "*".
	Fields []string `protobuf:"bytes,3,rep,name=fields,proto3" json:"fields,omitempty"`
	// must be > 0.
	Limit   int64  `protobuf:"varint,4,opt,name=limit,proto3" json:"limit,omitempty"`
	Offset  int64  `protobuf:"varint,5,opt,name=offset,proto3" json:"offset,omitempty"`
	OrderBy string `protobuf:"bytes,6,opt,name=order_by,json=orderBy,proto3" json:"order_by,omitempty"`
}

func (x *ListCommentsRequest) Reset() {
	*x = ListCommentsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListCommentsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListCommentsRequest) ProtoMessage() {}

func (x *ListCommentsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListCommentsRequest.ProtoReflect.Descriptor instead.
func (*ListCommentsRequest) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{5}
}

func (x *ListCommentsRequest) GetMode() ListCommentsRequest_Mode {
	if x != nil {
		return x.Mode
	}
	return ListCommentsRequest_Unspecified
}

func (x *ListCommentsRequest) GetPostId() int64 {
	if x != nil {
		return x.PostId
	}
	return 0
}

func (x *ListCommentsRequest) GetFields() []string {
	if x != nil {
		return x.Fields
	}
	return nil
}

func (x *ListCommentsRequest) GetLimit() int64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *ListCommentsRequest) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *ListCommentsRequest) GetOrderBy() string {
	if x != nil {
		return x.OrderBy
	}
	return ""
}

type ListCommentsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Comments []*Comment `protobuf:"bytes,1,rep,name=comments,proto3" json:"comments,omitempty"`
}

func (x *ListCommentsResponse) Reset() {
	*x = ListCommentsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListCommentsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListCommentsResponse) ProtoMessage() {}

func (x *ListCommentsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListCommentsResponse.ProtoReflect.Descriptor instead.
func (*ListCommentsResponse) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{6}
}

func (x *ListCommentsResponse) GetComments() []*Comment {
	if x != nil {
		return x.Comments
	}
	return nil
}

type SetCommentPostIDRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	PostId int64 `protobuf:"varint,2,opt,name=post_id,json=postId,proto3" json:"post_id,omitempty"`
}

func (x *SetCommentPostIDRequest) Reset() {
	*x = SetCommentPostIDRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetCommentPostIDRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetCommentPostIDRequest) ProtoMessage() {}

func (x *SetCommentPostIDRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetCommentPostIDRequest.ProtoReflect.Descriptor instead.
func (*SetCommentPostIDRequest) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{7}
}

func (x *SetCommentPostIDRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *SetCommentPostIDRequest) GetPostId() int64 {
	if x != nil {
		return x.PostId
	}
	return 0
}

type SetCommentPostIDResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SetCommentPostIDResponse) Reset() {
	*x = SetCommentPostIDResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetCommentPostIDResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetCommentPostIDResponse) ProtoMessage() {}

func (x *SetCommentPostIDResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetCommentPostIDResponse.ProtoReflect.Descriptor instead.
func (*SetCommentPostIDResponse) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{8}
}

type PreviewCommentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Markdown string `protobuf:"bytes,1,opt,name=markdown,proto3" json:"markdown,omitempty"`
	// 是否在新窗口中打开链接。
	// 对于正在编辑的评论来说，总是应该这样。
	OpenLinksInNewTab bool `protobuf:"varint,2,opt,name=open_links_in_new_tab,json=openLinksInNewTab,proto3" json:"open_links_in_new_tab,omitempty"`
	// 所属文章编号。不是必须的。
	// 但是为了计算图片的大小，建议加上。
	PostId int32 `protobuf:"varint,3,opt,name=post_id,json=postId,proto3" json:"post_id,omitempty"`
}

func (x *PreviewCommentRequest) Reset() {
	*x = PreviewCommentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PreviewCommentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreviewCommentRequest) ProtoMessage() {}

func (x *PreviewCommentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreviewCommentRequest.ProtoReflect.Descriptor instead.
func (*PreviewCommentRequest) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{9}
}

func (x *PreviewCommentRequest) GetMarkdown() string {
	if x != nil {
		return x.Markdown
	}
	return ""
}

func (x *PreviewCommentRequest) GetOpenLinksInNewTab() bool {
	if x != nil {
		return x.OpenLinksInNewTab
	}
	return false
}

func (x *PreviewCommentRequest) GetPostId() int32 {
	if x != nil {
		return x.PostId
	}
	return 0
}

type PreviewCommentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Html string `protobuf:"bytes,1,opt,name=html,proto3" json:"html,omitempty"`
}

func (x *PreviewCommentResponse) Reset() {
	*x = PreviewCommentResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocols_comment_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PreviewCommentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreviewCommentResponse) ProtoMessage() {}

func (x *PreviewCommentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocols_comment_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PreviewCommentResponse.ProtoReflect.Descriptor instead.
func (*PreviewCommentResponse) Descriptor() ([]byte, []int) {
	return file_protocols_comment_proto_rawDescGZIP(), []int{10}
}

func (x *PreviewCommentResponse) GetHtml() string {
	if x != nil {
		return x.Html
	}
	return ""
}

var File_protocols_comment_proto protoreflect.FileDescriptor

var file_protocols_comment_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x73, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x6d, 0x61, 0x73, 0x6b,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc1, 0x03, 0x0a, 0x07, 0x43, 0x6f, 0x6d, 0x6d, 0x65,
	0x6e, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x6f,
	0x6f, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x72, 0x6f, 0x6f, 0x74, 0x12, 0x17,
	0x0a, 0x07, 0x70, 0x6f, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x06, 0x70, 0x6f, 0x73, 0x74, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x12,
	0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x65, 0x6d, 0x61, 0x69, 0x6c, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x65, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18,
	0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x19,
	0x0a, 0x08, 0x69, 0x73, 0x5f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x69, 0x73, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x64, 0x61, 0x74,
	0x65, 0x5f, 0x66, 0x75, 0x7a, 0x7a, 0x79, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x64,
	0x61, 0x74, 0x65, 0x46, 0x75, 0x7a, 0x7a, 0x79, 0x12, 0x21, 0x0a, 0x0c, 0x67, 0x65, 0x6f, 0x5f,
	0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x67, 0x65, 0x6f, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x19, 0x0a, 0x08, 0x63,
	0x61, 0x6e, 0x5f, 0x65, 0x64, 0x69, 0x74, 0x18, 0x11, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x63,
	0x61, 0x6e, 0x45, 0x64, 0x69, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72,
	0x18, 0x12, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x1a,
	0x0a, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x18, 0x13, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x22, 0x23, 0x0a, 0x11, 0x47, 0x65,
	0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x22,
	0x81, 0x01, 0x0a, 0x14, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2c, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d,
	0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x07, 0x63,
	0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x3b, 0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x5f, 0x6d, 0x61, 0x73, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x4d, 0x61, 0x73, 0x6b, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4d,
	0x61, 0x73, 0x6b, 0x22, 0x26, 0x0a, 0x14, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x43, 0x6f, 0x6d,
	0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64, 0x22, 0x17, 0x0a, 0x15, 0x44,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0xf5, 0x01, 0x0a, 0x13, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6d,
	0x6d, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x37, 0x0a, 0x04,
	0x6d, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x23, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65,
	0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x4d, 0x6f, 0x64, 0x65, 0x52,
	0x04, 0x6d, 0x6f, 0x64, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x6f, 0x73, 0x74, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x70, 0x6f, 0x73, 0x74, 0x49, 0x64, 0x12, 0x16,
	0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06,
	0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x6f, 0x66,
	0x66, 0x73, 0x65, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x5f, 0x62, 0x79,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x42, 0x79, 0x22,
	0x2b, 0x0a, 0x04, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x6e, 0x73, 0x70, 0x65,
	0x63, 0x69, 0x66, 0x69, 0x65, 0x64, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x54, 0x72, 0x65, 0x65,
	0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x46, 0x6c, 0x61, 0x74, 0x10, 0x02, 0x22, 0x46, 0x0a, 0x14,
	0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x08, 0x63, 0x6f, 0x6d, 0x6d,
	0x65, 0x6e, 0x74, 0x73, 0x22, 0x42, 0x0a, 0x17, 0x53, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x65,
	0x6e, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x49, 0x44, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x17, 0x0a, 0x07, 0x70, 0x6f, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x06, 0x70, 0x6f, 0x73, 0x74, 0x49, 0x64, 0x22, 0x1a, 0x0a, 0x18, 0x53, 0x65, 0x74, 0x43,
	0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x7e, 0x0a, 0x15, 0x50, 0x72, 0x65, 0x76, 0x69, 0x65, 0x77, 0x43,
	0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x6d, 0x61, 0x72, 0x6b, 0x64, 0x6f, 0x77, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x6d, 0x61, 0x72, 0x6b, 0x64, 0x6f, 0x77, 0x6e, 0x12, 0x30, 0x0a, 0x15, 0x6f, 0x70, 0x65,
	0x6e, 0x5f, 0x6c, 0x69, 0x6e, 0x6b, 0x73, 0x5f, 0x69, 0x6e, 0x5f, 0x6e, 0x65, 0x77, 0x5f, 0x74,
	0x61, 0x62, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x11, 0x6f, 0x70, 0x65, 0x6e, 0x4c, 0x69,
	0x6e, 0x6b, 0x73, 0x49, 0x6e, 0x4e, 0x65, 0x77, 0x54, 0x61, 0x62, 0x12, 0x17, 0x0a, 0x07, 0x70,
	0x6f, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x70, 0x6f,
	0x73, 0x74, 0x49, 0x64, 0x22, 0x2c, 0x0a, 0x16, 0x50, 0x72, 0x65, 0x76, 0x69, 0x65, 0x77, 0x43,
	0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x68, 0x74, 0x6d, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x74,
	0x6d, 0x6c, 0x42, 0x24, 0x5a, 0x22, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x6d, 0x6f, 0x76, 0x73, 0x62, 0x2f, 0x74, 0x61, 0x6f, 0x62, 0x6c, 0x6f, 0x67, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protocols_comment_proto_rawDescOnce sync.Once
	file_protocols_comment_proto_rawDescData = file_protocols_comment_proto_rawDesc
)

func file_protocols_comment_proto_rawDescGZIP() []byte {
	file_protocols_comment_proto_rawDescOnce.Do(func() {
		file_protocols_comment_proto_rawDescData = protoimpl.X.CompressGZIP(file_protocols_comment_proto_rawDescData)
	})
	return file_protocols_comment_proto_rawDescData
}

var file_protocols_comment_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protocols_comment_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_protocols_comment_proto_goTypes = []interface{}{
	(ListCommentsRequest_Mode)(0),    // 0: protocols.ListCommentsRequest.Mode
	(*Comment)(nil),                  // 1: protocols.Comment
	(*GetCommentRequest)(nil),        // 2: protocols.GetCommentRequest
	(*UpdateCommentRequest)(nil),     // 3: protocols.UpdateCommentRequest
	(*DeleteCommentRequest)(nil),     // 4: protocols.DeleteCommentRequest
	(*DeleteCommentResponse)(nil),    // 5: protocols.DeleteCommentResponse
	(*ListCommentsRequest)(nil),      // 6: protocols.ListCommentsRequest
	(*ListCommentsResponse)(nil),     // 7: protocols.ListCommentsResponse
	(*SetCommentPostIDRequest)(nil),  // 8: protocols.SetCommentPostIDRequest
	(*SetCommentPostIDResponse)(nil), // 9: protocols.SetCommentPostIDResponse
	(*PreviewCommentRequest)(nil),    // 10: protocols.PreviewCommentRequest
	(*PreviewCommentResponse)(nil),   // 11: protocols.PreviewCommentResponse
	(*fieldmaskpb.FieldMask)(nil),    // 12: google.protobuf.FieldMask
}
var file_protocols_comment_proto_depIdxs = []int32{
	1,  // 0: protocols.UpdateCommentRequest.comment:type_name -> protocols.Comment
	12, // 1: protocols.UpdateCommentRequest.update_mask:type_name -> google.protobuf.FieldMask
	0,  // 2: protocols.ListCommentsRequest.mode:type_name -> protocols.ListCommentsRequest.Mode
	1,  // 3: protocols.ListCommentsResponse.comments:type_name -> protocols.Comment
	4,  // [4:4] is the sub-list for method output_type
	4,  // [4:4] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_protocols_comment_proto_init() }
func file_protocols_comment_proto_init() {
	if File_protocols_comment_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protocols_comment_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Comment); i {
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
		file_protocols_comment_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCommentRequest); i {
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
		file_protocols_comment_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateCommentRequest); i {
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
		file_protocols_comment_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteCommentRequest); i {
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
		file_protocols_comment_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteCommentResponse); i {
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
		file_protocols_comment_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListCommentsRequest); i {
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
		file_protocols_comment_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListCommentsResponse); i {
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
		file_protocols_comment_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetCommentPostIDRequest); i {
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
		file_protocols_comment_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetCommentPostIDResponse); i {
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
		file_protocols_comment_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PreviewCommentRequest); i {
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
		file_protocols_comment_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PreviewCommentResponse); i {
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
			RawDescriptor: file_protocols_comment_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protocols_comment_proto_goTypes,
		DependencyIndexes: file_protocols_comment_proto_depIdxs,
		EnumInfos:         file_protocols_comment_proto_enumTypes,
		MessageInfos:      file_protocols_comment_proto_msgTypes,
	}.Build()
	File_protocols_comment_proto = out.File
	file_protocols_comment_proto_rawDesc = nil
	file_protocols_comment_proto_goTypes = nil
	file_protocols_comment_proto_depIdxs = nil
}
