/*
Copyright 2017 The Kubernetes Authors.
Modification copyright (C) 2020 Alex Neo

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

package coordinator

import (
	"fmt"
	"time"

  log "github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
  listers "k8s.io/client-go/listers/core/v1"
  communication "github.com/alexnjh/epsilon/communication"
)

// PodController struct defines how a PodController should encapsulate
// logging, client connectivity, informing (list and watching)
// queueing, and handling of resource changes
type PodController struct {
	clientset kubernetes.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
  lister  listers.PodLister
	handler   Handler
}

// Run is the main path of execution for the PodController loop
func (c *PodController) Run(stopCh <-chan struct{}) {
	// handle a panic with logging and exiting
	defer utilruntime.HandleCrash()
	// ignore new items in the queue but when all goroutines
	// have completed existing items then shutdown
	defer c.queue.ShutDown()

	log.Info("PodController.Run: initiating")

	// do the initial synchronization (one time) to populate resources
	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Error syncing cache"))
		return
	}
	log.Info("PodController.Run: cache sync complete")

	// run the runWorker method every second with a stop channel
	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced allows us to satisfy the PodController interface
// by wiring up the informer's HasSynced method to it
func (c *PodController) HasSynced() bool {
	return (c.informer.HasSynced())
}

// runWorker executes the loop to process new items added to the queue
func (c *PodController) runWorker() {
	log.Info("PodController.runWorker: starting")

	// invoke processNextItem to fetch and consume the next change
	// to a watched or listed resource
	for c.processNextItem() {
		log.Info("PodController.runWorker: processing next item")
	}

	log.Info("PodController.runWorker: completed")
}

// processNextItem retrieves each queued item and takes the
// necessary handler action based off of if the item was
// created or deleted
func (c *PodController) processNextItem() bool {
	log.Info("PodController.processNextItem: start")

	// fetch the next item (blocking) from the queue to process or
	// if a shutdown is requested then return out of this to stop
	// processing
	key, quit := c.queue.Get()

	// stop the worker loop from running as this indicates we
	// have sent a shutdown message that the queue has indicated
	// from the Get method
	if quit {
		return false
	}

	defer c.queue.Done(key)

	// assert the string out of the key (format `namespace/name`)
	keyRaw := key.(string)

	log.Infof("PodController.processNextItem: object to process: %s", keyRaw)

	var err = c.handler.ObjectSync(keyRaw)
  if err != nil {
  	if c.queue.NumRequeues(key) < 5 {
  		log.Errorf("PodController.processNextItem: Failed processing item with key %s with error %v, retrying", key, err)
  		c.queue.AddRateLimited(key)
    }
  }else{
    c.queue.Forget(key)
  }
	// keep the worker loop running by returning true
	return true
}
