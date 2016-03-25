package models

import (
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "log"
)

type DBSession struct{
  *mgo.Session
}

func (s DBSession) C(name string) *mgo.Collection {
  return s.DB("").C(name)
}

func EnsureDBIndices(DB *mgo.Session) error {
  if err := DB.DB("").C("users").EnsureIndex(mgo.Index{
    Key: []string{"username"},
    Unique: true,
  }); err != nil {
    log.Println(err)
    return err
  }
  return nil
}

type Idx struct {
  ID string `bson:"_id"`
  Seq int
}

type Submission struct {
  ID bson.ObjectId  `bson:"_id"`
  Pid int
  Username string
  Verdict Verdict
  Lang int
  Content string
}

type Verdict struct {
  Result int
  Timeused int
  Memused int
}

const (
  QUEUED int = iota
  JUDGING
  AC
  WA
  RE
  TLE
  MLE
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
  TestdataCount int `bson:"testdatacount"`
}

type DiscussPost struct {
  ID bson.ObjectId   `bson:"_id"`
  Content string
  Username string
}

type User struct {
  ID bson.ObjectId   `bson:"_id"`
  Name string
  Username string
  Hashed_password []byte
  Role int
}

const (
  USERROLE int = iota
  ADMINROLE
)

func (u *User) IsAdmin() bool {
  return u.Role == ADMINROLE
}

