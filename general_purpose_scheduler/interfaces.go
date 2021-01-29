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

package general_purpose_scheduler

import (
 "github.com/streadway/amqp"
)

// Communication interface contains the methods that are required
type Communication interface {
 Send(message []byte, queue string) error
 Receive(queue string) (<-chan amqp.Delivery, error)
 Connect() error
}
