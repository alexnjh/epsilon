package main

import (
  "os"
  "fmt"
  "math"
  "time"
  "strconv"
  "bufio"
  "net/http"
  "strings"
  "errors"
  "context"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/util/retry"
  // "k8s.io/sample-controller/pkg/signals"
  "k8s.io/apimachinery/pkg/labels"
  rabbithole "github.com/michaelklishin/rabbit-hole/v2"
  kubeinformers "k8s.io/client-go/informers"
  corev1 "k8s.io/api/core/v1"
  appsv1 "k8s.io/api/apps/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  log "github.com/sirupsen/logrus"
  applisters "k8s.io/client-go/listers/apps/v1"
  configparser "github.com/bigkevmcd/go-configparser"
  rabbitplugin "github.com/alexnjh/autoscaler/plugins/rabbitmq"
  queueplugin "github.com/alexnjh/autoscaler/plugins/queue_theory"
  linearplugin "github.com/alexnjh/autoscaler/plugins/linear_regression"
  schedplugin "github.com/alexnjh/autoscaler/plugins/scheduler_prob"
)

const (
  DefaultConfigPath = "/go/src/app/config.cfg"
)


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

  var mqHost, mqManagePort, mqUser, mqPass, namespace, defaultQueue, pcURL, updateInterval string

  if err != nil {

    log.Errorf(err.Error())

    namespace = os.Getenv("POD_NAMESPACE")
    mqHost = os.Getenv("MQ_HOST")
    mqManagePort = os.Getenv("MQ_MANAGE_PORT")
    mqUser = os.Getenv("MQ_USER")
    mqPass = os.Getenv("MQ_PASS")
    defaultQueue = os.Getenv("DEFAULT_QUEUE")
    updateInterval = os.Getenv("INTERVAL")
    pcURL = os.Getenv("PC_METRIC_URL")

    if len(mqHost) == 0 ||
    len(mqManagePort) == 0 ||
    len(mqUser) == 0 ||
    len(mqPass) == 0 ||
    len(defaultQueue) == 0 ||
    len(namespace) == 0 ||
    len(pcURL) == 0 ||
    len(updateInterval) == 0{
  	   log.Fatalf("Config not found, Environment variables missing")
    }


  }else{

    mqHost, err = config.Get("QueueService", "hostname")
    if err != nil {
      log.Fatalf(err.Error())
    }
    mqManagePort, err = config.Get("QueueService", "management_port")
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
    namespace, err = config.Get("DEFAULTS", "namespace")
    if err != nil {
      log.Fatalf(err.Error())
    }
    pcURL, err = config.Get("CoordinatorService", "metrics_absolute_url")
    if err != nil {
      log.Fatalf(err.Error())
    }
    updateInterval, err = config.Get("DEFAULTS", "update_interval")
    if err != nil {
      log.Fatalf(err.Error())
    }
  }

  interval, err := strconv.Atoi(updateInterval)
  if err != nil {
	   log.Fatalf(err.Error())
  }

  queueList := map[string]bool {
    fmt.Sprintf(defaultQueue): true,
  }

  kubeClient := getKubernetesClient()
  kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
  nodeInformer := kubeInformerFactory.Core().V1().Nodes().Informer()
  nodeLister := kubeInformerFactory.Core().V1().Nodes().Lister()
  deployInformer := kubeInformerFactory.Apps().V1().Deployments().Informer()
  deployLister := kubeInformerFactory.Apps().V1().Deployments().Lister()

  // use a channel to synchronize the finalization for a graceful shutdown
	stopCh := make(chan struct{})
	defer close(stopCh)

  kubeInformerFactory.Start(stopCh)

  log.Infof("Waiting for cache to be populated")
  for {
    if nodeInformer.HasSynced() && deployInformer.HasSynced(){
        log.Infof("Cache is populated\n\n")
        break
    }
  }

  rmqc, _ := rabbithole.NewClient(fmt.Sprintf("http://%s:%s",mqHost,mqManagePort), mqUser, mqPass)

  res, err := rmqc.Overview()

  if err != nil{
    log.Fatalf(err.Error())
  }

  log.Infof("RabbitMQ Server Information")
  log.Infof("---------------------------")
  log.Infof("Management version: %s",res.ManagementVersion)
  log.Infof("Erlang version: %s",res.ErlangVersion)

  // Initialize Plugins

  pluginList := make(map[string]AutoScalerPlugin)

  qs, err := rmqc.ListQueues()

  if err != nil{
    log.Fatalf(err.Error())
  }

  for _ , queue := range(qs){
    if queue.Name == defaultQueue{

      // Initialize Plugins
      pluginList["rabbitmq"]=rabbitplugin.NewRabbitMQPlugin("rabbitmq",queue.Vhost,queue.Name,0.5,rmqc)
      pluginList["schedprob"]=schedplugin.NewSchedProbPlugin("schedprob",queue.Name,0.5)
      pluginList["reggression"]=linearplugin.NewLinearRegressionPlugin("reggression",queue.Name,5)
      pluginList["queuetheory"]=queueplugin.NewQueueTheoryPlugin("queuetheory",0.5,fmt.Sprintf("http://%s",pcURL))

      break
    }
  }

  temp := make([]ComputeResult,len(pluginList))

  for {

    qs, err := rmqc.ListQueues()

    if err != nil{
      log.Fatalf(err.Error())
    }

    for _ , queue := range(qs){
      if queueList[queue.Name] {

        log.Infof("Queue Name: %s\n------------------------------",queue.Name)

          nodeList, err := nodeLister.List(labels.NewSelector())

          if err != nil{
            log.Fatalf(err.Error())
          }

          noOfPendingPods := float64(queue.MessagesReady)
          noOfNodes := float64(len(nodeList))
          noOfSched := float64(queue.Consumers)

          var i = 0
          for key , plugin := range(pluginList){
            temp[i] = plugin.Compute(noOfPendingPods,noOfNodes,noOfSched)
            log.Infof("%s Decision:  %s",key,temp[i])
            i++
          }

          result := makeDecision(temp)
          UpdateDeployment(kubeClient,deployLister,namespace,queue.Name,result)
      }
    }

    log.Infof("Sleeping for %d seconds before testing again...",interval)
    time.Sleep(time.Duration(interval)*time.Second)

  }
}


