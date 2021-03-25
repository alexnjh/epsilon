![title](https://alexneo.net/epsilon/communication.png "comms")
## Communication Library

---

## :page_facing_up: Contents
- [Contents](#contents)
  - [Description](#desc)
  - [Deployment](#deploy)
  - [How does the retry microservice work?](#work)
  - [Directory and File description](#dir)
  - [Common questions](#qna)


<br>

<a name="desc"/></a> 
### :grey_exclamation: Description

Common communciation libarary used by all microservices in Epsilon. This libary contains message formats used by the different microservices and the communication handler implementaion for a microservice to communicate with the queue service. All communications in epsilon should be via the queue service.

<br>

---


<br>

<a name="deploy"/></a> 
### :grey_exclamation: Using the communication libary

The libarary can be imported by adding the link in the import section of the code

    import(
      communication "github.com/alexnjh/epsilon/communication"
    )

To connect to the queue service (The queue microservice should be running and accessible)

    import(
      communication "github.com/alexnjh/epsilon/communication"
    )

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
### :grey_exclamation: Directory and File Description

| Directory Name  | File name  | Description                                                     |
|-----------------|------------|-----------------------------------------------------------------|
| /               | main.go    | Implementation code of the main routine                         |
| /helper         | helper.go  | Contain helper methods use by the main routine                  |
| /docker         | Dockerfile | Used by docker to create a docker image                         |
| /yaml           | retry.yaml | Deployment file to deploy the scheduler in a Kubernetes cluster |

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
