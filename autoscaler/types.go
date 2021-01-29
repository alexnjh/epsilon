package main

import(
  "time"
  "math"
  "strconv"

  // "github.com/prometheus/client_golang/api"
  // "github.com/prometheus/common/config"
  // "github.com/RobinUS2/golang-moving-average"

  log "github.com/sirupsen/logrus"
  regression "github.com/sajari/regression"
  rabbithole "github.com/michaelklishin/rabbit-hole/v2"
  // v1 "github.com/prometheus/client_golang/api/prometheus/v1"

)

type QueueTheoryPlugin struct{
  Name string
  Threshold float64
  targetURL string
}

func NewQueueTheoryPlugin(name string,threshold float64,targetURL string) *QueueTheoryPlugin{
  return &QueueTheoryPlugin{
    Name: name,
    Threshold: threshold,
    targetURL: targetURL,
  }
}

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

type RabbitMQPlugin struct{
  Name string
  Vhost string
  QueueName string
  threshold float64
  client *rabbithole.Client
}

func NewRabbitMQPlugin(name,vhost,queueName string,threshold float64,client *rabbithole.Client) *RabbitMQPlugin{
  return &RabbitMQPlugin{
    Name: name,
    Vhost: vhost,
    QueueName: queueName,
    threshold: threshold,
    client: client,
  }
}

func (plugin *RabbitMQPlugin) Compute(_,_,_ float64) ComputeResult{

      qs, err := plugin.client.GetQueue(plugin.Vhost,plugin.QueueName)

      if err != nil{
        log.Fatalf(err.Error())
      }

      if qs.ConsumerUtilisation > plugin.threshold {
        return ScaleUp
      }else{
        return DoNotScale
      }

}

type SchedProbPlugin struct{
  Name string
  QueueName string
  threshold float64
}

func NewSchedProbPlugin(name,queueName string,threshold float64) *SchedProbPlugin{
  return &SchedProbPlugin{
    Name: name,
    QueueName: queueName,
    threshold: threshold,
  }
}

func (plugin *SchedProbPlugin) Compute(_, noOfNodes, noOfSched float64) ComputeResult{

  p := calProb(noOfNodes,noOfSched)

  if (p > plugin.threshold){
    return DoNotScale
  }else{
    return ScaleDown
  }

}

type LinearRegressionPlugin struct{
  Name string
  QueueName string
  MinimumAmtOfData int
  dataCount int
  r *regression.Regression
}

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
