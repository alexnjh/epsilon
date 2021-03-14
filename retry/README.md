![title](https://alexneo.net/epsilon/retry.png "Retry")
## Retry Microservice

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

The Retry service's goal is to reschedule pods that failed

<br>

---


<br>

<a name="deploy"/></a> 
### :grey_exclamation: Deployment of the retry service

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
      value: "epsilon.backoff"

<br>

---

<br>

<a name="work"/></a> 
### :grey_exclamation: Coordinator algorithm

**[STEP 1]**
<br>
The retry monitors the queue for new pods that failed.


<br>

---

<br>

<a name="dir"/></a> 
### :grey_exclamation: Directory Description

<dl>
  <dt>docker</dt>
  <dd>contain the dockerfile for generating the retry service docker image</dd>
  
  <dt>yaml</dt>
  <dd>contain the deployment yaml file</dd>
  
</dl>

<br>

---

<br>

<a name="file"/></a> 
### :grey_exclamation: File Description

<dl>
  <dt>main.go</dt>
  <dd>contain the main routine. All initialization of required variables including the waiting for new pods that failed is in this file</dd>

</dl>

<br>

---

<br>

<a name="qna"/></a> 
### :grey_exclamation: Common questions

<dl>
  <dt>test?</dt>
  <dd></dd>

</dl>

<br>

---
