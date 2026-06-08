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
	"sync"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorframeworkv1 "github.com/operator-framework/api/pkg/operators/v1"

	res "github.com/IBM/ibm-licensing-operator/controllers/resources"
	odlm "github.com/IBM/operand-deployment-lifecycle-manager/api/v1alpha1"
)

type opreqState int

const (
	opreqDisabled      opreqState = iota // not running, flag is false
	opreqWaitingForCRD                   // flag is true but OperandRequest CRD absent; polling for it
	opreqActive                          // controller + (maybe) discovery/cleaner running
)

// OperandRequestSubsystem owns the runtime lifecycle of the OperandRequest
// controller, discovery and the OperatorGroup cleaner, and starts/stops them
// in response to features.operandRequestsEnabled without restarting the operator
type OperandRequestSubsystem struct {
	// Immutable dependencies, set once at construction in main.go.
	parentCtx         context.Context // tied to the manager signal context; cancels on shutdown
	mgr               ctrl.Manager
	restConfig        *rest.Config
	scheme            *runtime.Scheme
	operatorNamespace string
	watchNamespaces   []string
	nssSemaphore      chan bool
	log               logr.Logger

	// activate/deactivate/crdExists are indirection points for unit testing the
	// state machine without a real cluster; defaults wired in the constructor
	activate   func(ctx context.Context) error
	deactivate func()
	crdExists  func(list client.ObjectList) (bool, error)

	// crdReconcileInterval controls how often waitForCRDThenActivate polls for
	// the OperandRequest CRD. Same knob the old RestartOnCRDCreation used
	crdReconcileInterval time.Duration

	mu     sync.Mutex
	state  opreqState
	cancel context.CancelFunc // cancels the current activation (controller+cache+goroutines+CRD waiter)
}

// NewOperandRequestSubsystem builds a subsystem with the real activation, deactivation and CRD-probe
// implementations wired in. The subsystem starts in the opreqDisabled state. The first IBMLicensing
// reconcile drives it to the desired state via Sync
func NewOperandRequestSubsystem(
	parentCtx context.Context,
	mgr ctrl.Manager,
	restConfig *rest.Config,
	operatorNamespace string,
	watchNamespaces []string,
	nssSemaphore chan bool,
	log logr.Logger,
) *OperandRequestSubsystem {
	s := &OperandRequestSubsystem{
		parentCtx:         parentCtx,
		mgr:               mgr,
		restConfig:        restConfig,
		scheme:            mgr.GetScheme(),
		operatorNamespace: operatorNamespace,
		watchNamespaces:   watchNamespaces,
		nssSemaphore:      nssSemaphore,
		log:               log,
		state:             opreqDisabled,
	}
	s.activate = s.realActivate
	s.deactivate = func() {} // real teardown happens through context cancellation
	s.crdExists = func(list client.ObjectList) (bool, error) {
		return res.DoesCRDExist(s.mgr.GetAPIReader(), list)
	}

	interval, err := res.GetCrdReconcileInterval()
	if err != nil {
		log.Error(err, "Incorrect CRD reconcile interval set. Defaulting to 300s")
	}
	s.crdReconcileInterval = interval

	return s
}

// Sync reconciles the running subsystem state toward `desired`. It is safe to call on every reconcile: it only acts
// on an actual transition. Returns an error only when starting the subsystem fails in a way worth requeuing
func (s *OperandRequestSubsystem) Sync(desired bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch {
	case !desired && s.state != opreqDisabled:
		s.log.Info("OperandRequest support disabled; stopping controller, discovery and OperatorGroup cleaner")
		s.stopLocked()
		return nil

	case desired && s.state == opreqDisabled:
		return s.startLocked()

	default:
		// desired matches current intent (disabled/disabled, or already
		// active/waiting while enabled) — nothing to do.
		return nil
	}
}

// startLocked probes the OperandRequest CRD and either activates the subsystem immediately or, when the CRD is absent,
// enters opreqWaitingForCRD and launches an in-process waiter
func (s *OperandRequestSubsystem) startLocked() error {
	ctx, cancel := context.WithCancel(s.parentCtx)
	s.cancel = cancel

	crdExists, err := s.crdExists(&odlm.OperandRequestList{})
	if err != nil {
		s.log.Error(err, "Checking OperandRequest CRD existence failed; will retry")
	}
	if !crdExists {
		s.state = opreqWaitingForCRD
		s.log.Info("OperandRequest CRD not present yet; watching for it to appear")
		go s.waitForCRDThenActivate(ctx)
		return nil
	}

	if err := s.activate(ctx); err != nil {
		cancel()
		s.cancel = nil
		return err
	}
	s.state = opreqActive
	return nil
}

// stopLocked tears down the current activation. Canceling the activation context stops the dedicated cache,
// which closes its informer's watch connections to the API server — this is what makes the operator stop issuing
// operandrequests calls
func (s *OperandRequestSubsystem) stopLocked() {
	if s.cancel != nil {
		s.cancel() // stops dedicated cache (closes the watch), controller, discovery, cleaner, CRD waiter
		s.cancel = nil
	}
	s.deactivate()
	s.state = opreqDisabled
}

