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


All functions inside this file are common functions that may be used by multiple
functions
*/

package main

import (
  "os"
  "fmt"
  "time"
  "context"
  "math/rand"
  "k8s.io/client-go/rest"
  "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
  "k8s.io/client-go/util/retry"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/k8s.io/kubernetes/pkg/features"

	log "github.com/sirupsen/logrus"
  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  framework "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/v1alpha1"
  utilfeature "k8s.io/apiserver/pkg/util/feature"
  configparser "github.com/bigkevmcd/go-configparser"
  communication "github.com/alexnjh/epsilon/communication"
)

// Get scheduler config file from config path
func getConfig(path string) (*configparser.ConfigParser, error){
  p, err := configparser.NewConfigParserFromFile(path)
  if err != nil {
    return nil,err
  }

  return p,nil
}


// Retrieve the Kubernetes cluster client from outside of the cluster
func getKubernetesClient() (kubernetes.Interface){
	// construct the path to resolve to `~/.kube/config`
  config, err := rest.InClusterConfig()
  if err != nil {
    kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
    // kubeConfigPath := "/etc/kubernetes/scheduler.conf"

    //create the config from the path
    config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
    if err != nil {
      log.Fatalf("getInClusterConfig: %v", err)
      panic("Failed to load kube config")
    }
  }

  // Same settings as the default scheduler
  config.QPS = QPS
  config.Burst = Burst

  // generate the client based off of the config
  client, err := kubernetes.NewForConfig(config)
  if err != nil {
    panic("Failed to create kube client")
  }

	log.Info("Successfully constructed k8s client")
	return client
}

// Use to send a message to a message queue
func SendToQueue(comm communication.Communication, message []byte, queue string){
  for {
    if err := comm.Send(message,queue); err != nil {

      for{
        err = comm.Connect()
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

// Add a new pod event
func AddPodEvent(
  client kubernetes.Interface,
  obj *corev1.Pod, message string,
  typeOfError string){

  // Update API server to inform admin that the pod cannot be deployed and require manual intervention
  timestamp := time.Now().UTC()
  client.CoreV1().Events(obj.Namespace).Create(context.TODO(), &corev1.Event{
    Count:          1,
    Message:        message,
    Reason:         "Error",
    LastTimestamp:  metav1.NewTime(timestamp),
    FirstTimestamp: metav1.NewTime(timestamp),
    Type:           typeOfError,
    Source: corev1.EventSource{
      Component: "custom",
    },
    InvolvedObject: corev1.ObjectReference{
      Kind:      "Pod",
      Name:      obj.Name,
      Namespace: obj.Namespace,
      UID:       obj.UID,
    },
    ObjectMeta: metav1.ObjectMeta{
      GenerateName: obj.Name + "-",
    },
  },metav1.CreateOptions{})
}

// Add a new status for a pod
func AddPodStatus(
  client kubernetes.Interface,
  pod *corev1.Pod,
  options metav1.UpdateOptions,
  ){

  retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {

    // Update Pod Status
    _ , err := client.CoreV1().Pods(pod.Namespace).UpdateStatus(context.TODO(), pod, metav1.UpdateOptions{});

    return err
  })

  if retryErr != nil {
    panic(fmt.Errorf("Update failed: %v", retryErr))
  }

}

// Use for preemption process only, updates preemption information to notify the other
// scheduler services
func AddNodeStatusofPreemption(
  client kubernetes.Interface,
  suggestedHost string,
  preemptorpod *corev1.Pod,
  nominatedpod *corev1.Pod,
  ){

  preemptorPodRequest := computePodResourceRequest(preemptorpod)
  cpu := preemptorPodRequest.MilliCPU
  mem := preemptorPodRequest.Memory
  disk := preemptorPodRequest.EphemeralStorage

  retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
    // Update Node Information
    node, err := client.CoreV1().Nodes().Get(context.TODO(),suggestedHost,metav1.GetOptions{});
    nodeCond := node.Status.Conditions;
    nodeCond = append(nodeCond,corev1.NodeCondition{
      Type: "Preemption",
      Status: "True",
      LastHeartbeatTime: metav1.Now(),
      LastTransitionTime: metav1.Now(),
      Reason: fmt.Sprintf("%d,%d,%d",cpu,mem,disk),
      Message: nominatedpod.Name,
    });

    node.Status.Conditions = nodeCond;

    _ , err = client.CoreV1().Nodes().UpdateStatus(context.TODO(), node, metav1.UpdateOptions{});

    return err
  })

  if retryErr != nil {
    panic(fmt.Errorf("Update failed: %v", retryErr))
  }

}

// Used to remove preemption update once pod is binded
func RemovePreemptionInfo(client kubernetes.Interface, nodeName string, key string){

  retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
    // Retrieve the latest version of pod object before attempting update
  	// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
    node, err := client.CoreV1().Nodes().Get(context.TODO(),nodeName,metav1.GetOptions{});

    nodeCond := node.Status.Conditions;

    for idx, elem := range nodeCond {

      if(len(nodeCond) <= idx){
        break
      }

      // Remove any preemption entry that
      if elem.Type == "Preemption" {
        if elem.Message == key{
          // Remove the element at index i from a.
          nodeCond[idx] = nodeCond[len(nodeCond)-1] // Copy last element to index i.
          nodeCond = nodeCond[:len(nodeCond)-1]   // Truncate slice.
        }
      }
    }

    node.Status.Conditions = nodeCond;
    _ , err = client.CoreV1().Nodes().UpdateStatus(context.TODO(), node, metav1.UpdateOptions{});
    return err
  })

  if retryErr != nil {
    panic(fmt.Errorf("Update failed: %v", retryErr))
  }

}

// Use to compute pod resource requriments
func computePodResourceRequest(pod *corev1.Pod) *framework.Resource {
	result := &framework.Resource{}
	for _, container := range pod.Spec.Containers {
		result.Add(container.Resources.Requests)
	}

	// take max_resource(sum_pod, any_init_container)
	for _, container := range pod.Spec.InitContainers {
		result.SetMaxResource(container.Resources.Requests)
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil && utilfeature.DefaultFeatureGate.Enabled(features.PodOverhead) {
		result.Add(pod.Spec.Overhead)
	}

	return result
}

// Use to bind the pod to the selected node
func bind(client kubernetes.Interface, p corev1.Pod, NodeName string, discoverTime time.Duration, schedTime time.Time) *framework.Status{

  if p.Annotations != nil {
    p.Annotations["epsilon.discover.time"]=discoverTime.String()
    p.Annotations["epsilon.scheduling.time"]=time.Since(schedTime).String()
  }else{
    p.Annotations=map[string]string{
      "epsilon.discover.time": discoverTime.String(),
      "epsilon.scheduling.time": time.Since(schedTime).String(),
    }
  }

	binding := &corev1.Binding{
		ObjectMeta: metav1.ObjectMeta{
      Namespace: p.Namespace,
      Name: p.Name,
      UID: p.UID,
      Annotations: p.Annotations,
    },
		Target:     corev1.ObjectReference{Kind: "Node", Name: NodeName},
	}

	err := client.CoreV1().Pods(binding.Namespace).Bind(context.TODO(), binding, metav1.CreateOptions{})
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	return nil
}
