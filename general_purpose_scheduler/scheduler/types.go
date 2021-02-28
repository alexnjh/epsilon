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

package scheduler

import (
  v1 "k8s.io/api/core/v1"
)


type ConflictReasons []ConflictReason

type ConflictReason string

const (
	// ErrReasonBindConflict is used for VolumeBindingNoMatch predicate error.
	ErrReasonBindConflict ConflictReason = "node(s) didn't find available persistent volumes to bind"
	// ErrReasonNodeConflict is used for VolumeNodeAffinityConflict predicate error.
	ErrReasonNodeConflict ConflictReason = "node(s) had volume node affinity conflict"
)

type SchedulerVolumeBinder interface {
  // FindPodVolumes checks if all of a Pod's PVCs can be satisfied by the node.
	FindPodVolumes(pod *v1.Pod, node *v1.Node) (reasons ConflictReasons, err error)
}

// Contains the suggested host to schedule a pod and if preemption is required the name of the norminated pod.
type ScheduleResult struct {

	// Name of the scheduler suggest host
	SuggestedHost string

  // Pod to terminate (Only used in preemption)
  NorminatedPod *v1.Pod
}
