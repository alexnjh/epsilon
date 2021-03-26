package queue_theory

import(
  "strconv"
  "time"
  "net/http"
  "bufio"
  "strings"
  "fmt"
  "os"
  "github.com/alexnjh/epsilon/autoscaler/interfaces"
  log "github.com/sirupsen/logrus"
)

// QueueTheoryPlugin decides based on approximation of the waiting time for all the pods current in the cluster wiating to be scheduled.
type QueueTheoryPlugin struct{
  Name string
  Threshold float64
  targetURL string
}

// Creates a new QueueTheoryPlugin
func NewQueueTheoryPlugin(name string,threshold float64,targetURL string) *QueueTheoryPlugin{
  return &QueueTheoryPlugin{
    Name: name,
    Threshold: threshold,
    targetURL: targetURL,
  }
}

// Compute processes the data and return a ComputeResult
func (plugin *QueueTheoryPlugin) Compute(_, _, noOfSched float64) interfaces.ComputeResult{

  metricMap := promToMap(plugin.targetURL)

  arrivalRate, err := strconv.ParseFloat(metricMap["pod_request_total_in_1min"], 64)
  if err != nil {
		log.Fatalf(err.Error())
	}

  serviceRate := noOfSched*(float64((1*time.Minute)/(25*time.Millisecond)))

  avgWaitingTime := arrivalRate/(serviceRate*(serviceRate-arrivalRate))

  if (avgWaitingTime < plugin.Threshold){
    return interfaces.DoNotScale
  }else{
    return interfaces.ScaleUp
  }

}

// Convert prometheus formatted metrics into a map
func promToMap(url string) map[string]string{

  metricMap := make(map[string]string)

  resp, err := http.Get(url)
  if err != nil {
    log.Fatalf(err.Error())
  }

  scanner := bufio.NewScanner(resp.Body)
  for scanner.Scan() {
    if len(scanner.Text()) > 0{
      if scanner.Text()[0] != '#'{
        s := strings.Split(scanner.Text(), " ")
        if (len(s) == 2){
          metricMap[s[0]]=s[1]
        }
      }
    }
  }
  if err := scanner.Err(); err != nil {
    fmt.Fprintln(os.Stderr, "reading standard input:", err)
  }

  return metricMap
}
