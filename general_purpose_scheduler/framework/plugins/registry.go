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
package plugins

import (
	"scheduler_unit/framework/plugins/tainttoleration"
  "scheduler_unit/framework/plugins/nodeaffinity"
  "scheduler_unit/framework/plugins/noderesources"
  "scheduler_unit/framework/plugins/nodename"
  "scheduler_unit/framework/plugins/nodestatus"
  "scheduler_unit/framework/plugins/nodeports"
  "scheduler_unit/framework/plugins/nodeunschedulable"
  "scheduler_unit/framework/plugins/interpodaffinity"
  "scheduler_unit/framework/plugins/imagelocality"
  "scheduler_unit/framework/plugins/volumebinding"
  "scheduler_unit/framework/plugins/volumerestrictions"
  "scheduler_unit/framework/plugins/resourcepriority"
  "scheduler_unit/framework/plugins/repeatpriority"
	framework "scheduler_unit/framework/v1alpha1"
)

// NewInTreeRegistry builds the registry with all the in-tree plugins.
// A scheduler that runs out of tree plugins can register additional plugins
// through the WithFrameworkOutOfTreeRegistry option.
func NewInTreeRegistry() framework.Registry {
	return framework.Registry{
		tainttoleration.Name:                       tainttoleration.New,
		nodeaffinity.Name:                          nodeaffinity.New,
		nodename.Name:                              nodename.New,
		nodestatus.Name:                            nodestatus.New,
    nodeports.Name:                             nodeports.New,
    nodeunschedulable.Name:                     nodeunschedulable.New,
    noderesources.FitName:                      noderesources.NewFit,
    interpodaffinity.Name:                      interpodaffinity.New,
    imagelocality.Name:                         imagelocality.New,
    volumebinding.Name:                         volumebinding.New,
    volumerestrictions.Name:                    volumerestrictions.New,
    resourcepriority.Name:                      resourcepriority.New,
    repeatpriority.Name:                        repeatpriority.New,
	}
}
