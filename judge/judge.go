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
  "io/ioutil"
)

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

var judgePath string

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

func main() {
  pid := os.Getpid()
  pids := strconv.Itoa(pid)
  judgePath = "/tmp/gioj-judge-" + pids + "/"
  if err := os.MkdirAll(judgePath, 0755); err != nil {
    log.Println(err)
    return
  }
  defer os.RemoveAll(judgePath)

  /* work queue buffer */
  bufferSize := 4
  workQueue := make(chan *models.Submission, bufferSize)

  /* spawn two workers */
  workerCount := 2
  for i := 1; i <= workerCount; i++ {
    go worker(workQueue, i)
  }
  log.Println("Start judge service...")
  log.Println("Judge directory: ", judgePath)
  log.Println("Worker count: ", workerCount)

  for {
    var submission models.Submission
    db := models.DBSession{DB.Copy()}
    _, err := db.C("submissions").
      Find(bson.M{"verdict.result": models.QUEUED}).
      Sort("_id").
      Apply(mgo.Change{Update: bson.M{"$set": bson.M{"verdict.result": models.JUDGING}}}, &submission)
    db.Close()
    if err == nil {
      log.Println("Recieved: ", submission.ID.Hex())
      workQueue <- &submission
    } else {
      time.Sleep(5 * time.Second)
    }
  }
}

func worker(c <-chan *models.Submission, workerId int) {
  defer func() {
    if err := recover(); err != nil {
      log.Println("Worker panic: ", err)
      log.Println("Restart worker: ", workerId)
      go worker(c, workerId)
    }
  } ()
  for {
    submission := <-c
    if submission != nil {
      /* can't be too careful */
      log.Println("Worker: ", workerId, submission.ID.Hex())
      judge(submission)
    }
  }
}

func judge(submission *models.Submission) {
  db := models.DBSession{DB.Copy()}
  defer db.Close()
  // Error handling
  defer func() {
    if err := recover(); err != nil {
      log.Println("Judge routine panic: ", err)
      db.C("submissions").Update(bson.M{"_id": submission.ID}, bson.M{"$set": bson.M{"verdict.result": models.ERR}})
    }
  } ()

  var problem models.Problem
  if err := db.C("problems").Find(bson.M{"_id": submission.Pid}).One(&problem); err != nil {
    panic(err)
  }

  path := judgePath + submission.ID.Hex()
  if err := os.Mkdir(path, 0755); err != nil {
    panic(err)
  }
  defer os.RemoveAll(path)
  filename := path + "/test.cpp"
  if f, err := os.Create(filename); err != nil {
    panic(err)
  } else {
    f.WriteString(submission.Content)
    f.Close()
  }

  var ws *websocket.Conn
  for {
    var err error
    ws, err = websocket.Dial("ws://localhost:2501/judge", "", "http://localhost:5050")
    if err == nil {
      break
    }
    time.Sleep(10 * time.Second)
  }
  defer ws.Close()

  res_path := os.Getenv("GOPATH") + "/src/github.com/silverneko/gioj/td/" + strconv.Itoa(problem.ID)
  var testCases []map[string]interface{}
  // Open readonly
  if metaFile, err := ioutil.ReadFile(res_path + "/meta.json"); err != nil {
    panic(err)
  } else {
    var testCasesData map[string][][]int
    if err := json.Unmarshal(metaFile, &testCasesData); err != nil {
      panic(err)
    }
    for i, v := range testCasesData["Testcase"] {
      testCases = append(testCases, map[string]interface{}{
	"test_idx": i+1,
	"timelimit": problem.Timelimit,
	"memlimit": problem.Memlimit << 10,
	"metadata": map[string]interface{}{
	  "data": v,
	},
      })
    }
  }

  var comp_type string
  switch submission.Lang {
    default:
      comp_type = "g++"
    case models.LANGCPP:
      comp_type = "g++"
    case models.LANGCPPCLANG:
      comp_type = "clang++"
    case models.LANGPYTHON3:
      comp_type = "python3"
    case models.LANGGHC:
      comp_type = "g++"
  }
  msg, _ := json.Marshal(map[string]interface{}{
    "chal_id": 1,  // What number is this doesn't really matter
    "code_path": filename,
    "res_path": res_path,
    "comp_type": comp_type,
    "check_type": "diff",
    "metadata": "",
    "test": testCases,
  })
  if _, err := ws.Write(msg); err != nil {
    panic(err)
  }
  rcv := make([]byte, 1 << 20)
  if n, err := ws.Read(rcv); err != nil {
    panic(err)
  } else {
    rcv = rcv[:n]
  }

  var response struct {
    Verdict string
    Result []struct{
      Test_idx int
      State int
      Runtime int
      Peakmem int
      Verdict string
    }
  }
  if err := json.Unmarshal(rcv, &response); err != nil {
    log.Println(err)
  }
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
  result.Memused >>= 10
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

  if err := db.C("submissions").Update(bson.M{"_id": submission.ID}, bson.M{"$set": bson.M{"verdict": result}}); err != nil {
    panic(err)
  }
}

