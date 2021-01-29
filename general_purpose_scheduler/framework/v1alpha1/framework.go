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

package v1alpha1

import (
  "fmt"
  "k8s.io/klog"
  "reflect"
  "k8s.io/apimachinery/pkg/util/sets"
  "context"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/internal/parallelize"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/k8s.io/kubernetes/pkg/controller/volume/scheduling"

  v1 "k8s.io/api/core/v1"
  clientset "k8s.io/client-go/kubernetes"
  config "github.com/alexnjh/epsilon/general_purpose_scheduler/scheduler/config"
)


// framework is the component responsible for initializing and running scheduler plugins.
type framework struct {
  highestRepeatFactor   int
	registry              Registry
	pluginNameToWeightMap map[string]int
  nodeUsageMap          map[string]int
  preFilterPlugins      []PreFilterPlugin
	filterPlugins         []FilterPlugin
  preScorePlugins       []PreScorePlugin
	scorePlugins          []ScorePlugin
	clientSet             clientset.Interface
  snapshotSharedLister  SharedLister
  volumeBinder          scheduling.SchedulerVolumeBinder


}

// Creates a new framework struct
func NewFramework(
  r Registry,
  client clientset.Interface,
  sharedLister SharedLister,
  volumeBinder scheduling.SchedulerVolumeBinder) (framework,error){

  f := &framework{
    highestRepeatFactor:   1,
    registry:              r,
    pluginNameToWeightMap: make(map[string]int),
    nodeUsageMap:          make(map[string]int),
    clientSet:             client,
    snapshotSharedLister:  sharedLister,
    volumeBinder:          volumeBinder,
  }


  pluginsMap := make(map[string]Plugin)

  for name, factory := range r {
    p, err := factory(nil, f)
    if err != nil {
      return *f, fmt.Errorf("error initializing plugin %q: %v", name, err)
    }
    pluginsMap[p.Name()] = p
    f.pluginNameToWeightMap[p.Name()] = 1
  }


  // Filter plugins
  pluginArr := make([]config.Plugin,0)
  pluginArr = append(pluginArr,
  config.Plugin{
    Name: "NodeStatus",
    Weight: 1,
  },config.Plugin{
    Name: "TaintToleration",
    Weight: 1,
  },config.Plugin{
    Name: "NodeAffinity",
    Weight: 1,
  },config.Plugin{
    Name: "NodeResourcesFit",
    Weight: 1,
  },config.Plugin{
    Name: "NodeName",
    Weight: 1,
  },config.Plugin{
    Name: "NodePorts",
    Weight: 1,
  },config.Plugin{
    Name: "NodeUnschedulable",
    Weight: 1,
  },config.Plugin{
    Name: "InterPodAffinity",
    Weight: 1,
  },config.Plugin{
    Name: "VolumeBinding",
    Weight: 1,
  },config.Plugin{
    Name: "VolumeRestrictions",
    Weight: 1,
  })

  // Prefilter plugins
  pluginArr2 := make([]config.Plugin,0)
  pluginArr2 = append(pluginArr2,config.Plugin{
    Name: "NodeResourcesFit",
    Weight: 1,
  },config.Plugin{
    Name: "NodePorts",
    Weight: 1,
  },config.Plugin{
    Name: "InterPodAffinity",
    Weight: 1,
  })

  // Prescore plugins
  pluginArr3 := make([]config.Plugin,0)
  pluginArr3 = append(pluginArr3,config.Plugin{
    Name: "TaintToleration",
    Weight: 1,
  },config.Plugin{
    Name: "InterPodAffinity",
    Weight: 1,
  })

  // Score plugins
  pluginArr4 := make([]config.Plugin,0)
  pluginArr4 = append(pluginArr4,config.Plugin{
    Name: "TaintToleration",
    Weight: 1,
  },config.Plugin{
    Name: "NodeAffinity",
    Weight: 1,
  },config.Plugin{
    Name: "ImageLocality",
    Weight: 1,
  },config.Plugin{
    Name: "ResourcePriority",
    Weight: 1,
  },config.Plugin{
    Name: "RepeatPriority",
    Weight: 1,
  },config.Plugin{
    Name: "InterPodAffinity",
    Weight: 1,
  })


  pluginSet := config.PluginSet{
    Enabled: pluginArr,
  }

  pluginSet2 := config.PluginSet{
    Enabled: pluginArr2,
  }

  pluginSet3 := config.PluginSet{
    Enabled: pluginArr3,
  }

  pluginSet4 := config.PluginSet{
    Enabled: pluginArr4,
  }

  plugins := config.Plugins{
    PreFilter: &pluginSet2,
    Filter: &pluginSet,
    PreScore: &pluginSet3,
    Score: &pluginSet4,
  }


  for _, e := range f.getExtensionPoints(&plugins) {
		if err := updatePluginList(e.slicePtr, e.plugins, pluginsMap); err != nil {
      fmt.Println(err)
			return *f, err
		}
	}


  return *f, nil

}

