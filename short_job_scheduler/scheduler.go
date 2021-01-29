package main

import(
  "fmt"
  "sort"
  // "context"
  "errors"
  "k8s.io/client-go/kubernetes"
  "k8s.io/apimachinery/pkg/labels"
  pcglib "github.com/MichaelTJones/pcg"
  // metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  corev1 "k8s.io/api/core/v1"
  corelisters "k8s.io/client-go/listers/core/v1"
)

const (
  // DefaultMilliCPURequest defines default milli cpu request number.
	DefaultMilliCPURequest int64 = 100 // 0.1 core

	// DefaultMemoryRequest defines default memory request size.
	DefaultMemoryRequest int64 = 200*1024*1024 // 200 MB
)

type ShortJobScheduler struct{
    previousNodeIndex   int
    clientset           kubernetes.Interface
    podlister           corelisters.PodLister
    nodelister          corelisters.NodeLister
}

func NewShortJobScheduler (clientset kubernetes.Interface, podlister corelisters.PodLister, nodelister corelisters.NodeLister) *ShortJobScheduler{

  // Get the pod resource with this namespace/name
  list, err := nodelister.List(labels.NewSelector())
  if err != nil {

  }

  pcg := pcglib.NewPCG64()
  index := pcg.Bounded(uint64(len(list)))

  return &ShortJobScheduler{
    previousNodeIndex:  int(index),
    clientset:          clientset,
    podlister:          podlister,
    nodelister:         nodelister,
  }



}

// ObjectDeleted is called when an object is deleted
func (s *ShortJobScheduler) Schedule(pod *corev1.Pod) (name string, err error){

  list, err := s.nodelister.List(labels.NewSelector())

  if err != nil {

  }

  // Sort slice to ensure correct order
  sort.SliceStable(list, func(i, j int) bool {
      return list[i].Name < list[j].Name
  })

  for _ = range(list){

      fmt.Println(list[s.previousNodeIndex].Name)

      index := (s.previousNodeIndex+1) % len(list)

      if s.checkNode(pod, list[index]){
        s.previousNodeIndex = index
        return list[index].Name, nil
      }

      s.previousNodeIndex = index

  }


  return "", errors.New("Unable to find a suitable node to schedule")



}

// UpdateNodeList is called when an object is deleted
func (s *ShortJobScheduler) checkNode(pod *corev1.Pod, node *corev1.Node) bool{


  if checkTaints(pod,node) {
    // pods, err := s.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
    //   FieldSelector: "spec.nodeName=" + node.Name,
    // })

    // if err != nil {
    //   return false
    // }
  //
  //   var totalCpu = int64(0)
  //   var totalMem = int64(0)
  //
  //   pods.Items := append(pods.Items, pod)
  //
  //   for _, p := range(pods.Items){
  //     for _, cont := range(p.Spec.Containers){
  //       cpu, _ :=      cont.Resources.Requests.Cpu().AsInt64()
  //       mem, _ :=      cont.Resources.Requests.Memory().AsInt64()
  //
  //       if cpu == 0 {
  //         cpu = DefaultMilliCPURequest
  //       }
  //
  //       if mem == 0 {
  //         cpu = DefaultMemoryRequest
  //       }
  //
  //       totalCpu+=cpu
  //       totalMem+=mem
  //
  //     }
  //   }

    var totalMem =  int64(0)
    capacity, _ :=   node.Status.Capacity.Memory().AsInt64()

    for _, p := range(pod.Spec.Containers){
      mem, _ :=   p.Resources.Requests.Memory().AsInt64()
      totalMem += mem
    }

    if totalMem > capacity {
      return false
    }

    return true

  }

  return false



}

func checkTaints(pod *corev1.Pod, node *corev1.Node) bool{

  var passTaints = 0

  for _, taint := range(node.Spec.Taints){
    if taint.Effect == corev1.TaintEffectNoSchedule{
        for _, tol := range(pod.Spec.Tolerations){
          if tol.Key == taint.Key && tol.Value == taint.Value && (tol.Effect == taint.Effect || len(tol.Effect) == 0){
            passTaints++
          }
        }
    }
  }

  if passTaints != len(node.Spec.Taints) {
    return false
  }

  return true

}
