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
	"github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/tainttoleration"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/nodeaffinity"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/noderesources"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/nodename"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/nodestatus"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/nodeports"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/nodeunschedulable"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/interpodaffinity"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/imagelocality"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/volumebinding"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/volumerestrictions"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/resourcepriority"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins/repeatpriority"
	framework "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/v1alpha1"
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
