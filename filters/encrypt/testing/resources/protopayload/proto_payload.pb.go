// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.3
// 	protoc        v5.29.3
// source: proto_payload.proto

package protopayload

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type WithTaggable struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// @inject_tag: `class:"public"`
	PublicString string `protobuf:"bytes,10,opt,name=public_string,json=publicString,proto3" json:"public_string,omitempty" class:"public"`
	// @inject_tag: `class:"sensitive"`
	SensitiveString string `protobuf:"bytes,20,opt,name=sensitive_string,json=sensitiveString,proto3" json:"sensitive_string,omitempty" class:"sensitive"`
	// @inject_tag: `class:"secret"`
	SecretString string `protobuf:"bytes,30,opt,name=secret_string,json=secretString,proto3" json:"secret_string,omitempty" class:"secret"`
	// intentionally unclassified
	UnclassifiedString string `protobuf:"bytes,40,opt,name=unclassified_string,json=unclassifiedString,proto3" json:"unclassified_string,omitempty"`
	// @inject_tag: `class:"public"`
	PublicBytes []byte `protobuf:"bytes,50,opt,name=public_bytes,json=publicBytes,proto3" json:"public_bytes,omitempty" class:"public"`
	// @inject_tag: `class:"sensitive"`
	SensitiveBytes []byte `protobuf:"bytes,60,opt,name=sensitive_bytes,json=sensitiveBytes,proto3" json:"sensitive_bytes,omitempty" class:"sensitive"`
	// @inject_tag: `class:"secret"`
	SecretBytes []byte `protobuf:"bytes,70,opt,name=secret_bytes,json=secretBytes,proto3" json:"secret_bytes,omitempty" class:"secret"`
	// intentionally unclassified
	UnclassifiedBytes []byte `protobuf:"bytes,80,opt,name=unclassified_bytes,json=unclassifiedBytes,proto3" json:"unclassified_bytes,omitempty"`
	// @inject_tag: `class:"public"`
	PublicStringValue *wrapperspb.StringValue `protobuf:"bytes,90,opt,name=public_string_value,json=publicStringValue,proto3" json:"public_string_value,omitempty" class:"public"`
	// @inject_tag: `class:"sensitive"`
	SensitiveStringValue *wrapperspb.StringValue `protobuf:"bytes,100,opt,name=sensitive_string_value,json=sensitiveStringValue,proto3" json:"sensitive_string_value,omitempty" class:"sensitive"`
	// @inject_tag: `class:"secret"`
	SecretStringValue *wrapperspb.StringValue `protobuf:"bytes,110,opt,name=secret_string_value,json=secretStringValue,proto3" json:"secret_string_value,omitempty" class:"secret"`
	// intentionally unclassified
	UnclassifiedStringValue *wrapperspb.StringValue `protobuf:"bytes,120,opt,name=unclassified_string_value,json=unclassifiedStringValue,proto3" json:"unclassified_string_value,omitempty"`
	// @inject_tag: `class:"public"`
	PublicBytesValue *wrapperspb.BytesValue `protobuf:"bytes,130,opt,name=public_bytes_value,json=publicBytesValue,proto3" json:"public_bytes_value,omitempty" class:"public"`
	// @inject_tag: `class:"sensitive"`
	SensitiveBytesValue *wrapperspb.BytesValue `protobuf:"bytes,140,opt,name=sensitive_bytes_value,json=sensitiveBytesValue,proto3" json:"sensitive_bytes_value,omitempty" class:"sensitive"`
	// @inject_tag: `class:"secret"`
	SecretBytesValue *wrapperspb.BytesValue `protobuf:"bytes,150,opt,name=secret_bytes_value,json=secretBytesValue,proto3" json:"secret_bytes_value,omitempty" class:"secret"`
	// intentionally unclassified
	UnclassifiedBytesValue *wrapperspb.BytesValue `protobuf:"bytes,160,opt,name=unclassified_bytes_value,json=unclassifiedBytesValue,proto3" json:"unclassified_bytes_value,omitempty"`
	// will always be unset and unclassified
	UnsetStringValue *wrapperspb.StringValue `protobuf:"bytes,170,opt,name=unset_string_value,json=unsetStringValue,proto3" json:"unset_string_value,omitempty"`
	// attributes will need to be redacted via Taggable
	TaggableAttributes    *structpb.Struct       `protobuf:"bytes,180,opt,name=taggable_attributes,json=taggableAttributes,proto3" json:"taggable_attributes,omitempty"`
	NontaggableAttributes *structpb.Struct       `protobuf:"bytes,190,opt,name=nontaggable_attributes,json=nontaggableAttributes,proto3" json:"nontaggable_attributes,omitempty"`
	EmbeddedTaggable      *EmbeddedTaggable      `protobuf:"bytes,200,opt,name=embedded_taggable,json=embeddedTaggable,proto3" json:"embedded_taggable,omitempty"`
	CreateTime            *timestamppb.Timestamp `protobuf:"bytes,210,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	unknownFields         protoimpl.UnknownFields
	sizeCache             protoimpl.SizeCache
}

func (x *WithTaggable) Reset() {
	*x = WithTaggable{}
	mi := &file_proto_payload_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WithTaggable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WithTaggable) ProtoMessage() {}

func (x *WithTaggable) ProtoReflect() protoreflect.Message {
	mi := &file_proto_payload_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WithTaggable.ProtoReflect.Descriptor instead.
func (*WithTaggable) Descriptor() ([]byte, []int) {
	return file_proto_payload_proto_rawDescGZIP(), []int{0}
}

func (x *WithTaggable) GetPublicString() string {
	if x != nil {
		return x.PublicString
	}
	return ""
}

func (x *WithTaggable) GetSensitiveString() string {
	if x != nil {
		return x.SensitiveString
	}
	return ""
}

func (x *WithTaggable) GetSecretString() string {
	if x != nil {
		return x.SecretString
	}
	return ""
}

func (x *WithTaggable) GetUnclassifiedString() string {
	if x != nil {
		return x.UnclassifiedString
	}
	return ""
}

func (x *WithTaggable) GetPublicBytes() []byte {
	if x != nil {
		return x.PublicBytes
	}
	return nil
}

func (x *WithTaggable) GetSensitiveBytes() []byte {
	if x != nil {
		return x.SensitiveBytes
	}
	return nil
}

func (x *WithTaggable) GetSecretBytes() []byte {
	if x != nil {
		return x.SecretBytes
	}
	return nil
}

func (x *WithTaggable) GetUnclassifiedBytes() []byte {
	if x != nil {
		return x.UnclassifiedBytes
	}
	return nil
}

func (x *WithTaggable) GetPublicStringValue() *wrapperspb.StringValue {
	if x != nil {
		return x.PublicStringValue
	}
	return nil
}

func (x *WithTaggable) GetSensitiveStringValue() *wrapperspb.StringValue {
	if x != nil {
		return x.SensitiveStringValue
	}
	return nil
}

func (x *WithTaggable) GetSecretStringValue() *wrapperspb.StringValue {
	if x != nil {
		return x.SecretStringValue
	}
	return nil
}

func (x *WithTaggable) GetUnclassifiedStringValue() *wrapperspb.StringValue {
	if x != nil {
		return x.UnclassifiedStringValue
	}
	return nil
}

func (x *WithTaggable) GetPublicBytesValue() *wrapperspb.BytesValue {
	if x != nil {
		return x.PublicBytesValue
	}
	return nil
}

func (x *WithTaggable) GetSensitiveBytesValue() *wrapperspb.BytesValue {
	if x != nil {
		return x.SensitiveBytesValue
	}
	return nil
}

func (x *WithTaggable) GetSecretBytesValue() *wrapperspb.BytesValue {
	if x != nil {
		return x.SecretBytesValue
	}
	return nil
}

func (x *WithTaggable) GetUnclassifiedBytesValue() *wrapperspb.BytesValue {
	if x != nil {
		return x.UnclassifiedBytesValue
	}
	return nil
}

func (x *WithTaggable) GetUnsetStringValue() *wrapperspb.StringValue {
	if x != nil {
		return x.UnsetStringValue
	}
	return nil
}

func (x *WithTaggable) GetTaggableAttributes() *structpb.Struct {
	if x != nil {
		return x.TaggableAttributes
	}
	return nil
}

func (x *WithTaggable) GetNontaggableAttributes() *structpb.Struct {
	if x != nil {
		return x.NontaggableAttributes
	}
	return nil
}

func (x *WithTaggable) GetEmbeddedTaggable() *EmbeddedTaggable {
	if x != nil {
		return x.EmbeddedTaggable
	}
	return nil
}

func (x *WithTaggable) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

type EmbeddedTaggable struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// @inject_tag: `class:"public"`
	EPublicString string `protobuf:"bytes,10,opt,name=e_public_string,json=ePublicString,proto3" json:"e_public_string,omitempty" class:"public"`
	// @inject_tag: `class:"secret"`
	ESecretString       string           `protobuf:"bytes,30,opt,name=e_secret_string,json=eSecretString,proto3" json:"e_secret_string,omitempty" class:"secret"`
	ETaggableAttributes *structpb.Struct `protobuf:"bytes,180,opt,name=e_taggable_attributes,json=eTaggableAttributes,proto3" json:"e_taggable_attributes,omitempty"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *EmbeddedTaggable) Reset() {
	*x = EmbeddedTaggable{}
	mi := &file_proto_payload_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EmbeddedTaggable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EmbeddedTaggable) ProtoMessage() {}

