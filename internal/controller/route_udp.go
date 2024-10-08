// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright (c) 2024 KhulnaSoft Ltd

package controller

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	gateway "github.com/khulnasoft/gateway/internal"
	"github.com/khulnasoft/gateway/internal/routechecks"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=udproutes,verbs=get;list;watch
// +kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=udproutes/status,verbs=patch;update

type UDPRouteReconciler struct {
	client.Client

	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

var _ reconcile.Reconciler = (*UDPRouteReconciler)(nil)

// SetupWithManager sets up the controller with the Manager.
func (r *UDPRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gatewayv1alpha2.UDPRoute{}, backendServiceIndex, func(o client.Object) []string {
		route, ok := o.(*gatewayv1alpha2.UDPRoute)
		if !ok {
			return nil
		}
		var backendServices []string
		for _, rule := range route.Spec.Rules {
			for _, backend := range rule.BackendRefs {
				backendServiceName, err := gateway.GetBackendServiceName(backend.BackendObjectReference)
				if err != nil {
					mgr.GetLogger().WithValues(
						"controller", "udp-route",
						"resource", client.ObjectKeyFromObject(o),
					).Error(err, "Failed to get backend service name")
					continue
				}

				backendServices = append(backendServices, types.NamespacedName{
					Namespace: gateway.NamespaceDerefOr(backend.Namespace, route.Namespace),
					Name:      backendServiceName,
				}.String())
			}
		}
		return backendServices
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gatewayv1alpha2.UDPRoute{}, gatewayIndex, func(o client.Object) []string {
		route, ok := o.(*gatewayv1alpha2.UDPRoute)
		if !ok {
			return nil
		}
		var gateways []string
		for _, parent := range route.Spec.ParentRefs {
			if !gateway.IsGateway(parent) {
				continue
			}
			gateways = append(gateways, types.NamespacedName{
				Namespace: gateway.NamespaceDerefOr(parent.Namespace, route.Namespace),
				Name:      string(parent.Name),
			}.String())
		}
		return gateways
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1alpha2.UDPRoute{}).
		Watches(&corev1.Service{}, r.enqueueRequestForBackendService()).
		Watches(&gatewayv1beta1.ReferenceGrant{}, r.enqueueRequestForReferenceGrant()).
		Watches(
			&gatewayv1.Gateway{},
			r.enqueueRequestForGateway(),
			builder.WithPredicates(predicate.NewPredicateFuncs(r.hasMatchingController(ctx))),
		).
		Complete(r)
}

func (r *UDPRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	original := &gatewayv1alpha2.UDPRoute{}
	if err := r.Client.Get(ctx, req.NamespacedName, original); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to get UDPRoute")
		return ctrl.Result{}, err
	}

	// Check if the UDPRoute is being deleted.
	if original.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	route := original.DeepCopy()

	grants := &gatewayv1beta1.ReferenceGrantList{}
	if err := r.Client.List(ctx, grants); err != nil {
		return r.handleReconcileErrorWithStatus(ctx, fmt.Errorf("failed to retrieve reference grants: %w", err), original, route)
	}

	// input for the validators
	i := &routechecks.UDPRouteInput{
		Ctx:      ctx,
		Client:   r.Client,
		Grants:   grants,
		UDPRoute: route,
	}

	// gateway validators
	for _, parent := range route.Spec.ParentRefs {
		// set acceptance to okay, this wil be overwritten in checks if needed
		i.SetParentCondition(parent, metav1.Condition{
			Type:    string(gatewayv1.RouteConditionAccepted),
			Status:  metav1.ConditionTrue,
			Reason:  string(gatewayv1.RouteReasonAccepted),
			Message: "Accepted UDPRoute",
		})

		// set status to okay, this wil be overwritten in checks if needed
		i.SetAllParentCondition(metav1.Condition{
			Type:    string(gatewayv1.RouteConditionResolvedRefs),
			Status:  metav1.ConditionTrue,
			Reason:  string(gatewayv1.RouteReasonResolvedRefs),
			Message: "Service reference is valid",
		})

		// run the actual validators
		for _, fn := range []routechecks.CheckGatewayFunc{
			routechecks.CheckGatewayAllowedForNamespace,
			routechecks.CheckGatewayRouteKindAllowed,
			routechecks.CheckGatewayMatchingPorts,
			routechecks.CheckGatewayMatchingSection,
		} {
			continueCheck, err := fn(i, parent)
			if err != nil {
				return r.handleReconcileErrorWithStatus(ctx, fmt.Errorf("failed to apply Gateway check: %w", err), original, route)
			}
			if !continueCheck {
				break
			}
		}
	}

	for _, fn := range []routechecks.CheckRuleFunc{
		routechecks.CheckAgainstCrossNamespaceBackendReferences,
		routechecks.CheckBackend,
		routechecks.CheckBackendIsExistingService,
	} {
		continueCheck, err := fn(i)
		if err != nil {
			return r.handleReconcileErrorWithStatus(ctx, fmt.Errorf("failed to apply Backend check: %w", err), original, route)
		}
		if !continueCheck {
			break
		}
	}

	if err := r.updateStatus(ctx, original, route); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update UDPRoute status: %w", err)
	}

	log.Info("Reconciled UDPRoute")
	return ctrl.Result{}, nil
}

