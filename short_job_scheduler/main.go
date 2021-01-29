package main

import (
  "context"
  "time"
  "fmt"
  "os"
  "strconv"
  "math/rand"

	"github.com/streadway/amqp"
  "k8s.io/client-go/tools/cache"
  "k8s.io/client-go/kubernetes"

  corev1 "k8s.io/api/core/v1"
  log "github.com/sirupsen/logrus"
  kubeinformers "k8s.io/client-go/informers"
  jsoniter "github.com/json-iterator/go"
  corelisters "k8s.io/client-go/listers/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  configparser "github.com/bigkevmcd/go-configparser"
)

const (

  // Maximum backoff time before the scheduler stop trying to schedule the pod
  MaxBackOffTime = 256
  PodBackoffExceeded corev1.PodPhase = "PodBackoffExceeded"
  DefaultConfigPath = "/go/src/app/config.cfg"

)

// Initialize json encoder
var json = jsoniter.ConfigCompatibleWithStandardLibrary


func main() {


  // Get required values
  confDir := os.Getenv("CONFIG_DIR")

  var config *configparser.ConfigParser
  var err error

  if len(confDir) != 0 {
    config, err = getConfig(confDir)
  }else{
    config, err = getConfig(DefaultConfigPath)
  }


  var mqHost, mqPort, mqUser, mqPass, receiveQueue, backoffQueue, hostname string
  var maxBackOff = MaxBackOffTime

  if err != nil {

    log.Errorf(err.Error())

    hostname = os.Getenv("HOSTNAME")
    mqHost = os.Getenv("MQ_HOST")
    mqPort = os.Getenv("MQ_PORT")
    mqUser = os.Getenv("MQ_USER")
    mqPass = os.Getenv("MQ_PASS")
    receiveQueue = os.Getenv("RECEIVE_QUEUE")
    backoffQueue = os.Getenv("RETRY_QUEUE")

    if len(mqHost) == 0 ||
    len(mqPort) == 0 ||
    len(mqUser) == 0 ||
    len(mqPass) == 0 ||
    len(hostname) == 0 ||
    len(receiveQueue) == 0 ||
    len(backoffQueue) == 0{
  	   log.Fatalf("Config not found, Environment variables missing")
    }


  }else{

    mqHost, err = config.Get("QueueService", "hostname")
    if err != nil {
      log.Fatalf(err.Error())
    }
    mqPort, err = config.Get("QueueService", "port")
    if err != nil {
      log.Fatalf(err.Error())
    }
    mqUser, err = config.Get("QueueService", "user")
    if err != nil {
      log.Fatalf(err.Error())
    }
    mqPass, err = config.Get("QueueService", "pass")
    if err != nil {
      log.Fatalf(err.Error())
    }
    hostname, err = config.Get("DEFAULTS", "hostname")
    if err != nil {
      log.Fatalf(err.Error())
    }
    receiveQueue, err = config.Get("DEFAULTS", "receive_queue")
    if err != nil {
      log.Fatalf(err.Error())
    }
    backoffQueue, err = config.Get("DEFAULTS", "retry_queue")
    if err != nil {
      log.Fatalf(err.Error())
    }
    // Get max back off duration if exist
    p, err := config.Get("DEFAULTS", "maximum_backoff_time")
    if err == nil {
      val, err := strconv.Atoi(p)
      if err == nil {
        maxBackOff = val
      }else{
        log.Errorf(err.Error())
      }
    }
  }


  // Get the Kubernetes client for communicating with API server
	client := getKubernetesClient()

  // Create the required resource listers and informers
  kubefactory := kubeinformers.NewSharedInformerFactory(client, time.Second*30)
  node_lister := kubefactory.Core().V1().Nodes().Lister()
  pod_lister := kubefactory.Core().V1().Pods().Lister()

  // Attempt to connect to the rabbitMQ server
  comm, err := NewRabbitMQCommunication(fmt.Sprintf("amqp://%s:%s@%s:%s/",mqUser, mqPass, mqHost, mqPort))
  if err != nil {
    log.Fatalf(err.Error())
  }

  err = comm.QueueDeclare(receiveQueue)
  if err != nil {
    log.Fatalf(err.Error())
  }

  msgs, err := comm.Receive(receiveQueue)

  // Use a channel if goroutine closes
  retryCh := make(chan bool)
  defer close(retryCh)

  // use a channel to synchronize the finalization for a graceful shutdown
  stopCh := make(chan struct{})
  defer close(stopCh)
  kubefactory.Start(stopCh)


  // Do the initial synchronization (one time) to populate resources
  kubefactory.WaitForCacheSync(stopCh)

  // Create scheduler object
  main_sched := NewShortJobScheduler(client, pod_lister, node_lister)

  // Scheduler initialization failed
  if err != nil {
    log.Fatalf(err.Error())
  }

  // Start go routine to start consuming messages
	go ScheduleProcess(&comm, main_sched, client, pod_lister, msgs, retryCh, receiveQueue, backoffQueue, hostname, maxBackOff)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

  // Check for connection failures and reconnect
  for {

    if status := <-retryCh; status == true {
      log.Errorf("Disconnected from message server and attempting to reconnect")
      for{
        err = comm.Connect()
        if err != nil{
          log.Errorf(err.Error())
        }else{

          err = comm.QueueDeclare(receiveQueue)
          if err != nil {
            log.Errorf(err.Error())
          }else{
            msgs, err = comm.Receive(receiveQueue)
            if(err != nil){
              log.Errorf(err.Error())
            }else{
              // Start go routine to start consuming messages
              go ScheduleProcess(&comm, main_sched, client, pod_lister, msgs, retryCh, receiveQueue, backoffQueue, hostname, maxBackOff)
              log.Infof("Reconnected to message server")
              break
            }
          }
        }
        // Sleep for a random time before trying again
        time.Sleep(time.Duration(rand.Intn(10))*time.Second)
      }
    }
  }
}

