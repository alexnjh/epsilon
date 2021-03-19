![title](https://alexneo.net/epsilon/scheduler.png "scheduler")
## General Purpose Scheduler Microservice

---

## :page_facing_up: Contents
- [Contents](#contents)
  - [Description](#desc)
  - [Deployment](#deploy)
  - [How does the retry microservice work?](#work)
  - [Directory description](#dir)
  - [File description](#file)
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

Before deploying the retry.yaml file, please configure the environment variables to the correct values used by the queue microservice.

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

<dl>
  <dt>docker</dt>
  <dd>contain the dockerfile for generating the scheduler service docker image</dd>
  | File Name                 | Description                        |
  |---------------------------|------------------------------------|
  | Dockerfile                | Use by docker to build the image   |
  | general_purpose_scheduler | Executable binary of the scheduler |
  
  <dt>framework</dt>
  <dd>contains all the scheduling plugin implementations</dd>

  <dt>internal</dt>
  <dd>contains the cache implementaion used by the default kubernetes scheduler (Kube-Scheduler)</dd>

  <dt>yaml</dt>
  <dd>contains the deployment yaml file</dd>
  
  <dt>k8s.io/kubernetes/pkg</dt>
  <dd>contains Kubernetes libary files</dd>
  
  <dt>scheduler</dt>
  <dd>contains the implementation of the scheduler service</dd>
  
</dl>

<br>

---

<br>

<a name="file"/></a> 
### :grey_exclamation: Key files to take note

<dl>
  <dt>main.go</dt>
  <dd>contain the main routine. All initialization of required variables including the waiting for new pods that failed is in this file</dd>

</dl>

<dl>
  <dt>helper.go</dt>
  <dd>contains helper functions used by the main routine</dd>

</dl>

<br>

---

<br>

<a name="qna"/></a> 
### :grey_exclamation: Common questions

<dl>
  <dt>How to change the retry algorithm?</dt>
  <dd>The function WaitAndSend() in line 176 of main.go, contains the implementation of the retry algorithm. By modifiying this function the retry algorithm can be modified.</dd>

</dl>

<br>

---
