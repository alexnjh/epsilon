![title](https://alexneo.net/epsilon/coordinator.png "Coordinator")
## Coordinator Microservice

---

## :page_facing_up: Contents
- [Contents](#contents)
  - [Description](#desc)
  - [Deployment](#deploy)
  - [How does the coordinator work?](#algo)
  - [Directory and File description](#dir)
  - [Common questions](#qna)


<br>

<a name="desc"/></a> 
### :grey_exclamation: Description

The coordinator's goal is to monitor the kubernetes cluster for newly created pods. Once a pod is created the coordinator will proceeed to to examine the details of the pod and send it to the queue service.

<br>

---


<br>

<a name="deploy"/></a> 
### :grey_exclamation: Deployment of the coodinator service

Before deploying the coordinator.yaml file, please configure the environment variables to the correct values used by the queue microservice.

    env:
    - name: MQ_HOST
      value: "sched-rabbitmq-0.sched-rabbitmq.custom-scheduler.svc.cluster.local"
    - name: MQ_PORT
      value: "5672"
    - name: MQ_MANAGE_PORT
      value: "15672"
    - name: MQ_USER
      value: "guest"
    - name: MQ_PASS
      value: "guest"
    - name: DEFAULT_QUEUE
      value: "epsilon.distributed"

<br>

The **DEFAULT_QUEUE** is the queue used by the general-purpose scheduler. In Epsilon, atleast one scheduler service need to act as the default scheduler.

---

<br>

<a name="work"/></a> 
### :grey_exclamation: Coordinator algorithm

**[STEP 1]**
<br>
The coordinator monitors for the creation of a new pod.

**[STEP 2]**
<br>
When a new pod is created the coodinator will check the **SchedulerName** field of the pod configuration to ensure that the pod is configured to be scheduled by the Epsilon scheduler.

The function that checks for the scheduler name can be found in **main.go at line 193-195**

**[STEP 3]**
<br>
Once a pod passes the check in STEP 2. The coordinator will put the pod in a FIFO queue till the handler processes it.

**[STEP 4]**
<br>
The coordinator handler will fetch the pod from the queue and check  if a specific shceduler queue is specified inside the pod configuration. If there isn't a scheduler queue specified in the pod labels the coordinator will send the pod to the default scheduler queue.

The **handle.go** file contains the pod handling algorithm and all the coodinator handler functions. The function that fetches a pod from the queue and processes it is called **ObjectSync(). (Line 88-92 of handler.go)**

<br>

---

<br>

<a name="dir"/></a> 
### :grey_exclamation: Directory and File Description

| Directory Name  | File name        | Description                                                                                                           |
|-----------------|------------------|-----------------------------------------------------------------------------------------------------------------------|
| /               | main.go          | Implementation code of the main routine                                                                               |
| /               | controller.go    | Implementation code containing a queue implementation for buffering pods that are created and waiting to be scheduled |
| /               | handler.go       | Contains the implementation coordinator logic and how it select which scheduler to send the pod to                    |
| /helper         | helper.go        | Contain helper methods use by the main routine                                                                        |
| /docker         | Dockerfile       | Used by docker to create a docker image                                                                               |
| /yaml           | coordinator.yaml | Deployment file to deploy the scheduler in a Kubernetes cluster                                                       |

<br>

---

<br>

<a name="qna"/></a> 
### :grey_exclamation: Common questions

<dl>
  <dt>How to change the SchedulerName used by Epsilon to something else?</dt>
  <dd>The SchedulerName can be changed by changing the name to check in the if statement in main.go at line 193.</dd>

</dl>

<br>

---
