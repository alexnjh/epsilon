/*
Copyright 2015 The Kubernetes Authors.

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

package util

import (
	"fmt"
	"k8s.io/klog"
	v1 "k8s.io/api/core/v1"
  "k8s.io/apimachinery/pkg/labels"
	v1helper "scheduler_unit/k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

// CheckNodeAffinity looks at the PV node affinity, and checks if the node has the same corresponding labels
// This ensures that we don't mount a volume that doesn't belong to this node
func CheckNodeAffinity(pv *v1.PersistentVolume, nodeLabels map[string]string) error {
	return checkVolumeNodeAffinity(pv, nodeLabels)
}

func checkVolumeNodeAffinity(pv *v1.PersistentVolume, nodeLabels map[string]string) error {
	if pv.Spec.NodeAffinity == nil {
		return nil
	}

	if pv.Spec.NodeAffinity.Required != nil {
		terms := pv.Spec.NodeAffinity.Required.NodeSelectorTerms
		klog.V(10).Infof("Match for Required node selector terms %+v", terms)
		if !v1helper.MatchNodeSelectorTerms(terms, labels.Set(nodeLabels), nil) {
			return fmt.Errorf("No matching NodeSelectorTerms")
		}
	}

	return nil
}
