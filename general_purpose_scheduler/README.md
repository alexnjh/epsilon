![title](https://alexneo.net/epsilon/scheduler.png "scheduler")
## General Purpose Scheduler Microservice

---

## :page_facing_up: Contents
- [Contents](#contents)
  - [Description](#desc)
  - [Deployment](#deploy)
  - [How does the general purpose scheduler microservice work?](#work)
  - [Directory and file description](#dir)
  - [Common questions](#qna)


<br>

<a name="desc"/></a> 
### :grey_exclamation: Description

The Scheduler service's goal is to schedule newly created pods.

<br>

---


<br>

<a name="deploy"/></a> 
### :grey_exclamation: Deployment of the scheduler service

Before deploying the scheduler.yaml file, please configure the environment variables to the correct values used by the queue microservice.

    env:
    - name: MQ_HOST
      value: "sched-rabbitmq-0.sched-rabbitmq.custom-scheduler.svc.cluster.local"
    - name: MQ_PORT
      value: "5672"
    - name: MQ_USER
      value: "guest"
    - name: MQ_PASS
      value: "guest"
    - name: RECEIVE_QUEUE
      value: "epsilon.distributed"
    - name: RETRY_QUEUE
      value: "epsilon.backoff"

**RECEIVE_QUEUE** indicates the queue the scheduler is going to be listening to for new pods send by the coordinator service.
**RETRY_QUEUE** indicates the queue the scheduler is going to send failed pods to.

<br>

---

<br>

<a name="work"/></a> 
### :grey_exclamation: The General Purpose Scheduler algorithm

![schedLifecycle](https://alexneo.net/epsilon/sched_lifecycle.JPG "scedLifecycle")


### 1. FETCH Stage
1. Wait for new pods to be send by the coordinator
2. When a new pod is received, proceed with fetching the details of the received pod from the Kube API Server.
3. Once the details of the pod is fetched form the Kube API server. The pod details can be send to the PreFilter stage.

### 2. PRE FILTER Stage

1. Send the pod through a list of preconfigured PreFilter Plugins
2. Once the pod passes all the checks by the PreFilter Plugins, the pod will be sent to the Filter Stage

### 3. FILTER Stage

1. Send the pod through a list of preconfigured Filter Plugins
2. Once the pod passes all the checks by the Filter Plugins, the pod will be sent to the PreScore Stage

### 4. PRE SCORE Stage
1. Send the pod through a list of preconfigured PreScore Plugins
2. Once all the PreScore plugins intitnlizes the required variables for use on the next stage, the pod will be sent to the Score Stage

### 5. SCORE Stage

1. Send the pod through a list of preconfigured Score Plugins
2. Once all the Score plugins return the  score value, the pod will be sent to the Score Stage

### 6. BIND Stage
1. During the stage the scheduler will commit the changes to the cluster and ends the scheduling lifecycle. Only during this stage the pod is considered to be deployed.

---


<a name="dir"/></a> 
### :grey_exclamation: Directory and File description

| Directory name                        | File name              | Description                                                                                                                   |
|---------------------------------------|------------------------|-------------------------------------------------------------------------------------------------------------------------------|
| /                                     | main.go                | Implementation code of the main routine                                                                                       |
| /                                     | processes.go           | Implementation code of the processes used by the scheduler during operation                                                   |
| /                                     | helper.go              | Implementation code containing common functions used by the different processes                                               |
| /                                     | experiment.go          | Used for experiments only. Does not affect the scheduler and can be removed                                                   |
| /                                     | event_handlers.go      | Implementation code that updates the local state of the scheduler                                                             |
| /framework/plugins                    | registry.go            | Implementation code of the plugin registry                                                                                    |
| /framework/plugins/helper             | node_affinity.go       | Helper functions used by node affinity plugin                                                                                 |
| /framework/plugins/helper             | taints.go              | Helper functions used by node affinity plugin                                                                                 |
| /framework/plugins/helper             | normalize_score.go     | Implementation of normalizing different scores returned by the score plugins                                                  |
| /framework/plugins/imagelocality      | image_locality.go      | Implementation code of the image locality plugin                                                                              |
| /framework/plugins/interpodaffinity   | filtering.go           | Implementation code of the inter pod affinity plugin                                                                          |
| /framework/plugins/interpodaffinity   | plugin.go              | Implementation code of the inter pod affinity plugin                                                                          |
| /framework/plugins/interpodaffinity   | scoring.go             | Implementation code of the inter pod affinity plugin                                                                          |
| /framework/plugins/nodeaffinity       | node_affinity.go       | Implementation code of the node affinity plugin                                                                               |
| /framework/plugins/nodename           | node_name.go           | Implementation code of the node name plugin                                                                                   |
| /framework/plugins/nodeports          | node_ports.go          | Implementation code of the node ports plugin                                                                                  |
| /framework/plugins/noderesources      | fit.go                 | Implementation code of the node resources plugin                                                                              |
| /framework/plugins/nodestatus         | node_status.go         | Implementation code of the node status plugin                                                                                 |
| /framework/plugins/nodeunschedulable  | node_unschedulable.go  | Implementation code of the node unschedulable plugin                                                                          |
| /framework/plugins/repeatpriority     | repeat_priority.go     | Implementation code of the repeat priority plugin                                                                             |
| /framework/plugins/resourcepriority   | resource_priority.go   | Implementation code of the resource priority plugin                                                                           |
| /framework/plugins/tainttoleration    | taint_toleration.go    | Implementation code of the taints and tolerations plugin                                                                      |
| /framework/plugins/volumebinding      | volume_binding.go      | Implementation code of the volume binding plugin                                                                              |
| /framework/plugins/volumerestrictions | volume_restrictions.go | Implementation code of the volume restrictions plugin                                                                         |
| /framework/v1alpha1                   | framework.go           | Contains scheduling framework implementation. Edit this file if changing plugin execution order or enabling/disabling plugins |
| /framework/v1alpha1                   | interface.go           | Contains Interface of scheduling framework                                                                                    |
| /framework/v1alpha1                   | registry.go            | Contains Interface of registry                                                                                                |
| /framework/v1alpha1                   | listers.go             | Contains Interface of custom listers used by Kube-Scheduler                                                                   |
| /internal/cache                       | cache.go               | Contains cache implementation of Kube-Scheduler, only modify this if you know what your doing                                 |
| /internal/parallelize                 | parallelism.go         | Contains parallel execution implementation of Kube-Scheduler, only modify this if you know what your doing                    |
| /k8s.io                               | *                      | This are Kube-Scheduler library files, only modify this if you know what your doing                                           |
| /scheduler                            | helper.go              | Contain helper functions used by the scheduler implementation                                                                 |
| /scheduler                            | scheduler.go           | Implementation of the Epsilon scheduling lifecycle                                                                            |
| /scheduler                            | types.go               | Contain struct types used by the scheduler                                                                                    |
| /scheduler/config                     | *                      | This are Kube-Scheduler library files, only modify this if you know what your doing                                           |
| /scheduler/util                       | *                      | This are Kube-Scheduler library files, only modify this if you know what your doing                                           |
| /scheduler/metrics                    | *                      | This are Kube-Scheduler library files, only modify this if you know what your doing                                           |
| /yaml                                 | scheduler.yaml         | Deployment file to deploy the scheduler in a Kubernetes cluster                                                               |

---

<br>

<a name="qna"/></a> 
### :grey_exclamation: Common questions

<dl>
  
  <dt>Changing the list of plugins used by the scheduler</dt>
  <dd>Modification of plugin details can be achieved by editing the <b>/framework/v1alpha1/framework.go</b> file under the <b>NewFramework()</b> function</dd>

  <dt>Change a scheduler plugin's weight</dt>
  <dd>Same as changing the list of plugins used by the scheduler shown above</dd>
  
  <dt>Change scheduling lifecycle</dt>
  <dd>Modification of the scheduling lifecycle can be achieved by editing the <b>/scheduler/scheduler.go</b> file under the <b>Schedule()</b> function</dd>
  
  <dt>Adding a new scheduler plugin</dt>
  <dd>
  
  In order to add a new scheduler plugin, follow the steps shown below
  
      1. Create a new folder in /framework/plugins
      2. Write the plugin implementation and store the file in the new folder created in 1
      3. Open /framework/plugins/registry.go and import the new folder that was created in 2
      4. Initilize the newly added plugin by calling its New function inside the NewInTreeRegistry() function
      5. Open /framework/v1alpha1/framework.go and add the plugin into the correct plugin list. 
         Each list is for a different stage, make sure to add the new plugin to the correct list

  </dd>
  
</dl>

<br>

---
