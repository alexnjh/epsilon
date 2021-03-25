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
	"os"
  "time"
  "fmt"
  "syscall"
  "net/http"
  "os/signal"
  "sync/atomic"
	"k8s.io/client-go/tools/cache"
  "k8s.io/client-go/util/workqueue"
  "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
  "github.com/alexnjh/epsilon/coordinator/helper"

  kubeinformers "k8s.io/client-go/informers"
  log "github.com/sirupsen/logrus"
  corev1 "k8s.io/api/core/v1"
  configparser "github.com/bigkevmcd/go-configparser"
  communication "github.com/alexnjh/epsilon/communication"
)

const (
  // Default config path if not config path given
  DefaultConfigPath = "/go/src/app/config.cfg"
)

// main code path
func main() {

  // Get required values
  confDir := os.Getenv("CONFIG_DIR")

  var config *configparser.ConfigParser
  var err error

  if len(confDir) != 0 {
    config, err = helper.GetConfig(confDir)
  }else{
    config, err = helper.GetConfig(DefaultConfigPath)
  }

  var mqHost, mqPort, mqUser, mqPass, mqManagePort, defaultQueue, hostName string
  if err != nil {


    log.Errorf(err.Error())

    hostName = os.Getenv("HOSTNAME")
    mqHost = os.Getenv("MQ_HOST")
    mqManagePort = os.Getenv("MQ_MANAGE_PORT")
    mqPort = os.Getenv("MQ_PORT")
    mqUser = os.Getenv("MQ_USER")
    mqPass = os.Getenv("MQ_PASS")
    defaultQueue = os.Getenv("DEFAULT_QUEUE")

    if len(mqHost) == 0 ||
    len(mqPort) == 0 ||
    len(mqManagePort) == 0 ||
    len(mqUser) == 0 ||
    len(mqPass) == 0 ||
    len(hostName) == 0 ||
    len(defaultQueue) == 0{
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
    mqManagePort, err = config.Get("QueueService", "management_port")
    if err != nil {
      log.Fatalf(err.Error())
    }
    hostName, err = config.Get("DEFAULTS", "hostname")
    if err != nil {
      log.Fatalf(err.Error())
    }
    defaultQueue, err = config.Get("DEFAULTS", "default_queue")
    if err != nil {
      log.Fatalf(err.Error())
    }
  }

  // declare the counter as unsigned int
  var requestsCounter uint64 = 0

	// get the Kubernetes client for communicating with the kubernetes API server
	client := helper.GetKubernetesClient()

  // Create the required informer and listers for kubernetes resources
  kubefactory := kubeinformers.NewSharedInformerFactory(client, time.Second*30)
  pod_informer := kubefactory.Core().V1().Pods().Informer()
  pod_lister := kubefactory.Core().V1().Pods().Lister()

  // Create a new workqueue internally to buffer pos creation request
  queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

  // Attempt to connect to the rabbitMQ server
  comm, err := communication.NewCommunicationClient(fmt.Sprintf("amqp://%s:%s@%s:%s/",mqUser, mqPass, mqHost, mqPort))
  if err != nil {
    log.Fatalf(err.Error())
  }

  newCounter := prometheus.NewCounter(prometheus.CounterOpts{
    Name: "pod_request_processed",
    Help: "How many pod requests processed by the pod coordinator",
  })

  newCounter2 := prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "pod_request_total_in_1min",
    Help: "How many Pod requests processed in the last 1 min (Updates every 1 minute)",
  })
  // register counter in Prometheus collector
  prometheus.MustRegister(prometheus.NewCounterFunc(
    prometheus.CounterOpts{
        Name: "pod_request_total",
        Help: "Counts number of pod requests received",
    },
    func() float64 {
        return float64(atomic.LoadUint64(&requestsCounter))
    }))

  // Metrics have to be registered to be exposed:
	prometheus.MustRegister(newCounter)
	prometheus.MustRegister(newCounter2)

  // Start metric server
  go recordPodCountEvery(1*time.Minute,newCounter2,&requestsCounter)
  go metricsServer()

  // Create a pod controller
  controller := PodController{
  clientset: client,
  informer: pod_informer,
  lister: pod_lister,
  queue: queue,
  handler: &PodHandler{
      defaultQueue: defaultQueue,
      hostname: hostName,
      clientset: client,
      lister: pod_lister,
      comm: &comm,
      metricCounter: newCounter,
    },
  }


  // Add a event handler to listen for new pods
	pod_informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

      key, err := cache.MetaNamespaceKeyFunc(obj)

      if err == nil {

        obj := obj.(*corev1.Pod)

        if obj.Spec.SchedulerName != "custom" ||  obj.Spec.NodeName != ""{
          return
        }

        // somewhere in your code
        atomic.AddUint64(&requestsCounter, 1)
        // Add to workqueue
        queue.Add(key)
      }
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
		},
		DeleteFunc: func(obj interface{}) {
		},
	})

	// use a channel to synchronize the finalization for a graceful shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)

  // Start the informers
  kubefactory.Start(stopCh)

  // run the controller loop to process items
  controller.Run(stopCh)

	// use a channel to handle OS signals to terminate and gracefully shut
	// down processing
	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm
}

/*

Creates a prometheus based metrics server exporting coordinator metrics

*/
func metricsServer(){
  // The Handler function provides a default handler to expose metrics
  // via an HTTP server. "/metrics" is the usual endpoint for that.
  http.Handle("/metrics", promhttp.Handler())
  log.Fatal(http.ListenAndServe(":8080", nil))
}

func recordPodCountEvery(d time.Duration, gauge prometheus.Gauge, currentPodReqCount *uint64) {

  var previousCount = uint64(0)

	for _ = range time.Tick(d) {
    if (*currentPodReqCount != previousCount){
      gauge.Set(float64(*currentPodReqCount-previousCount))
      previousCount = *currentPodReqCount
    }else{
      gauge.Set(float64(0))
    }
	}

}
