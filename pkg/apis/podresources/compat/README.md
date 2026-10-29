# Internal PodResources compatibility

`NewClient(conn, Kubewharf18)` decodes the fixed Kubewharf 1.18 wire layout and maps every field explicitly to the current v1 DTO. `Kubewharf35` uses the current generated client. Deployment must choose the protocol; RPC errors or an empty node are not reliable version indicators. Both use `/v1.PodResourcesLister/*`.

| Response | kubewharf-18 | current kubewharf-35 |
| --- | --- | --- |
| PodResources.containers | 5 | 3 |
| PodResources.labels / annotations | 6 / 7 | 14 / 15 |
| ContainerResources.resources | 4 | 15 (4 is memory) |
| AllocatableResourcesResponse.resources | 3 | 15 (3 is memory) |

The isolated protobuf namespace avoids descriptor registration conflicts. The reference is internal Kubernetes commit `a16bf49244d3159432493461f6fe7697f295b051`, `staging/src/k8s.io/kubelet/pkg/apis/podresources/v1/api.proto`. Field numbers must not be regenerated from a different server branch without review. Labels, annotations, devices, CPU IDs, capacity, original quantities, and NUMA/socket topology survive conversion. The current server's new memory fields remain unchanged.

The 1.18 client implements Get through List using both namespace and name, because 1.18 has no Get RPC. It returns NotFound if absent. This package fixes client-side reporting; it does not modify kubelet allocation, ResourcePlugin topology/device support, or native QoS policy.

Regenerate internal118 messages with protoc 29.0 and protoc-gen-go v1.36.4:

```sh
protoc --go_out=. --go_opt=module=k8s.io/kubelet pkg/apis/podresources/internal118/api.proto
```

Run the real gRPC regression (including successful but incomplete modern decoding of old responses):

```sh
go test -race ./pkg/apis/podresources/compat
```
