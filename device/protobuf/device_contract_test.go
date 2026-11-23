/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */

package protobuf

// The contract of orbtr.device.ApplyBillingHold, held to what its two sides
// rely on: the serving device domain registers the request and response types
// this package generates, and the billing hold rule dials the operation with
// them under a control identity. The descriptors this package carries answer
// for the fields and the operation; the proto source answers for the comments,
// because protoc-gen-go keeps no comment in a descriptor.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// deviceProto is the proto this package is generated from, in the same tree,
// so a run in an export reads that export's own source.
var deviceProto = filepath.Join("..", "..", "protobuf", "contracts", "device", "v1", "device.proto")

func readSource(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// commentAbove returns the comment lines directly above the first line
// matching decl, joined into one space-separated text. With open empty the
// search runs over the whole file; otherwise it runs inside the top-level
// block whose opening line is open and stops at that block's closing brace.
func commentAbove(t *testing.T, src, open, decl string) string {
	t.Helper()
	lines := strings.Split(src, "\n")
	start, end := -1, len(lines)
	if open != "" {
		for i, l := range lines {
			if strings.TrimSpace(l) == open {
				start = i
				break
			}
		}
		if start < 0 {
			t.Fatalf("no block opens with %q", open)
		}
		for j := start + 1; j < len(lines); j++ {
			if lines[j] == "}" {
				end = j
				break
			}
		}
	}
	re := regexp.MustCompile(decl)
	for j := start + 1; j < end; j++ {
		if !re.MatchString(strings.TrimSpace(lines[j])) {
			continue
		}
		var parts []string
		for k := j - 1; k > start; k-- {
			l := strings.TrimSpace(lines[k])
			if !strings.HasPrefix(l, "//") {
				break
			}
			parts = append([]string{strings.TrimSpace(strings.TrimPrefix(l, "//"))}, parts...)
		}
		return strings.Join(parts, " ")
	}
	t.Fatalf("nothing matches %q in block %q", decl, open)
	return ""
}

// axisField matches a field name that would carry an axis on the wire under a
// name other than partitionId. The request names the held partition in its one
// partitionId field; a second copy under another name would be a value the server
// either read, letting a caller act on two partitions at once, or ignored.
var axisField = regexp.MustCompile(`(?i)(tenant|org|owner|subject|user|customer|account|scope)`)

func plainField(f protoreflect.FieldDescriptor, kind protoreflect.Kind, num protoreflect.FieldNumber) bool {
	return f.Number() == num && f.Kind() == kind && f.Cardinality() == protoreflect.Optional &&
		!f.HasPresence() && f.ContainingOneof() == nil
}

func TestTheHoldRequestNamesThePartitionAndTheDirection(t *testing.T) {
	m := File_contracts_device_v1_device_proto.Messages().ByName("ApplyBillingHoldRequest")
	if m == nil {
		t.Fatal("ApplyBillingHoldRequest is not declared")
	}
	if m.Fields().Len() != 2 {
		t.Errorf("ApplyBillingHoldRequest declares %d fields, want hold and partitionId alone", m.Fields().Len())
	}
	for i := 0; i < m.Fields().Len(); i++ {
		if f := m.Fields().Get(i); axisField.MatchString(string(f.Name())) {
			t.Errorf("ApplyBillingHoldRequest.%s names an axis under another name; the held partition is the partitionId field", f.Name())
		}
	}
	if m.ReservedRanges().Len() != 0 || m.Oneofs().Len() != 0 {
		t.Errorf("ApplyBillingHoldRequest reserves %d ranges and declares %d oneofs, want none",
			m.ReservedRanges().Len(), m.Oneofs().Len())
	}
	f := m.Fields().ByName("hold")
	if f == nil {
		t.Fatal("ApplyBillingHoldRequest.hold is not declared")
	}
	if !plainField(f, protoreflect.BoolKind, 1) {
		t.Errorf("hold is field %d %v %v presence=%v oneof=%v, want one plain bool at field 1",
			f.Number(), f.Cardinality(), f.Kind(), f.HasPresence(), f.ContainingOneof() != nil)
	}
	p := m.Fields().ByName("partitionId")
	if p == nil {
		t.Fatal("ApplyBillingHoldRequest.partitionId is not declared")
	}
	if !plainField(p, protoreflect.StringKind, 2) {
		t.Errorf("partitionId is field %d %v %v presence=%v oneof=%v, want one plain string at field 2",
			p.Number(), p.Cardinality(), p.Kind(), p.HasPresence(), p.ContainingOneof() != nil)
	}
}

func TestTheHoldResponseCountsHeldAndReleased(t *testing.T) {
	m := File_contracts_device_v1_device_proto.Messages().ByName("ApplyBillingHoldResponse")
	if m == nil {
		t.Fatal("ApplyBillingHoldResponse is not declared")
	}
	want := []struct {
		name string
		num  protoreflect.FieldNumber
	}{{"held", 1}, {"released", 2}}
	if m.Fields().Len() != len(want) {
		t.Errorf("ApplyBillingHoldResponse declares %d fields, want held and released alone", m.Fields().Len())
	}
	for _, w := range want {
		f := m.Fields().ByName(protoreflect.Name(w.name))
		if f == nil {
			t.Errorf("ApplyBillingHoldResponse.%s is not declared", w.name)
			continue
		}
		if !plainField(f, protoreflect.Int32Kind, w.num) {
			t.Errorf("%s is field %d %v %v presence=%v oneof=%v, want one plain int32 at field %d",
				w.name, f.Number(), f.Cardinality(), f.Kind(), f.HasPresence(), f.ContainingOneof() != nil, w.num)
		}
	}
}

func TestApplyBillingHoldIsAUnaryOperationOfTheDeviceService(t *testing.T) {
	fd := File_contracts_device_v1_device_proto
	if fd.Package() != "contracts.device.v1" || fd.Path() != "contracts/device/v1/device.proto" {
		t.Errorf("the file is %s at %s, want contracts.device.v1 at contracts/device/v1/device.proto",
			fd.Package(), fd.Path())
	}
	s := fd.Services().ByName("DeviceService")
	if s == nil {
		t.Fatal("DeviceService is not declared")
	}
	md := s.Methods().ByName("ApplyBillingHold")
	if md == nil {
		t.Fatal("DeviceService.ApplyBillingHold is not declared")
	}
	if md.Input().FullName() != "contracts.device.v1.ApplyBillingHoldRequest" ||
		md.Output().FullName() != "contracts.device.v1.ApplyBillingHoldResponse" {
		t.Errorf("ApplyBillingHold takes %s and answers %s, want its own request and response",
			md.Input().FullName(), md.Output().FullName())
	}
	if md.IsStreamingClient() || md.IsStreamingServer() {
		t.Errorf("ApplyBillingHold streams (client %v, server %v); the rule dials one call per partition",
			md.IsStreamingClient(), md.IsStreamingServer())
	}
}

func TestTheCommentsStateWhoDialsAndWhatThePartitionIs(t *testing.T) {
	src := readSource(t, deviceProto)
	for _, c := range []struct {
		open, decl string
		want       []string
	}{
		{"", `^service DeviceService \{`, []string{
			"control partition", "names the partition", "control-plane registry"}},
		{"service DeviceService {", `^rpc ApplyBillingHold\(`, []string{
			"billing hold rule", "control identity", "orbtr.control", "no partition member's grant",
			"outside the control partition", "never reads or writes",
			"keeps its identity, its enrolment and its heartbeat", "held = 0 and released = 0"}},
		{"", `^message ApplyBillingHoldRequest \{`, []string{"partition it acts on", "direction"}},
		{"message ApplyBillingHoldRequest {", `^bool hold = 1;`, []string{
			"enrolment instant", "earliest first", "free limit", "every later one under it",
			"every device of the partition", "carried on no field here"}},
		{"message ApplyBillingHoldRequest {", `^string partitionId = 2;`, []string{
			"partition whose devices", "empty", "refused", "control partition"}},
		{"message ApplyBillingHoldResponse {", `^int32 held = 1;`, []string{
			"put under the hold", "not under it before"}},
		{"message ApplyBillingHoldResponse {", `^int32 released = 2;`, []string{
			"took out from under the hold", "inside the free limit", "every device that carried one"}},
	} {
		got := commentAbove(t, src, c.open, c.decl)
		for _, w := range c.want {
			if !strings.Contains(got, w) {
				t.Errorf("%s %s: the comment does not state %q:\n%s", c.open, c.decl, w, got)
			}
		}
	}
}

// recordWords name a record kept outside the repository or frame a history.
// A comment states only the live contract, so no comment line may match.
var recordWords = regexp.MustCompile(
	`(?i)\b(rulings?|brief|adr|owner|commit|phase|previously|legacy|no longer|retired|corrected)\b` +
		`|\b(BIL|ENT|DEV|UWD|CAP|OTL)\s*[0-9]|\b(S[1-4]|W[5-7])\b|\b20\d\d-\d\d-\d\d\b`)

func TestNoCommentNamesARecordOutsideTheRepository(t *testing.T) {
	for i, l := range strings.Split(readSource(t, deviceProto), "\n") {
		if c := strings.Index(l, "//"); c >= 0 && recordWords.MatchString(l[c:]) {
			t.Errorf("device.proto:%d names a record outside the repository or frames a history: %s",
				i+1, strings.TrimSpace(l))
		}
	}
}
