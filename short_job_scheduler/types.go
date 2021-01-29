package main

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
