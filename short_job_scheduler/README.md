![title](https://alexneo.net/epsilon/sjsched.png "SJSched")
## Short Job Scheduler Microservice

---

## :page_facing_up: Contents
- [Contents](#contents)
  - [Description](#desc)
  - [Deployment](#deploy)
  - [How does the Short Job Scheduler work?](#algo)
  - [Directory and File description](#dir)
  - [Common questions](#qna)


<br>

<a name="desc"/></a> 
### :grey_exclamation: Description

The short job scheduler is a lightweight scheduler designed to schedule pods very quickly and is used for demonstration purposes and also a template for creating different scheduler microservices. 

<br>

---


<br>

<a name="deploy"/></a> 
### :grey_exclamation: Deployment of the Short Job Scheduler microservice

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
        value: "epsilon.shortjob"
      - name: RETRY_QUEUE
        value: "epsilon.backoff"
      - name: HOSTNAME
        valueFrom:
          fieldRef:
            fieldPath: metadata.name

<br>
<b>RECEIVE_QUEUE<b> indicates the queue the scheduler is going to be listening to for new pods send by the coordinator service.
<br>
<b>RETRY_QUEUE<b> indicates the queue the scheduler is going to send failed pods to.

---

<br>

<a name="work"/></a> 
### :grey_exclamation: How does the scheduler operates?

![schedLifecycle](https://alexneo.net/epsilon/sj.png "scedLifecycle")

---

<br>

<a name="dir"/></a> 
### :grey_exclamation: Directory and File Description

| Directory Name  | File name      | Description                                                     |
|-----------------|----------------|-----------------------------------------------------------------|
| /               | main.go        | Implementation code of the main routine                         |
| /               | helper.go      | Contain helper methods use by the main routine                  |
| /               | scheduler.go   | Contains the implementation of the scheduling logic             |
| /yaml           | scheduler.yaml | Deployment file to deploy the scheduler in a Kubernetes cluster |
| /docker         | Dockerfile     | Used by docker to create a docker image                         |
<br>

---

<br>

<a name="qna"/></a> 
### :grey_exclamation: Common questions

<dl>
  <dt></dt>

</dl>

<br>

---
