package scheduler_prob

import(
  "math"
  "github.com/alexnjh/epsilon/autoscaler/interfaces"
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
func (plugin *SchedProbPlugin) Compute(_, noOfNodes, noOfSched float64) interfaces.ComputeResult{

  p := calProb(noOfNodes,noOfSched)

  if (p > plugin.threshold){
    return interfaces.DoNotScale
  }else{
    return interfaces.ScaleDown
  }

}

// Calculate the scheduler conflict probability
// N = No of schedulers
// K = No of nodes
func calProb(N,K float64) float64{
  return Factorial(N)/(Factorial(N-K)*math.Pow(N,K))
}

func Factorial(n float64)(result float64) {
	if (n > 0) {
		result = n * Factorial(n-1)
		return result
	}
	return 1
}
