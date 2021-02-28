package main

import (
  "os"

  "k8s.io/client-go/rest"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/clientcmd"
  log "github.com/sirupsen/logrus"
  configparser "github.com/bigkevmcd/go-configparser"
)

// Get scheduler config file from config path
func getConfig(path string) (*configparser.ConfigParser, error){
  p, err := configparser.NewConfigParserFromFile(path)
  if err != nil {
    return nil,err
  }

  return p,nil
}

// Retrieve the Kubernetes cluster client either from outside the cluster or inside the cluster
func getKubernetesClient() (kubernetes.Interface){
	// construct the path to resolve to `~/.kube/config`
  config, err := rest.InClusterConfig()
  if err != nil {
    kubeConfigPath := os.Getenv("HOME") + "/.kube/config"

    //create the config from the path
    config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
    if err != nil {
      log.Fatalf("getInClusterConfig: %v", err)
      panic("Failed to load kube config")
    }
  }

  // generate the client based off of the config
  client, err := kubernetes.NewForConfig(config)
  if err != nil {
    panic("Failed to create kube client")
  }

	log.Info("Successfully constructed k8s client")
	return client
}
