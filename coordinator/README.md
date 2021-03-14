![title](https://alexneo.net/epsilon/coordinator.png "Coordinator")
## Coordinator Microservice

---

## :page_facing_up: Contents
- [Contents](#contents)
  - [Description](#desc)
  - [Deployment](#deploy)
  - [Coordinator Scheduling Algorithm](#algo)


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

The **DEFAULT_QUEUE** is the queue used by the general-purpose scheduler. In Epsilon, atleast one scheduler service needs to act as the default scheduler.

