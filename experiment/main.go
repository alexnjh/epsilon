package experiment

import (
  "os"
  "fmt"
  "time"
  "math/rand"
  "net/http"
  "database/sql"
  "github.com/streadway/amqp"
  log "github.com/sirupsen/logrus"
  jsoniter "github.com/json-iterator/go"
  configparser "github.com/bigkevmcd/go-configparser"
	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

const (
    DefaultConfigPath = "/go/src/app/config.cfg"
    CoordinatorTableEntry = `CREATE TABLE coordinator(
        "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    		"received_at" TEXT,
    		"processed_at" TEXT,
        "elapsed_time" TEXT,
    		"pod_name" TEXT
    	  );`
    SchedulerTableEntry = `CREATE TABLE scheduler(
        "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    		"received_at" TEXT,
    		"processed_at" TEXT,
        "elapsed_time" TEXT,
    		"pod_name" TEXT,
        "node_name" TEXT,
        "host_name" TEXT
    	  );`
    JobTableEntry = `CREATE TABLE job(
        "id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    		"received_at" TEXT,
    		"processed_at" TEXT,
        "elapsed_time" TEXT,
    		"pod_name" TEXT
    	  );`
)


// Initialize json encoder
var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main(){


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

  if !DBFileExists("sqlite-database.db") {


    //SQLite database init
    log.Println("Creating sqlite-database.db...")
  	file, err := os.Create("sqlite-database.db") // Create SQLite file
  	if err != nil {
  		log.Fatal(err.Error())
  	}
  	file.Close()
  	log.Println("sqlite-database.db created")

  }

  sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File
  defer sqliteDatabase.Close() // Defer Closing the database


  dbHandler := &DatabaseHandler{sqliteDatabase}




  if(!dbHandler.CheckIfTableExists("coordinator")){
    dbHandler.CreateTable(CoordinatorTableEntry)
  }

  if(!dbHandler.CheckIfTableExists("scheduler")){
    dbHandler.CreateTable(SchedulerTableEntry)
  }

  if(!dbHandler.CheckIfTableExists("job")){
    dbHandler.CreateTable(JobTableEntry)
  }

  // Attempt to connect to the rabbitMQ server
  comm, err := NewRabbitMQCommunication(fmt.Sprintf("amqp://%s:%s@%s:%s/",mqUser, mqPass, mqHost, mqPort))
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

  go ExperimentProcess(dbHandler, &comm, msgs, retryCh)
  go WebProcess(dbHandler)

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
              go ExperimentProcess(dbHandler, &comm, msgs, retryCh)
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

func ExperimentProcess(handler *DatabaseHandler, comm Communication, msgs <-chan amqp.Delivery, closed chan<- bool){

  // Loop through all the messages in the queue
  for d := range msgs {

    log.Infof("Message received")

    // Convert json message to schedule request object
    var payload ExperimentPayload
    if err := json.Unmarshal(d.Body, &payload); err != nil {
        panic(err)
    }

    if payload.Type == "Coordinator"{
      handler.InsertEntry("coordinator",payload.Hostname,payload.InTime,payload.OutTime,payload.Pod)
    }else if payload.Type == "Scheduler"{
      handler.InsertEntry("scheduler",payload.Hostname,payload.InTime,payload.OutTime,payload.Pod)
    }else if payload.Type == "Completed_Jobs"{
      handler.InsertEntry("job",payload.Hostname,payload.InTime,payload.OutTime,payload.Pod)
    }

    d.Ack(true)
  }

  closed <- true

}

func WebProcess(handler *DatabaseHandler){
    http.HandleFunc("/coordinator", handler.CoordinatorHandler)
    http.HandleFunc("/scheduler", handler.SchedulerHandler)
    http.HandleFunc("/job", handler.JobHandler)
    http.ListenAndServe(":8080", nil)
}
