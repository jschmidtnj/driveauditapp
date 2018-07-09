package main

import (
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "log"
    "time"
)

var (
  mongoLogger *log.Logger
)

type MongoWriter struct {
  sess *mgo.Session
}

//creates a log file system in MongoDB, as another collection, and saves to a global variable

func (mw *MongoWriter) Write(p []byte) (n int, err error) {
  if len(p) > 0 && p[len(p)-1] == '\n' {
    p = p[:len(p)-1] // Cut terminating newline
  }
  c := mw.sess.DB(getConfig.Database_Name).C(getConfig.Log_Data_Collection)
  err = c.Insert(bson.M{
    "created": time.Now(),
    "msg":     string(p),
  })
  if err != nil {
    return
  }
  return len(p), nil
}

func createMainLogger(session *mgo.Session) {
  mw := &MongoWriter{session}
  mongoLogger = log.New(mw, "", 0)
  //log.SetOutput(mw) //or do this for the default logger
  // Now the default Logger of the log package uses our MongoWriter.
  // Generate a log message that will be inserted into MongoDB:
  //mongoLogger.Println("Starting Logging")
}
