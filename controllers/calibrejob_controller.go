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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	calibrev1 "calibre.siemens.com/calibrejob/api/v1"
)

func newPodForCR(cr *calibrev1.CalibreJob) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:    "busybox",
				Image:   "busybox",
				Command: strings.Split(cr.Spec.Command, " "),
			}},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

func timeUntilSchedule(schedule string) (time.Duration, error) {
	now := time.Now().UTC()
	layout := "2006-01-02T15:04:05Z"
	s, err := time.Parse(layout, schedule)
	if err != nil {
		return time.Duration(0), err
	}
	return s.Sub(now), nil
}

type CalibreJobReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=calibre.siemens.com,resources=calibrejobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=calibre.siemens.com,resources=calibrejobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=calibre.siemens.com,resources=calibrejobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

func (r *CalibreJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	instance := &calibrev1.CalibreJob{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.Status.Phase == "" {
		instance.Status.Phase = calibrev1.PhasePending
	}

	switch instance.Status.Phase {
	case calibrev1.PhasePending:
		r.Recorder.Event(instance, "Normal", "PhaseChange", calibrev1.PhasePending)
		d, err := timeUntilSchedule(instance.Spec.Schedule)
		if err != nil {
			return reconcile.Result{}, err
		}
		if d > 0 {
			return reconcile.Result{RequeueAfter: d}, nil
		}
		instance.Status.Phase = calibrev1.PhaseRunning

	case calibrev1.PhaseRunning:
		r.Recorder.Event(instance, "Running", "PhaseChange", calibrev1.PhaseRunning)
		pod := newPodForCR(instance)
		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}
		found := &corev1.Pod{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			err = r.Create(context.TODO(), pod)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err
		} else if found.Status.Phase == corev1.PodFailed || found.Status.Phase == corev1.PodSucceeded {
			instance.Status.Phase = calibrev1.PhaseDone
		} else {
			return reconcile.Result{}, nil
		}

	case calibrev1.PhaseDone:
		r.Recorder.Event(instance, "Done", "PhaseChange", calibrev1.PhaseDone)
		return reconcile.Result{}, nil

	default:
		return reconcile.Result{}, nil
	}

	err = r.Status().Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CalibreJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&calibrev1.CalibreJob{}).
		Owns(&calibrev1.CalibreJob{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
