package queue_theory

import(

)

// QueueTheoryPlugin decides based on approximation of the waiting time for all the pods current in the cluster wiating to be scheduled.
type QueueTheoryPlugin struct{
  Name string
  Threshold float64
  targetURL string
}

// Creates a new QueueTheoryPlugin
func NewQueueTheoryPlugin(name string,threshold float64,targetURL string) *QueueTheoryPlugin{
  return &QueueTheoryPlugin{
    Name: name,
    Threshold: threshold,
    targetURL: targetURL,
  }
}

// Compute processes the data and return a ComputeResult
func (plugin *QueueTheoryPlugin) Compute(_, _, noOfSched float64) ComputeResult{

  metricMap := promToMap(plugin.targetURL)

  arrivalRate, err := strconv.ParseFloat(metricMap["pod_request_total_in_1min"], 64)
  if err != nil {
		log.Fatalf(err.Error())
	}

  serviceRate := noOfSched*(float64((1*time.Minute)/(25*time.Millisecond)))

  avgWaitingTime := arrivalRate/(serviceRate*(serviceRate-arrivalRate))

  if (avgWaitingTime < plugin.Threshold){
    return DoNotScale
  }else{
    return ScaleUp
  }

}
