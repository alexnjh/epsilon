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
### :grey_exclamation: Retry algorithm

**[STEP 1]**
<br>
The retry service monitors the queue for new pods that failed.
<br>

**[STEP 2]**
<br>
When a pod that failed is recevied, the retry service will generate a backoff timer and wait for the backoff timer to pass
<br>

**[STEP 3]**
<br>
Once the backoff duration had past, the retry service will send the failed pod back to its respective scheduling queue
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
