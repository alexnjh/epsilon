![title](https://alexneo.net/epsilon/autoscaler.png "Autoscaler")
## Autoscaler Microservice

---

## :page_facing_up: Contents
- [Contents](#contents)
  - [Description](#desc)
  - [Deployment](#deploy)
  - [How does the autoscaler work?](#algo)
  - [Directory and File description](#dir)
  - [Common questions](#qna)


<br>

<a name="desc"/></a> 
### :grey_exclamation: Description

The autoscaler's goal is to scale scheduler replicas depending on cluster load to ensure performance is up to user requirement.

<br>

---


<br>

<a name="deploy"/></a> 
### :grey_exclamation: Deployment of the autoscaler service

Before deploying the coordinator.yaml file, please configure the environment variables to the correct values used by the queue microservice.

      env:
      - name: PC_METRIC_URL
        value: "pod-coordinator.custom-scheduler.svc.cluster.local:8080/metrics"
      - name: MQ_HOST
        value: "sched-rabbitmq-0.sched-rabbitmq.custom-scheduler.svc.cluster.local"
      - name: MQ_MANAGE_PORT
        value: "15672"
      - name: MQ_USER
        value: "guest"
      - name: MQ_PASS
        value: "guest"
      - name: INTERVAL
        value: "300"
      - name: DEFAULT_QUEUE
        value: "epsilon.distributed"
      - name: POD_NAMESPACE

<br>
The **DEFAULT_QUEUE** is the queue used by the general-purpose scheduler. In Epsilon, atleast one scheduler service need to act as the default scheduler.
<br>
The **PC_METRIC_URL** is the hostname of the coodinator service (This can be ignored if the QueueTheory plugin is not enabled)

---

<br>

<a name="work"/></a> 
### :grey_exclamation: How does the autoscaler operates?

**[STEP 1]**
<br>
The autoscaler will first get cluster metrics that the plugins require based on a specified time interval.

**[STEP 2]**
<br>
After getting the metrics the autoscaler will proceed to send the information to the different plugins and wait for their reply.

**[STEP 3]**
<br>
Once all the plugin's replies are consolidated the autoscaler will make a decision based on majority vote. 

**[STEP 4]**
<br>
The autoscaler will not attempt to scale up or down the scheduler replicas if there is a tie and will try again on the next time interval.
<br>

---

<br>

<a name="dir"/></a> 
### :grey_exclamation: Directory and File Description

| Directory Name             | File name       | Description                                                       |
|----------------------------|-----------------|-------------------------------------------------------------------|
| /                          | main.go         | Implementation code of the main routine                           |
| /                          | helper.go       | Contain helper methods use by the main routine                    |
| /interfaces                | interface.go    | Contains the auto scaler plugin interface definition              |
| /plugins/linear_regression | plugin.go       | Contains the linear regression plugin implementation              |
| /plugins/queue_theory      | plugin.go       | Contains the queue theory plugin implementation                   |
| /plugins/rabbitmq          | plugin.go       | Contains the rabbitmq plugin implementation                       |
| /plugins/scheduler_prob    | plugin.go       | Contains the scheduler conflict probability plugin implementation |
| /yaml                      | autoscaler.yaml | Deployment file to deploy the scheduler in a Kubernetes cluster   |
| /docker                    | Dockerfile      | Used by docker to create a docker image                           |

<br>

---

<br>

<a name="qna"/></a> 
### :grey_exclamation: Common questions

<dl>
  <dt>How to add a new autoscaler plugin</dt>
  
      1. Create a new folder in /plugins
      2. Write the plugin implementation and store the file in the new folder created in 1
      3. Open main.go and intialize the plugin. Intialization of plugins can be found at line 168. 

</dl>

<br>

---