func (x *EmbeddedTaggable) ProtoReflect() protoreflect.Message {
	mi := &file_proto_payload_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EmbeddedTaggable.ProtoReflect.Descriptor instead.
func (*EmbeddedTaggable) Descriptor() ([]byte, []int) {
	return file_proto_payload_proto_rawDescGZIP(), []int{1}
}

func (x *EmbeddedTaggable) GetEPublicString() string {
	if x != nil {
		return x.EPublicString
	}
	return ""
}

func (x *EmbeddedTaggable) GetESecretString() string {
	if x != nil {
		return x.ESecretString
	}
	return ""
}

func (x *EmbeddedTaggable) GetETaggableAttributes() *structpb.Struct {
	if x != nil {
		return x.ETaggableAttributes
	}
	return nil
}

var File_proto_payload_proto protoreflect.FileDescriptor

var file_proto_payload_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x5f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x1a, 0x1c, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72,
	0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xdd, 0x0a, 0x0a,
	0x0c, 0x57, 0x69, 0x74, 0x68, 0x54, 0x61, 0x67, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x23, 0x0a,
	0x0d, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x0a,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x12, 0x29, 0x0a, 0x10, 0x73, 0x65, 0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x5f,
	0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x14, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x73, 0x65,
	0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x23, 0x0a,
	0x0d, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x1e,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x12, 0x2f, 0x0a, 0x13, 0x75, 0x6e, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69,
	0x65, 0x64, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x28, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x12, 0x75, 0x6e, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x65, 0x64, 0x53, 0x74, 0x72,
	0x69, 0x6e, 0x67, 0x12, 0x21, 0x0a, 0x0c, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x62, 0x79,
	0x74, 0x65, 0x73, 0x18, 0x32, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x70, 0x75, 0x62, 0x6c, 0x69,
	0x63, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x27, 0x0a, 0x0f, 0x73, 0x65, 0x6e, 0x73, 0x69, 0x74,
	0x69, 0x76, 0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x3c, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x0e, 0x73, 0x65, 0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12,
	0x21, 0x0a, 0x0c, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18,
	0x46, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x42, 0x79, 0x74,
	0x65, 0x73, 0x12, 0x2d, 0x0a, 0x12, 0x75, 0x6e, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69,
	0x65, 0x64, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x18, 0x50, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x11,
	0x75, 0x6e, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x65, 0x64, 0x42, 0x79, 0x74, 0x65,
	0x73, 0x12, 0x4c, 0x0a, 0x13, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x73, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x5a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x11, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12,
	0x52, 0x0a, 0x16, 0x73, 0x65, 0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x73, 0x74, 0x72,
	0x69, 0x6e, 0x67, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x64, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x14, 0x73,
	0x65, 0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x12, 0x4c, 0x0a, 0x13, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x73, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x6e, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x11,
	0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x12, 0x58, 0x0a, 0x19, 0x75, 0x6e, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x65,
	0x64, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x78,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x52, 0x17, 0x75, 0x6e, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66, 0x69, 0x65, 0x64,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x4a, 0x0a, 0x12, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x82, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x73,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x10, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x42, 0x79, 0x74,
	0x65, 0x73, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x50, 0x0a, 0x15, 0x73, 0x65, 0x6e, 0x73, 0x69,
	0x74, 0x69, 0x76, 0x65, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x8c, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x52, 0x13, 0x73, 0x65, 0x6e, 0x73, 0x69, 0x74, 0x69, 0x76, 0x65, 0x42,
	0x79, 0x74, 0x65, 0x73, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x4a, 0x0a, 0x12, 0x73, 0x65, 0x63,
	0x72, 0x65, 0x74, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x96, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x52, 0x10, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x42, 0x79, 0x74, 0x65, 0x73,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x56, 0x0a, 0x18, 0x75, 0x6e, 0x63, 0x6c, 0x61, 0x73, 0x73,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0xa0, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x73,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x16, 0x75, 0x6e, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x69, 0x66,
	0x69, 0x65, 0x64, 0x42, 0x79, 0x74, 0x65, 0x73, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x4b, 0x0a,
	0x12, 0x75, 0x6e, 0x73, 0x65, 0x74, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0xaa, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72,
	0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x10, 0x75, 0x6e, 0x73, 0x65, 0x74, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x49, 0x0a, 0x13, 0x74, 0x61,
	0x67, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65,
	0x73, 0x18, 0xb4, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63,
	0x74, 0x52, 0x12, 0x74, 0x61, 0x67, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x41, 0x74, 0x74, 0x72, 0x69,
	0x62, 0x75, 0x74, 0x65, 0x73, 0x12, 0x4f, 0x0a, 0x16, 0x6e, 0x6f, 0x6e, 0x74, 0x61, 0x67, 0x67,
	0x61, 0x62, 0x6c, 0x65, 0x5f, 0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x18,
	0xbe, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52,
	0x15, 0x6e, 0x6f, 0x6e, 0x74, 0x61, 0x67, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x41, 0x74, 0x74, 0x72,
	0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x12, 0x56, 0x0a, 0x11, 0x65, 0x6d, 0x62, 0x65, 0x64, 0x64,
	0x65, 0x64, 0x5f, 0x74, 0x61, 0x67, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x18, 0xc8, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x28, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x2e, 0x45, 0x6d, 0x62, 0x65,
	0x64, 0x64, 0x65, 0x64, 0x54, 0x61, 0x67, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x52, 0x10, 0x65, 0x6d,
	0x62, 0x65, 0x64, 0x64, 0x65, 0x64, 0x54, 0x61, 0x67, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x3c,
	0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0xd2, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x22, 0xb0, 0x01, 0x0a,
	0x10, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x64, 0x65, 0x64, 0x54, 0x61, 0x67, 0x67, 0x61, 0x62, 0x6c,
	0x65, 0x12, 0x26, 0x0a, 0x0f, 0x65, 0x5f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x73, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x65, 0x50, 0x75, 0x62,
	0x6c, 0x69, 0x63, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x26, 0x0a, 0x0f, 0x65, 0x5f, 0x73,
	0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x1e, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0d, 0x65, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x53, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0x12, 0x4c, 0x0a, 0x15, 0x65, 0x5f, 0x74, 0x61, 0x67, 0x67, 0x61, 0x62, 0x6c, 0x65, 0x5f,
	0x61, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x18, 0xb4, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x13, 0x65, 0x54, 0x61, 0x67,
	0x67, 0x61, 0x62, 0x6c, 0x65, 0x41, 0x74, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x73, 0x42,
	0x5e, 0x5a, 0x5c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x68, 0x61,
	0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x6c, 0x6f, 0x67,
	0x67, 0x65, 0x72, 0x2f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x2f, 0x65, 0x6e, 0x63, 0x72,
	0x79, 0x70, 0x74, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x70, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_payload_proto_rawDescOnce sync.Once
	file_proto_payload_proto_rawDescData = file_proto_payload_proto_rawDesc
)

