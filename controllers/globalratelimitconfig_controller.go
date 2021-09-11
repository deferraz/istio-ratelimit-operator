/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ratelimitv1alpha1 "github.com/zufardhiyaulhaq/istio-ratelimit-operator/api/v1alpha1"
)

// GlobalRateLimitConfigReconciler reconciles a GlobalRateLimitConfig object
type GlobalRateLimitConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ratelimit.zufardhiyaulhaq.com,resources=globalratelimitconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ratelimit.zufardhiyaulhaq.com,resources=globalratelimitconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ratelimit.zufardhiyaulhaq.com,resources=globalratelimitconfigs/finalizers,verbs=update

func (r *GlobalRateLimitConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Start GlobalRateLimitConfig Reconciler")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GlobalRateLimitConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ratelimitv1alpha1.GlobalRateLimitConfig{}).
		Complete(r)
}
