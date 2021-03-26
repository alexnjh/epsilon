package scheduler_prob

import(

)

// SchedProbPlugin decides based on the scheduler conflict probability based on current cluster state.
type SchedProbPlugin struct{
  Name string
  QueueName string
  threshold float64
}

// Creates a new SchedProbPlugin
func NewSchedProbPlugin(name,queueName string,threshold float64) *SchedProbPlugin{
  return &SchedProbPlugin{
    Name: name,
    QueueName: queueName,
    threshold: threshold,
  }
}

// Compute processes the data and return a ComputeResult
func (plugin *SchedProbPlugin) Compute(_, noOfNodes, noOfSched float64) ComputeResult{

  p := calProb(noOfNodes,noOfSched)

  if (p > plugin.threshold){
    return DoNotScale
  }else{
    return ScaleDown
  }

}
