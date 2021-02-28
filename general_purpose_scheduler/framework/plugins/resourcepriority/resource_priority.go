/*
Copyright (C) 2020 Alex Neo

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

package resourcepriority

import (
  "math"
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/v1alpha1"
)

const (
  // The maximum priority value that the plugin returns
  MaxPriority = 100

  // Name is the name of the plugin used in the plugin registry and configurations.
  Name = "ResourcePriority"
)

// ResourcePriority is a score plugin that favors nodes that have the most available resources.
type ResourcePriority struct {
	handle framework.FrameworkHandle
}

var _ framework.ScorePlugin = &ResourcePriority{}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *ResourcePriority) Name() string {
	return Name
}

// Score invoked at the score extension point.
func (pl *ResourcePriority) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {

  nodeInfo, err := pl.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}

  node := nodeInfo.Node()

  capacity, status := node.Status.Capacity.Memory().AsInt64()

  if !status {
    return 0, nil
  }

  if nodeInfo.NonZeroRequested.Memory != 0 && capacity != 0{

    nodeRF := float64(nodeInfo.NonZeroRequested.Memory)/float64(capacity)
    score := math.Exp(-5*float64(nodeRF))
    return int64(score*100), nil

  }else{
    return MaxPriority, nil
  }

}

// ScoreExtensions of the Score plugin.
func (pl *ResourcePriority) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	return &ResourcePriority{handle: h}, nil
}
