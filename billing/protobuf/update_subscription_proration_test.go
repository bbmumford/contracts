/*
 * Copyright (c) 2026 ORBTR AI Pty Ltd. All Rights Reserved.
 * Queries: licensing@orbtr.ai
 */

package protobuf

// UpdateSubscriptionRequest's proration choice, held to the shape the billing
// host reads it by: an unset field is told from an explicit false, so a caller
// that does not choose is answered with proration while a caller can still ask
// for none. The descriptor answers for the field's number, kind and presence;
// the proto source answers for its comment, which protoc-gen-go keeps out of
// the descriptor, and the generated file is read too, so a comment changed in
// the proto and not regenerated fails here.

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestTheUpdateRequestCarriesAProrationChoiceWithPresence(t *testing.T) {
	msg := (&UpdateSubscriptionRequest{}).ProtoReflect()
	fd := msg.Descriptor().Fields().ByName("prorate")
	if fd == nil {
		t.Fatal("UpdateSubscriptionRequest has no prorate field")
	}
	if fd.Number() != 4 || fd.Kind() != protoreflect.BoolKind || fd.Cardinality() != protoreflect.Optional {
		t.Fatalf("prorate = number %d, kind %v, cardinality %v; want 4, bool, optional", fd.Number(), fd.Kind(), fd.Cardinality())
	}
	if !fd.HasPresence() {
		t.Fatal("prorate has no presence: an explicit false would read as unset, and unset is proration")
	}
	for name, num := range map[protoreflect.Name]protoreflect.FieldNumber{"subscriptionId": 1, "items": 2, "status": 3} {
		if f := msg.Descriptor().Fields().ByName(name); f == nil || f.Number() != num {
			t.Errorf("UpdateSubscriptionRequest.%s is not field %d", name, num)
		}
	}

	for _, set := range []bool{false, true} {
		m := (&UpdateSubscriptionRequest{SubscriptionId: "sub_1"}).ProtoReflect()
		m.Set(fd, protoreflect.ValueOfBool(set))
		b, err := proto.Marshal(m.Interface())
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		back := (&UpdateSubscriptionRequest{}).ProtoReflect()
		if err := proto.Unmarshal(b, back.Interface()); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !back.Has(fd) || back.Get(fd).Bool() != set {
			t.Errorf("prorate=%v did not survive the wire: has %v, value %v", set, back.Has(fd), back.Get(fd).Bool())
		}
	}
	b, err := proto.Marshal(&UpdateSubscriptionRequest{SubscriptionId: "sub_1"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	unset := (&UpdateSubscriptionRequest{}).ProtoReflect()
	if err := proto.Unmarshal(b, unset.Interface()); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if unset.Has(fd) {
		t.Error("a request that does not set prorate arrives with it set")
	}
}

func TestTheProrationCommentStatesTheChoiceAndTheUnsetAnswer(t *testing.T) {
	src := readSource(t, billingProto)
	lines, trailing := commentAbove(t, src, "message UpdateSubscriptionRequest {", `^optional bool prorate = 4;`)
	got := strings.Join(lines, " ")
	requireStates(t, "UpdateSubscriptionRequest.prorate", got, []string{
		"true prorates", "false", "no proration", "Unset", "presence", "SubscriptionUpdater"})
	if trailing != "" {
		t.Errorf("UpdateSubscriptionRequest.prorate carries a trailing comment %q", trailing)
	}
	for _, l := range lines {
		if recordWords.MatchString(l) {
			t.Errorf("UpdateSubscriptionRequest.prorate: the comment names a record outside the repository or frames a history: %s", l)
		}
	}
	gen := readSource(t, "billing.pb.go")
	want := "// " + strings.Join(lines, "\n// ")
	var b strings.Builder
	for _, l := range strings.Split(gen, "\n") {
		b.WriteString(strings.TrimSpace(l))
		b.WriteString("\n")
	}
	if !strings.Contains(b.String(), want) {
		t.Errorf("billing.pb.go does not carry the comment of UpdateSubscriptionRequest.prorate:\n%s", want)
	}
}

// TestTheProrationCommentClaimsNoUnreadField holds the comment to the billing
// host as it reads the field: its UpdateSubscription hands the request's choice to
// the payment rail (prorateOf in the host), so a comment saying the host prorates
// every change or does not read the field tells a caller its explicit false is
// ignored when it is not.
func TestTheProrationCommentClaimsNoUnreadField(t *testing.T) {
	src := readSource(t, billingProto)
	lines, _ := commentAbove(t, src, "message UpdateSubscriptionRequest {", `^optional bool prorate = 4;`)
	for _, l := range lines {
		if strings.HasPrefix(l, "TODO") || strings.Contains(l, "does not read") || strings.Contains(l, "prorates every change") {
			t.Errorf("UpdateSubscriptionRequest.prorate says the billing host ignores the field, which it reads: %s", l)
		}
	}
}
