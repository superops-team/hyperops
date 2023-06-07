package event

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const (
	ETMainSaidHello   = Type("main:SaidHello")
	ETMainOpSucceeded = Type("main:OperationSucceeded")
	ETMainOpFailed    = Type("main:OperationFailed")
)

func Example() {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	bus := NewBus(ctx)

	makeDoneHandler := func(label string) Handler {
		return func(ctx context.Context, e Event) error {
			fmt.Printf("%s handler called\n", label)
			return nil
		}
	}

	bus.SubscribeTypes(makeDoneHandler("first"), ETMainSaidHello, ETMainOpSucceeded)
	bus.SubscribeTypes(makeDoneHandler("second"), ETMainSaidHello)
	bus.SubscribeTypes(makeDoneHandler("third"), ETMainSaidHello)

	_ = bus.Publish(ctx, ETMainSaidHello, "hello")
	_ = bus.Publish(ctx, ETMainOpSucceeded, "operation worked!")

	// Output: first handler called
	// second handler called
	// third handler called
	// first handler called
}

func TestEventSubscribeTypes(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	counter := 0
	prevNowFunc := NowFunc
	NowFunc = func() time.Time {
		counter++
		return time.Unix(int64(1234567000+counter), 0)
	}
	defer func() { NowFunc = prevNowFunc }()

	bus := NewBus(ctx)

	var gotNumEvents int
	var gotTimestamp int64
	var gotPayload interface{}
	handler := func(ctx context.Context, e Event) error {
		gotNumEvents++
		gotTimestamp = e.Timestamp
		gotPayload = e.Payload
		return nil
	}

	bus.SubscribeTypes(handler, ETMainSaidHello)

	_ = bus.Publish(ctx, ETMainOpFailed, "ignore me")
	_ = bus.Publish(ctx, ETMainSaidHello, "hello")
	_ = bus.Publish(ctx, ETMainOpSucceeded, "ignore me too")

	// Got 1 event
	expectNum := 1
	if diff := cmp.Diff(expectNum, gotNumEvents); diff != "" {
		t.Errorf("num events (-want +got):\n%s", diff)
	}
	// Timestamp has 2 seconds from the initial value
	expectTs := int64(1234567002000000000)
	if diff := cmp.Diff(expectTs, gotTimestamp); diff != "" {
		t.Errorf("timestamp (-want +got):\n%s", diff)
	}
	// Only type we care about sets the payload value
	expectPayload := "hello"
	if diff := cmp.Diff(expectPayload, gotPayload); diff != "" {
		t.Errorf("payload (-want +got):\n%s", diff)
	}
}

func TestEventSubscribeID(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	counter := 0
	prevNowFunc := NowFunc
	NowFunc = func() time.Time {
		counter++
		return time.Unix(int64(1234567000+counter), 0)
	}
	defer func() { NowFunc = prevNowFunc }()

	bus := NewBus(ctx)

	var gotNumEvents int
	var gotTimestamp int64
	var gotPayload interface{}
	handler := func(ctx context.Context, e Event) error {
		gotNumEvents++
		gotTimestamp = e.Timestamp
		gotPayload = e.Payload
		return nil
	}

	bus.SubscribeID(handler, "789")

	_ = bus.PublishID(ctx, ETMainSaidHello, "123", "hi1")
	_ = bus.PublishID(ctx, ETMainSaidHello, "456", "hi2")
	_ = bus.PublishID(ctx, ETMainSaidHello, "789", "hi3")
	_ = bus.PublishID(ctx, ETMainSaidHello, "321", "hi4")

	// Got 1 event
	expectNum := 1
	if diff := cmp.Diff(expectNum, gotNumEvents); diff != "" {
		t.Errorf("num events (-want +got):\n%s", diff)
	}
	// Timestamp has 3 seconds from the initial value
	expectTs := int64(1234567003000000000)
	if diff := cmp.Diff(expectTs, gotTimestamp); diff != "" {
		t.Errorf("timestamp (-want +got):\n%s", diff)
	}
	// Only id we care about sets the payload value
	expectPayload := "hi3"
	if diff := cmp.Diff(expectPayload, gotPayload); diff != "" {
		t.Errorf("payload (-want +got):\n%s", diff)
	}
}

func TestEventSubscribeAll(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	counter := 0
	prevNowFunc := NowFunc
	NowFunc = func() time.Time {
		counter++
		return time.Unix(int64(1234567000+counter), 0)
	}
	defer func() { NowFunc = prevNowFunc }()

	bus := NewBus(ctx)

	var gotNumEvents int
	handler := func(ctx context.Context, e Event) error {
		gotNumEvents++
		return nil
	}

	bus.SubscribeAll(handler)

	_ = bus.Publish(ctx, ETMainOpFailed, "ignore me")
	_ = bus.Publish(ctx, ETMainSaidHello, "hello")
	_ = bus.Publish(ctx, ETMainOpSucceeded, "ignore me too")
	_ = bus.PublishID(ctx, ETMainSaidHello, "123", "hi1")

	// Got all 4 events
	expectNum := 4
	if diff := cmp.Diff(expectNum, gotNumEvents); diff != "" {
		t.Errorf("num events (-want +got):\n%s", diff)
	}
}
