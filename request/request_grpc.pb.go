// Copyright 2021 The sacloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package request

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

// ScalingServiceClient is the client API for ScalingService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ScalingServiceClient interface {
	// Up スケールアップ or スケールアウトのリクエスト
	Up(ctx context.Context, in *ScalingRequest, opts ...grpc.CallOption) (*ScalingResponse, error)
	// Down スケールダウン or スケールインのリクエスト
	Down(ctx context.Context, in *ScalingRequest, opts ...grpc.CallOption) (*ScalingResponse, error)
}

type scalingServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewScalingServiceClient(cc grpc.ClientConnInterface) ScalingServiceClient {
	return &scalingServiceClient{cc}
}

func (c *scalingServiceClient) Up(ctx context.Context, in *ScalingRequest, opts ...grpc.CallOption) (*ScalingResponse, error) {
	out := new(ScalingResponse)
	err := c.cc.Invoke(ctx, "/autoscaler.ScalingService/Up", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *scalingServiceClient) Down(ctx context.Context, in *ScalingRequest, opts ...grpc.CallOption) (*ScalingResponse, error) {
	out := new(ScalingResponse)
	err := c.cc.Invoke(ctx, "/autoscaler.ScalingService/Down", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ScalingServiceServer is the server API for ScalingService service.
// All implementations must embed UnimplementedScalingServiceServer
// for forward compatibility
type ScalingServiceServer interface {
	// Up スケールアップ or スケールアウトのリクエスト
	Up(context.Context, *ScalingRequest) (*ScalingResponse, error)
	// Down スケールダウン or スケールインのリクエスト
	Down(context.Context, *ScalingRequest) (*ScalingResponse, error)
	mustEmbedUnimplementedScalingServiceServer()
}

// UnimplementedScalingServiceServer must be embedded to have forward compatible implementations.
type UnimplementedScalingServiceServer struct {
}

func (UnimplementedScalingServiceServer) Up(context.Context, *ScalingRequest) (*ScalingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Up not implemented")
}
func (UnimplementedScalingServiceServer) Down(context.Context, *ScalingRequest) (*ScalingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Down not implemented")
}
func (UnimplementedScalingServiceServer) mustEmbedUnimplementedScalingServiceServer() {}

// UnsafeScalingServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ScalingServiceServer will
// result in compilation errors.
type UnsafeScalingServiceServer interface {
	mustEmbedUnimplementedScalingServiceServer()
}

func RegisterScalingServiceServer(s grpc.ServiceRegistrar, srv ScalingServiceServer) {
	s.RegisterService(&ScalingService_ServiceDesc, srv)
}

func _ScalingService_Up_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ScalingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScalingServiceServer).Up(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/autoscaler.ScalingService/Up",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScalingServiceServer).Up(ctx, req.(*ScalingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ScalingService_Down_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ScalingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScalingServiceServer).Down(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/autoscaler.ScalingService/Down",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScalingServiceServer).Down(ctx, req.(*ScalingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ScalingService_ServiceDesc is the grpc.ServiceDesc for ScalingService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ScalingService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "autoscaler.ScalingService",
	HandlerType: (*ScalingServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Up",
			Handler:    _ScalingService_Up_Handler,
		},
		{
			MethodName: "Down",
			Handler:    _ScalingService_Down_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "request.proto",
}
