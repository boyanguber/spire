// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package hostservices

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// IdentityProviderClient is the client API for IdentityProvider service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IdentityProviderClient interface {
	FetchX509Identity(ctx context.Context, in *FetchX509IdentityRequest, opts ...grpc.CallOption) (*FetchX509IdentityResponse, error)
}

type identityProviderClient struct {
	cc grpc.ClientConnInterface
}

func NewIdentityProviderClient(cc grpc.ClientConnInterface) IdentityProviderClient {
	return &identityProviderClient{cc}
}

func (c *identityProviderClient) FetchX509Identity(ctx context.Context, in *FetchX509IdentityRequest, opts ...grpc.CallOption) (*FetchX509IdentityResponse, error) {
	out := new(FetchX509IdentityResponse)
	err := c.cc.Invoke(ctx, "/spire.server.hostservices.IdentityProvider/FetchX509Identity", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IdentityProviderServer is the server API for IdentityProvider service.
// All implementations must embed UnimplementedIdentityProviderServer
// for forward compatibility
type IdentityProviderServer interface {
	FetchX509Identity(context.Context, *FetchX509IdentityRequest) (*FetchX509IdentityResponse, error)
	mustEmbedUnimplementedIdentityProviderServer()
}

// UnimplementedIdentityProviderServer must be embedded to have forward compatible implementations.
type UnimplementedIdentityProviderServer struct {
}

func (UnimplementedIdentityProviderServer) FetchX509Identity(context.Context, *FetchX509IdentityRequest) (*FetchX509IdentityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchX509Identity not implemented")
}
func (UnimplementedIdentityProviderServer) mustEmbedUnimplementedIdentityProviderServer() {}

// UnsafeIdentityProviderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IdentityProviderServer will
// result in compilation errors.
type UnsafeIdentityProviderServer interface {
	mustEmbedUnimplementedIdentityProviderServer()
}

func RegisterIdentityProviderServer(s grpc.ServiceRegistrar, srv IdentityProviderServer) {
	s.RegisterService(&_IdentityProvider_serviceDesc, srv)
}

func _IdentityProvider_FetchX509Identity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchX509IdentityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IdentityProviderServer).FetchX509Identity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/spire.server.hostservices.IdentityProvider/FetchX509Identity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IdentityProviderServer).FetchX509Identity(ctx, req.(*FetchX509IdentityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _IdentityProvider_serviceDesc = grpc.ServiceDesc{
	ServiceName: "spire.server.hostservices.IdentityProvider",
	HandlerType: (*IdentityProviderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchX509Identity",
			Handler:    _IdentityProvider_FetchX509Identity_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "spire/server/hostservices/identityprovider.proto",
}