// waitForCRDThenActivate polls for the OperandRequest CRD until it appears or ctx is canceled; on success it re-takes the lock,
// confirms the subsystem is still waiting (and not canceled), activates and transitions to opreqActive
func (s *OperandRequestSubsystem) waitForCRDThenActivate(ctx context.Context) {
	ticker := time.NewTicker(s.crdReconcileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			exists, err := s.crdExists(&odlm.OperandRequestList{})
			if err != nil {
				s.log.Error(err, "Checking OperandRequest CRD existence failed; will retry")
				continue
			}
			if !exists {
				continue
			}

			s.mu.Lock()
			// Guard against ctx cancellation (disable while waiting) racing the lock.
			if s.state != opreqWaitingForCRD || ctx.Err() != nil {
				s.mu.Unlock()
				return
			}
			if err := s.activate(ctx); err != nil {
				s.log.Error(err, "Activating OperandRequest subsystem failed after CRD appeared; will retry")
				s.mu.Unlock()
				continue
			}
			s.state = opreqActive
			s.log.Info("OperandRequest CRD detected; OperandRequest support activated")
			s.mu.Unlock()
			return
		}
	}
}

// realActivate is the default activate implementation: it stands up an unmanaged OperandRequest controller backed by
// its own cancelable cache, plus discovery and the OperatorGroup cleaner. Canceling ctx stops all of them
func (s *OperandRequestSubsystem) realActivate(ctx context.Context) error {
	opreqCache, err := cache.New(s.restConfig, cache.Options{Scheme: s.scheme})
	if err != nil {
		return err
	}
	go func() {
		if err := opreqCache.Start(ctx); err != nil {
			s.log.Error(err, "OperandRequest dedicated cache stopped with error")
		}
	}()

	reconciler := &OperandRequestReconciler{
		Client:            s.mgr.GetClient(),
		Reader:            s.mgr.GetAPIReader(),
		Log:               s.log,
		Scheme:            s.scheme,
		OperatorNamespace: s.operatorNamespace,
	}

	c, err := controller.NewUnmanaged("operandrequest-controller", controller.Options{
		Reconciler:         reconciler,
		SkipNameValidation: ptr.To(true),
		LogConstructor:     func(*reconcile.Request) logr.Logger { return reconciler.Log },
	})
	if err != nil {
		return err
	}

	src := source.Kind(
		opreqCache,
		client.Object(&odlm.OperandRequest{}),
		&handler.EnqueueRequestForObject{},
		ignoreDeletionPredicate(),
	)
	if err := c.Watch(src); err != nil {
		return err
	}
	go func() {
		if err := c.Start(ctx); err != nil {
			s.log.Error(err, "OperandRequest controller stopped with error")
		}
	}()

	s.startDiscoveryAndCleaner(ctx)
	return nil
}

// startDiscoveryAndCleaner starts the discovery poll-loop and the OperatorGroup cleaner under ctx, applying the same
// Namespace Scope Operator / OperatorGroup CRD gating as the old startup path
func (s *OperandRequestSubsystem) startDiscoveryAndCleaner(ctx context.Context) {
	// In Cloud Pak 2.0/3.0 coexistence scenario, License Service Operator 4.x.x leverages Namespace Scope Operator and must not modify OperatorGroup.
	isNssActive, err := res.IsNamespaceScopeOperatorAvailable(ctx, s.mgr.GetAPIReader(), s.operatorNamespace)
	if err != nil {
		s.log.Error(err, "Error occurred while detecting Namespace Scope Operator")
	}
	if isNssActive {
		s.log.Info("Namespace Scope ConfigMap detected. operandrequest-discovery disabled")
		return
	}

	// On clusters without OLM installed, skip both operandrequest-discovery and the stale-namespace cleanup task,
	// otherwise every run produces a "no matches for kind OperatorGroup" error
	operatorGroupCRDExists, err := s.crdExists(&operatorframeworkv1.OperatorGroupList{})
	if err != nil {
		s.log.Error(err, "An error occurred while checking for OperatorGroup CRD existence. operandrequest-discovery and operatorgroup-namespaces-watcher will not be started")
	}
	if !operatorGroupCRDExists {
		s.log.Info("OperatorGroup CRD not found in cluster. operandrequest-discovery and operatorgroup-namespaces-watcher disabled")
		return
	}

	discLog := s.log.WithName("operandrequest-discovery")
	go DiscoverOperandRequests(ctx, &discLog, s.mgr.GetClient(), s.mgr.GetAPIReader(), s.watchNamespaces, s.nssSemaphore)

	cleanerLog := s.log.WithName("operatorgroup-namespaces-watcher")
	go RunRemoveStaleNamespacesFromOperatorGroupTask(ctx, &cleanerLog, s.mgr.GetClient(), s.mgr.GetAPIReader())
}