func file_proto_payload_proto_rawDescGZIP() []byte {
	file_proto_payload_proto_rawDescOnce.Do(func() {
		file_proto_payload_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_payload_proto_rawDescData)
	})
	return file_proto_payload_proto_rawDescData
}

var file_proto_payload_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_payload_proto_goTypes = []any{
	(*WithTaggable)(nil),           // 0: resources.protopayload.WithTaggable
	(*EmbeddedTaggable)(nil),       // 1: resources.protopayload.EmbeddedTaggable
	(*wrapperspb.StringValue)(nil), // 2: google.protobuf.StringValue
	(*wrapperspb.BytesValue)(nil),  // 3: google.protobuf.BytesValue
	(*structpb.Struct)(nil),        // 4: google.protobuf.Struct
	(*timestamppb.Timestamp)(nil),  // 5: google.protobuf.Timestamp
}
var file_proto_payload_proto_depIdxs = []int32{
	2,  // 0: resources.protopayload.WithTaggable.public_string_value:type_name -> google.protobuf.StringValue
	2,  // 1: resources.protopayload.WithTaggable.sensitive_string_value:type_name -> google.protobuf.StringValue
	2,  // 2: resources.protopayload.WithTaggable.secret_string_value:type_name -> google.protobuf.StringValue
	2,  // 3: resources.protopayload.WithTaggable.unclassified_string_value:type_name -> google.protobuf.StringValue
	3,  // 4: resources.protopayload.WithTaggable.public_bytes_value:type_name -> google.protobuf.BytesValue
	3,  // 5: resources.protopayload.WithTaggable.sensitive_bytes_value:type_name -> google.protobuf.BytesValue
	3,  // 6: resources.protopayload.WithTaggable.secret_bytes_value:type_name -> google.protobuf.BytesValue
	3,  // 7: resources.protopayload.WithTaggable.unclassified_bytes_value:type_name -> google.protobuf.BytesValue
	2,  // 8: resources.protopayload.WithTaggable.unset_string_value:type_name -> google.protobuf.StringValue
	4,  // 9: resources.protopayload.WithTaggable.taggable_attributes:type_name -> google.protobuf.Struct
	4,  // 10: resources.protopayload.WithTaggable.nontaggable_attributes:type_name -> google.protobuf.Struct
	1,  // 11: resources.protopayload.WithTaggable.embedded_taggable:type_name -> resources.protopayload.EmbeddedTaggable
	5,  // 12: resources.protopayload.WithTaggable.create_time:type_name -> google.protobuf.Timestamp
	4,  // 13: resources.protopayload.EmbeddedTaggable.e_taggable_attributes:type_name -> google.protobuf.Struct
	14, // [14:14] is the sub-list for method output_type
	14, // [14:14] is the sub-list for method input_type
	14, // [14:14] is the sub-list for extension type_name
	14, // [14:14] is the sub-list for extension extendee
	0,  // [0:14] is the sub-list for field type_name
}

func init() { file_proto_payload_proto_init() }
func file_proto_payload_proto_init() {
	if File_proto_payload_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_payload_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_payload_proto_goTypes,
		DependencyIndexes: file_proto_payload_proto_depIdxs,
		MessageInfos:      file_proto_payload_proto_msgTypes,
	}.Build()
	File_proto_payload_proto = out.File
	file_proto_payload_proto_rawDesc = nil
	file_proto_payload_proto_goTypes = nil
	file_proto_payload_proto_depIdxs = nil
}
