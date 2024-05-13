// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: proto/shortener.proto

package proto

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ShortenerService_SetURL_FullMethodName         = "/proto.ShortenerService/SetURL"
	ShortenerService_SetURLsBatch_FullMethodName   = "/proto.ShortenerService/SetURLsBatch"
	ShortenerService_GetURL_FullMethodName         = "/proto.ShortenerService/GetURL"
	ShortenerService_GetUserURLs_FullMethodName    = "/proto.ShortenerService/GetUserURLs"
	ShortenerService_DeleteUserURLs_FullMethodName = "/proto.ShortenerService/DeleteUserURLs"
)

// ShortenerServiceClient is the client API for ShortenerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ShortenerServiceClient interface {
	SetURL(ctx context.Context, in *SetURLRequest, opts ...grpc.CallOption) (*SetURLResponse, error)
	SetURLsBatch(ctx context.Context, in *SetURLsBatchRequest, opts ...grpc.CallOption) (*SetURLsBatchResponse, error)
	GetURL(ctx context.Context, in *GetURLRequest, opts ...grpc.CallOption) (*GetURLResponse, error)
	GetUserURLs(ctx context.Context, in *GetUserURLsRequest, opts ...grpc.CallOption) (*GetUserURLsResponse, error)
	DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*DeleteUserURLsResponse, error)
}

type shortenerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewShortenerServiceClient(cc grpc.ClientConnInterface) ShortenerServiceClient {
	return &shortenerServiceClient{cc}
}

func (c *shortenerServiceClient) SetURL(ctx context.Context, in *SetURLRequest, opts ...grpc.CallOption) (*SetURLResponse, error) {
	out := new(SetURLResponse)
	err := c.cc.Invoke(ctx, ShortenerService_SetURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) SetURLsBatch(ctx context.Context, in *SetURLsBatchRequest, opts ...grpc.CallOption) (*SetURLsBatchResponse, error) {
	out := new(SetURLsBatchResponse)
	err := c.cc.Invoke(ctx, ShortenerService_SetURLsBatch_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) GetURL(ctx context.Context, in *GetURLRequest, opts ...grpc.CallOption) (*GetURLResponse, error) {
	out := new(GetURLResponse)
	err := c.cc.Invoke(ctx, ShortenerService_GetURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) GetUserURLs(ctx context.Context, in *GetUserURLsRequest, opts ...grpc.CallOption) (*GetUserURLsResponse, error) {
	out := new(GetUserURLsResponse)
	err := c.cc.Invoke(ctx, ShortenerService_GetUserURLs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) DeleteUserURLs(ctx context.Context, in *DeleteUserURLsRequest, opts ...grpc.CallOption) (*DeleteUserURLsResponse, error) {
	out := new(DeleteUserURLsResponse)
	err := c.cc.Invoke(ctx, ShortenerService_DeleteUserURLs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ShortenerServiceServer is the server API for ShortenerService service.
// All implementations must embed UnimplementedShortenerServiceServer
// for forward compatibility
type ShortenerServiceServer interface {
	SetURL(context.Context, *SetURLRequest) (*SetURLResponse, error)
	SetURLsBatch(context.Context, *SetURLsBatchRequest) (*SetURLsBatchResponse, error)
	GetURL(context.Context, *GetURLRequest) (*GetURLResponse, error)
	GetUserURLs(context.Context, *GetUserURLsRequest) (*GetUserURLsResponse, error)
	DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*DeleteUserURLsResponse, error)
	mustEmbedUnimplementedShortenerServiceServer()
}

// UnimplementedShortenerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedShortenerServiceServer struct {
}

func (UnimplementedShortenerServiceServer) SetURL(context.Context, *SetURLRequest) (*SetURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetURL not implemented")
}
func (UnimplementedShortenerServiceServer) SetURLsBatch(context.Context, *SetURLsBatchRequest) (*SetURLsBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetURLsBatch not implemented")
}
func (UnimplementedShortenerServiceServer) GetURL(context.Context, *GetURLRequest) (*GetURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetURL not implemented")
}
func (UnimplementedShortenerServiceServer) GetUserURLs(context.Context, *GetUserURLsRequest) (*GetUserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserURLs not implemented")
}
func (UnimplementedShortenerServiceServer) DeleteUserURLs(context.Context, *DeleteUserURLsRequest) (*DeleteUserURLsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUserURLs not implemented")
}
func (UnimplementedShortenerServiceServer) mustEmbedUnimplementedShortenerServiceServer() {}

// UnsafeShortenerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ShortenerServiceServer will
// result in compilation errors.
type UnsafeShortenerServiceServer interface {
	mustEmbedUnimplementedShortenerServiceServer()
}

func RegisterShortenerServiceServer(s grpc.ServiceRegistrar, srv ShortenerServiceServer) {
	s.RegisterService(&ShortenerService_ServiceDesc, srv)
}

func _ShortenerService_SetURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).SetURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_SetURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).SetURL(ctx, req.(*SetURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_SetURLsBatch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetURLsBatchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).SetURLsBatch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_SetURLsBatch_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).SetURLsBatch(ctx, req.(*SetURLsBatchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_GetURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).GetURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_GetURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).GetURL(ctx, req.(*GetURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_GetUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).GetUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_GetUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).GetUserURLs(ctx, req.(*GetUserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_DeleteUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUserURLsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).DeleteUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_DeleteUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).DeleteUserURLs(ctx, req.(*DeleteUserURLsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ShortenerService_ServiceDesc is the grpc.ServiceDesc for ShortenerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ShortenerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ShortenerService",
	HandlerType: (*ShortenerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetURL",
			Handler:    _ShortenerService_SetURL_Handler,
		},
		{
			MethodName: "SetURLsBatch",
			Handler:    _ShortenerService_SetURLsBatch_Handler,
		},
		{
			MethodName: "GetURL",
			Handler:    _ShortenerService_GetURL_Handler,
		},
		{
			MethodName: "GetUserURLs",
			Handler:    _ShortenerService_GetUserURLs_Handler,
		},
		{
			MethodName: "DeleteUserURLs",
			Handler:    _ShortenerService_DeleteUserURLs_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/shortener.proto",
}
