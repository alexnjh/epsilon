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

package scheduler

import (
  "sort"
  "context"
  "errors"
  "sync"
  "time"
  "math/rand"
  "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/plugins"
  "k8s.io/client-go/informers"

  log "github.com/sirupsen/logrus"
  v1 "k8s.io/api/core/v1"
  corelisters "k8s.io/client-go/listers/core/v1"
  framework "github.com/alexnjh/epsilon/general_purpose_scheduler/framework/v1alpha1"
  internalcache "github.com/alexnjh/epsilon/general_purpose_scheduler/internal/cache"
  clientset "k8s.io/client-go/kubernetes"
  pcglib "github.com/MichaelTJones/pcg"

  "github.com/alexnjh/epsilon/general_purpose_scheduler/k8s.io/kubernetes/pkg/controller/volume/scheduling"
)

const (
	// BindTimeoutSeconds defines the default bind timeout
	BindTimeoutSeconds = 100
	// SchedulerError is the reason recorded for events when an error occurs during scheduling a pod.
	SchedulerError = "SchedulerError"
)

// Scheduler is responsible for scheduling pods
type Scheduler struct {

  // Kubernetes client interface (Use to fetch information from kube-api server).
  client clientset.Interface

  // Plugin registry containing all the plugins constructors.
  registry framework.Registry

  // Framework instance containing all the plugins that are initialized.
  fw framework.Framework

  // Snapshot of the current cluster state.
  snapshot internalcache.Snapshot

  // Volume binder for scheduler to bind volumes if needed.
  volumeBinder scheduling.SchedulerVolumeBinder

  // It is expected that changes made via SchedulerCache will be observed
	// by NodeLister and Algorithm.
	SchedulerCache internalcache.Cache

  // Node lister that is used to list nodes from the local cache
  nodeLister corelisters.NodeLister

  // Pod lister that is used to list pods from the local cache
  podLister corelisters.PodLister

	// Disable pod preemption or not.
	DisablePreemption bool

  //Percentage to node score
  percentageNodeScore int
}

