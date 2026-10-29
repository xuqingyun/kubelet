// Copyright The Kubernetes Authors. SPDX-License-Identifier: Apache-2.0
// Explicit mapping preserves internal118 semantics without remarshalling old wire bytes.
package compat

import (
	legacy "k8s.io/kubelet/pkg/apis/podresources/internal118"
	v1 "k8s.io/kubelet/pkg/apis/podresources/v1"
)

func convertListPodResourcesResponse(src *legacy.ListPodResourcesResponse) *v1.ListPodResourcesResponse {
	if src == nil {
		return nil
	}
	dst := &v1.ListPodResourcesResponse{}
	if src.PodResources != nil {
		dst.PodResources = make([]*v1.PodResources, 0, len(src.PodResources))
		for _, v := range src.PodResources {
			dst.PodResources = append(dst.PodResources, convertPodResources(v))
		}
	}
	return dst
}

func convertPodResources(src *legacy.PodResources) *v1.PodResources {
	if src == nil {
		return nil
	}
	dst := &v1.PodResources{}
	dst.Name = src.Name
	dst.Namespace = src.Namespace
	dst.PodRole = src.PodRole
	dst.PodType = src.PodType
	if src.Containers != nil {
		dst.Containers = make([]*v1.ContainerResources, 0, len(src.Containers))
		for _, v := range src.Containers {
			dst.Containers = append(dst.Containers, convertContainerResources(v))
		}
	}
	if src.Labels != nil {
		dst.Labels = make(map[string]string, len(src.Labels))
		for k, v := range src.Labels {
			dst.Labels[k] = v
		}
	}
	if src.Annotations != nil {
		dst.Annotations = make(map[string]string, len(src.Annotations))
		for k, v := range src.Annotations {
			dst.Annotations[k] = v
		}
	}
	return dst
}

func convertContainerResources(src *legacy.ContainerResources) *v1.ContainerResources {
	if src == nil {
		return nil
	}
	dst := &v1.ContainerResources{}
	dst.Name = src.Name
	if src.Devices != nil {
		dst.Devices = make([]*v1.ContainerDevices, 0, len(src.Devices))
		for _, v := range src.Devices {
			dst.Devices = append(dst.Devices, convertContainerDevices(v))
		}
	}
	dst.CpuIds = append([]int64(nil), src.CpuIds...)
	if src.Resources != nil {
		dst.Resources = make([]*v1.TopologyAwareResource, 0, len(src.Resources))
		for _, v := range src.Resources {
			dst.Resources = append(dst.Resources, convertTopologyAwareResource(v))
		}
	}
	return dst
}

func convertContainerDevices(src *legacy.ContainerDevices) *v1.ContainerDevices {
	if src == nil {
		return nil
	}
	dst := &v1.ContainerDevices{}
	dst.ResourceName = src.ResourceName
	dst.DeviceIds = append([]string(nil), src.DeviceIds...)
	dst.Topology = convertTopologyInfo(src.Topology)
	return dst
}

func convertTopologyInfo(src *legacy.TopologyInfo) *v1.TopologyInfo {
	if src == nil {
		return nil
	}
	dst := &v1.TopologyInfo{}
	if src.Nodes != nil {
		dst.Nodes = make([]*v1.NUMANode, 0, len(src.Nodes))
		for _, v := range src.Nodes {
			dst.Nodes = append(dst.Nodes, convertNUMANode(v))
		}
	}
	return dst
}

func convertNUMANode(src *legacy.NUMANode) *v1.NUMANode {
	if src == nil {
		return nil
	}
	dst := &v1.NUMANode{}
	dst.ID = src.ID
	return dst
}

