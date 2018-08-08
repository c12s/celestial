// Code generated by protoc-gen-go. DO NOT EDIT.
// source: celestial.proto

/*
Package service is a generated protocol buffer package.

It is generated from these files:
	celestial.proto

It has these top-level messages:
	KV
	Label
	Data
	ListReq
	MutateReq
	ListResp
	MutateResp
*/
package service

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Other messages
type KV struct {
	Key   string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
}

func (m *KV) Reset()                    { *m = KV{} }
func (m *KV) String() string            { return proto.CompactTextString(m) }
func (*KV) ProtoMessage()               {}
func (*KV) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *KV) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *KV) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type Label struct {
	Labels []*KV `protobuf:"bytes,1,rep,name=labels" json:"labels,omitempty"`
}

func (m *Label) Reset()                    { *m = Label{} }
func (m *Label) String() string            { return proto.CompactTextString(m) }
func (*Label) ProtoMessage()               {}
func (*Label) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Label) GetLabels() []*KV {
	if m != nil {
		return m.Labels
	}
	return nil
}

type Data struct {
	Data []*KV `protobuf:"bytes,1,rep,name=data" json:"data,omitempty"`
}

func (m *Data) Reset()                    { *m = Data{} }
func (m *Data) String() string            { return proto.CompactTextString(m) }
func (*Data) ProtoMessage()               {}
func (*Data) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Data) GetData() []*KV {
	if m != nil {
		return m.Data
	}
	return nil
}

// Requests
type ListReq struct {
	RegionId  string `protobuf:"bytes,1,opt,name=region_id,json=regionId" json:"region_id,omitempty"`
	ClusterId string `protobuf:"bytes,2,opt,name=cluster_id,json=clusterId" json:"cluster_id,omitempty"`
	Labels    *Label `protobuf:"bytes,3,opt,name=labels" json:"labels,omitempty"`
	Kind      string `protobuf:"bytes,4,opt,name=kind" json:"kind,omitempty"`
}

func (m *ListReq) Reset()                    { *m = ListReq{} }
func (m *ListReq) String() string            { return proto.CompactTextString(m) }
func (*ListReq) ProtoMessage()               {}
func (*ListReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ListReq) GetRegionId() string {
	if m != nil {
		return m.RegionId
	}
	return ""
}

func (m *ListReq) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *ListReq) GetLabels() *Label {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *ListReq) GetKind() string {
	if m != nil {
		return m.Kind
	}
	return ""
}

type MutateReq struct {
	RegionId  string `protobuf:"bytes,1,opt,name=region_id,json=regionId" json:"region_id,omitempty"`
	ClusterId string `protobuf:"bytes,2,opt,name=cluster_id,json=clusterId" json:"cluster_id,omitempty"`
	Label     *Label `protobuf:"bytes,3,opt,name=label" json:"label,omitempty"`
	Data      *Data  `protobuf:"bytes,4,opt,name=data" json:"data,omitempty"`
	Kind      string `protobuf:"bytes,5,opt,name=kind" json:"kind,omitempty"`
}

func (m *MutateReq) Reset()                    { *m = MutateReq{} }
func (m *MutateReq) String() string            { return proto.CompactTextString(m) }
func (*MutateReq) ProtoMessage()               {}
func (*MutateReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *MutateReq) GetRegionId() string {
	if m != nil {
		return m.RegionId
	}
	return ""
}

func (m *MutateReq) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *MutateReq) GetLabel() *Label {
	if m != nil {
		return m.Label
	}
	return nil
}

func (m *MutateReq) GetData() *Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *MutateReq) GetKind() string {
	if m != nil {
		return m.Kind
	}
	return ""
}

// Replys
type ListResp struct {
	Error string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
	Data  []*KV  `protobuf:"bytes,2,rep,name=data" json:"data,omitempty"`
}

func (m *ListResp) Reset()                    { *m = ListResp{} }
func (m *ListResp) String() string            { return proto.CompactTextString(m) }
func (*ListResp) ProtoMessage()               {}
func (*ListResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ListResp) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *ListResp) GetData() []*KV {
	if m != nil {
		return m.Data
	}
	return nil
}

type MutateResp struct {
	Error string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
}

func (m *MutateResp) Reset()                    { *m = MutateResp{} }
func (m *MutateResp) String() string            { return proto.CompactTextString(m) }
func (*MutateResp) ProtoMessage()               {}
func (*MutateResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *MutateResp) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterType((*KV)(nil), "service.KV")
	proto.RegisterType((*Label)(nil), "service.Label")
	proto.RegisterType((*Data)(nil), "service.Data")
	proto.RegisterType((*ListReq)(nil), "service.ListReq")
	proto.RegisterType((*MutateReq)(nil), "service.MutateReq")
	proto.RegisterType((*ListResp)(nil), "service.ListResp")
	proto.RegisterType((*MutateResp)(nil), "service.MutateResp")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Secrets service

