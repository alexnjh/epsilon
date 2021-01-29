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

package repeatpriority

import (
  // "fmt"
  "math"
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/v1alpha1"
)

// The two thresholds are used as bounds for the image score range. They correspond to a reasonable size range for
// container images compressed and stored in registries; 90%ile of images on dockerhub drops into this range.
const (
  MaxPriority = 100
)

// ImageLocality is a score plugin that favors nodes that already have requested pod container's images.
type RepeatPriority struct {
	handle framework.FrameworkHandle
}

var _ framework.ScorePlugin = &RepeatPriority{}

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "RepeatPriority"

// Name returns name of the plugin. It is used in logs, etc.
func (pl *RepeatPriority) Name() string {
	return Name
}


// Score invoked at the score extension point.
func (pl *RepeatPriority) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {

  rnode := pl.handle.GetNodeUsageFactor(nodeName)
  rhigh := pl.handle.GetHighestUsageFactor()

  nodeRF := float64(rnode)/float64(rhigh)
  score := math.Exp(-5*float64(nodeRF))
  return int64(score*100), nil
}

// ScoreExtensions of the Score plugin.
func (pl *RepeatPriority) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	return &RepeatPriority{handle: h}, nil
}
