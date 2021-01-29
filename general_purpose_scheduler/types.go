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

package general_purpose_scheduler

import(
    "time"
    corev1 "k8s.io/api/core/v1"
)

type RetryRequest struct {
  Req  ScheduleRequest
  Queue string
}

type ScheduleRequest struct {
  Key  string
  LastBackOffTime int
  ProcessedTime time.Duration
  Message string // Left empty unless there is an error
}

type CommitRequest struct {
  Status  string
  Description string
  NodeName string
  Pod corev1.Pod
}

type ExperimentPayload struct {
  Type string
  Hostname string
  InTime time.Time
  OutTime time.Time
  Pod *corev1.Pod
}