type SecretsClient interface {
	List(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (*ListResp, error)
	Mutate(ctx context.Context, in *MutateReq, opts ...grpc.CallOption) (*MutateResp, error)
}

type secretsClient struct {
	cc *grpc.ClientConn
}

func NewSecretsClient(cc *grpc.ClientConn) SecretsClient {
	return &secretsClient{cc}
}

func (c *secretsClient) List(ctx context.Context, in *ListReq, opts ...grpc.CallOption) (*ListResp, error) {
	out := new(ListResp)
	err := grpc.Invoke(ctx, "/service.Secrets/List", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *secretsClient) Mutate(ctx context.Context, in *MutateReq, opts ...grpc.CallOption) (*MutateResp, error) {
	out := new(MutateResp)
	err := grpc.Invoke(ctx, "/service.Secrets/Mutate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Secrets service

type SecretsServer interface {
	List(context.Context, *ListReq) (*ListResp, error)
	Mutate(context.Context, *MutateReq) (*MutateResp, error)
}

func RegisterSecretsServer(s *grpc.Server, srv SecretsServer) {
	s.RegisterService(&_Secrets_serviceDesc, srv)
}

func _Secrets_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SecretsServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.Secrets/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SecretsServer).List(ctx, req.(*ListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Secrets_Mutate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MutateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SecretsServer).Mutate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/service.Secrets/Mutate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SecretsServer).Mutate(ctx, req.(*MutateReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Secrets_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service.Secrets",
	HandlerType: (*SecretsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _Secrets_List_Handler,
		},
		{
			MethodName: "Mutate",
			Handler:    _Secrets_Mutate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "celestial.proto",
}

func init() { proto.RegisterFile("celestial.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 337 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x92, 0xcf, 0x4e, 0xc2, 0x40,
	0x10, 0xc6, 0x2d, 0xb4, 0x40, 0x87, 0xa8, 0x38, 0x7a, 0x68, 0x30, 0x46, 0x5c, 0x8d, 0x72, 0x40,
	0x0e, 0xf0, 0x04, 0x26, 0x5e, 0x08, 0x78, 0xc1, 0x84, 0xab, 0x59, 0xda, 0x89, 0x69, 0xa8, 0x14,
	0x77, 0x17, 0x12, 0x8f, 0xbe, 0x8b, 0x0f, 0x6a, 0xf6, 0x0f, 0xab, 0x51, 0x39, 0x79, 0x9b, 0xfd,
	0xbe, 0xe9, 0xcc, 0x6f, 0xbf, 0x2d, 0x1c, 0xa6, 0x54, 0x90, 0x54, 0x39, 0x2f, 0xfa, 0x2b, 0x51,
	0xaa, 0x12, 0xeb, 0x92, 0xc4, 0x26, 0x4f, 0x89, 0xf5, 0xa0, 0x32, 0x9e, 0x61, 0x0b, 0xaa, 0x0b,
	0x7a, 0x4b, 0x82, 0x4e, 0xd0, 0x8d, 0xa7, 0xba, 0xc4, 0x13, 0x88, 0x36, 0xbc, 0x58, 0x53, 0x52,
	0x31, 0x9a, 0x3d, 0xb0, 0x1e, 0x44, 0x13, 0x3e, 0xa7, 0x02, 0x2f, 0xa1, 0x56, 0xe8, 0x42, 0x26,
	0x41, 0xa7, 0xda, 0x6d, 0x0e, 0x9a, 0x7d, 0x37, 0xb0, 0x3f, 0x9e, 0x4d, 0x9d, 0xc5, 0x6e, 0x20,
	0xbc, 0xe7, 0x8a, 0xe3, 0x39, 0x84, 0x19, 0x57, 0xfc, 0xaf, 0x56, 0x63, 0xb0, 0xf7, 0x00, 0xea,
	0x93, 0x5c, 0xaa, 0x29, 0xbd, 0xe2, 0x29, 0xc4, 0x82, 0x9e, 0xf3, 0x72, 0xf9, 0x94, 0x67, 0x0e,
	0xa8, 0x61, 0x85, 0x51, 0x86, 0x67, 0x00, 0x69, 0xb1, 0x96, 0x8a, 0x84, 0x76, 0x2d, 0x5a, 0xec,
	0x94, 0x51, 0x86, 0xd7, 0x9e, 0xaa, 0xda, 0x09, 0xba, 0xcd, 0xc1, 0x81, 0x5f, 0x65, 0xa8, 0xb7,
	0x60, 0x88, 0x10, 0x2e, 0xf2, 0x65, 0x96, 0x84, 0x66, 0x80, 0xa9, 0xd9, 0x47, 0x00, 0xf1, 0xc3,
	0x5a, 0x71, 0x45, 0xff, 0xa5, 0xb8, 0x82, 0xc8, 0xec, 0xd9, 0x01, 0x61, 0x4d, 0xbc, 0x70, 0xa1,
	0x84, 0xa6, 0x69, 0xdf, 0x37, 0xe9, 0xc4, 0x6c, 0x2c, 0x1e, 0x33, 0xfa, 0x86, 0x79, 0x07, 0x0d,
	0x9b, 0x94, 0x5c, 0xe9, 0x37, 0x22, 0x21, 0x4a, 0xe1, 0x00, 0xed, 0xc1, 0xa7, 0x5d, 0xd9, 0x95,
	0x36, 0x03, 0xd8, 0x5e, 0x74, 0xd7, 0x90, 0xc1, 0x0b, 0xd4, 0x1f, 0x29, 0x15, 0xa4, 0x24, 0xde,
	0x42, 0xa8, 0x37, 0x62, 0xeb, 0xeb, 0x1e, 0xf6, 0xa9, 0xda, 0x47, 0x3f, 0x14, 0xb9, 0x62, 0x7b,
	0x38, 0x84, 0x9a, 0x9d, 0x8e, 0xe8, 0x6d, 0x9f, 0x6b, 0xfb, 0xf8, 0x97, 0xa6, 0x3f, 0x9a, 0xd7,
	0xcc, 0x5f, 0x39, 0xfc, 0x0c, 0x00, 0x00, 0xff, 0xff, 0xec, 0x60, 0x86, 0xb1, 0xa8, 0x02, 0x00,
	0x00,
}