// extensionPoint encapsulates desired and applied set of plugins at a specific extension
// point. This is used to simplify iterating over all extension points supported by the
// framework.
type extensionPoint struct {
	// the set of plugins to be configured at this extension point.
	plugins *config.PluginSet
	// a pointer to the slice storing plugins implementations that will run at this
	// extension point.
	slicePtr interface{}
}

func (f *framework) getExtensionPoints(plugins *config.Plugins) []extensionPoint {
	return []extensionPoint{
		{plugins.PreFilter, &f.preFilterPlugins},
		{plugins.Filter, &f.filterPlugins},
		{plugins.PreScore, &f.preScorePlugins},
		{plugins.Score, &f.scorePlugins},
	}
}

// Update list of plugins in the framework
func updatePluginList(pluginList interface{}, pluginSet *config.PluginSet, pluginsMap map[string]Plugin) error {
	if pluginSet == nil {
		return nil
	}

	plugins := reflect.ValueOf(pluginList).Elem()
	pluginType := plugins.Type().Elem()
	set := sets.NewString()
	for _, ep := range pluginSet.Enabled {
		pg, ok := pluginsMap[ep.Name]
		if !ok {
			return fmt.Errorf("%s %q does not exist", pluginType.Name(), ep.Name)
		}

		if !reflect.TypeOf(pg).Implements(pluginType) {
			return fmt.Errorf("plugin %q does not extend %s plugin", ep.Name, pluginType.Name())
		}

		if set.Has(ep.Name) {
			return fmt.Errorf("plugin %q already registered as %q", ep.Name, pluginType.Name())
		}

		set.Insert(ep.Name)

		newPlugins := reflect.Append(plugins, reflect.ValueOf(pg))
		plugins.Set(newPlugins)
	}
	return nil
}

// RunPreFilterPlugins runs the set of configured PreFilter plugins. It returns
// *Status and its code is set to non-success if any of the plugins returns
// anything but Success. If a non-success status is returned, then the scheduling
// cycle is aborted.
func (f *framework) RunPreFilterPlugins(ctx context.Context, state *CycleState, pod *v1.Pod) (status *Status) {

	for _, pl := range f.preFilterPlugins {
		status = f.runPreFilterPlugin(ctx, pl, state, pod)
		if !status.IsSuccess() {
			if status.IsUnschedulable() {
				msg := fmt.Sprintf("rejected by %q at prefilter: %v", pl.Name(), status.Message())
				klog.V(4).Infof(msg)
				return NewStatus(status.Code(), msg)
			}
			msg := fmt.Sprintf("error while running %q prefilter plugin for pod %q: %v", pl.Name(), pod.Name, status.Message())
			klog.Error(msg)
			return NewStatus(Error, msg)
		}
	}

	return nil
}

func (f *framework) runPreFilterPlugin(ctx context.Context, pl PreFilterPlugin, state *CycleState, pod *v1.Pod) *Status {
	status := pl.PreFilter(ctx, state, pod)
	return status
}

// RunFilterPlugins runs the set of configured Filter plugins for pod on
// the given node. If any of these plugins doesn't return "Success", the
// given node is not suitable for running pod.
// Meanwhile, the failure message and status are set for the given node.
func (f *framework) RunFilterPlugins(
	ctx context.Context,
	state *CycleState,
	pod *v1.Pod,
	nodeInfo *NodeInfo,
) PluginToStatus {
	var firstFailedStatus *Status
	statuses := make(PluginToStatus)
	for _, pl := range f.filterPlugins {
		pluginStatus := f.runFilterPlugin(ctx, pl, state, pod, nodeInfo)
		if len(statuses) == 0 {
			firstFailedStatus = pluginStatus
		}
		if !pluginStatus.IsSuccess() {
			if !pluginStatus.IsUnschedulable() {
				// Filter plugins are not supposed to return any status other than
				// Success or Unschedulable.
				firstFailedStatus = NewStatus(Error, fmt.Sprintf("running %q filter plugin for pod %q: %v", pl.Name(), pod.Name, pluginStatus.Message()))
				return map[string]*Status{pl.Name(): firstFailedStatus}
			}
			statuses[pl.Name()] = pluginStatus
		}
	}

	return statuses
}


func (f *framework) runFilterPlugin(ctx context.Context, pl FilterPlugin, state *CycleState, pod *v1.Pod, nodeInfo *NodeInfo) *Status {
	status := pl.Filter(ctx, state, pod, nodeInfo)
	return status
}

func (f *framework) ClientSet() (clientset.Interface){
  return f.clientSet
}

func (f *framework) HasFilterPlugins() (bool){
  return (len(f.filterPlugins) > 0)
}

func (f *framework) HasScorePlugins() (bool){
  return (len(f.scorePlugins) > 0)
}

func (f *framework) GetHighestUsageFactor() (int){
  return f.highestRepeatFactor
}

func (f *framework) GetNodeUsageFactor(nodeName string) (int){
  return f.nodeUsageMap[nodeName]
}

