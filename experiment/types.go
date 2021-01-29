package experiment

import (
  "time"
  corev1 "k8s.io/api/core/v1"
)

/*
Message structure for communicating with the Experiment microservice
*/
type ExperimentPayload struct {
  Type string
  Hostname string
  InTime time.Time
  OutTime time.Time
  Pod *corev1.Pod
}
