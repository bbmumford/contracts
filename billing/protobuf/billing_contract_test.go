/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */

package protobuf

// The comments of the subscription status surfaces, held to what the billing
// host does with them: library's billinghost answers UpdateSubscriptionStatus
// without writing a status, and the status a Subscription carries is the one
// the provider names, written at creation from the provider's answer and then
// by HandleWebhook's applier alone. The proto source answers for the comments,
// because protoc-gen-go keeps no comment in a descriptor, and the generated
// files are read too, so a comment changed in the proto and not regenerated
// fails here.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// billingProto is the proto this package is generated from, in the same tree,
// so a run in an export reads that export's own source.
var billingProto = filepath.Join("..", "..", "protobuf", "contracts", "billing", "v1", "billing.proto")

func readSource(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// commentAbove returns the comment lines directly above the first line
// matching decl inside the top-level block whose opening line is open, one
// entry per line with the comment marker removed, together with that line's
// own text after any trailing comment marker ("" when it has none).
func commentAbove(t *testing.T, src, open, decl string) (lines []string, trailing string) {
	t.Helper()
	all := strings.Split(src, "\n")
	start, end := -1, len(all)
	for i, l := range all {
		if strings.TrimSpace(l) == open {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("no block opens with %q", open)
	}
	for j := start + 1; j < len(all); j++ {
		if all[j] == "}" {
			end = j
			break
		}
	}
	re := regexp.MustCompile(decl)
	for j := start + 1; j < end; j++ {
		l := strings.TrimSpace(all[j])
		if !re.MatchString(l) {
			continue
		}
		if c := strings.Index(l, "//"); c >= 0 {
			trailing = strings.TrimSpace(l[c+2:])
		}
		for k := j - 1; k > start; k-- {
			p := strings.TrimSpace(all[k])
			if !strings.HasPrefix(p, "//") {
				break
			}
			lines = append([]string{strings.TrimSpace(strings.TrimPrefix(p, "//"))}, lines...)
		}
		return lines, trailing
	}
	t.Fatalf("nothing matches %q in block %q", decl, open)
	return nil, ""
}

func requireStates(t *testing.T, where, got string, want []string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("%s: the comment does not state %q:\n%s", where, w, got)
		}
	}
}

// statusValues are the subscription status words a comment could list. The
// field's comment names the set's one home, billing.ValidSubscriptionStatus,
// which Service.Subscribe refuses a provider answer outside of, and lists none
// of them: the applier stores whatever value the provider names, so a list here
// is read by nothing and drifts from both.
var statusValues = regexp.MustCompile(
	`\b(trialing|active|past_due|unpaid|incomplete|incomplete_expired|canceled|cancelled|paused|suspended|downgrade_pending)\b`)

func TestTheStatusOperationSaysItWritesNoStatus(t *testing.T) {
	lines, trailing := commentAbove(t, readSource(t, billingProto),
		"service BillingService {", `^rpc UpdateSubscriptionStatusRPC\(`)
	got := strings.Join(lines, " ")
	requireStates(t, "rpc UpdateSubscriptionStatusRPC", got, []string{
		"writes no status", "caller's partition holds", "verified provider event", "HandleWebhook",
		"FQNUpdateSubscriptionStatus", "No caller dials it"})
	for _, claim := range []string{"webhook processing", "enforcement"} {
		if strings.Contains(got, claim) {
			t.Errorf("rpc UpdateSubscriptionStatusRPC: the comment says the operation serves %q, which writes no status:\n%s", claim, got)
		}
	}
	if trailing != "" {
		t.Errorf("rpc UpdateSubscriptionStatusRPC carries a trailing comment %q", trailing)
	}
}

func TestTheSubscriptionStatusNamesTheProviderAndTheSetsOneHome(t *testing.T) {
	lines, trailing := commentAbove(t, readSource(t, billingProto),
		"message Subscription {", `^string status = 4;`)
	got := strings.Join(lines, " ")
	requireStates(t, "Subscription.status", got, []string{
		"as the provider names it", "billing.ValidSubscriptionStatus", "Service.Subscribe",
		"HandleWebhook", "verified provider event"})
	if v := statusValues.FindAllString(got+" "+trailing, -1); len(v) != 0 {
		t.Errorf("Subscription.status: the comment lists status values %v; the set's one home is billing.ValidSubscriptionStatus:\n%s %s",
			v, got, trailing)
	}
	if trailing != "" {
		t.Errorf("Subscription.status carries a trailing comment %q", trailing)
	}
}

// TestTheGeneratedFilesCarryTheProtoComments reads each comment line the two
// tests above hold back out of the generated file protoc-gen-go or
// protoc-gen-go-grpc writes it into, so the package's Go source says what its
// proto says.
func TestTheGeneratedFilesCarryTheProtoComments(t *testing.T) {
	src := readSource(t, billingProto)
	for _, c := range []struct {
		open, decl, generated string
	}{
		{"service BillingService {", `^rpc UpdateSubscriptionStatusRPC\(`, "billing_grpc.pb.go"},
		{"message Subscription {", `^string status = 4;`, "billing.pb.go"},
	} {
		lines, _ := commentAbove(t, src, c.open, c.decl)
		if len(lines) == 0 {
			t.Errorf("%s %s has no comment to carry", c.open, c.decl)
			continue
		}
		gen := readSource(t, c.generated)
		want := "// " + strings.Join(lines, "\n// ")
		var b strings.Builder
		for _, l := range strings.Split(gen, "\n") {
			b.WriteString(strings.TrimSpace(l))
			b.WriteString("\n")
		}
		if !strings.Contains(b.String(), want) {
			t.Errorf("%s does not carry the comment of %s %s:\n%s", c.generated, c.open, c.decl, want)
		}
	}
}

// recordWords name a record kept outside the repository or frame a history.
// A comment states only the live contract, so neither comment may match.
var recordWords = regexp.MustCompile(
	`(?i)\b(rulings?|brief|adr|owner|commit|phase|previously|legacy|no longer|retired|corrected|was|now)\b` +
		`|\b(BIL|ENT|DEV|UWD|CAP|OTL)[-\s]*[A-Z]?[0-9]|\b20\d\d-\d\d-\d\d\b`)

func TestTheStatusCommentsNameNoRecordOutsideTheRepository(t *testing.T) {
	src := readSource(t, billingProto)
	for _, c := range []struct{ open, decl string }{
		{"service BillingService {", `^rpc UpdateSubscriptionStatusRPC\(`},
		{"message Subscription {", `^string status = 4;`},
	} {
		lines, trailing := commentAbove(t, src, c.open, c.decl)
		for _, l := range append(lines, trailing) {
			if recordWords.MatchString(l) {
				t.Errorf("%s %s: the comment names a record outside the repository or frames a history: %s", c.open, c.decl, l)
			}
		}
	}
}
