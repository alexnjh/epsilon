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

package communication

import (
 "errors"
 "github.com/streadway/amqp"
)

type RabbitMQCommunication struct{
  host string
  conn *amqp.Connection
  ch *amqp.Channel
}

func NewRabbitMQCommunication(host string) (RabbitMQCommunication, error){

 var comms = RabbitMQCommunication{host,nil,nil}

 err := comms.Connect()
 if(err!=nil){
   return comms,err
 }

 return comms,nil
}

func (c *RabbitMQCommunication) QueueDeclare(queue string) error{

 _, err := c.ch.QueueDeclare(
   queue, // name
   true,   // durable
   false,   // delete when unused
   false,   // exclusive
   false,   // no-wait
   nil,     // arguments
 )

 if err != nil{
   return errors.New(
     `Failed to declare queue maybe connection is down?
     Consider running Connect() again to reconnect to queue service`)
 }

 return nil
}

// Send schedule request to the schedulers
func (c *RabbitMQCommunication) Send(message []byte, queue string) error{

 err := c.ch.Publish(
  "",     // exchange
  queue, // routing key
  false,  // mandatory
  false,  // immediate
  amqp.Publishing {
    ContentType: "text/json",
    Body:        message,
  })

 if err != nil{
   return errors.New(
     `Failed to send message maybe connection is down?
     Consider running Connect() again to reconnect to queue service`)
 }

 return nil
}

// Send schedule request to the schedulers
func (c *RabbitMQCommunication) Receive(queue string) (<-chan amqp.Delivery, error){

  msgs, err := c.ch.Consume(
   queue,   // queue
   "",      // consumer
   false,   // auto-ack
   false,   // exclusive
   false,   // no-local
   false,   // no-wait
   nil,     // args
 )

if err != nil{

 return nil,errors.New(
   `Failed to create channel to receive message maybe connection is down?
   Consider running Connect() again to reconnect to queue service`)

}
return msgs,nil
}

func (c *RabbitMQCommunication) Connect() error{

 if c.conn != nil{
   if c.conn.IsClosed() {
     conn, err := amqp.Dial(c.host)

     if err != nil {
       return err
     }

     ch, err := conn.Channel()

     if err != nil {
       return err
     }

     c.conn = conn
     c.ch = ch

     return nil
   }else{

     ch, err := c.conn.Channel()

     if err != nil {
       return err
     }

     c.ch = ch

     return nil
   }
 }else{
   conn, err := amqp.Dial(c.host)

   if err != nil {
     return err
   }

   ch, err := conn.Channel()

   if err != nil {
     return err
   }

   c.conn = conn
   c.ch = ch

   return nil
 }


}
