/*
Copyright 2019 Polyaxon, Inc.

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

package utils

import (
	"reflect"

	corev1alpha1 "github.com/polyaxon/polyaxon-operator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultMaxRetries for PlxJobs
	DefaultMaxRetries = 1
	// DefaultRestartPolicy for PlxJobs
	DefaultRestartPolicy = "Never"
)

// CopyJobFields copies the owned fields from one Job to another
// Returns true if the fields copied from don't match to.
func CopyJobFields(from, to *batchv1.Job) bool {
	requireUpdate := false
	for k, v := range to.Labels {
		if from.Labels[k] != v {
			requireUpdate = true
		}
	}
	to.Labels = from.Labels

	if !reflect.DeepEqual(to.Spec.Template.Spec, from.Spec.Template.Spec) {
		requireUpdate = true
	}
	to.Spec.Template.Spec = from.Spec.Template.Spec

	return requireUpdate
}

// GetPlxJobCondition returns PolyaxonBaseJobCondition given a JobCondition
func GetPlxJobCondition(jc batchv1.JobCondition) corev1alpha1.PolyaxonBaseJobCondition {
	var jtype = ""
	var reason = jc.Reason
	var msg = jc.Message

	if jc.Type == "ReplicaFailure" {
		jtype = "warning"
	} else if jc.Type == "Progressing" {
		jtype = "starting"
	} else if jc.Type == "Available" {
		if jc.Status == "True" {
			jtype = "running"
		} else {
			jtype = "warning"
		}
	}

	newCondition := corev1alpha1.PolyaxonBaseJobCondition{
		Type:          jtype,
		LastProbeTime: metav1.Now(),
		Reason:        reason,
		Message:       msg,
	}
	return newCondition
}

// GeneratePlxJob returns a deployment given a PolyaxonBaseJobSpec
func GeneratePlxJob(
	name string,
	namespace string,
	labels map[string]string,
	maxRetries *int32,
	podSpec corev1.PodSpec,
) *batchv1.Job {
	backoffLimit := int32(DefaultMaxRetries)
	if maxRetries != nil {
		backoffLimit = *maxRetries
	}

	if podSpec.RestartPolicy == "" {
		podSpec.RestartPolicy = DefaultRestartPolicy
	}

	PlxJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
				Spec:       podSpec,
			},
		},
	}
	// copy all of the labels to the pod including poddefault related labels
	l := &PlxJob.Spec.Template.ObjectMeta.Labels
	for k, v := range labels {
		(*l)[k] = v
	}

	return PlxJob
}