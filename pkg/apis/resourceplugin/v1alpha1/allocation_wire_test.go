/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"encoding/hex"
	"encoding/json"
	"google.golang.org/protobuf/proto"
	"testing"
)

// Fixed ResourceAllocationInfo wire bytes from the internal pre.3 contract:
// field 9 contains NUMA assignments, field 11 contains a DeviceSpec. The gap at
// field 10 must not renumber devices or reinterpret another legacy field.
func TestInternalAllocationTopologyAndDeviceWire(t *testing.T) {
	wire, err := hex.DecodeString("4a04080010024a04080110035a100a042f64657612042f6465761a027277")
	if err != nil {
		t.Fatal(err)
	}
	got := &ResourceAllocationInfo{}
	if err = proto.Unmarshal(wire, got); err != nil {
		t.Fatal(err)
	}
	if got.TopologyAssignments[0] != 2 || got.TopologyAssignments[1] != 3 {
		t.Fatalf("NUMA assignments lost: %v", got.TopologyAssignments)
	}
	if len(got.Devices) != 1 || got.Devices[0].HostPath != "/dev" || got.Devices[0].ContainerPath != "/dev" || got.Devices[0].Permissions != "rw" {
		t.Fatalf("device mapping lost: %v", got.Devices)
	}
	checkpoint, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	restored := &ResourceAllocationInfo{}
	if err = json.Unmarshal(checkpoint, restored); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(got, restored) {
		t.Fatalf("checkpoint lost allocation: %v", restored)
	}
}