func (f *framework) IncreaseNodeUsageFactor(nodeName string){

  f.nodeUsageMap[nodeName] +=1

  if f.nodeUsageMap[nodeName] > f.highestRepeatFactor {
    f.highestRepeatFactor+=1
  }

  // Special case, reset the map if this happens
  if f.nodeUsageMap[nodeName] == 0 {
    f.nodeUsageMap = make(map[string]int)
  }
}

// SnapshotSharedLister returns the scheduler's SharedLister of the latest NodeInfo
// snapshot. The snapshot is taken at the beginning of a scheduling cycle and remains
// unchanged until a pod finishes "Reserve". There is no guarantee that the information
// remains unchanged after "Reserve".
func (f *framework) SnapshotSharedLister() SharedLister {
	return f.snapshotSharedLister
}

// WithSnapshotSharedLister sets the SharedLister of the snapshot.
func (f *framework) WithSnapshotSharedLister(snapshotSharedLister SharedLister) {
	f.snapshotSharedLister = snapshotSharedLister
}

// VolumeBinder returns the volume binder used by scheduler.
func (f *framework) VolumeBinder() scheduling.SchedulerVolumeBinder {
	return f.volumeBinder
}

// RunPreScorePlugins runs the set of configured pre-score plugins. If any
// of these plugins returns any status other than "Success", the given pod is rejected.
func (f *framework) RunPreScorePlugins(
	ctx context.Context,
	state *CycleState,
	pod *v1.Pod,
	nodes []*v1.Node,
) (status *Status) {

	for _, pl := range f.preScorePlugins {
		status = f.runPreScorePlugin(ctx, pl, state, pod, nodes)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("error while running %q prescore plugin for pod %q: %v", pl.Name(), pod.Name, status.Message())
			klog.Error(msg)
			return NewStatus(Error, msg)
		}
	}

	return nil
}

func (f *framework) runPreScorePlugin(ctx context.Context, pl PreScorePlugin, state *CycleState, pod *v1.Pod, nodes []*v1.Node) *Status {
	status := pl.PreScore(ctx, state, pod, nodes)
	return status
}

// RunScorePlugins runs the set of configured scoring plugins. It returns a list that
// stores for each scoring plugin name the corresponding NodeScoreList(s).
// It also returns *Status, which is set to non-success if any of the plugins returns
// a non-success status.
func (f *framework) RunScorePlugins(ctx context.Context, state *CycleState, pod *v1.Pod, nodes []*v1.Node) (ps PluginToNodeScores, status *Status) {
	pluginToNodeScores := make(PluginToNodeScores, len(f.scorePlugins))
	for _, pl := range f.scorePlugins {
		pluginToNodeScores[pl.Name()] = make(NodeScoreList, len(nodes))
	}
	ctx, cancel := context.WithCancel(ctx)
	errCh := parallelize.NewErrorChannel()

	// Run Score method for each node in parallel.
	parallelize.Until(ctx, len(nodes), func(index int) {
		for _, pl := range f.scorePlugins {
			nodeName := nodes[index].Name
			s, status := f.runScorePlugin(ctx, pl, state, pod, nodeName)
			if !status.IsSuccess() {
				errCh.SendErrorWithCancel(fmt.Errorf(status.Message()), cancel)
				return
			}
			pluginToNodeScores[pl.Name()][index] = NodeScore{
				Name:  nodeName,
				Score: int64(s),
			}
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		msg := fmt.Sprintf("error while running score plugin for pod %q: %v", pod.Name, err)
		klog.Error(msg)
		return nil, NewStatus(Error, msg)
	}

	// Apply score defaultWeights for each ScorePlugin in parallel.
	parallelize.Until(ctx, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		// Score plugins' weight has been checked when they are initialized.
		weight := f.pluginNameToWeightMap[pl.Name()]
		nodeScoreList := pluginToNodeScores[pl.Name()]

		for i, nodeScore := range nodeScoreList {
			// return error if score plugin returns invalid score.
			if nodeScore.Score > int64(MaxNodeScore) || nodeScore.Score < int64(MinNodeScore) {
				err := fmt.Errorf("score plugin %q returns an invalid score %v, it should in the range of [%v, %v] after normalizing", pl.Name(), nodeScore.Score, MinNodeScore, MaxNodeScore)
				errCh.SendErrorWithCancel(err, cancel)
				return
			}
			nodeScoreList[i].Score = nodeScore.Score * int64(weight)
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		msg := fmt.Sprintf("error while applying score defaultWeights for pod %q: %v", pod.Name, err)
		klog.Error(msg)
		return nil, NewStatus(Error, msg)
	}

	return pluginToNodeScores, nil
}

func (f *framework) runScorePlugin(ctx context.Context, pl ScorePlugin, state *CycleState, pod *v1.Pod, nodeName string) (int64, *Status) {
	s, status := pl.Score(ctx, state, pod, nodeName)
	return s, status
}
