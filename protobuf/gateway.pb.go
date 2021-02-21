// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: protobuf/gateway.proto

package protobuf

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// SendEventRequest contains the OPCode, details about identifying the
// specific shard and the data that will be sent to the gateway.
type SendEventRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GatewayOPCode uint32 `protobuf:"varint,1,opt,name=GatewayOPCode,proto3" json:"GatewayOPCode,omitempty"`
	Manager       string `protobuf:"bytes,2,opt,name=Manager,proto3" json:"Manager,omitempty"`
	ShardGroup    int32  `protobuf:"varint,3,opt,name=ShardGroup,proto3" json:"ShardGroup,omitempty"`
	ShardID       int32  `protobuf:"varint,4,opt,name=ShardID,proto3" json:"ShardID,omitempty"`
	Data          []byte `protobuf:"bytes,5,opt,name=Data,proto3" json:"Data,omitempty"`
}

func (x *SendEventRequest) Reset() {
	*x = SendEventRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_gateway_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendEventRequest) ProtoMessage() {}

func (x *SendEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_gateway_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendEventRequest.ProtoReflect.Descriptor instead.
func (*SendEventRequest) Descriptor() ([]byte, []int) {
	return file_protobuf_gateway_proto_rawDescGZIP(), []int{0}
}

func (x *SendEventRequest) GetGatewayOPCode() uint32 {
	if x != nil {
		return x.GatewayOPCode
	}
	return 0
}

func (x *SendEventRequest) GetManager() string {
	if x != nil {
		return x.Manager
	}
	return ""
}

func (x *SendEventRequest) GetShardGroup() int32 {
	if x != nil {
		return x.ShardGroup
	}
	return 0
}

func (x *SendEventRequest) GetShardID() int32 {
	if x != nil {
		return x.ShardID
	}
	return 0
}

func (x *SendEventRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

// SendEventResponse replies with a boolean if the Shard could be found
// and an error is encountered during SendEvent(op, data).
type SendEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FoundShard bool   `protobuf:"varint,1,opt,name=FoundShard,proto3" json:"FoundShard,omitempty"`
	Error      string `protobuf:"bytes,2,opt,name=Error,proto3" json:"Error,omitempty"`
}

func (x *SendEventResponse) Reset() {
	*x = SendEventResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_gateway_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendEventResponse) ProtoMessage() {}

func (x *SendEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_gateway_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendEventResponse.ProtoReflect.Descriptor instead.
func (*SendEventResponse) Descriptor() ([]byte, []int) {
	return file_protobuf_gateway_proto_rawDescGZIP(), []int{1}
}

func (x *SendEventResponse) GetFoundShard() bool {
	if x != nil {
		return x.FoundShard
	}
	return false
}

func (x *SendEventResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_protobuf_gateway_proto protoreflect.FileDescriptor

var file_protobuf_gateway_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x67, 0x61, 0x74, 0x65, 0x77,
	0x61, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61,
	0x79, 0x22, 0xa0, 0x01, 0x0a, 0x10, 0x53, 0x65, 0x6e, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61,
	0x79, 0x4f, 0x50, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0d, 0x47,
	0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x4f, 0x50, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x4d,
	0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x53, 0x68, 0x61, 0x72, 0x64, 0x47,
	0x72, 0x6f, 0x75, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x53, 0x68, 0x61, 0x72,
	0x64, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x53, 0x68, 0x61, 0x72, 0x64, 0x49,
	0x44, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x53, 0x68, 0x61, 0x72, 0x64, 0x49, 0x44,
	0x12, 0x12, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x44, 0x61, 0x74, 0x61, 0x22, 0x49, 0x0a, 0x11, 0x53, 0x65, 0x6e, 0x64, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x46, 0x6f, 0x75,
	0x6e, 0x64, 0x53, 0x68, 0x61, 0x72, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x46,
	0x6f, 0x75, 0x6e, 0x64, 0x53, 0x68, 0x61, 0x72, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x32,
	0x58, 0x0a, 0x07, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x12, 0x4d, 0x0a, 0x12, 0x53, 0x65,
	0x6e, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x6f, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x12, 0x19, 0x2e, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x67, 0x61,
	0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x32, 0x5a, 0x30, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x68, 0x65, 0x52, 0x6f, 0x63, 0x6b, 0x65,
	0x74, 0x74, 0x65, 0x6b, 0x2f, 0x53, 0x61, 0x6e, 0x64, 0x77, 0x69, 0x63, 0x68, 0x2d, 0x44, 0x61,
	0x65, 0x6d, 0x6f, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protobuf_gateway_proto_rawDescOnce sync.Once
	file_protobuf_gateway_proto_rawDescData = file_protobuf_gateway_proto_rawDesc
)

func file_protobuf_gateway_proto_rawDescGZIP() []byte {
	file_protobuf_gateway_proto_rawDescOnce.Do(func() {
		file_protobuf_gateway_proto_rawDescData = protoimpl.X.CompressGZIP(file_protobuf_gateway_proto_rawDescData)
	})
	return file_protobuf_gateway_proto_rawDescData
}

var file_protobuf_gateway_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protobuf_gateway_proto_goTypes = []interface{}{
	(*SendEventRequest)(nil),  // 0: gateway.SendEventRequest
	(*SendEventResponse)(nil), // 1: gateway.SendEventResponse
}
var file_protobuf_gateway_proto_depIdxs = []int32{
	0, // 0: gateway.Gateway.SendEventToGateway:input_type -> gateway.SendEventRequest
	1, // 1: gateway.Gateway.SendEventToGateway:output_type -> gateway.SendEventResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protobuf_gateway_proto_init() }
func file_protobuf_gateway_proto_init() {
	if File_protobuf_gateway_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protobuf_gateway_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendEventRequest); i {
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
		file_protobuf_gateway_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendEventResponse); i {
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
			RawDescriptor: file_protobuf_gateway_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_protobuf_gateway_proto_goTypes,
		DependencyIndexes: file_protobuf_gateway_proto_depIdxs,
		MessageInfos:      file_protobuf_gateway_proto_msgTypes,
	}.Build()
	File_protobuf_gateway_proto = out.File
	file_protobuf_gateway_proto_rawDesc = nil
	file_protobuf_gateway_proto_goTypes = nil
	file_protobuf_gateway_proto_depIdxs = nil
}
