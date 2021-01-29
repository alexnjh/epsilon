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

All functions inside this file is for  experiments and not part of the normal
operations of the scheduler.

*/

package general_purpose_scheduler

import (
  "time"
  "math/rand"
  corev1 "k8s.io/api/core/v1"
  log "github.com/sirupsen/logrus"
  communication "github.com/alexnjh/epsilon/communication"
  ex "github.com/alexnjh/epsilon/experiment"
)

// Use for sending experiment data (Not used in normal operations)
func SendExperimentPayload(comm communication.Communication, obj *corev1.Pod, in time.Time, out time.Time, queueName string, suggestedHost string, hostname string){

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

// Use for sending experiment data (Not used in normal operations)
func sendExperimentPayload(comm communication.Communication, pod *corev1.Pod, in time.Time, out time.Time, queueName string, hostname string) bool{

  respBytes, err := json.Marshal(ex.ExperimentPayload{Type:"Scheduler",InTime:in,OutTime:out,Pod:pod,Hostname: hostname})
  if err != nil {
    log.Fatalf("%s", err)
  }

  err = comm.Send(respBytes,queueName)

  if err != nil{
    return false
  }

  return true
}
