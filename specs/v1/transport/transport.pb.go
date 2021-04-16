//
//Copyright 2021 The JamJar Relay Server Authors.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.6
// source: v1/transport/transport.proto

package transport

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

type Payload_FlagType int32

const (
	Payload_REQUEST_RELAY_MESSAGE        Payload_FlagType = 0
	Payload_REQUEST_CONNECT              Payload_FlagType = 1
	Payload_REQUEST_RECONNECT            Payload_FlagType = 2
	Payload_REQUEST_LIST                 Payload_FlagType = 3
	Payload_REQUEST_KICK                 Payload_FlagType = 4
	Payload_REQUEST_GRANT_HOST           Payload_FlagType = 5
	Payload_RESPONSE_RELAY_MESSAGE       Payload_FlagType = 6
	Payload_RESPONSE_CONNECT             Payload_FlagType = 7
	Payload_RESPONSE_ASSIGN_HOST         Payload_FlagType = 8
	Payload_RESPONSE_BEGIN_HOST_MIGRATE  Payload_FlagType = 9
	Payload_RESPONSE_FINISH_HOST_MIGRATE Payload_FlagType = 10
	Payload_RESPONSE_LIST                Payload_FlagType = 11
	Payload_RESPONSE_KICK                Payload_FlagType = 12
	Payload_RESPONSE_ERROR               Payload_FlagType = 13
	Payload_RESPONSE_CLIENT_CONNECT      Payload_FlagType = 14
)

// Enum value maps for Payload_FlagType.
var (
	Payload_FlagType_name = map[int32]string{
		0:  "REQUEST_RELAY_MESSAGE",
		1:  "REQUEST_CONNECT",
		2:  "REQUEST_RECONNECT",
		3:  "REQUEST_LIST",
		4:  "REQUEST_KICK",
		5:  "REQUEST_GRANT_HOST",
		6:  "RESPONSE_RELAY_MESSAGE",
		7:  "RESPONSE_CONNECT",
		8:  "RESPONSE_ASSIGN_HOST",
		9:  "RESPONSE_BEGIN_HOST_MIGRATE",
		10: "RESPONSE_FINISH_HOST_MIGRATE",
		11: "RESPONSE_LIST",
		12: "RESPONSE_KICK",
		13: "RESPONSE_ERROR",
		14: "RESPONSE_CLIENT_CONNECT",
	}
	Payload_FlagType_value = map[string]int32{
		"REQUEST_RELAY_MESSAGE":        0,
		"REQUEST_CONNECT":              1,
		"REQUEST_RECONNECT":            2,
		"REQUEST_LIST":                 3,
		"REQUEST_KICK":                 4,
		"REQUEST_GRANT_HOST":           5,
		"RESPONSE_RELAY_MESSAGE":       6,
		"RESPONSE_CONNECT":             7,
		"RESPONSE_ASSIGN_HOST":         8,
		"RESPONSE_BEGIN_HOST_MIGRATE":  9,
		"RESPONSE_FINISH_HOST_MIGRATE": 10,
		"RESPONSE_LIST":                11,
		"RESPONSE_KICK":                12,
		"RESPONSE_ERROR":               13,
		"RESPONSE_CLIENT_CONNECT":      14,
	}
)

func (x Payload_FlagType) Enum() *Payload_FlagType {
	p := new(Payload_FlagType)
	*p = x
	return p
}

func (x Payload_FlagType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Payload_FlagType) Descriptor() protoreflect.EnumDescriptor {
	return file_v1_transport_transport_proto_enumTypes[0].Descriptor()
}

func (Payload_FlagType) Type() protoreflect.EnumType {
	return &file_v1_transport_transport_proto_enumTypes[0]
}

func (x Payload_FlagType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Payload_FlagType.Descriptor instead.
func (Payload_FlagType) EnumDescriptor() ([]byte, []int) {
	return file_v1_transport_transport_proto_rawDescGZIP(), []int{0, 0}
}

type Payload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Flag Payload_FlagType `protobuf:"varint,1,opt,name=Flag,proto3,enum=v1_transport.Payload_FlagType" json:"Flag,omitempty"`
	Data []byte           `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
}

func (x *Payload) Reset() {
	*x = Payload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_transport_transport_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Payload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Payload) ProtoMessage() {}

func (x *Payload) ProtoReflect() protoreflect.Message {
	mi := &file_v1_transport_transport_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Payload.ProtoReflect.Descriptor instead.
func (*Payload) Descriptor() ([]byte, []int) {
	return file_v1_transport_transport_proto_rawDescGZIP(), []int{0}
}

func (x *Payload) GetFlag() Payload_FlagType {
	if x != nil {
		return x.Flag
	}
	return Payload_REQUEST_RELAY_MESSAGE
}

func (x *Payload) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code    int32  `protobuf:"varint,1,opt,name=Code,proto3" json:"Code,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_transport_transport_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_v1_transport_transport_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_v1_transport_transport_proto_rawDescGZIP(), []int{1}
}

func (x *Error) GetCode() int32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *Error) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_v1_transport_transport_proto protoreflect.FileDescriptor

