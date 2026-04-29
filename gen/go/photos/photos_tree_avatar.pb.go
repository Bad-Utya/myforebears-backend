package photospb

import (
	proto "google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
)

type UploadTreeAvatarRequest struct {
	state         protoimpl.MessageState
	TreeId        string `protobuf:"bytes,1,opt,name=tree_id,json=treeId,proto3" json:"tree_id,omitempty"`
	FileName      string `protobuf:"bytes,2,opt,name=file_name,json=fileName,proto3" json:"file_name,omitempty"`
	MimeType      string `protobuf:"bytes,3,opt,name=mime_type,json=mimeType,proto3" json:"mime_type,omitempty"`
	Content       []byte `protobuf:"bytes,4,opt,name=content,proto3" json:"content,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UploadTreeAvatarRequest) Reset() {
	*x = UploadTreeAvatarRequest{}
	mi := &file_photos_photos_tree_avatar_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UploadTreeAvatarRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*UploadTreeAvatarRequest) ProtoMessage()    {}

func (x *UploadTreeAvatarRequest) ProtoReflect() protoreflect.Message {
	mi := &file_photos_photos_tree_avatar_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*UploadTreeAvatarRequest) Descriptor() ([]byte, []int) {
	return nil, []int{0}
}

func (x *UploadTreeAvatarRequest) GetTreeId() string {
	if x != nil {
		return x.TreeId
	}
	return ""
}

func (x *UploadTreeAvatarRequest) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

func (x *UploadTreeAvatarRequest) GetMimeType() string {
	if x != nil {
		return x.MimeType
	}
	return ""
}

func (x *UploadTreeAvatarRequest) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

type UploadTreeAvatarResponse struct {
	state         protoimpl.MessageState
	Photo         *Photo `protobuf:"bytes,1,opt,name=photo,proto3" json:"photo,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UploadTreeAvatarResponse) Reset() {
	*x = UploadTreeAvatarResponse{}
	mi := &file_photos_photos_tree_avatar_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UploadTreeAvatarResponse) String() string { return protoimpl.X.MessageStringOf(x) }
func (*UploadTreeAvatarResponse) ProtoMessage()    {}

func (x *UploadTreeAvatarResponse) ProtoReflect() protoreflect.Message {
	mi := &file_photos_photos_tree_avatar_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*UploadTreeAvatarResponse) Descriptor() ([]byte, []int) {
	return nil, []int{1}
}

func (x *UploadTreeAvatarResponse) GetPhoto() *Photo {
	if x != nil {
		return x.Photo
	}
	return nil
}

type GetTreeAvatarRequest struct {
	state         protoimpl.MessageState
	TreeId        string `protobuf:"bytes,1,opt,name=tree_id,json=treeId,proto3" json:"tree_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetTreeAvatarRequest) Reset() {
	*x = GetTreeAvatarRequest{}
	mi := &file_photos_photos_tree_avatar_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetTreeAvatarRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*GetTreeAvatarRequest) ProtoMessage()    {}

func (x *GetTreeAvatarRequest) ProtoReflect() protoreflect.Message {
	mi := &file_photos_photos_tree_avatar_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetTreeAvatarRequest) Descriptor() ([]byte, []int) {
	return nil, []int{2}
}

func (x *GetTreeAvatarRequest) GetTreeId() string {
	if x != nil {
		return x.TreeId
	}
	return ""
}

var File_photos_photos_tree_avatar_proto protoreflect.FileDescriptor

var file_photos_photos_tree_avatar_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_photos_photos_tree_avatar_proto_goTypes = []any{
	(*UploadTreeAvatarRequest)(nil),  // 0: photos.UploadTreeAvatarRequest
	(*UploadTreeAvatarResponse)(nil), // 1: photos.UploadTreeAvatarResponse
	(*GetTreeAvatarRequest)(nil),     // 2: photos.GetTreeAvatarRequest
	(*Photo)(nil),                    // 3: photos.Photo
}
var file_photos_photos_tree_avatar_proto_depIdxs = []int32{
	3, // 0: photos.UploadTreeAvatarResponse.photo:type_name -> photos.Photo
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_photos_photos_tree_avatar_proto_init() }

func file_photos_photos_tree_avatar_proto_init() {
	if File_photos_photos_tree_avatar_proto != nil {
		return
	}

	fdProto := &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Name:    proto.String("photos/photos_tree_avatar.proto"),
		Package: proto.String("photos"),
		Dependency: []string{
			"photos/photos.proto",
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("UploadTreeAvatarRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("tree_id"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("treeId")},
					{Name: proto.String("file_name"), Number: proto.Int32(2), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("fileName")},
					{Name: proto.String("mime_type"), Number: proto.Int32(3), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("mimeType")},
					{Name: proto.String("content"), Number: proto.Int32(4), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum(), JsonName: proto.String("content")},
				},
			},
			{
				Name: proto.String("UploadTreeAvatarResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("photo"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".photos.Photo"), JsonName: proto.String("photo")},
				},
			},
			{
				Name: proto.String("GetTreeAvatarRequest"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("tree_id"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(), JsonName: proto.String("treeId")},
				},
			},
		},
	}

	rawDesc, err := proto.Marshal(fdProto)
	if err != nil {
		panic(err)
	}

	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_photos_photos_tree_avatar_proto_goTypes,
		DependencyIndexes: file_photos_photos_tree_avatar_proto_depIdxs,
		MessageInfos:      file_photos_photos_tree_avatar_proto_msgTypes,
	}.Build()

	File_photos_photos_tree_avatar_proto = out.File
	file_photos_photos_tree_avatar_proto_goTypes = nil
	file_photos_photos_tree_avatar_proto_depIdxs = nil
}
