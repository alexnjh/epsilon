package communication

import(
    "time"
    corev1 "k8s.io/api/core/v1"
)

/*
Message structure for communicating with the Scheduler microservice
*/
type ScheduleRequest struct {
  Key  string // A string containing pod details in the following format [pod name]@[namespace]
  LastBackOffTime int // Previous backoff duration
  ProcessedTime time.Duration // Total time taken to complete scheduling
  Message string // Supporting information if required [optional]
}

type CommitRequest struct {
  Status  string
  Description string
  NodeName string
  Pod corev1.Pod
}

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

/*
Message structure for communicating with the Retry microservice
*/
type RetryRequest struct {
  Req  ScheduleRequest
  Queue string //Name of the queue the Retry microservice is using
}
