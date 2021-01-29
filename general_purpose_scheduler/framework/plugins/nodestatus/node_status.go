/*
Copyright 2019 The Kubernetes Authors.

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

package nodestatus

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/v1alpha1"
)

// NodeName is a plugin that checks if a pod spec node name matches the current node.
type NodeStatus struct{}

var _ framework.FilterPlugin = &NodeStatus{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = "NodeStatus"

	// ErrReason returned when node name doesn't match.
	ErrReason = "node(s) not ready to accept pods"
)

// Name returns name of the plugin. It is used in logs, etc.
func (pl *NodeStatus) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *NodeStatus) Filter(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {

  if nodeInfo.Node() == nil {
		return framework.NewStatus(framework.Error, "node not found")
	}
	if !Fits(pod, nodeInfo) {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReason)
	}
	return nil
}

// Fits actually checks if the pod fits the node.
func Fits(pod *v1.Pod, nodeInfo *framework.NodeInfo) bool {

  arrayOfStatus := nodeInfo.Node().Status.Conditions

  for _, s := range arrayOfStatus {
    if(s.Type == v1.NodeReady){
      if(s.Status != v1.ConditionTrue){
        return false
      }else{
        return true
      }
    }
  }

  return false


}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.FrameworkHandle) (framework.Plugin, error) {
	return &NodeStatus{}, nil
}
