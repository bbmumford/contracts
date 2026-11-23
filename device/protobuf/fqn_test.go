/*
 * Copyright (c) 2026 Brandon Miller-Mumford. Licensed under the MIT License; see LICENSE.
 */

package protobuf

import (
	"strings"
	"testing"
)

// TestTheFQNNamesTheDeclaredOperationUnderTheDeviceDomain pins the name both
// sides dial and declare: its operation segment is the one method DeviceService
// declares, read from the descriptor, and its prefix is the device domain under
// the bare namespace every device operation is registered under.
func TestTheFQNNamesTheDeclaredOperationUnderTheDeviceDomain(t *testing.T) {
	s := File_contracts_device_v1_device_proto.Services().ByName("DeviceService")
	if s == nil {
		t.Fatal("DeviceService is not declared")
	}
	if s.Methods().Len() != 1 {
		t.Fatalf("DeviceService declares %d methods, want the one this constant names", s.Methods().Len())
	}
	method := string(s.Methods().Get(0).Name())
	if got, want := FQNApplyBillingHold, "orbtr.device."+method; got != want {
		t.Fatalf("FQNApplyBillingHold = %q, want %q", got, want)
	}
	if strings.Count(FQNApplyBillingHold, ".") != 2 {
		t.Fatalf("FQNApplyBillingHold = %q, want namespace.domain.Operation", FQNApplyBillingHold)
	}
}
