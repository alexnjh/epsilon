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

All functions inside this file is for execution in a goroutine/thread or seperate process

*/

package main

import (
  "fmt"
  "time"
  "context"
  "math/rand"
	"github.com/streadway/amqp"
  "k8s.io/client-go/tools/cache"
  "k8s.io/client-go/kubernetes"
  "k8s.io/apimachinery/pkg/api/errors"

  corev1 "k8s.io/api/core/v1"
  log "github.com/sirupsen/logrus"
  sched "github.com/alexnjh/epsilon/general_purpose_scheduler/scheduler"
  corelisters "k8s.io/client-go/listers/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  communication "github.com/alexnjh/epsilon/communication"
)

/*

The schedule process consist of the following steps:

1. Check for new pods assigned by the coordinator by monitoring the queue
2. Once a new pod is received, get details of the pod from the local state
3. Send pod for scheduling by running the Schedule() method of the scheduler struct
4. Once the Scheduler() function returns check if the NorminatedPod is nil
5. If NorminatedPod is nil proceed to bind the pod to the node if not execute preemption process
6. Repeat step 1

*/
func ScheduleProcess(
  comm communication.Communication,
  s *sched.Scheduler,
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
    var req communication.ScheduleRequest

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
    if err == nil{

      // log.Infof("Scheduling %s",obj.Name)

      // Start scheduling the pod
      result, err := s.Schedule(context.TODO(), obj)

      if err != nil {

        // Print the error in the event the scheduler is unable to schedule the pod
        log.Errorf("%s", err)

        // Check scheduling request last back off time and check if it exceeds the maximum backoff time
        if (req.NextBackOffTime > maxBackOff){

          go AddPodEvent(client,obj,fmt.Sprintf("Scheduler will not retry scheduling; Reason: %s",err.Error()),"Fatal")

          obj.Status.Phase = PodBackoffExceeded

          go AddPodStatus(client,obj,metav1.UpdateOptions{})

        }else{

          req.Message = err.Error()

          respBytes, err := json.Marshal(communication.RetryRequest{req,receiveQueue})
          if err != nil {
            log.Fatalf("%s", err)
          }

          go AddPodEvent(client,obj,fmt.Sprintf("Scheduler will retry in %d seconds; Reason: %s",req.NextBackOffTime,req.Message),"Warning")

          // Attempt to send message to retry service
          go SendToQueue(comm,respBytes,backoffQueue)
        }

        d.Ack(true)

      }else{

        // If no error detected is preemption required.
        // This can be known by checking if the NorminatedPod is nil

        // Run premption process?
        if (result.NorminatedPod != nil){
          PreemptionProcess(client,result.SuggestedHost,obj,result.NorminatedPod,int64(30),req.ProcessedTime,timestamp,result.NorminatedPod.Name)
          // //Use for experiment only
          // go SendExperimentPayload(comm,obj,timestamp,time.Now(),"epsilon.experiment",result.SuggestedHost,hostname)

        }else{
          // log.Infof("Scheduling Pod %s to %s", name, result.SuggestedHost)
          go bind(client,*obj,result.SuggestedHost,req.ProcessedTime,timestamp)
          // //Use for experiment only
          // go SendExperimentPayload(comm,obj,timestamp,time.Now(),"epsilon.experiment",result.SuggestedHost,hostname)
        }

        d.Ack(true)
      }
    }else{
      d.Ack(true)
    }
  }

  closed <- true

}


/*

The preemption process consist of the following steps:

1. Inform the other scehduler services of the preeemption by updating the node details
2. Once update is complete, delete the victim pod from the node
3. Once the pod is deleted, proceed to deploy the preemptor pod
4. Once preemptor is deployed, inform the other scheduler services the preemption had completed

*/
func PreemptionProcess(
  client kubernetes.Interface,
  suggestedHost string,
  preemptorpod *corev1.Pod,
  nominatedpod *corev1.Pod,
  gracePeriod int64,
  discoverTime time.Duration,
  schedTime time.Time,
  key string){

  log.Infof("Starting preemption logic")

  // Add a resource reservation by updating node status
  AddNodeStatusofPreemption(client,suggestedHost,preemptorpod,nominatedpod);

  // Delete the nominated pod based on given grace period and delete in foreground mode
  deletePolicy := metav1.DeletePropagationBackground
  if err := client.CoreV1().Pods(nominatedpod.Namespace).Delete(context.TODO(), nominatedpod.Name, metav1.DeleteOptions{
    PropagationPolicy: &deletePolicy,
    GracePeriodSeconds: &gracePeriod,
  }); err != nil {
    panic(err)
  }

  // Bind preemptor pod to cluster
  for{

    _, err := client.CoreV1().Pods(nominatedpod.Namespace).Get(context.TODO(), nominatedpod.Name, metav1.GetOptions{})

    if err != nil && errors.IsNotFound(err) {
      bind(client,*preemptorpod,suggestedHost,discoverTime,schedTime)
      RemovePreemptionInfo(client,suggestedHost,key)
      return
    }

    time.Sleep(time.Duration(rand.Intn(10))*time.Second)

  }

}
