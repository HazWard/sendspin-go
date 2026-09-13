// ABOUTME: Shared-clock tests: receivers reuse one ClockSync per server
// ABOUTME: Samples survive reconnects; a server change resets the clock
package sendspin

import "testing"

func TestPlayerBuildReceiverSharesClock(t *testing.T) {
	p, err := NewPlayer(PlayerConfig{
		ServerAddr: "localhost:8927",
		PlayerName: "clock-share-test",
		Output:     &fakeVolumeOutput{},
	})
	if err != nil {
		t.Fatalf("NewPlayer: %v", err)
	}

	r1, err := p.buildReceiver("localhost:8927")
	if err != nil {
		t.Fatalf("buildReceiver: %v", err)
	}
	r2, err := p.buildReceiver("localhost:8927")
	if err != nil {
		t.Fatalf("buildReceiver: %v", err)
	}
	if r1.clockSync != r2.clockSync {
		t.Error("receivers for the same server must share one ClockSync")
	}

	// One sync sample through the shared clock...
	r1.clockSync.ProcessSyncResponse(1_000_000, 500_000, 500_100, 1_000_200)
	if !r1.clockSync.Synced() {
		t.Fatal("shared clock should be synced after a sample")
	}

	// ...must survive a rebuild for the same server...
	r3, err := p.buildReceiver("localhost:8927")
	if err != nil {
		t.Fatalf("buildReceiver: %v", err)
	}
	if !r3.clockSync.Synced() {
		t.Error("same-server rebuild must keep accumulated sync samples")
	}

	// ...but a different server is a different clock domain and resets it.
	r4, err := p.buildReceiver("otherhost:8927")
	if err != nil {
		t.Fatalf("buildReceiver: %v", err)
	}
	if r4.clockSync != r1.clockSync {
		t.Error("must keep sharing the same clock instance")
	}
	if r4.clockSync.Synced() {
		t.Error("server change must reset the shared clock")
	}
}