// Does scheduling operations and should be executed in a goroutine
func ScheduleProcess(
  comm Communication,
  s *ShortJobScheduler,
  client kubernetes.Interface,
  podLister corelisters.PodLister,
  msgs <-chan amqp.Delivery,
  closed chan<- bool,
  receiveQueue string,
  backoffQueue string,
  hostname string,
  maxBackOff int,){

  // Loop through all the messages in the queue
  for d := range msgs {

    // Record time of processing
    timestamp := time.Now()

    // Convert json message to schedule request object
    var req ScheduleRequest

    if err := json.Unmarshal(d.Body, &req); err != nil {
        panic(err)
    }

    // Extract the pod name and namespace from the request
    key := string(req.Key);

    // Convert the namespace/name string into a distinct namespace and name
    namespace, name, err := cache.SplitMetaNamespaceKey(key)
    if err != nil {
      log.Errorf("%s", err)
    }

    obj, err := client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})


    // Check if pod still exist in the kube-api server if not ignore it
    if err == nil && obj != nil{

      log.Infof("Scheduling %s",obj.Name)

      // Start scheduling the pod
      result, err := s.Schedule(obj)

      if err != nil {

        // Print the error in the event the scheduler is unable to schedule the pod
        log.Errorf("%s", err)

        // Check scheduling request last back off time and check if it exceeds the maximum backoff time
        if (req.LastBackOffTime >= maxBackOff){

          go AddPodEvent(client,obj,fmt.Sprintf("Scheduler will not retry scheduling; Reason: %s",err.Error()),"Fatal")

          obj.Status.Phase = PodBackoffExceeded

          go AddPodStatus(client,obj,metav1.UpdateOptions{})

        }else{

          // If backoff time not exceeded, multiply the last backoff time by 2 and send it to backoff queue
          req.LastBackOffTime = req.LastBackOffTime*2
          req.Message = err.Error()

          respBytes, err := json.Marshal(RetryRequest{req,receiveQueue})
          if err != nil {
            log.Fatalf("%s", err)
          }

          go AddPodEvent(client,obj,fmt.Sprintf("Scheduler will retry in %d seconds; Reason: %s",req.LastBackOffTime,req.Message),"Warning")

          // Attempt to send message to retry service
          go SendToQueue(comm,respBytes,backoffQueue)
        }

        d.Ack(true)

      }else if len(result) != 0{

        log.Infof("Scheduling Pod %s to %s", name, result)
        bind(client,*obj,result,req.ProcessedTime,timestamp)
        //Use for experiment only
        go SendExperimentPayload(comm,obj,timestamp,time.Now(),"epsilon.experiment",result,hostname)
        d.Ack(true)

      }else{
        d.Nack(true, true)
      }
    }else{
      d.Ack(true)
    }
  }

  closed <- true

}

func SendExperimentPayload(comm Communication, obj *corev1.Pod, in time.Time, out time.Time, queueName string, suggestedHost string, hostname string){

  // Deep copy as modifications will be made to the pod
  pod := obj.DeepCopy()

  pod.Spec.NodeName = suggestedHost

  for {
    if sendExperimentPayload(comm, pod, in, out, "epsilon.experiment", hostname) == false {

      for{
        err := comm.Connect()
        if err == nil{
          break
        }
        // Sleep for a random time before trying again
        time.Sleep(time.Duration(rand.Intn(10))*time.Second)
      }
    }else{
      break
    }
  }
}

func sendExperimentPayload(comm Communication, pod *corev1.Pod, in time.Time, out time.Time, queueName string, hostname string) bool{

  respBytes, err := json.Marshal(ExperimentPayload{Type:"Scheduler",InTime:in,OutTime:out,Pod:pod,Hostname: hostname})
  if err != nil {
    log.Fatalf("%s", err)
  }

  err = comm.Send(respBytes,queueName)

  if err != nil{
    return false
  }

  return true
}
