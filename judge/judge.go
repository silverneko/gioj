package main

import (
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "log"
  "time"
  "os"
  "golang.org/x/net/websocket"
  "encoding/json"
  "strconv"
  "github.com/silverneko/gioj/models"
)

func main() {
  log.Println("Start judge service...")

  os.RemoveAll("/tmp/gioj-judge")
  os.Mkdir("/tmp/gioj-judge", 0755)

  bufferSize := 2
  workQueue := make(chan *models.Submission, bufferSize)

  /* spawn two workers */
  go worker(workQueue, 1)
  go worker(workQueue, 2)

  for {
    db := models.DBSession{DB.Copy()}
    var submission models.Submission
    _, err := db.C("submissions").
      Find(bson.M{"verdict.result": models.QUEUED}).
      Sort("_id").
      Apply(mgo.Change{Update: bson.M{"$set": bson.M{"verdict.result": models.JUDGING}}}, &submission)
    if err == nil {
      log.Println("Recieved: ", submission.ID.Hex())
      workQueue <- &submission
    } else {
      log.Println("Nothing to judge")
      time.Sleep(5 * time.Second)
    }
    db.Close()
  }
}

func worker(c <-chan *models.Submission, workerId int) {
  for {
    submission := <-c
    log.Println("Worker: ", workerId)
    judge(submission)
  }
}

func judge(submission *models.Submission) {
  path := "/tmp/gioj-judge/" + submission.ID.Hex()
  os.Mkdir(path, 0755)
  filename := path + "/test.cpp"
  f, err := os.Create(filename)
  if err != nil {
    log.Println("File create: ", err)
    return
  }
  defer os.Remove(filename)
  f.WriteString(submission.Content)
  f.Close()

  db := models.DBSession{DB.Copy()}
  defer db.Close()

  var problem models.Problem
  err = db.C("problems").Find(bson.M{"_id": submission.Pid}).One(&problem)
  if err != nil {
    log.Println("Problem retrieve: ", err)
    return
  }

  cf, _ := websocket.NewConfig("ws://localhost:2501/judge", "http://localhost:5050")
  ws, err := websocket.DialConfig(cf)
  if err != nil {
    log.Println("Ws create: ", err)
    return
  }
  defer ws.Close()

  dataset := make([]int, problem.TestdataCount)
  for i, _ := range dataset {
    dataset[i] = i+1
  }
  msg, _ := json.Marshal(map[string]interface{}{
    "chal_id": 1,  // What number is this doesn't really matter
    "code_path": filename,
    "res_path": "/home/silvernegi/Projects/gioj/td/" + strconv.Itoa(submission.Pid),
    "comp_type": "g++",
    "check_type": "diff",
    "metadata": "",
    "test": []map[string]interface{}{
      {
	"test_idx": 1,
	"timelimit": problem.Timelimit,
	"memlimit": problem.Memlimit * 2 << 10,
	"metadata": map[string]interface{}{
	  "data": dataset,
	},
      },
    },
  })
  ws.Write(msg)

  rcv := make([]byte, 2 << 20)
  n, err := ws.Read(rcv)
  rcv = rcv[:n]

  var response Response
  err = json.Unmarshal(rcv, &response)
  log.Println(err)
  log.Println(string(rcv))

  var result models.Verdict
  var status int = STATUS_NONE
  for _, res := range response.Result {
    if res.Peakmem > result.Memused {
      result.Memused = res.Peakmem
    }
    if res.State > status {
      status = res.State
    }
    result.Timeused += res.Runtime
  }
  switch status {
    case STATUS_NONE, STATUS_ERR:
      result.Result = models.ERR
    case STATUS_AC:
      result.Result = models.AC
    case STATUS_WA:
      result.Result = models.WA
    case STATUS_RE:
      result.Result = models.RE
    case STATUS_TLE:
      result.Result = models.TLE
    case STATUS_MLE:
      result.Result = models.MLE
    case STATUS_CE:
      result.Result = models.CE
  }

  err = db.C("submissions").Update(bson.M{"_id": submission.ID}, bson.M{"$set": bson.M{"verdict": result}})
  if err != nil {
    log.Println("Judge Error: ", err)
    return
  }
}

type Response struct {
  Verdict string
  Result []struct{
    Test_idx int
    State int
    Runtime int
    Peakmem int
    Verdict string
  }
}

var DB *mgo.Session

func init() {
  var err error
  DB, err = mgo.DialWithInfo(&mgo.DialInfo{
    Addrs: []string{"localhost"},
    Database: "gioj_test",
    Username: "gioj",
    Password: "gioj",
  })
  if err != nil {
    log.Fatal("mgo.Dial: ", err)
  }

}

const (
  STATUS_NONE int = iota
  STATUS_AC
  STATUS_WA
  STATUS_RE
  STATUS_TLE
  STATUS_MLE
  STATUS_CE
  STATUS_ERR
)

