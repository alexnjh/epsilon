package main

import (
  "os"
  "fmt"
  "time"
  "errors"
  "context"
  "math/rand"
  "k8s.io/client-go/rest"
  "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
  "k8s.io/client-go/util/retry"

	log "github.com/sirupsen/logrus"
  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  configparser "github.com/bigkevmcd/go-configparser"
)


func getConfig(path string) (*configparser.ConfigParser, error){
  p, err := configparser.NewConfigParserFromFile(path)
  if err != nil {
    return nil,err
  }

  return p,nil
}


// retrieve the Kubernetes cluster client from outside of the cluster
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
  config.QPS = 100
  config.Burst = 200

  // generate the client based off of the config
  client, err := kubernetes.NewForConfig(config)
  if err != nil {
    panic("Failed to create kube client")
  }

	log.Info("Successfully constructed k8s client")
	return client
}

func SendToQueue(comm Communication, message []byte, queue string){
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

// Add a new pod event to the pod
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

func bind(client kubernetes.Interface, p corev1.Pod, NodeName string, discoverTime time.Duration, schedTime time.Time) error{

  if p.Annotations != nil {
    p.Annotations["epsilon.discover.time"]=discoverTime.String()
    p.Annotations["epsilon.scheduling.time"]=time.Since(schedTime).String()
  }else{
    p.Annotations=map[string]string{
      "epsilon.discover.time": discoverTime.String(),
      "epsilon.scheduling.time": time.Since(schedTime).String(),
    }
  }

  log.Infof("Attempting to bind %v/%v to %v", p.Namespace, p.Name, NodeName)
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
    //key, _ := cache.MetaNamespaceKeyFunc(p)


    // Resend pod for reschedule

		return errors.New(err.Error())
	}
	return nil
}
