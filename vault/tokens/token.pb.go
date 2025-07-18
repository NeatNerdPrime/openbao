// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.31.1
// source: vault/tokens/token.proto

package tokens

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// SignedToken
type SignedToken struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TokenVersion  uint64                 `protobuf:"varint,1,opt,name=token_version,json=tokenVersion,proto3" json:"token_version,omitempty"` // always 1 for now
	Hmac          []byte                 `protobuf:"bytes,2,opt,name=hmac,proto3" json:"hmac,omitempty"`                                      // HMAC of token
	Token         []byte                 `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`                                    // protobuf-marshalled Token message
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SignedToken) Reset() {
	*x = SignedToken{}
	mi := &file_vault_tokens_token_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SignedToken) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedToken) ProtoMessage() {}

func (x *SignedToken) ProtoReflect() protoreflect.Message {
	mi := &file_vault_tokens_token_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedToken.ProtoReflect.Descriptor instead.
func (*SignedToken) Descriptor() ([]byte, []int) {
	return file_vault_tokens_token_proto_rawDescGZIP(), []int{0}
}

func (x *SignedToken) GetTokenVersion() uint64 {
	if x != nil {
		return x.TokenVersion
	}
	return 0
}

func (x *SignedToken) GetHmac() []byte {
	if x != nil {
		return x.Hmac
	}
	return nil
}

func (x *SignedToken) GetToken() []byte {
	if x != nil {
		return x.Token
	}
	return nil
}

type Token struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Random        string                 `protobuf:"bytes,1,opt,name=random,proto3" json:"random,omitempty"`                            // unencoded equiv of former $randbase62
	LocalIndex    uint64                 `protobuf:"varint,2,opt,name=local_index,json=localIndex,proto3" json:"local_index,omitempty"` // required storage state to have this token
	IndexEpoch    uint32                 `protobuf:"varint,3,opt,name=index_epoch,json=indexEpoch,proto3" json:"index_epoch,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Token) Reset() {
	*x = Token{}
	mi := &file_vault_tokens_token_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Token) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Token) ProtoMessage() {}

func (x *Token) ProtoReflect() protoreflect.Message {
	mi := &file_vault_tokens_token_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Token.ProtoReflect.Descriptor instead.
func (*Token) Descriptor() ([]byte, []int) {
	return file_vault_tokens_token_proto_rawDescGZIP(), []int{1}
}

func (x *Token) GetRandom() string {
	if x != nil {
		return x.Random
	}
	return ""
}

func (x *Token) GetLocalIndex() uint64 {
	if x != nil {
		return x.LocalIndex
	}
	return 0
}

func (x *Token) GetIndexEpoch() uint32 {
	if x != nil {
		return x.IndexEpoch
	}
	return 0
}

var File_vault_tokens_token_proto protoreflect.FileDescriptor

const file_vault_tokens_token_proto_rawDesc = "" +
	"\n" +
	"\x18vault/tokens/token.proto\x12\x06tokens\"\\\n" +
	"\vSignedToken\x12#\n" +
	"\rtoken_version\x18\x01 \x01(\x04R\ftokenVersion\x12\x12\n" +
	"\x04hmac\x18\x02 \x01(\fR\x04hmac\x12\x14\n" +
	"\x05token\x18\x03 \x01(\fR\x05token\"a\n" +
	"\x05Token\x12\x16\n" +
	"\x06random\x18\x01 \x01(\tR\x06random\x12\x1f\n" +
	"\vlocal_index\x18\x02 \x01(\x04R\n" +
	"localIndex\x12\x1f\n" +
	"\vindex_epoch\x18\x03 \x01(\rR\n" +
	"indexEpochB)Z'github.com/openbao/openbao/vault/tokensb\x06proto3"

var (
	file_vault_tokens_token_proto_rawDescOnce sync.Once
	file_vault_tokens_token_proto_rawDescData []byte
)

func file_vault_tokens_token_proto_rawDescGZIP() []byte {
	file_vault_tokens_token_proto_rawDescOnce.Do(func() {
		file_vault_tokens_token_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_vault_tokens_token_proto_rawDesc), len(file_vault_tokens_token_proto_rawDesc)))
	})
	return file_vault_tokens_token_proto_rawDescData
}

var file_vault_tokens_token_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_vault_tokens_token_proto_goTypes = []any{
	(*SignedToken)(nil), // 0: tokens.SignedToken
	(*Token)(nil),       // 1: tokens.Token
}
var file_vault_tokens_token_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_vault_tokens_token_proto_init() }
func file_vault_tokens_token_proto_init() {
	if File_vault_tokens_token_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_vault_tokens_token_proto_rawDesc), len(file_vault_tokens_token_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_vault_tokens_token_proto_goTypes,
		DependencyIndexes: file_vault_tokens_token_proto_depIdxs,
		MessageInfos:      file_vault_tokens_token_proto_msgTypes,
	}.Build()
	File_vault_tokens_token_proto = out.File
	file_vault_tokens_token_proto_goTypes = nil
	file_vault_tokens_token_proto_depIdxs = nil
}
