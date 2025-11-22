package baggageprocessor

import (
	"testing"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/baggage"
	"go.uber.org/zap"
)

func newTestProcessor(cfg *Config) *baggageProcessor {
	if cfg == nil {
		cfg = &Config{AttributePrefix: "baggage.", MaxBaggageSize: 8192}
	}
	return newBaggageProcessor(zap.NewNop(), cfg)
}

func TestExtractBaggage_DefaultPrefixAndProperties(t *testing.T) {
	// Prepare baggage with member and properties
	p1, _ := baggage.NewKeyValueProperty("source", "test")
	p2, _ := baggage.NewKeyValueProperty("env", "dev")
	m, err := baggage.NewMember("user.id", "123", p1, p2)
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}
	bg, err := baggage.New(m)
	if err != nil {
		t.Fatalf("failed to create baggage: %v", err)
	}

	attrs := pcommon.NewMap()
	bp := newTestProcessor(&Config{AttributePrefix: "baggage.", Actions: nil})

	action := Action{Key: "user.id", Action: Extract, FromContext: true}
	if err := bp.extractBaggage(action, attrs, bg); err != nil {
		t.Fatalf("extractBaggage returned error: %v", err)
	}

	// Assert attribute value with default prefix
	v, ok := attrs.Get("baggage.user.id")
	if !ok || v.AsString() != "123" {
		t.Fatalf("expected attribute baggage.user.id=123, got ok=%v val=%v", ok, v.AsString())
	}
	// Assert properties attribute exists and contains both properties
	pv, ok := attrs.Get("baggage.user.id_properties")
	if !ok {
		t.Fatalf("expected properties attribute to be set")
	}
	got := pv.AsString()
	if !(containsAll(got, []string{"source=test", "env=dev"})) {
		t.Fatalf("expected properties string to contain both properties, got %q", got)
	}
}

func TestExtractBaggage_ToAttributeOverride(t *testing.T) {
	m, _ := baggage.NewMember("session.id", "s-42")
	bg, _ := baggage.New(m)

	attrs := pcommon.NewMap()
	bp := newTestProcessor(nil)
	action := Action{Key: "session.id", Action: Extract, FromContext: true, ToAttribute: "custom.session.id"}
	if err := bp.extractBaggage(action, attrs, bg); err != nil {
		t.Fatalf("extractBaggage returned error: %v", err)
	}
	v, ok := attrs.Get("custom.session.id")
	if !ok || v.AsString() != "s-42" {
		t.Fatalf("expected custom.session.id to be s-42, got ok=%v val=%v", ok, v.AsString())
	}
}

func TestInjectBaggage_FromAttributeWithProperties(t *testing.T) {
	attrs := pcommon.NewMap()
	attrs.PutStr("service.version", "1.2.3")

	cfg := &Config{DropInvalidBaggage: false}
	bp := newTestProcessor(cfg)
	bg, _ := baggage.New()

	action := Action{Key: "service.version", Action: Inject, FromAttribute: "service.version", Properties: map[string]string{"type": "semantic", "format": "semver"}}
	if err := bp.injectBaggage(action, attrs, &bg); err != nil {
		t.Fatalf("injectBaggage returned error: %v", err)
	}

	mem := bg.Member("service.version")
	if mem.Key() == "" || mem.Value() != "1.2.3" {
		t.Fatalf("expected baggage member service.version=1.2.3, got %q=%q", mem.Key(), mem.Value())
	}
	// Ensure properties are present
	if len(mem.Properties()) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(mem.Properties()))
	}
}

func TestInjectBaggage_InvalidPropertyBehavior(t *testing.T) {
	attrs := pcommon.NewMap()
	attrs.PutStr("foo", "bar")

	// invalid property key (contains space) should error when DropInvalidBaggage=false
	bp := newTestProcessor(&Config{DropInvalidBaggage: false})
	bg, _ := baggage.New()
	action := Action{Key: "k", Action: Inject, FromAttribute: "foo", Properties: map[string]string{"bad key": "x"}}
	if err := bp.injectBaggage(action, attrs, &bg); err == nil {
		t.Fatalf("expected error for invalid property when DropInvalidBaggage=false")
	}

	// With DropInvalidBaggage=true, property is skipped and member still added
	bp2 := newTestProcessor(&Config{DropInvalidBaggage: true})
	bg2, _ := baggage.New()
	if err := bp2.injectBaggage(action, attrs, &bg2); err != nil {
		t.Fatalf("did not expect error with DropInvalidBaggage=true, got %v", err)
	}
	mem := bg2.Member("k")
	if mem.Key() == "" || mem.Value() != "bar" {
		t.Fatalf("expected member k=bar to be added when skipping invalid property")
	}
}