// Update the replica count of the scheduler services
func UpdateDeployment(client kubernetes.Interface, lister applisters.DeploymentLister, namespace string, queueName string, decision ComputeResult){

  if decision == DoNotScale {
    return
  }

  labelmap := map[string]string{
    "epsilon.queue" : queueName,
  }

  var deployment []*appsv1.Deployment

  retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {

    if len(namespace) == 0{
      deployment, _ = lister.List(labels.SelectorFromSet(labelmap))
    }else{
      deployment, _ = lister.Deployments(namespace).List(labels.SelectorFromSet(labelmap))
    }

    if len(deployment) > 1 {
      return errors.New("More than one deployment found for queue name")
    }

    if len(deployment) == 0 {
      return errors.New("No deployment found for queue name")
    }

    obj := deployment[0]

    if decision == ScaleUp {
      *obj.Spec.Replicas +=1
    }else if decision == ScaleDown && *obj.Spec.Replicas > 1{
      *obj.Spec.Replicas -=1
    }else{
      return nil
    }

    _, updateErr := client.AppsV1().Deployments(obj.Namespace).Update(context.TODO(), obj, metav1.UpdateOptions{});
    return updateErr
  })
  if retryErr != nil {
    panic(fmt.Errorf("Update failed: %v", retryErr))
  }

}

// Convert prometheus formatted metrics into a map
func promToMap(url string) map[string]string{

  metricMap := make(map[string]string)

  resp, err := http.Get(url)
  if err != nil {
    log.Fatalf(err.Error())
  }

  scanner := bufio.NewScanner(resp.Body)
  for scanner.Scan() {
    if len(scanner.Text()) > 0{
      if scanner.Text()[0] != '#'{
        s := strings.Split(scanner.Text(), " ")
        if (len(s) == 2){
          metricMap[s[0]]=s[1]
        }
      }
    }
  }
  if err := scanner.Err(); err != nil {
    fmt.Fprintln(os.Stderr, "reading standard input:", err)
  }

  return metricMap
}

// Consolidate all the decisions made by the plugins and decide on the next course of action.
func makeDecision(result []ComputeResult) ComputeResult{
  count := make(map[ComputeResult]int)

  for _, r := range(result){
    count[r] +=1
  }


  largest := ScaleUp

  if(count[ScaleUp] > count[ScaleDown]){
    largest = ScaleUp
  }else{
    largest = ScaleDown
  }

  if(count[ScaleUp] == count[ScaleDown]){
    largest = DoNotScale
  }

  return largest

}

// Update kube-api server of the autoscaler's scale down operations
func addScaleDownEvent(client kubernetes.Interface, obj *appsv1.Deployment){
  client.CoreV1().Events(obj.Namespace).Create(context.TODO(), &corev1.Event{
    Count:          1,
    Message:        "Scheduler replica reduced by 1",
    Reason:         "ScaleDown",
    LastTimestamp:  metav1.Now(),
    FirstTimestamp: metav1.Now(),
    Type:           "Information",
    Source: corev1.EventSource{
      Component: "autoscaler",
    },
    InvolvedObject: corev1.ObjectReference{
      Kind:      "Deployment",
      Name:      obj.Name,
      Namespace: obj.Namespace,
      UID:       obj.UID,
    },
    ObjectMeta: metav1.ObjectMeta{
      GenerateName: obj.Name + "-",
    },
  },metav1.CreateOptions{})
}

// Update kube-api server of the autoscaler's scale up operations
func addScaleUpEvent(client kubernetes.Interface, obj *appsv1.Deployment){
  client.CoreV1().Events(obj.Namespace).Create(context.TODO(), &corev1.Event{
    Count:          1,
    Message:        "Scheduler replica increased by 1",
    Reason:         "ScaleUp",
    LastTimestamp:  metav1.Now(),
    FirstTimestamp: metav1.Now(),
    Type:           "Information",
    Source: corev1.EventSource{
      Component: "autoscaler",
    },
    InvolvedObject: corev1.ObjectReference{
      Kind:      "Deployment",
      Name:      obj.Name,
      Namespace: obj.Namespace,
      UID:       obj.UID,
    },
    ObjectMeta: metav1.ObjectMeta{
      GenerateName: obj.Name + "-",
    },
  },metav1.CreateOptions{})
}

// Calculate the scheduler conflict probability
// N = No of schedulers
// K = No of nodes
func calProb(N,K float64) float64{
  return Factorial(N)/(Factorial(N-K)*math.Pow(N,K))
}

func Factorial(n float64)(result float64) {
	if (n > 0) {
		result = n * Factorial(n-1)
		return result
	}
	return 1
}
