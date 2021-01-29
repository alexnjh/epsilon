package main

import(
  "fmt"
  "time"
  "net/http"
  "database/sql"
  corev1 "k8s.io/api/core/v1"
  log "github.com/sirupsen/logrus"
  _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

type DatabaseHandler struct{
  db *sql.DB
}

const (
  RFC3339Micro = "2006-01-02T15:04:05.999999Z07:00"
)

func (dh *DatabaseHandler) CoordinatorHandler(w http.ResponseWriter, r *http.Request) {

  var response = "ID,RECEIVED_AT,PROCESSED_AT,ELAPSED_TIME,POD_NAME\n"

  row, err := dh.db.Query("SELECT * FROM coordinator ORDER BY id")
	if err != nil {
		log.Errorf(err.Error())
	}else{
    log.Infof("Checking rows of data")
    for row.Next() { // Iterate and fetch the records from result cursor
      var id int
      var received_at string
      var processed_at string
      var elapsed_time string
      var pod_name string
      row.Scan(&id, &received_at, &processed_at, &elapsed_time, &pod_name)
      response = fmt.Sprintf("%s%d,%s,%s,%s,%s\n",response,id,received_at,processed_at,elapsed_time,pod_name)
    }
    defer row.Close()
  }

  fmt.Fprintf(w, "%s", response)
}

func (dh *DatabaseHandler) JobHandler(w http.ResponseWriter, r *http.Request) {

  var response = "ID,RECEIVED_AT,PROCESSED_AT,ELAPSED_TIME,POD_NAME\n"

  row, err := dh.db.Query("SELECT * FROM job ORDER BY id")
	if err != nil {
		log.Errorf(err.Error())
	}else{
    log.Infof("Checking rows of data")
    for row.Next() { // Iterate and fetch the records from result cursor
      var id int
      var received_at string
      var processed_at string
      var elapsed_time string
      var pod_name string
      row.Scan(&id, &received_at, &processed_at, &elapsed_time, &pod_name)
      response = fmt.Sprintf("%s%d,%s,%s,%s,%s\n",response,id,received_at,processed_at,elapsed_time,pod_name)
    }
    defer row.Close()
  }

  fmt.Fprintf(w, "%s", response)
}

func (dh *DatabaseHandler) SchedulerHandler(w http.ResponseWriter, r *http.Request) {

  var response = "ID,RECEIVED_AT,PROCESSED_AT,ELAPSED_TIME,POD_NAME,NODE_NAME,HOST_NAME\n"

  row, err := dh.db.Query("SELECT * FROM scheduler ORDER BY id")
	if err != nil {
		log.Errorf(err.Error())
	}else{
    log.Infof("Checking rows of data")
    for row.Next() { // Iterate and fetch the records from result cursor
      var id int
      var received_at string
      var processed_at string
      var elapsed_time string
      var pod_name string
      var node_name string
      var host_name string
      row.Scan(&id, &received_at, &processed_at, &elapsed_time, &pod_name, &node_name, &host_name)
      response = fmt.Sprintf("%s%d,%s,%s,%s,%s,%s,%s\n",response,id,received_at,processed_at,elapsed_time,pod_name,node_name,host_name)
    }
    defer row.Close()
  }

  fmt.Fprintf(w, "%s", response)
}

func (dh *DatabaseHandler) CheckIfTableExists(name string) bool{
  row, err := dh.db.Query(fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s'",name))
  if err != nil {
		log.Errorf(err.Error())
    return false
	}
  defer row.Close()

  if row.Next(){
      return true
  }else{
      return false
  }

}

func (dh *DatabaseHandler) CreateTable(sql string){
	log.Println("Creating table...")
	statement, err := dh.db.Prepare(sql) // Prepare SQL Statement
	if err != nil {
		log.Fatalf(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("Table created")
}

func (dh *DatabaseHandler) InsertEntry(name string, hostname string, received_at time.Time, processed_at time.Time, pod *corev1.Pod) {
	log.Println("Inserting entry ...")

  elapsedTime := processed_at.Sub(received_at).Microseconds()

  //init the loc
  loc, _ := time.LoadLocation("Asia/Singapore")

  if name == "coordinator"{

    // creationTime := pod.CreationTimestamp.In(loc).Format(RFC3339Micro)

    insertSQL := fmt.Sprintf("INSERT INTO %s(received_at, processed_at, elapsed_time, pod_name) VALUES (?, ?, ?, ?)",name)
    statement, err := dh.db.Prepare(insertSQL) // Prepare statement.
                                                     // This is good to avoid SQL injections
  	if err != nil {
  		log.Fatalf(err.Error())
  	}
  	_, err = statement.Exec(received_at.In(loc).Format(RFC3339Micro), processed_at.In(loc).Format(RFC3339Micro), elapsedTime, pod.Name)
  	if err != nil {
  		log.Fatalf(err.Error())
  	}
  }else if name == "scheduler"{
    insertSQL := fmt.Sprintf("INSERT INTO %s(received_at, processed_at, elapsed_time, pod_name, node_name, host_name) VALUES (?, ?, ?, ?, ?, ?)",name)
    statement, err := dh.db.Prepare(insertSQL) // Prepare statement.
                                                     // This is good to avoid SQL injections
  	if err != nil {
  		log.Fatalf(err.Error())
  	}

    var nodeName = "NA"

    if len(pod.Spec.NodeName) != 0 {
      nodeName = pod.Spec.NodeName
    }

  	_, err = statement.Exec(received_at.In(loc).Format(RFC3339Micro), processed_at.In(loc).Format(RFC3339Micro), elapsedTime, pod.Name, nodeName, hostname)
  	if err != nil {
  		log.Fatalf(err.Error())
  	}
  }else{
    insertSQL := fmt.Sprintf("INSERT INTO %s(received_at, processed_at, elapsed_time, pod_name) VALUES (?, ?, ?, ?)",name)
    statement, err := dh.db.Prepare(insertSQL) // Prepare statement.
                                                     // This is good to avoid SQL injections
  	if err != nil {
  		log.Fatalf(err.Error())
  	}
  	_, err = statement.Exec(received_at.In(loc).Format(RFC3339Micro), processed_at.In(loc).Format(RFC3339Micro),elapsedTime, pod.Name)
  	if err != nil {
  		log.Fatalf(err.Error())
  	}
  }

}
