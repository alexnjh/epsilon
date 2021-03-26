package linear_regression

import(
  regression "github.com/sajari/regression"
)

// LinearRegressionPlugin decides based on LinearRegression.
// (NOTE this is not very effective and was use as a demo and should be replaced by a more robust machine learning algorithm)
type LinearRegressionPlugin struct{
  Name string
  QueueName string
  MinimumAmtOfData int
  dataCount int
  r *regression.Regression
}

// Creates a new LinearRegressionPlugin
func NewLinearRegressionPlugin(name,queueName string,threshold int) *LinearRegressionPlugin{

  r := new(regression.Regression)
  r.SetObserved("Number of pending pods")
  r.SetVar(0, "Number of nodes")
	r.SetVar(1, "Number of schedulers")


  return &LinearRegressionPlugin{
    Name: name,
    QueueName: queueName,
    MinimumAmtOfData: threshold,
    dataCount: 0,
    r: r,
  }
}

// Compute processes the data and return a ComputeResult
func (plugin *LinearRegressionPlugin) Compute(noOfPendingPods, noOfNodes, noOfSched float64) ComputeResult{

  plugin.dataCount+=1

  plugin.r.Train(
    regression.DataPoint(noOfPendingPods, []float64{noOfNodes, noOfSched}),
  )

  if (plugin.dataCount > plugin.MinimumAmtOfData) {

    plugin.r.Run()

    prediction, err := plugin.r.Predict([]float64{noOfNodes, noOfSched})

    if err != nil{
      log.Fatalf(err.Error())
    }

    var diff = noOfPendingPods-prediction

    if diff < 0 {

      if (math.Abs(diff)/noOfPendingPods) > 0.5 {
        return ScaleUp
      }

    }else{

      if (math.Abs(diff)/noOfPendingPods) > 0.5{
        return ScaleDown
      }

    }

  }
  return DoNotScale

}