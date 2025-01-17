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
	"fmt"
	"time"

	"github.com/zufardhiyaulhaq/istio-ratelimit-operator/pkg/global/config"
	"github.com/zufardhiyaulhaq/istio-ratelimit-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	funk "github.com/thoas/go-funk"
	ratelimitv1alpha1 "github.com/zufardhiyaulhaq/istio-ratelimit-operator/api/v1alpha1"
	clientnetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// GlobalRateLimitConfigReconciler reconciles a GlobalRateLimitConfig object
type GlobalRateLimitConfigReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ratelimit.zufardhiyaulhaq.com,resources=globalratelimitconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ratelimit.zufardhiyaulhaq.com,resources=globalratelimitconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ratelimit.zufardhiyaulhaq.com,resources=globalratelimitconfigs/finalizers,verbs=update

func (r *GlobalRateLimitConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Start GlobalRateLimitConfig Reconciler")

	globalRateLimitConfig := &ratelimitv1alpha1.GlobalRateLimitConfig{}
	err := r.Client.Get(ctx, req.NamespacedName, globalRateLimitConfig)
	if err != nil {
		return ctrl.Result{}, nil
	}

	// if GlobalRateLimitConfig refer to RateLimitService, fetch the Kubernetes Service
	// and update the information
	if globalRateLimitConfig.Spec.Ratelimit.Spec.Service.Type == "service" {
		rateLimitService := &ratelimitv1alpha1.RateLimitService{}
		err := r.Client.Get(ctx, types.NamespacedName{
			Name:      globalRateLimitConfig.Spec.Ratelimit.Spec.Service.Name,
			Namespace: req.Namespace,
		}, rateLimitService)
		if err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}

		rateLimitServiceObject := &corev1.Service{}
		err = r.Client.Get(ctx, types.NamespacedName{
			Name:      globalRateLimitConfig.Spec.Ratelimit.Spec.Service.Name,
			Namespace: req.Namespace,
		}, rateLimitServiceObject)
		if err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}

		globalRateLimitConfig.Spec.Ratelimit.Spec.Service.Address = rateLimitServiceObject.Name + "." + rateLimitServiceObject.Namespace + ".svc.cluster.local"
		globalRateLimitConfig.Spec.Ratelimit.Spec.Service.Port = 8081
	}

	log.Info("Build globalratelimitconfig envoyfilters")
	envoyFilters, err := config.NewConfigBuilder().
		SetConfig(*globalRateLimitConfig).
		Build()
	if err != nil {
		return ctrl.Result{}, err
	}

	if len(envoyFilters) == 0 {
		return ctrl.Result{}, fmt.Errorf("empty globalratelimitconfig envoyfilter from builder")
	}

	allVersionEnvoyFilterNames := utils.BuildEnvoyFilterNamesAllVersion(globalRateLimitConfig.Name)
	EnvoyFilterNames := utils.BuildEnvoyFilterNames(globalRateLimitConfig.Name, globalRateLimitConfig.Spec.Selector.IstioVersion)

	deleteEnvoyFilters, _ := funk.DifferenceString(allVersionEnvoyFilterNames, EnvoyFilterNames)
	for _, deleteEnvoyFilterName := range deleteEnvoyFilters {
		deleteEnvoyFilter := &clientnetworking.EnvoyFilter{}
		err := r.Client.Get(ctx, types.NamespacedName{Name: deleteEnvoyFilterName, Namespace: req.Namespace}, deleteEnvoyFilter)
		if err != nil {
			continue
		}

		log.Info("delete unused globalratelimitconfig envoyfilter")
		err = r.Client.Delete(ctx, deleteEnvoyFilter, &client.DeleteOptions{})
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("create globalratelimitconfig envoyfilters")
	for _, envoyFilter := range envoyFilters {
		err := ctrl.SetControllerReference(globalRateLimitConfig, envoyFilter, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}

		log.Info("get globalratelimitconfig envoyfilter")
		createdEnvoyFilter := &clientnetworking.EnvoyFilter{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: envoyFilter.Name, Namespace: envoyFilter.Namespace}, createdEnvoyFilter)
		if err != nil {
			if errors.IsNotFound(err) {
				log.Info("create globalratelimitconfig envoyfilter")
				err := r.Client.Create(ctx, envoyFilter, &client.CreateOptions{})
				if err != nil {
					return ctrl.Result{}, err
				}

				return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
			} else {
				return ctrl.Result{}, err
			}
		}

		if !equality.Semantic.DeepEqual(createdEnvoyFilter.Spec, envoyFilter.Spec) {
			createdEnvoyFilter.Spec = envoyFilter.Spec

			log.Info("update globalratelimitconfig envoyfilter")
			err := r.Client.Update(ctx, createdEnvoyFilter, &client.UpdateOptions{})
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GlobalRateLimitConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ratelimitv1alpha1.GlobalRateLimitConfig{}).
		Complete(r)
}
