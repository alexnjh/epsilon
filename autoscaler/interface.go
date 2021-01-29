package main

import (

)

type ComputeResult string

const (
	ScaleUp ComputeResult = "ScaleUp"
	DoNotScale ComputeResult = "DoNotScale"
	ScaleDown ComputeResult = "ScaleDown"
)

// Autoscaler plugin used by the autoscaler to decide if scaling is necessary
type AutoScalerPlugin interface{
  Compute(float64,float64,float64) ComputeResult
}
