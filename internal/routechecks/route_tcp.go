// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright (c) 2024 KhulnaSoft Ltd

package routechecks

import (
	"context"
	"fmt"
	"reflect"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	gateway "github.com/khulnasoft/gateway/internal"
)

type TCPRouteInput struct {
	Ctx      context.Context
	Client   client.Client
	Grants   *gatewayv1beta1.ReferenceGrantList
	TCPRoute *gatewayv1alpha2.TCPRoute

	gateways map[gatewayv1.ParentReference]*gatewayv1.Gateway
}

func (h *TCPRouteInput) SetParentCondition(ref gatewayv1.ParentReference, condition metav1.Condition) {
	// fill in the condition
	condition.LastTransitionTime = metav1.NewTime(time.Now())
	condition.ObservedGeneration = h.TCPRoute.GetGeneration()

	h.mergeStatusConditions(ref, []metav1.Condition{
		condition,
	})
}

func (h *TCPRouteInput) SetAllParentCondition(condition metav1.Condition) {
	// fill in the condition
	condition.LastTransitionTime = metav1.NewTime(time.Now())
	condition.ObservedGeneration = h.TCPRoute.GetGeneration()

	for _, parent := range h.TCPRoute.Spec.ParentRefs {
		h.mergeStatusConditions(parent, []metav1.Condition{
			condition,
		})
	}
}

func (h *TCPRouteInput) mergeStatusConditions(parentRef gatewayv1.ParentReference, updates []metav1.Condition) {
	index := -1
	for i, parent := range h.TCPRoute.Status.RouteStatus.Parents {
		if reflect.DeepEqual(parent.ParentRef, parentRef) {
			index = i
			break
		}
	}
	if index != -1 {
		h.TCPRoute.Status.RouteStatus.Parents[index].Conditions = merge(h.TCPRoute.Status.RouteStatus.Parents[index].Conditions, updates...)
		return
	}
	h.TCPRoute.Status.RouteStatus.Parents = append(h.TCPRoute.Status.RouteStatus.Parents, gatewayv1.RouteParentStatus{
		ParentRef:      parentRef,
		ControllerName: gateway.ControllerName,
		Conditions:     updates,
	})
}

func (h *TCPRouteInput) GetGrants() []gatewayv1beta1.ReferenceGrant {
	return h.Grants.Items
}

func (h *TCPRouteInput) GetNamespace() string {
	return h.TCPRoute.GetNamespace()
}

func (h *TCPRouteInput) GetGVK() schema.GroupVersionKind {
	return gatewayv1alpha2.SchemeGroupVersion.WithKind("TCPRoute")
}

func (h *TCPRouteInput) GetRules() []GenericRule {
	rules := make([]GenericRule, len(h.TCPRoute.Spec.Rules))
	for i, rule := range h.TCPRoute.Spec.Rules {
		rules[i] = &TCPRouteRule{rule}
	}
	return rules
}

func (h *TCPRouteInput) GetClient() client.Client {
	return h.Client
}

func (h *TCPRouteInput) GetContext() context.Context {
	return h.Ctx
}

func (h *TCPRouteInput) GetHostnames() []gatewayv1.Hostname {
	return nil
}

func (h *TCPRouteInput) GetGateway(parent gatewayv1.ParentReference) (*gatewayv1.Gateway, error) {
	if h.gateways == nil {
		h.gateways = make(map[gatewayv1.ParentReference]*gatewayv1.Gateway)
	}
	if gw, exists := h.gateways[parent]; exists {
		return gw, nil
	}

	ns := gateway.NamespaceDerefOr(parent.Namespace, h.GetNamespace())
	gw := &gatewayv1.Gateway{}
	if err := h.Client.Get(h.Ctx, client.ObjectKey{Namespace: ns, Name: string(parent.Name)}, gw); err != nil {
		if !apierrors.IsNotFound(err) {
			// if it is not just a not found error, we should return the error as something is bad
			return nil, fmt.Errorf("error while getting gateway: %w", err)
		}
		// Gateway does not exist skip further checks
		return nil, fmt.Errorf("gateway %q (%q) does not exist: %w", parent.Name, ns, err)
	}

	h.gateways[parent] = gw
	return gw, nil
}

// TCPRouteRule is used to implement the GenericRule interface for TLSRoute
type TCPRouteRule struct {
	Rule gatewayv1alpha2.TCPRouteRule
}

func (t *TCPRouteRule) GetBackendRefs() []gatewayv1.BackendRef {
	return t.Rule.BackendRefs
}
