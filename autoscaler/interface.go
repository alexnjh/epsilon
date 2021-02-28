package main

import (

)

// ComputeResult indicate the different decision decided by the autoscaler plugins.
type ComputeResult string

const (
  // Scale up the number of scheduler replicas by 1
	ScaleUp ComputeResult = "ScaleUp"
  // Do not scale up/down scheduler replicas at all
	DoNotScale ComputeResult = "DoNotScale"
  // Scale down the number of scheduler replicas by 1
	ScaleDown ComputeResult = "ScaleDown"
)

// Autoscaler plugin used by the autoscaler to decide if scaling is necessary
type AutoScalerPlugin interface{
  Compute(float64,float64,float64) ComputeResult
}