// Invokes the scheduling routine
func (s *Scheduler) Schedule(con context.Context, pod *v1.Pod) (scheduleResult ScheduleResult, err error){


  s.SchedulerCache.UpdateSnapshot(&s.snapshot)

  nodeList, err := s.snapshot.NodeInfos().List()

  if err != nil {
    return ScheduleResult{}, err
  }

  // Create list of viable nodes
  viableNodes := make([]*v1.Node, 0)
  var lenOfArr float64 = float64(len(nodeList))


  // Check number of nodes and select a subset if length of node list exceed 50
  if(lenOfArr > 50){

    // Generate random seed using current time
    rand.Seed(time.Now().UnixNano())

    // Fisher–Yates shuffle
    for i := len(nodeList) - 1; i > 0; i-- {
        j := rand.Intn(i + 1)
        nodeList[i], nodeList[j] = nodeList[j], nodeList[i]
    }

    lenOfArr = lenOfArr*(float64(s.percentageNodeScore)/100.0)+1

    if(!s.processSubset(&nodeList,pod,&viableNodes,int(lenOfArr))){
      return ScheduleResult{}, errors.New("Fail to schedule pod, Fail to select node from a subset of nodelist")
    }

  }else{

    if(!s.processFullset(&nodeList,pod,&viableNodes)){
      return ScheduleResult{}, errors.New("Fail to schedule pod, Fail to select node from nodelist")
    }

  }

  if(len(viableNodes) == 0){

      if *pod.Spec.Priority == 0 {
        return ScheduleResult{}, errors.New("Pod priority too low to preempt other pods")
      }

      if pod.Spec.PreemptionPolicy != nil && *pod.Spec.PreemptionPolicy == v1.PreemptNever {
        return ScheduleResult{}, errors.New("Pod preeemption policy do not allow preemption")
      }

      // Create array to store the nodes for selection
      viableNodes := make([]*framework.NodeInfo, 0)

      // Check number of nodes and select a subset if length of node list exceed 50
      if(lenOfArr > 50){
        s.processSubsetPreemption(&nodeList,pod,&viableNodes,int(lenOfArr))
      }else{
        // Search for suitable nodes to select pods for premption
        s.processFullsetPreemption(&nodeList,pod,&viableNodes)
      }

      // Check if a suitable node is found
      if (len(viableNodes) == 0){
        return ScheduleResult{}, errors.New("Preemption not possible, no suitable node found")
      }else{

        // Search all the pods in each viable node to select pod to preempt
        for _ , nodeinfo := range viableNodes{
          nominated_pod, err := getPodToPreempt(nodeinfo,pod)
          if (err != nil){
            log.Infof(err.Error())
            continue
          }else{
            log.Infof("Pod %s selected for preemption to deploy %s",nominated_pod.Name,pod.Name)

            return ScheduleResult{
              SuggestedHost: nodeinfo.Node().Name,
              NorminatedPod: nominated_pod,
              }, nil
          }
        }

        return ScheduleResult{}, errors.New("Preemption not possible, No viable pod found to preempt")

      }


    }

  // Get each viable node's priority value
  cyclestate := framework.NewCycleState()
  results, err := s.prioritizeNodes(context.TODO(), cyclestate, pod, viableNodes)

  if err != nil {
    return ScheduleResult{}, err
  }

  // Sort node list in descending order
  sort.Slice(results, func(i, j int) bool {
    return results[i].Score > results[j].Score
  })

  if(len(results) == 1){
    selectedNode := results[0].Name
    go s.fw.IncreaseNodeUsageFactor(selectedNode)
    return ScheduleResult{results[0].Name,nil},nil
  }

  if(results[0].Score != results[1].Score){
    selectedNode := results[0].Name
    go s.fw.IncreaseNodeUsageFactor(selectedNode)
    return ScheduleResult{results[0].Name,nil},nil
  }

  // Select node randomly based on PCG
  pcg := pcglib.NewPCG64()

  if results[0].Score != results[(len(results)-1)].Score {
    for idx, value := range results[1:] {
      if (results[0].Score != value.Score){

        index := pcg.Bounded(uint64(idx))
        selectedNode := results[index].Name
        go s.fw.IncreaseNodeUsageFactor(selectedNode)


        // selectedNode := s.selectNodeBasedOnRepeatScore(results[:idx+1])

        return ScheduleResult{
          SuggestedHost: selectedNode,
          NorminatedPod: nil,
        }, nil
      }
    }
  }

  index := pcg.Bounded(uint64((len(results)-1)))
  selectedNode := results[index].Name
  go s.fw.IncreaseNodeUsageFactor(selectedNode)


  return ScheduleResult{
    SuggestedHost: selectedNode,
    NorminatedPod: nil,
  }, nil
}

// Function use to generate priority values for each node
func (s *Scheduler) prioritizeNodes(
  ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	nodes []*v1.Node,
) (framework.NodeScoreList, error) {


  status := s.fw.RunPreScorePlugins(context.TODO(), state, pod, nodes)

  if status.IsSuccess() {
    ps, status := s.fw.RunScorePlugins(context.TODO(), state, pod, nodes)

    if !status.IsSuccess() {
      log.Errorf(status.Message())
    }

    scoreList := make(map[string]int64, 0)

    for _, pg := range ps {

      for _, p := range pg {
          scoreList[p.Name] += p.Score
      }
    }

    scoreResults := make([]framework.NodeScore, 0)

    for key, value := range scoreList {
      scoreResults = append(scoreResults, framework.NodeScore{
        Name: key,
        Score: value,
      })
    }

    return scoreResults, nil


  }else{
    log.Errorf(status.Message())
  }


  return framework.NodeScoreList{}, nil

}

// Create new scheduler object
func New(
  volumeBinder scheduling.SchedulerVolumeBinder,
  client clientset.Interface,
  cache internalcache.Cache,
  kubefactory informers.SharedInformerFactory,
  node_lister corelisters.NodeLister,
  pod_lister  corelisters.PodLister,
  disablePreemption bool,
  percentageNodeScore int,
  ) (*Scheduler, error){

registry := plugins.NewInTreeRegistry()
snapshot := internalcache.NewEmptySnapshot()
fw, _ := framework.NewFramework(registry,client,snapshot,volumeBinder)

return &Scheduler{
  client: client,
  registry: registry,
  fw: &fw,
  snapshot: *snapshot,
  volumeBinder: volumeBinder,
  SchedulerCache: cache,
  nodeLister: node_lister,
  podLister: pod_lister,
  DisablePreemption: disablePreemption,
  percentageNodeScore: percentageNodeScore,
}, nil


}


