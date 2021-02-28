/*
Copyright 2019 The Kubernetes Authors.
Modifications copyright (C) 2020 Alex Neo

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

package scheduler

import (
  "github.com/alexnjh/epsilon/general_purpose_scheduler/k8s.io/kubernetes/pkg/features"
  corev1 "k8s.io/api/core/v1"
  framework "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/v1alpha1"
  utilfeature "k8s.io/apiserver/pkg/util/feature"
)

// Compute pod resource requirements similar to the one used by the resource plugin
func computePodResourceRequest(pod *corev1.Pod) *framework.Resource {
	result := &framework.Resource{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Requests)
	}

	// take max_resource(sum_pod, any_init_container)
	for _, container := range pod.Spec.InitContainers {
		result.SetMaxResource(container.Resources.Requests)
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil && utilfeature.DefaultFeatureGate.Enabled(features.PodOverhead) {
		result.Add(pod.Spec.Overhead)
	}

	return result
}
