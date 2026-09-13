// ABOUTME: Clock-domain regression tests: unix client vs monotonic server
// ABOUTME: Covers the aiosendspin case where server timestamps live near boot time
package sync

import (
	"testing"
	"time"
)

func TestClockSyncSyncedPassthrough(t *testing.T) {
	cs := NewClockSync()
	if cs.Synced() {
		t.Error("fresh clock must report unsynced")
	}
	cs.ProcessSyncResponse(1_000_000, 500_000, 500_100, 1_000_200)
	if !cs.Synced() {
		t.Error("expected synced after first sample")
	}
}

// TestClockSyncMonotonicServerDomain pins the Music Assistant case: the
// client clock is unix-epoch microseconds while the server clock is
// monotonic microseconds since boot (~10 days here). A converged filter
// must map server timestamps back near local wall-clock time — the
// scheduler drops anything it places in 1970 as impossibly late.
func TestClockSyncMonotonicServerDomain(t *testing.T) {
	cs := NewClockSync()

	clientNow := time.Now().UnixMicro()
	serverNow := int64(861_714_428_096) // ~10 days monotonic, like aiosendspin

	for i := 0; i < 5; i++ {
		ct := clientNow + int64(i*1_000_000) // 1s apart
		st := serverNow + int64(i*1_000_000)
		cs.ProcessSyncResponse(ct, st+500, st+1500, ct+2000) // ~1ms RTT
	}

	// A chunk timestamped 200ms past a 5s horizon must map near local now+5.2s.
	futureServer := serverNow + 5*1_000_000 + 200_000
	got := cs.ServerToLocalTime(futureServer)
	want := time.UnixMicro(clientNow + 5*1_000_000 + 200_000)
	if d := got.Sub(want); d < -50*time.Millisecond || d > 50*time.Millisecond {
		t.Errorf("ServerToLocalTime(%d) = %v, want near %v (diff %v)",
			futureServer, got, want, d)
	}

	// ServerMicrosNow must track the server domain, not unix epoch.
	if sn := cs.ServerMicrosNow(); abs64(sn-(serverNow+5*1_000_000)) > 5*1_000_000 {
		t.Errorf("ServerMicrosNow() = %d, want near %d", sn, serverNow+5*1_000_000)
	}
}