// Get a subset of nodes from nodelist
func getSubset(a *[]*framework.NodeInfo, numSamples int) []*framework.NodeInfo {

  if (len(*a) == 0){
    return *a
  }

  if len(*a) <= numSamples {
    var temp = *a
    *a = nil;
    return temp
  }

  // Create list of viable nodes
  viableNodes := make([]*framework.NodeInfo, 0)

  for i := (len(*a)-1); i >= (len(*a)-numSamples); i-- {
    viableNodes = append(viableNodes, (*a)[i])

	}

  *a = (*a)[:len(*a)-numSamples]   // Truncate slice.

  return viableNodes

}

// Run PreFilter and Filter plugins for a subset of nodes from nodelist
func (s *Scheduler) processSubset(nodeList *[]*framework.NodeInfo, pod *v1.Pod, viableNodes *[]*v1.Node, numOfViable int) bool{

  if(len(*nodeList) == 0){
    return true
  }

  var a = 0;
  var mux sync.Mutex
  subset := getSubset(nodeList,numOfViable*2)

  for _ , n := range subset {

      go func(s *Scheduler, pod *v1.Pod, nodeInfo *framework.NodeInfo,viableNodes *[]*v1.Node, mux *sync.Mutex, numOfViable int) {

        cyclestate := framework.NewCycleState()

        status := s.fw.RunPreFilterPlugins(context.TODO(), cyclestate, pod)

        if status.IsSuccess() {
          node := *nodeInfo.Node()

          status := s.fw.RunFilterPlugins(context.TODO(), cyclestate, pod, nodeInfo)

          if status.Merge().IsSuccess() {

            if len(*viableNodes) <= numOfViable{
              *viableNodes = append(*viableNodes, &node)
            }
          }

          mux.Lock()
          a++
          mux.Unlock()



        }

      }(s,pod, n, viableNodes, &mux, numOfViable)

    }

    for {

          if len(*viableNodes) >= numOfViable  {
            return true
          }else if a == len(subset) && (len(*viableNodes) < numOfViable) {
            s.processSubset(nodeList, pod, viableNodes, numOfViable)
            return true
          }
    }

}


// processSubsetPremption is used for finding viable nodes for premption by searching the a subset of nodes from the nodelist
// DO NOT RUN THIS UNLESS DOING PREMPTION!!!
func (s *Scheduler) processSubsetPreemption(nodeList *[]*framework.NodeInfo, pod *v1.Pod, viableNodes *[]*framework.NodeInfo, numOfViable int) bool{

  if(len(*nodeList) == 0){
    return true
  }

  var a = 0;
  var mux sync.Mutex
  subset := getSubset(nodeList,numOfViable*2)

  for _ , n := range subset {

      go func(s *Scheduler, pod *v1.Pod, nodeInfo *framework.NodeInfo,viableNodes *[]*framework.NodeInfo, mux *sync.Mutex, numOfViable int) {

        cyclestate := framework.NewCycleState()

        status := s.fw.RunPreFilterPlugins(context.TODO(), cyclestate, pod)

        if status.IsSuccess() {

          status := s.fw.RunFilterPlugins(context.TODO(), cyclestate, pod, nodeInfo)

          if checkIfPreemptable(status) {
            *viableNodes = append(*viableNodes, nodeInfo)
          }

          mux.Lock()
          a++
          mux.Unlock()

        }

      }(s,pod, n, viableNodes, &mux, numOfViable)

    }

    for {

          if len(*viableNodes) >= numOfViable  {
            return true
          }else if a == len(subset) && (len(*viableNodes) < numOfViable) {
            s.processSubsetPreemption(nodeList, pod, viableNodes, numOfViable)
            return true
          }
    }

}


