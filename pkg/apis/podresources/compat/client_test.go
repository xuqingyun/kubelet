package compat

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	legacy "k8s.io/kubelet/pkg/apis/podresources/internal118"
	v1 "k8s.io/kubelet/pkg/apis/podresources/v1"
)

func testConn(t *testing.T, register func(*grpc.Server)) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	register(server)
	go server.Serve(lis)
	t.Cleanup(server.Stop)
	conn, err := grpc.DialContext(context.Background(), "passthrough:///buf", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func fixture() (*legacy.ListPodResourcesResponse, *legacy.AllocatableResourcesResponse) {
	quantity := &legacy.TopologyAwareQuantity{ResourceValue: 2.5, Node: 1, Name: "cpu", Type: "Core", TopologyLevel: legacy.TopologyLevel_SOCKET, Annotations: map[string]string{"kind": "original"}}
	resource := &legacy.TopologyAwareResource{ResourceName: "cpu", IsScalarResource: true, AggregatedQuantity: 2.5, OriginalAggregatedQuantity: 3, TopologyAwareQuantityList: []*legacy.TopologyAwareQuantity{quantity}, OriginalTopologyAwareQuantityList: []*legacy.TopologyAwareQuantity{quantity}}
	device := &legacy.ContainerDevices{ResourceName: "example/device", DeviceIds: []string{"device-1"}, Topology: &legacy.TopologyInfo{Nodes: []*legacy.NUMANode{{ID: 1}}}}
	list := &legacy.ListPodResourcesResponse{PodResources: []*legacy.PodResources{{Name: "pod", Namespace: "ns", Labels: map[string]string{"app": "test"}, Annotations: map[string]string{"numa_binding": "true"}, Containers: []*legacy.ContainerResources{{Name: "main", CpuIds: []int64{2, 3}, Devices: []*legacy.ContainerDevices{device}, Resources: []*legacy.TopologyAwareResource{resource}}}}}}
	alloc := &legacy.AllocatableResourcesResponse{CpuIds: []int64{0, 1, 2, 3}, Devices: []*legacy.ContainerDevices{device}, Resources: []*legacy.AllocatableTopologyAwareResource{{ResourceName: "cpu", IsNodeResource: true, IsScalarResource: true, AggregatedAllocatableQuantity: 4, AggregatedCapacityQuantity: 8, TopologyAwareAllocatableQuantityList: []*legacy.TopologyAwareQuantity{quantity}, TopologyAwareCapacityQuantityList: []*legacy.TopologyAwareQuantity{quantity}}}}
	return list, alloc
}

func legacyConn(t *testing.T, list *legacy.ListPodResourcesResponse, alloc *legacy.AllocatableResourcesResponse, rpcErr error) *grpc.ClientConn {
	return testConn(t, func(s *grpc.Server) {
		type service interface{}
		s.RegisterService(&grpc.ServiceDesc{ServiceName: "v1.PodResourcesLister", HandlerType: (*service)(nil), Methods: []grpc.MethodDesc{
			{MethodName: "List", Handler: func(_ interface{}, _ context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
				if err := dec(new(legacy.ListPodResourcesRequest)); err != nil {
					return nil, err
				}
				return list, rpcErr
			}},
			{MethodName: "GetAllocatableResources", Handler: func(_ interface{}, _ context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
				if err := dec(new(legacy.AllocatableResourcesRequest)); err != nil {
					return nil, err
				}
				return alloc, rpcErr
			}},
		}}, struct{}{})
	})
}

func TestLegacyWireSuccessStillLosesFieldsWithCurrentDecoder(t *testing.T) {
	list, alloc := fixture()
	conn := legacyConn(t, list, alloc, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	raw := v1.NewPodResourcesListerClient(conn)
	wrong, err := raw.List(ctx, &v1.ListPodResourcesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(wrong.PodResources) != 1 || len(wrong.PodResources[0].Containers) != 0 {
		t.Fatalf("fixture must reproduce successful but incomplete decoding: %v", wrong)
	}
	wrongAlloc, err := raw.GetAllocatableResources(ctx, &v1.AllocatableResourcesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(wrongAlloc.Resources) != 0 {
		t.Fatal("current decoder unexpectedly preserved old resource field")
	}
	client, err := NewClient(conn, Kubewharf18)
	if err != nil {
		t.Fatal(err)
	}
	correct, err := client.List(ctx, &v1.ListPodResourcesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	quantity := &v1.TopologyAwareQuantity{ResourceValue: 2.5, Node: 1, Name: "cpu", Type: "Core", TopologyLevel: v1.TopologyLevel_SOCKET, Annotations: map[string]string{"kind": "original"}}
	device := &v1.ContainerDevices{ResourceName: "example/device", DeviceIds: []string{"device-1"}, Topology: &v1.TopologyInfo{Nodes: []*v1.NUMANode{{ID: 1}}}}
	expectedList := &v1.ListPodResourcesResponse{PodResources: []*v1.PodResources{{Name: "pod", Namespace: "ns", Labels: map[string]string{"app": "test"}, Annotations: map[string]string{"numa_binding": "true"}, Containers: []*v1.ContainerResources{{Name: "main", CpuIds: []int64{2, 3}, Devices: []*v1.ContainerDevices{device}, Resources: []*v1.TopologyAwareResource{{ResourceName: "cpu", IsScalarResource: true, AggregatedQuantity: 2.5, OriginalAggregatedQuantity: 3, TopologyAwareQuantityList: []*v1.TopologyAwareQuantity{quantity}, OriginalTopologyAwareQuantityList: []*v1.TopologyAwareQuantity{quantity}}}}}}}}
	if !proto.Equal(correct, expectedList) {
		t.Fatalf("lost List fields: %v", correct)
	}
	correctAlloc, err := client.GetAllocatableResources(ctx, &v1.AllocatableResourcesRequest{})
	if err != nil {
		t.Fatal(err)
	}
	expectedAlloc := &v1.AllocatableResourcesResponse{CpuIds: []int64{0, 1, 2, 3}, Devices: []*v1.ContainerDevices{device}, Resources: []*v1.AllocatableTopologyAwareResource{{ResourceName: "cpu", IsNodeResource: true, IsScalarResource: true, AggregatedAllocatableQuantity: 4, AggregatedCapacityQuantity: 8, TopologyAwareAllocatableQuantityList: []*v1.TopologyAwareQuantity{quantity}, TopologyAwareCapacityQuantityList: []*v1.TopologyAwareQuantity{quantity}}}}
	if !proto.Equal(correctAlloc, expectedAlloc) {
		t.Fatalf("lost capacity fields: %v", correctAlloc)
	}
	resource := correct.PodResources[0].Containers[0].Resources[0]
	if resource.OriginalAggregatedQuantity != 3 || resource.TopologyAwareQuantityList[0].TopologyLevel != v1.TopologyLevel_SOCKET || correct.PodResources[0].Annotations["numa_binding"] != "true" {
		t.Fatal("resource semantics or annotations were not preserved")
	}
}

func TestLegacyGetUsesNamespaceAndNameWithoutNewRPC(t *testing.T) {
	list, alloc := fixture()
	list.PodResources = append(list.PodResources, &legacy.PodResources{Name: "pod", Namespace: "other"})
	client, _ := NewClient(legacyConn(t, list, alloc, nil), Kubewharf18)
	got, err := client.Get(context.Background(), &v1.GetPodResourcesRequest{PodName: "pod", PodNamespace: "ns"})
	if err != nil || got.PodResources.Namespace != "ns" {
		t.Fatalf("lookup failed: %v %v", got, err)
	}
	_, err = client.Get(context.Background(), &v1.GetPodResourcesRequest{PodName: "missing", PodNamespace: "ns"})
	if status.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
}

type currentServer struct {
	v1.UnimplementedPodResourcesListerServer
	list  *v1.ListPodResourcesResponse
	alloc *v1.AllocatableResourcesResponse
}

func (s *currentServer) List(context.Context, *v1.ListPodResourcesRequest) (*v1.ListPodResourcesResponse, error) {
	return s.list, nil
}
func (s *currentServer) GetAllocatableResources(context.Context, *v1.AllocatableResourcesRequest) (*v1.AllocatableResourcesResponse, error) {
	return s.alloc, nil
}

func TestCurrentProtocolPreservesNewMemoryFields(t *testing.T) {
	srv := &currentServer{list: &v1.ListPodResourcesResponse{PodResources: []*v1.PodResources{{Name: "pod", Containers: []*v1.ContainerResources{{Name: "main", Memory: []*v1.ContainerMemory{{MemoryType: "memory", Size: 1024}}}}}}}, alloc: &v1.AllocatableResourcesResponse{Memory: []*v1.ContainerMemory{{MemoryType: "memory", Size: 8192}}}}
	client, _ := NewClient(testConn(t, func(s *grpc.Server) { v1.RegisterPodResourcesListerServer(s, srv) }), Kubewharf35)
	list, err := client.List(context.Background(), &v1.ListPodResourcesRequest{})
	if err != nil || !proto.Equal(list, srv.list) {
		t.Fatalf("List changed: %v %v", list, err)
	}
	alloc, err := client.GetAllocatableResources(context.Background(), &v1.AllocatableResourcesRequest{})
	if err != nil || !proto.Equal(alloc, srv.alloc) {
		t.Fatalf("capacity changed: %v %v", alloc, err)
	}
}

func TestEmptyLegacyResponseAndRPCErrorsDoNotSelectOtherProtocol(t *testing.T) {
	client, _ := NewClient(legacyConn(t, &legacy.ListPodResourcesResponse{}, &legacy.AllocatableResourcesResponse{}, nil), Kubewharf18)
	got, err := client.List(context.Background(), &v1.ListPodResourcesRequest{})
	if err != nil || len(got.PodResources) != 0 {
		t.Fatalf("empty node failed: %v %v", got, err)
	}
	client, _ = NewClient(legacyConn(t, nil, nil, status.Error(codes.Unavailable, "offline")), Kubewharf18)
	_, err = client.GetAllocatableResources(context.Background(), &v1.AllocatableResourcesRequest{})
	if status.Code(err) != codes.Unavailable {
		t.Fatal(err)
	}
	if _, err := ParseProtocol("auto"); err == nil {
		t.Fatal("unknown protocol accepted")
	}
	if _, err := ParseProtocol("internal118"); err == nil {
		t.Fatal("retired protocol name accepted")
	}
}

func TestLegacyPodRoleAndTypeArePreserved(t *testing.T) {
	list, alloc := fixture()
	list.PodResources[0].PodRole = "role"
	list.PodResources[0].PodType = "type"
	client, _ := NewClient(legacyConn(t, list, alloc, nil), Kubewharf18)
	got, err := client.List(context.Background(), &v1.ListPodResourcesRequest{})
	if err != nil || got.PodResources[0].PodRole != "role" || got.PodResources[0].PodType != "type" {
		t.Fatalf("string fields lost: %v %v", got, err)
	}
}