func convertTopologyAwareResource(src *legacy.TopologyAwareResource) *v1.TopologyAwareResource {
	if src == nil {
		return nil
	}
	dst := &v1.TopologyAwareResource{}
	dst.ResourceName = src.ResourceName
	dst.IsNodeResource = src.IsNodeResource
	dst.IsScalarResource = src.IsScalarResource
	dst.AggregatedQuantity = src.AggregatedQuantity
	dst.OriginalAggregatedQuantity = src.OriginalAggregatedQuantity
	if src.TopologyAwareQuantityList != nil {
		dst.TopologyAwareQuantityList = make([]*v1.TopologyAwareQuantity, 0, len(src.TopologyAwareQuantityList))
		for _, v := range src.TopologyAwareQuantityList {
			dst.TopologyAwareQuantityList = append(dst.TopologyAwareQuantityList, convertTopologyAwareQuantity(v))
		}
	}
	if src.OriginalTopologyAwareQuantityList != nil {
		dst.OriginalTopologyAwareQuantityList = make([]*v1.TopologyAwareQuantity, 0, len(src.OriginalTopologyAwareQuantityList))
		for _, v := range src.OriginalTopologyAwareQuantityList {
			dst.OriginalTopologyAwareQuantityList = append(dst.OriginalTopologyAwareQuantityList, convertTopologyAwareQuantity(v))
		}
	}
	return dst
}

func convertTopologyAwareQuantity(src *legacy.TopologyAwareQuantity) *v1.TopologyAwareQuantity {
	if src == nil {
		return nil
	}
	dst := &v1.TopologyAwareQuantity{}
	dst.ResourceValue = src.ResourceValue
	dst.Node = src.Node
	dst.Name = src.Name
	dst.Type = src.Type
	dst.TopologyLevel = v1.TopologyLevel(src.TopologyLevel)
	if src.Annotations != nil {
		dst.Annotations = make(map[string]string, len(src.Annotations))
		for k, v := range src.Annotations {
			dst.Annotations[k] = v
		}
	}
	return dst
}

func convertAllocatableResourcesResponse(src *legacy.AllocatableResourcesResponse) *v1.AllocatableResourcesResponse {
	if src == nil {
		return nil
	}
	dst := &v1.AllocatableResourcesResponse{}
	if src.Devices != nil {
		dst.Devices = make([]*v1.ContainerDevices, 0, len(src.Devices))
		for _, v := range src.Devices {
			dst.Devices = append(dst.Devices, convertContainerDevices(v))
		}
	}
	dst.CpuIds = append([]int64(nil), src.CpuIds...)
	if src.Resources != nil {
		dst.Resources = make([]*v1.AllocatableTopologyAwareResource, 0, len(src.Resources))
		for _, v := range src.Resources {
			dst.Resources = append(dst.Resources, convertAllocatableTopologyAwareResource(v))
		}
	}
	return dst
}

func convertAllocatableTopologyAwareResource(src *legacy.AllocatableTopologyAwareResource) *v1.AllocatableTopologyAwareResource {
	if src == nil {
		return nil
	}
	dst := &v1.AllocatableTopologyAwareResource{}
	dst.ResourceName = src.ResourceName
	dst.IsNodeResource = src.IsNodeResource
	dst.IsScalarResource = src.IsScalarResource
	dst.AggregatedAllocatableQuantity = src.AggregatedAllocatableQuantity
	if src.TopologyAwareAllocatableQuantityList != nil {
		dst.TopologyAwareAllocatableQuantityList = make([]*v1.TopologyAwareQuantity, 0, len(src.TopologyAwareAllocatableQuantityList))
		for _, v := range src.TopologyAwareAllocatableQuantityList {
			dst.TopologyAwareAllocatableQuantityList = append(dst.TopologyAwareAllocatableQuantityList, convertTopologyAwareQuantity(v))
		}
	}
	dst.AggregatedCapacityQuantity = src.AggregatedCapacityQuantity
	if src.TopologyAwareCapacityQuantityList != nil {
		dst.TopologyAwareCapacityQuantityList = make([]*v1.TopologyAwareQuantity, 0, len(src.TopologyAwareCapacityQuantityList))
		for _, v := range src.TopologyAwareCapacityQuantityList {
			dst.TopologyAwareCapacityQuantityList = append(dst.TopologyAwareCapacityQuantityList, convertTopologyAwareQuantity(v))
		}
	}
	return dst
}
