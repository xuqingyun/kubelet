/*
Copyright 2026 The Kubernetes Authors.

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
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestResourcePluginOptionsExtraResources(t *testing.T) {
	want := []string{"resource.example.com/foo", "resource.example.com/bar"}
	options := &ResourcePluginOptions{ExtraResources: want}

	data, err := proto.Marshal(options)
	if err != nil {
		t.Fatalf("marshal ResourcePluginOptions: %v", err)
	}

	got := &ResourcePluginOptions{}
	if err := proto.Unmarshal(data, got); err != nil {
		t.Fatalf("unmarshal ResourcePluginOptions: %v", err)
	}

	if !proto.Equal(options, got) {
		t.Fatalf("round-trip mismatch: got %v, want %v", got.GetExtraResources(), want)
	}
}

func TestLegacyCompatibilityContract(t *testing.T) {
	file := File_staging_src_k8s_io_kubelet_pkg_apis_resourceplugin_v1alpha1_api_proto

	for _, messageName := range []protoreflect.Name{
		"PodResourceRequest",
		"PodResourceHintsResponse",
		"PodResourceAllocationResponse",
		"DeviceSpec",
	} {
		if file.Messages().ByName(messageName) == nil {
			t.Errorf("legacy message %q is missing", messageName)
		}
	}

	legacyFields := []struct {
		message protoreflect.Name
		field   protoreflect.Name
		number  protoreflect.FieldNumber
	}{
		{"ResourceRequest", "native_qos_class", 14},
		{"ResourceHintsResponse", "native_qos_class", 13},
		{"ResourceAllocationResponse", "native_qos_class", 13},
		{"ResourceAllocationInfo", "topology_assignments", 9},
		{"ResourceAllocationInfo", "devices", 11},
	}
	for _, legacyField := range legacyFields {
		message := file.Messages().ByName(legacyField.message)
		field := message.Fields().ByName(legacyField.field)
		if field == nil {
			t.Errorf("legacy field %s.%s is missing", legacyField.message, legacyField.field)
			continue
		}
		if field.Number() != legacyField.number {
			t.Errorf("legacy field %s.%s has number %d, want %d", legacyField.message, legacyField.field, field.Number(), legacyField.number)
		}
	}

	resourcePlugin := file.Services().ByName("ResourcePlugin")
	for _, methodName := range []protoreflect.Name{"GetPodTopologyHints", "AllocateForPod"} {
		if resourcePlugin.Methods().ByName(methodName) == nil {
			t.Errorf("legacy RPC %q is missing", methodName)
		}
	}
}