// Run PreFilter and Filter plugins for all nodes in the nodelist
func (s *Scheduler) processFullset(nodeList *[]*framework.NodeInfo, pod *v1.Pod, viableNodes *[]*v1.Node) bool{

var wg sync.WaitGroup
wg.Add(len(*nodeList))

for _ , n := range *nodeList {

    go func(s *Scheduler, pod *v1.Pod, nodeInfo *framework.NodeInfo,viableNodes *[]*v1.Node) {


      defer wg.Done()

      cyclestate := framework.NewCycleState()

      status := s.fw.RunPreFilterPlugins(context.TODO(), cyclestate, pod)

      if status.IsSuccess() {
        node := *nodeInfo.Node()

        status := s.fw.RunFilterPlugins(context.TODO(), cyclestate, pod, nodeInfo)

        if status.Merge().IsSuccess() {
          *viableNodes = append(*viableNodes, &node)
        }


      }

    }(s,pod, n, viableNodes)

  }

  wg.Wait()
  return true
}


// processFullsetPremption is used for finding viable nodes for premption by searching the whole list of nodes
// DO NOT RUN THIS UNLESS DOING PREMPTION!!!
func (s *Scheduler) processFullsetPreemption(nodeList *[]*framework.NodeInfo, pod *v1.Pod, viableNodes *[]*framework.NodeInfo) bool{

var wg sync.WaitGroup
wg.Add(len(*nodeList))

for _ , n := range *nodeList {

    go func(s *Scheduler, pod *v1.Pod, nodeInfo *framework.NodeInfo, viableNodes *[]*framework.NodeInfo) {


      defer wg.Done()

      cyclestate := framework.NewCycleState()

      status := s.fw.RunPreFilterPlugins(context.TODO(), cyclestate, pod)

      if status.IsSuccess() {

        status := s.fw.RunFilterPlugins(context.TODO(), cyclestate, pod, nodeInfo)

        if checkIfPreemptable(status) {
          *viableNodes = append(*viableNodes, nodeInfo)
        }
      }

    }(s,pod, n,viableNodes)

  }

  wg.Wait()
  return true
}

// Check if a node is capable to be used as a preemption node
func checkIfPreemptable(status framework.PluginToStatus) bool{

  if(len(status) != 1){
    return false
  }

  for k , _ := range status {
    if (k != "NodeResourcesFit"){
        return false
    }
  }

  return true


}

// Check if pod is preemptable
func getPodToPreempt(nodeInfo *framework.NodeInfo, preemptorpod *v1.Pod) (*v1.Pod,error){

  pods := nodeInfo.Pods

  preemptorPodRequest := computePodResourceRequest(preemptorpod)
  preemptorPriority := *preemptorpod.Spec.Priority

  ignore_pod := make(map[string]int)

  // Do not preempt a pod that is used to preempt another pod
  for _ , n := range nodeInfo.Node().Status.Conditions {

    if (n.Type == "Preemption"){
      ignore_pod[n.Message]=1
    }

  }


  // Sort the pod by priority
  sort.Slice(pods, func(i, j int) bool {
    return *pods[i].Pod.Spec.Priority < *pods[j].Pod.Spec.Priority
  })

  // Search for a pod to preempt
  for _ , podinfo := range pods{

    pod := podinfo.Pod

    if ignore_pod[pod.Name] == 0 &&
      *pod.Spec.Priority <= preemptorPriority &&
      pod.Namespace == preemptorpod.Namespace &&
      pod.Status.Phase == v1.PodRunning{

      nominatedPodRequest := computePodResourceRequest(podinfo.Pod)

      // If pod do not state resource requirements preempt the pod based on the assumption
      // that the resources is enough to deploy the preemptor pod
      if nominatedPodRequest.MilliCPU == 0 &&
    		nominatedPodRequest.Memory == 0 &&
    		nominatedPodRequest.EphemeralStorage == 0 {
        log.Infof("Pod to be preempted: %s", pod.Name)
    		return podinfo.Pod,nil
    	}

      // Ensure pod resources and free resources on the node is enough for the preemptor pod to be deployed
      if (nominatedPodRequest.Memory+nodeInfo.Allocatable.Memory > preemptorPodRequest.Memory)&&
      (nominatedPodRequest.EphemeralStorage+nodeInfo.Allocatable.EphemeralStorage > preemptorPodRequest.EphemeralStorage){
        log.Infof("Pod to be preempted: %s", pod.Name)
        return podinfo.Pod,nil
      }

    }
  }

  return nil, errors.New("No pods viable")
}
