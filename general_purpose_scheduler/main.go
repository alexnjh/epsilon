/*
Copyright (C) 2020 Alex Neo

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
  "time"
  "fmt"
  "os"
  "strconv"
  "math/rand"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/k8s.io/kubernetes/pkg/controller/volume/scheduling"
  corev1 "k8s.io/api/core/v1"
  log "github.com/sirupsen/logrus"
  sched "github.com/alexnjh/epsilon/general_purpose_scheduler/scheduler"
  kubeinformers "k8s.io/client-go/informers"
  jsoniter "github.com/json-iterator/go"
  internalcache "github.com/alexnjh/epsilon/general_purpose_scheduler/internal/cache"
  configparser "github.com/bigkevmcd/go-configparser"
  communication "github.com/alexnjh/epsilon/communication"
)

const (

  // Maximum backoff time before the scheduler stop trying to schedule the pod
  MaxBackOffTime = 256
  QPS = 100
  Burst = 200
  PodBackoffExceeded corev1.PodPhase = "PodBackoffExceeded"
  DefaultConfigPath = "/go/src/app/config.cfg"

)

// Initialize json encoder
var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*

The main routing of the scheduler microservice.

The scheduler will first attempt to get configuration variables via the config file.
If not config file is found the autoscaler will attempt to load configuration variables
from the Environment variables.

Once the configuration variables are loaded the scheduler will create the scheduler struct
and initialize the local state by populating the locat state with information fetched from
the kube-api server

Once all the require variables are created the ScheduleProcess() method will be invoked which
starts the scheduling lifecycle.

*/
func main() {

  var mqHost, mqPort, mqUser, mqPass, receiveQueue, backoffQueue, hostname string
  var maxBackOff = MaxBackOffTime
  var config *configparser.ConfigParser
  var err error

  // Get config files directory
  confDir := os.Getenv("CONFIG_DIR")

  // If no config path defined attempt to get config from default path
  if len(confDir) != 0 {
    config, err = getConfig(confDir)
  }else{
    config, err = getConfig(DefaultConfigPath)
  }

  // If fail to get config file attempt to get configuration details from OS Environment variables
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
  node_informer := kubefactory.Core().V1().Nodes().Informer()
  pod_informer := kubefactory.Core().V1().Pods().Informer()
  node_lister := kubefactory.Core().V1().Nodes().Lister()
  pod_lister := kubefactory.Core().V1().Pods().Lister()

  // Create volume binders for the different storage options
  volumeBinder := scheduling.NewVolumeBinder(
		client,
		kubefactory.Core().V1().Nodes(),
		kubefactory.Storage().V1().CSINodes(),
		kubefactory.Core().V1().PersistentVolumeClaims(),
		kubefactory.Core().V1().PersistentVolumes(),
		kubefactory.Storage().V1().StorageClasses(),
		time.Duration(10)*time.Second,
	)

  // Connect to RabbitMQ Server
  comm, err := communication.NewCommunicationClient(fmt.Sprintf("amqp://%s:%s@%s:%s/",mqUser, mqPass, mqHost, mqPort))
  if err != nil {
    log.Fatalf(err.Error())
  }

  // Declare queue to to receive messages from
  err = comm.QueueDeclare(receiveQueue)
  if err != nil {
    log.Fatalf(err.Error())
  }

  // Initilize a recevier to receive messages from queue
  msgs, err := comm.Receive(receiveQueue)

  // Use a channel to reconnect in the event the connection to the message broker is down
  retryCh := make(chan bool)
  defer close(retryCh)

  // Use a channel to synchronize the finalization for a graceful shutdown
  stopCh := make(chan struct{})
  defer close(stopCh)
  kubefactory.Start(stopCh)

  // Do the initial synchronization (one time) to populate resources
  kubefactory.WaitForCacheSync(stopCh)

  // Create a cache for the scheduler
  schedulerCache := internalcache.New(30*time.Second, stopCh)

  // Create scheduler object
  main_sched, err := sched.New(volumeBinder, client, schedulerCache, kubefactory, node_lister, pod_lister, false, 10.0)

  // Scheduler initialization failed
  if err != nil {
    log.Fatalf(err.Error())
  }

  // Add event handlers to update local state
  addAllEventHandlers(main_sched,kubefactory,node_informer,pod_informer)

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