func TestUpdateAndUpsertBehavior(t *testing.T) {
	// Existing baggage with key "a"=old
	m, _ := baggage.NewMember("a", "old")
	existing, _ := baggage.New(m)

	attrs := pcommon.NewMap()
	attrs.PutStr("newval", "new")
	bp := newTestProcessor(nil)

	// Update should change existing member
	bg1 := existing
	if err := bp.updateBaggage(Action{Key: "a", Action: Update, FromAttribute: "newval"}, attrs, &bg1, false); err != nil {
		t.Fatalf("updateBaggage returned error: %v", err)
	}
	if v := bg1.Member("a").Value(); v != "new" {
		t.Fatalf("expected updated value 'new', got %q", v)
	}

	// Update for missing key should be a no-op
	bg2, _ := baggage.New()
	if err := bp.updateBaggage(Action{Key: "missing", Action: Update, FromAttribute: "newval"}, attrs, &bg2, false); err != nil {
		t.Fatalf("updateBaggage returned error: %v", err)
	}
	if mem := bg2.Member("missing"); mem.Key() != "" {
		t.Fatalf("expected no member to be added for update when key missing")
	}

	// Upsert for missing key should add
	if err := bp.updateBaggage(Action{Key: "missing", Action: Upsert, FromAttribute: "newval"}, attrs, &bg2, true); err != nil {
		t.Fatalf("upsert returned error: %v", err)
	}
	if mem := bg2.Member("missing"); mem.Key() == "" || mem.Value() != "new" {
		t.Fatalf("expected upsert to add missing=new, got %q=%q", mem.Key(), mem.Value())
	}
}

func TestDeleteBaggage(t *testing.T) {
	m1, _ := baggage.NewMember("k1", "v1")
	m2, _ := baggage.NewMember("k2", "v2")
	bg, _ := baggage.New(m1, m2)

	bp := newTestProcessor(nil)
	if err := bp.deleteBaggage(Action{Key: "k1", Action: Delete}, &bg); err != nil {
		t.Fatalf("deleteBaggage returned error: %v", err)
	}
	if mem := bg.Member("k1"); mem.Key() != "" {
		t.Fatalf("expected k1 to be deleted")
	}
	if mem := bg.Member("k2"); mem.Key() == "" {
		t.Fatalf("expected k2 to remain")
	}
}

func TestMaxBaggageSizeEnforced(t *testing.T) {
	// Create a very large value to exceed the size limit
	large := make([]byte, 5000)
	for i := range large {
		large[i] = 'a'
	}
	largeVal := string(large)

	attrs := pcommon.NewMap()
	attrs.PutStr("big", largeVal)

	// Set small max size so that adding key will exceed
	bp := newTestProcessor(&Config{MaxBaggageSize: 100, DropInvalidBaggage: false})
	bg, _ := baggage.New()
	err := bp.injectBaggage(Action{Key: "big", Action: Inject, FromAttribute: "big"}, attrs, &bg)
	if err == nil {
		t.Fatalf("expected error when baggage exceeds max size with DropInvalidBaggage=false")
	}

	// When DropInvalidBaggage=true, should skip without error and not modify baggage
	bp2 := newTestProcessor(&Config{MaxBaggageSize: 100, DropInvalidBaggage: true})
	bg2, _ := baggage.New()
	if err := bp2.injectBaggage(Action{Key: "big", Action: Inject, FromAttribute: "big"}, attrs, &bg2); err != nil {
		t.Fatalf("did not expect error when DropInvalidBaggage=true: %v", err)
	}
	if mem := bg2.Member("big"); mem.Key() != "" {
		t.Fatalf("expected no member to be added when exceeding size and dropping invalid")
	}
}

// containsAll reports whether s contains all substrings in subs.
func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

// Simple substring search to avoid importing strings to reduce noise
func indexOf(s, sub string) int {
	n := len(s)
	m := len(sub)
	if m == 0 {
		return 0
	}
	for i := 0; i <= n-m; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}
