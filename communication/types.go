package communication

import(
    "time"
    corev1 "k8s.io/api/core/v1"
)

/*
Message structure used by the Scheduler microservice
*/
type ScheduleRequest struct {
  // A string containing pod details in the following format [pod name]@[namespace]
  Key  string
  // Next backoff duration if the pod fails to schedule
  NextBackOffTime int
  // Total time taken to complete scheduling
  ProcessedTime time.Duration
  // Supporting information if required [optional]
  Message string // Supporting information if required [optional]
}

/*
Message structure for communicating with the Experiment microservice
*/
type ExperimentPayload struct {
  // Type of microservice
  Type string
  // The hostname of the microservice that send this message
  Hostname string
  // The time a microservice receive the pod
  InTime time.Time
  // The the a microservice finish processing a pod
  OutTime time.Time
  // The pod getting processed
  Pod *corev1.Pod
}

/*
Message structure for communicating with the Retry microservice
*/
type RetryRequest struct {
  Req  ScheduleRequest
  //Name of the queue the Retry microservice is using
  Queue string
}
