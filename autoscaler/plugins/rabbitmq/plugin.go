package rabbitmq

import(
  rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

// RabbitMQPlugin decides based on current queue utilization value given by the RabbitMQ service.
type RabbitMQPlugin struct{
  Name string
  Vhost string
  QueueName string
  threshold float64
  client *rabbithole.Client
}

// Creates a new RabbitMQPlugin
func NewRabbitMQPlugin(name,vhost,queueName string,threshold float64,client *rabbithole.Client) *RabbitMQPlugin{
  return &RabbitMQPlugin{
    Name: name,
    Vhost: vhost,
    QueueName: queueName,
    threshold: threshold,
    client: client,
  }
}

// Compute processes the data and return a ComputeResult
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
