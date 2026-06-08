//
// Copyright 2026 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package controllers

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// subsystemHarness wires an OperandRequestSubsystem with injectable activate /
// deactivate / crdExists hooks so the state machine can be exercised without a
// real cluster.
type subsystemHarness struct {
	s               *OperandRequestSubsystem
	activateCount   atomic.Int32
	deactivateCount atomic.Int32
	crdPresent      atomic.Bool
	lastActivateCtx atomic.Value // context.Context passed to the most recent activate
}

func newSubsystemHarness() *subsystemHarness {
	h := &subsystemHarness{}
	h.s = &OperandRequestSubsystem{
		parentCtx:            context.Background(),
		log:                  logr.Discard(),
		state:                opreqDisabled,
		crdReconcileInterval: 5 * time.Millisecond,
	}
	h.s.activate = func(ctx context.Context) error {
		h.lastActivateCtx.Store(ctx)
		h.activateCount.Add(1)
		return nil
	}
	h.s.deactivate = func() {
		h.deactivateCount.Add(1)
	}
	h.s.crdExists = func(client.ObjectList) (bool, error) {
		return h.crdPresent.Load(), nil
	}
	return h
}

func (h *subsystemHarness) state() opreqState {
	h.s.mu.Lock()
	defer h.s.mu.Unlock()
	return h.s.state
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("condition not met within %s", timeout)
}

func TestSyncDisabledStaysDisabled(t *testing.T) {
	h := newSubsystemHarness()

	if err := h.s.Sync(false); err != nil {
		t.Fatalf("Sync(false) returned error: %v", err)
	}
	if got := h.activateCount.Load(); got != 0 {
		t.Errorf("activate called %d times, want 0", got)
	}
	if st := h.state(); st != opreqDisabled {
		t.Errorf("state = %d, want opreqDisabled", st)
	}
}

func TestSyncEnableWithCRDActivatesOnce(t *testing.T) {
	h := newSubsystemHarness()
	h.crdPresent.Store(true)

	if err := h.s.Sync(true); err != nil {
		t.Fatalf("Sync(true) returned error: %v", err)
	}
	if got := h.activateCount.Load(); got != 1 {
		t.Fatalf("activate called %d times, want 1", got)
	}
	if st := h.state(); st != opreqActive {
		t.Fatalf("state = %d, want opreqActive", st)
	}

	// Repeated Sync(true) is idempotent — activate must not be called again.
	if err := h.s.Sync(true); err != nil {
		t.Fatalf("second Sync(true) returned error: %v", err)
	}
	if got := h.activateCount.Load(); got != 1 {
		t.Errorf("activate called %d times after idempotent Sync(true), want 1", got)
	}
}

func TestSyncDisableStopsActiveSubsystem(t *testing.T) {
	h := newSubsystemHarness()
	h.crdPresent.Store(true)

	if err := h.s.Sync(true); err != nil {
		t.Fatalf("Sync(true) returned error: %v", err)
	}
	activationCtx := h.lastActivateCtx.Load().(context.Context)

	if err := h.s.Sync(false); err != nil {
		t.Fatalf("Sync(false) returned error: %v", err)
	}
	if st := h.state(); st != opreqDisabled {
		t.Errorf("state = %d, want opreqDisabled", st)
	}
	if got := h.deactivateCount.Load(); got != 1 {
		t.Errorf("deactivate called %d times, want 1", got)
	}
	if activationCtx.Err() == nil {
		t.Error("activation context was not cancelled on disable")
	}
}

func TestSyncRapidFlipsAreBalanced(t *testing.T) {
	h := newSubsystemHarness()
	h.crdPresent.Store(true)

	var ctxs []context.Context
	for i := 0; i < 3; i++ {
		if err := h.s.Sync(true); err != nil {
			t.Fatalf("Sync(true) iteration %d: %v", i, err)
		}
		ctxs = append(ctxs, h.lastActivateCtx.Load().(context.Context))
		if err := h.s.Sync(false); err != nil {
			t.Fatalf("Sync(false) iteration %d: %v", i, err)
		}
	}

	if a, d := h.activateCount.Load(), h.deactivateCount.Load(); a != d || a != 3 {
		t.Errorf("activate=%d deactivate=%d, want both 3", a, d)
	}
	if st := h.state(); st != opreqDisabled {
		t.Errorf("state = %d, want opreqDisabled", st)
	}
	// No leaked activation: every activation context is cancelled.
	for i, c := range ctxs {
		if c.Err() == nil {
			t.Errorf("activation context %d not cancelled", i)
		}
	}
}

func TestSyncEnableWaitsForCRDThenActivates(t *testing.T) {
	h := newSubsystemHarness()
	// CRD absent at enable time.

	if err := h.s.Sync(true); err != nil {
		t.Fatalf("Sync(true) returned error: %v", err)
	}
	if st := h.state(); st != opreqWaitingForCRD {
		t.Fatalf("state = %d, want opreqWaitingForCRD", st)
	}
	if got := h.activateCount.Load(); got != 0 {
		t.Fatalf("activate called %d times while waiting for CRD, want 0", got)
	}

	// CRD appears — the waiter should activate on the next poll.
	h.crdPresent.Store(true)
	waitFor(t, 2*time.Second, func() bool { return h.activateCount.Load() == 1 })
	waitFor(t, time.Second, func() bool { return h.state() == opreqActive })
}

func TestSyncDisableWhileWaitingStopsWaiter(t *testing.T) {
	h := newSubsystemHarness()
	// CRD absent — enable enters the waiting state.

	if err := h.s.Sync(true); err != nil {
		t.Fatalf("Sync(true) returned error: %v", err)
	}
	if st := h.state(); st != opreqWaitingForCRD {
		t.Fatalf("state = %d, want opreqWaitingForCRD", st)
	}

	if err := h.s.Sync(false); err != nil {
		t.Fatalf("Sync(false) returned error: %v", err)
	}
	if st := h.state(); st != opreqDisabled {
		t.Fatalf("state = %d, want opreqDisabled", st)
	}

	// Even if the CRD now appears, the cancelled waiter must not activate.
	h.crdPresent.Store(true)
	time.Sleep(50 * time.Millisecond)
	if got := h.activateCount.Load(); got != 0 {
		t.Errorf("activate called %d times after disable-while-waiting, want 0", got)
	}
}
