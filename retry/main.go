/*

Copyright (C) 2020 Alex Neo

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package retry

import (
  "os"
  "fmt"
  "time"
  "math/rand"
  "github.com/streadway/amqp"
  log "github.com/sirupsen/logrus"
  jsoniter "github.com/json-iterator/go"
  configparser "github.com/bigkevmcd/go-configparser"
  communication "github.com/alexnjh/epsilon/communication"
)

const (
    DefaultConfigPath = "/go/src/app/config.cfg"
)


// Initialize json encoder
var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {

  // Get required values
  confDir := os.Getenv("CONFIG_DIR")

  var config *configparser.ConfigParser
  var err error

  if len(confDir) != 0 {
    config, err = getConfig(confDir)
  }else{
    config, err = getConfig(DefaultConfigPath)
  }

  var mqHost, mqPort, mqUser, mqPass, receiveQueue string

  if err != nil {

    log.Errorf(err.Error())

    mqHost = os.Getenv("MQ_HOST")
    mqPort = os.Getenv("MQ_PORT")
    mqUser = os.Getenv("MQ_USER")
    mqPass = os.Getenv("MQ_PASS")
    receiveQueue = os.Getenv("RECEIVE_QUEUE")

    if len(mqHost) == 0 ||
    len(mqPort) == 0 ||
    len(mqUser) == 0 ||
    len(mqPass) == 0 ||
    len(receiveQueue) == 0{
  	   log.Fatalf("Config not found, Environment variables missing")
    }


  }else{

    mqHost, err = config.Get("QueueService", "hostname")
    if err != nil {
      log.Fatalf(err.Error())
    }
    mqPort, err = config.Get("QueueService", "port")
    if err != nil {
      log.Fatalf(err.Error())
    }
    mqUser, err = config.Get("QueueService", "user")
    if err != nil {
      log.Fatalf(err.Error())
    }
    mqPass, err = config.Get("QueueService", "pass")
    if err != nil {
      log.Fatalf(err.Error())
    }
    receiveQueue, err = config.Get("DEFAULTS", "receive_queue")
    if err != nil {
      log.Fatalf(err.Error())
    }
  }

  // Attempt to connect to the rabbitMQ server
  comm, err := communication.NewRabbitMQCommunication(fmt.Sprintf("amqp://%s:%s@%s:%s/",mqUser, mqPass, mqHost, mqPort))
  if err != nil {
    log.Fatalf(err.Error())
  }

  err = comm.QueueDeclare(receiveQueue)
  if err != nil {
    log.Fatalf(err.Error())
  }

  msgs, err := comm.Receive(receiveQueue)

  // Use a channel if goroutine closes
  retryCh := make(chan bool)
  defer close(retryCh)

  go RetryProcess(&comm, msgs, retryCh)

  log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

  // Check for connection failures and reconnect
  for {

    if status := <-retryCh; status == true {
      log.Errorf("Disconnected from message server and attempting to reconnect")
      for{
        err = comm.Connect()
        if err != nil{
          log.Errorf(err.Error())
        }else{
          err = comm.QueueDeclare(receiveQueue)
          if err != nil {
            log.Errorf(err.Error())
          }else{
            msgs, err = comm.Receive(receiveQueue)
            if(err != nil){
              log.Errorf(err.Error())
            }else{
              // Start go routine to start consuming messages
              go RetryProcess(&comm, msgs, retryCh)
              log.Infof("Reconnected to message server")
              break
            }
          }
        }
        // Sleep for a random time before trying again
        time.Sleep(time.Duration(rand.Intn(10))*time.Second)
      }
    }
  }


}

func RetryProcess(comm communication.Communication, msgs <-chan amqp.Delivery, closed chan<- bool){

  // Loop through all the messages in the queue
  for d := range msgs {
    // Convert json message to schedule request object
    var req RetryRequest
    if err := json.Unmarshal(d.Body, &req); err != nil {
        panic(err)
    }

    go WaitAndSend(comm, req,time.Duration(req.Req.LastBackOffTime)*time.Second)
    d.Ack(true)
  }

  closed <- true

}

func WaitAndSend(comm communication.Communication, obj RetryRequest, duration time.Duration){

  time.Sleep(duration)

  respBytes, err := json.Marshal(obj.Req)
  if err != nil {
    log.Errorf(err.Error())
  }

  err = comm.Send(respBytes,obj.Queue)

  if err != nil{
    log.Errorf(err.Error())
  }

}