// enqueueRequestForBackendService .
// TODO: document
func (r *UDPRouteReconciler) enqueueRequestForBackendService() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(r.enqueueFromIndex(backendServiceIndex))
}

// enqueueRequestForGateway .
// TODO: document
func (r *UDPRouteReconciler) enqueueRequestForGateway() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(r.enqueueFromIndex(gatewayIndex))
}

// enqueueRequestForReferenceGrant .
// TODO: document
func (r *UDPRouteReconciler) enqueueRequestForReferenceGrant() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(r.enqueueAll())
}

// enqueueFromIndex .
// TODO: document
func (r *UDPRouteReconciler) enqueueFromIndex(index string) handler.MapFunc {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		return r.enqueue(ctx, &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(index, client.ObjectKeyFromObject(o).String()),
		})
	}
}

// enqueueAll .
// TODO
func (r *UDPRouteReconciler) enqueueAll() handler.MapFunc {
	return func(ctx context.Context, _ client.Object) []reconcile.Request {
		return r.enqueue(ctx)
	}
}

// enqueue .
// TODO
func (r *UDPRouteReconciler) enqueue(ctx context.Context, opts ...client.ListOption) []reconcile.Request {
	log := log.FromContext(ctx)

	list := &gatewayv1alpha2.UDPRouteList{}
	if err := r.Client.List(ctx, list, opts...); err != nil {
		log.Error(err, "Failed to get UDPRoute")
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(list.Items))
	for i, item := range list.Items {
		route := types.NamespacedName{
			Namespace: item.GetNamespace(),
			Name:      item.GetName(),
		}
		requests[i] = reconcile.Request{
			NamespacedName: route,
		}
		log.Info("Enqueued UDPRoute for resource", "route", route)
	}
	return requests
}

// hasMatchingController .
// TODO
func (r *UDPRouteReconciler) hasMatchingController(ctx context.Context) func(object client.Object) bool {
	return hasMatchingController(ctx, r.Client)
}

// updateStatus .
// TODO
func (r *UDPRouteReconciler) updateStatus(ctx context.Context, original, new *gatewayv1alpha2.UDPRoute) error {
	oldStatus := original.Status.DeepCopy()
	newStatus := new.Status.DeepCopy()

	opts := cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")
	if cmp.Equal(oldStatus, newStatus, opts) {
		return nil
	}
	return r.Client.Status().Update(ctx, new)
}

// handleReconcileErrorWithStatus .
// TODO
func (r *UDPRouteReconciler) handleReconcileErrorWithStatus(ctx context.Context, reconcileErr error, original, modified *gatewayv1alpha2.UDPRoute) (ctrl.Result, error) {
	if err := r.updateStatus(ctx, original, modified); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update UDPRoute status while handling the reconcile error %w: %w", reconcileErr, err)
	}
	return ctrl.Result{}, reconcileErr
}
