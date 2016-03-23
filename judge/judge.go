package main

import (
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "log"
  "time"
  "os"
  "golang.org/x/net/websocket"
  "encoding/json"
)

func main() {
  log.Println("Start judge service...")

  os.RemoveAll("/tmp/gioj-judge")
  os.Mkdir("/tmp/gioj-judge", 0755)
  for {
    db := DBSession{DB.Copy()}
    var submission Submission
    _, err := db.C("submissions").
      Find(bson.M{"verdict.result": QUEUED}).
      Sort("_id").
      Apply(mgo.Change{Update: bson.M{"$set": bson.M{"verdict.result": JUDGING}}}, &submission)
    if err == nil {
      log.Println("Recieved: ", submission.ID.Hex())
      go judge(&submission)
    } else {
      log.Println("Nothing to judge")
      time.Sleep(5 * time.Second)
    }
    db.Close()
  }
}

func judge(submission *Submission) {
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

  cf, _ := websocket.NewConfig("ws://localhost:2501/judge", "http://localhost:5050/")
  ws, err := websocket.DialConfig(cf)
  if err != nil {
    log.Println("Ws create: ", err)
    return
  }
  defer ws.Close()

  msg, _ := json.Marshal(map[string]interface{}{
    //"chal_id": submission.ID.Hex(),
    "chal_id": 1,
    "code_path": filename,
    "res_path": "/home/silvernegi/Projects/gioj/td/1",
    "comp_type": "g++",
    "check_type": "diff",
    "metadata": "",
    "test": []map[string]interface{}{
      {
	"test_idx": 1,
	"timelimit": 1000,
	"memlimit": 64 * 2 << 20,
	"metadata": map[string]interface{}{
	  "data": []int{1},
	},
      },
    },
  })
  ws.Write(msg)
  //log.Println(string(msg))

  rcv := make([]byte, 2 << 20)
  ws.Read(rcv)
  log.Println(string(rcv))

  var result verdict
  db := DBSession{DB.Copy()}
  defer db.Close()
  err = db.C("submissions").Update(bson.M{"_id": submission.ID}, bson.M{"$set": bson.M{"verdict": result}})
  if err != nil {
    log.Println("Judge Error: ", err)
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

type DBSession struct {
  *mgo.Session
}

func (s DBSession) C(name string) *mgo.Collection {
  return s.DB("").C(name)
}

type Submission struct {
  ID bson.ObjectId  `bson:"_id"`
  Pid int
  Username string
  Verdict verdict
  Lang int
  Content string
}

type verdict struct {
  Result int
  Timeused int
  Memused int
}

const (
  QUEUED int = iota
  JUDGING
  AC
  WA
  TLE
  MLE
  RE
  CE
  ERR
)

const (
  LANGCPP int = iota
  LANGC
  LANGGHC
  LANGSIZE
)

type Problem struct {
  ID int   `bson:"_id"`
  Name string
  Content string
  AuthorName string `bson:"authorname"`
  Memlimit int
  Timelimit int
  Testdata string
}

