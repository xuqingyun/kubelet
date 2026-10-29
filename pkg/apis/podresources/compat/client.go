// Copyright The Kubernetes Authors. SPDX-License-Identifier: Apache-2.0
// Package compat explicitly selects the internal kubelet PodResources wire layout.
package compat

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	legacy "k8s.io/kubelet/pkg/apis/podresources/internal118"
	v1 "k8s.io/kubelet/pkg/apis/podresources/v1"
)

type Protocol string

const (
	Kubewharf18     Protocol = "kubewharf-18"
	Kubewharf35     Protocol = "kubewharf-35"
	DefaultProtocol          = Kubewharf35
)

// ParseProtocol defaults existing callers to the current SDK layout.
func ParseProtocol(value string) (Protocol, error) {
	if value == "" {
		return DefaultProtocol, nil
	}
	p := Protocol(value)
	if p != Kubewharf18 && p != Kubewharf35 {
		return "", fmt.Errorf("unsupported PodResources protocol %q: use kubewharf-18 or kubewharf-35", value)
	}
	return p, nil
}

type legacyClient struct{ conn grpc.ClientConnInterface }

// NewClient binds both read methods to one layout, never inferring it from RPC success or empty data.
func NewClient(conn grpc.ClientConnInterface, protocol Protocol) (v1.PodResourcesListerClient, error) {
	p, err := ParseProtocol(string(protocol))
	if err != nil {
		return nil, err
	}
	if p == Kubewharf35 {
		return v1.NewPodResourcesListerClient(conn), nil
	}
	return &legacyClient{conn: conn}, nil
}

func (c *legacyClient) List(ctx context.Context, _ *v1.ListPodResourcesRequest, opts ...grpc.CallOption) (*v1.ListPodResourcesResponse, error) {
	out := new(legacy.ListPodResourcesResponse)
	if err := c.conn.Invoke(ctx, "/v1.PodResourcesLister/List", &legacy.ListPodResourcesRequest{}, out, opts...); err != nil {
		return nil, err
	}
	return convertListPodResourcesResponse(out), nil
}

func (c *legacyClient) GetAllocatableResources(ctx context.Context, _ *v1.AllocatableResourcesRequest, opts ...grpc.CallOption) (*v1.AllocatableResourcesResponse, error) {
	out := new(legacy.AllocatableResourcesResponse)
	if err := c.conn.Invoke(ctx, "/v1.PodResourcesLister/GetAllocatableResources", &legacy.AllocatableResourcesRequest{}, out, opts...); err != nil {
		return nil, err
	}
	return convertAllocatableResourcesResponse(out), nil
}

// Get uses a List lookup on 1.18, which has no single-pod Get RPC.
func (c *legacyClient) Get(ctx context.Context, in *v1.GetPodResourcesRequest, opts ...grpc.CallOption) (*v1.GetPodResourcesResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "nil pod resources request")
	}
	list, err := c.List(ctx, &v1.ListPodResourcesRequest{}, opts...)
	if err != nil {
		return nil, err
	}
	for _, pod := range list.PodResources {
		if pod != nil && pod.Name == in.PodName && pod.Namespace == in.PodNamespace {
			return &v1.GetPodResourcesResponse{PodResources: pod}, nil
		}
	}
	return nil, status.Error(codes.NotFound, "pod not found in kubewharf-18 PodResources list")
}
