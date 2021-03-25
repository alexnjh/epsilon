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
<br>
<dl>
  <dt>1. The libarary can be imported by adding the link in the import section of the code</dt>
</dd>

    import(
      communication "github.com/alexnjh/epsilon/communication"
    )


<dl>
  <dt>2. The library can be imported by adding the link in the import section of the code.</dt>
  <br>
  <dd>To connect to the queue service (The queue microservice should be running and accessible).<dd>
  <dd><b>mqUser</b> is the username of the user that the microservice will be using to authenticate with the queue microservice.<dd>
  <dd><b>mqPass</b> is the password of the user that the microservice will be using to authenticate with the queue microservice.<dd>
  <dd><b>mqHost</b> is the hostname of the queue microservice.<dd>
  <dd><b>mqPort</b> is the port that the queue microservice is listening on.<dd>
</dd>

    comm, err := communication.NewCommunicationClient(fmt.Sprintf("amqp://%s:%s@%s:%s/",mqUser, mqPass, mqHost, mqPort))
    if err != nil {
      log.Fatalf(err.Error())
    }

<dl>
  <dt>3. Creating a queue</dt>
  <br>
  <dd><b>queueName</b> refers the the name of the queue<dd>
</dd>
<br>

    err = comm.QueueDeclare(queueName)
    if err != nil {
      log.Fatalf(err.Error())
    }


<dl>
  <dt>4. Sending a message to a queue</dt>
  <br>
  <dd><b>bytes</b> refers to the message in byte array format to send into the queue<dd>
  <dd><b>queueName</b> the name of the queue to send the message<dd>
</dd>
<br>

    err = t.comm.Send(bytes,queueName)

    if err != nil{
      return false
    }
    
 <dl>
  <dt>5. Receiving a message from a queue</dt>
  <br>
  <dd><b>msgs</b> refers to a channel that messages will be piped through<dd>
  <dd><b>receiveQueue</b> the name of the queue to receive messages from<dd>
</dd>
<br>

    msgs, err := comm.Receive(receiveQueue)
    for d := range msgs {
      // Do something to the message 
    }
---

<br>

<a name="dir"/></a> 
### :grey_exclamation: Directory and File Description

| Directory Name  | File name         | Description                                                         |
|-----------------|-------------------|---------------------------------------------------------------------|
| /               | interfaces.go     | Communication interface declaration used between all microservices  |
| /               | types.go          | Contain message types used by the various microservices             |
| /               | communications.go | Implementation of the communication interface                       |

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
