package main

import (
  "log"
  "os"
  //"time"
)

var (
  Log *log.Logger
)

func NewLog(filelocation string) {
  f, err := os.OpenFile(filelocation, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  //f, err := os.Create(filelocation)
  if err != nil {
    Log.Printf("error opening file: %v", err)
  }
  //defer f.Close()
  Log = log.New(f, "", log.LstdFlags|log.Lshortfile)
}