var file_v1_transport_transport_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x76, 0x31, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c,
	0x76, 0x31, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x22, 0xc7, 0x03, 0x0a,
	0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x32, 0x0a, 0x04, 0x46, 0x6c, 0x61, 0x67,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x76, 0x31, 0x5f, 0x74, 0x72, 0x61, 0x6e,
	0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x46, 0x6c,
	0x61, 0x67, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x46, 0x6c, 0x61, 0x67, 0x12, 0x12, 0x0a, 0x04,
	0x44, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x44, 0x61, 0x74, 0x61,
	0x22, 0xf3, 0x02, 0x0a, 0x08, 0x46, 0x6c, 0x61, 0x67, 0x54, 0x79, 0x70, 0x65, 0x12, 0x19, 0x0a,
	0x15, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x5f, 0x52, 0x45, 0x4c, 0x41, 0x59, 0x5f, 0x4d,
	0x45, 0x53, 0x53, 0x41, 0x47, 0x45, 0x10, 0x00, 0x12, 0x13, 0x0a, 0x0f, 0x52, 0x45, 0x51, 0x55,
	0x45, 0x53, 0x54, 0x5f, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43, 0x54, 0x10, 0x01, 0x12, 0x15, 0x0a,
	0x11, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x5f, 0x52, 0x45, 0x43, 0x4f, 0x4e, 0x4e, 0x45,
	0x43, 0x54, 0x10, 0x02, 0x12, 0x10, 0x0a, 0x0c, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x5f,
	0x4c, 0x49, 0x53, 0x54, 0x10, 0x03, 0x12, 0x10, 0x0a, 0x0c, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53,
	0x54, 0x5f, 0x4b, 0x49, 0x43, 0x4b, 0x10, 0x04, 0x12, 0x16, 0x0a, 0x12, 0x52, 0x45, 0x51, 0x55,
	0x45, 0x53, 0x54, 0x5f, 0x47, 0x52, 0x41, 0x4e, 0x54, 0x5f, 0x48, 0x4f, 0x53, 0x54, 0x10, 0x05,
	0x12, 0x1a, 0x0a, 0x16, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x52, 0x45, 0x4c,
	0x41, 0x59, 0x5f, 0x4d, 0x45, 0x53, 0x53, 0x41, 0x47, 0x45, 0x10, 0x06, 0x12, 0x14, 0x0a, 0x10,
	0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x43, 0x4f, 0x4e, 0x4e, 0x45, 0x43, 0x54,
	0x10, 0x07, 0x12, 0x18, 0x0a, 0x14, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x41,
	0x53, 0x53, 0x49, 0x47, 0x4e, 0x5f, 0x48, 0x4f, 0x53, 0x54, 0x10, 0x08, 0x12, 0x1f, 0x0a, 0x1b,
	0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x42, 0x45, 0x47, 0x49, 0x4e, 0x5f, 0x48,
	0x4f, 0x53, 0x54, 0x5f, 0x4d, 0x49, 0x47, 0x52, 0x41, 0x54, 0x45, 0x10, 0x09, 0x12, 0x20, 0x0a,
	0x1c, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x46, 0x49, 0x4e, 0x49, 0x53, 0x48,
	0x5f, 0x48, 0x4f, 0x53, 0x54, 0x5f, 0x4d, 0x49, 0x47, 0x52, 0x41, 0x54, 0x45, 0x10, 0x0a, 0x12,
	0x11, 0x0a, 0x0d, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x4c, 0x49, 0x53, 0x54,
	0x10, 0x0b, 0x12, 0x11, 0x0a, 0x0d, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x4b,
	0x49, 0x43, 0x4b, 0x10, 0x0c, 0x12, 0x12, 0x0a, 0x0e, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53,
	0x45, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x10, 0x0d, 0x12, 0x1b, 0x0a, 0x17, 0x52, 0x45, 0x53,
	0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x43, 0x4c, 0x49, 0x45, 0x4e, 0x54, 0x5f, 0x43, 0x4f, 0x4e,
	0x4e, 0x45, 0x43, 0x54, 0x10, 0x0e, 0x22, 0x35, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12,
	0x12, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x43,
	0x6f, 0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x3e, 0x5a,
	0x3c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x61, 0x6d, 0x6a,
	0x61, 0x72, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x6a, 0x61, 0x6d, 0x6a, 0x61, 0x72, 0x2d, 0x72, 0x65,
	0x6c, 0x61, 0x79, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x73, 0x70, 0x65, 0x63, 0x73,
	0x2f, 0x76, 0x31, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_v1_transport_transport_proto_rawDescOnce sync.Once
	file_v1_transport_transport_proto_rawDescData = file_v1_transport_transport_proto_rawDesc
)

func file_v1_transport_transport_proto_rawDescGZIP() []byte {
	file_v1_transport_transport_proto_rawDescOnce.Do(func() {
		file_v1_transport_transport_proto_rawDescData = protoimpl.X.CompressGZIP(file_v1_transport_transport_proto_rawDescData)
	})
	return file_v1_transport_transport_proto_rawDescData
}

var file_v1_transport_transport_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_v1_transport_transport_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_v1_transport_transport_proto_goTypes = []interface{}{
	(Payload_FlagType)(0), // 0: v1_transport.Payload.FlagType
	(*Payload)(nil),       // 1: v1_transport.Payload
	(*Error)(nil),         // 2: v1_transport.Error
}
var file_v1_transport_transport_proto_depIdxs = []int32{
	0, // 0: v1_transport.Payload.Flag:type_name -> v1_transport.Payload.FlagType
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_v1_transport_transport_proto_init() }
func file_v1_transport_transport_proto_init() {
	if File_v1_transport_transport_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_v1_transport_transport_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Payload); i {
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
		file_v1_transport_transport_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Error); i {
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
			RawDescriptor: file_v1_transport_transport_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_v1_transport_transport_proto_goTypes,
		DependencyIndexes: file_v1_transport_transport_proto_depIdxs,
		EnumInfos:         file_v1_transport_transport_proto_enumTypes,
		MessageInfos:      file_v1_transport_transport_proto_msgTypes,
	}.Build()
	File_v1_transport_transport_proto = out.File
	file_v1_transport_transport_proto_rawDesc = nil
	file_v1_transport_transport_proto_goTypes = nil
	file_v1_transport_transport_proto_depIdxs = nil
}
