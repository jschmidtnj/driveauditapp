package main

import (
    "gopkg.in/mgo.v2"
    //"crypto/tls"
    //"net"
    "time"
)

var (
  mainsession *mgo.Session
)
//blocked by firewall, so will not deploy on local computer

//This function connects to mongodb database using mgo. Needs to work on Virtual Machine because corporate firewalls block the data

func mongoconnect() {
  /*
  //old method with URI (for mlab or mongodb atlas connection)
  //URI without ssl=true
  var mongoURI = getConfig.Mongo_URI
  dialInfo, err := mgo.ParseURL(mongoURI)
  if err != nil {
    panic(err)
  }
  //Below part is similar to above.
  tlsConfig := &tls.Config{}
  dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
      conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
      return conn, err
  }
  */
  var hosts = getConfig.Mongo_Hosts
  var database = getConfig.Mongo_Database
  var username = getConfig.Mongo_Username
  var password = getConfig.Mongo_Password
  dialInfo := &mgo.DialInfo{
    Addrs:    []string{hosts},
    Timeout:  999999 * time.Hour,
    Database: database,
    Username: username,
    Password: password,
  }
  session, err := mgo.DialWithInfo(dialInfo)
  if err != nil {
    panic(err)
  }
  mainsession = session
  //defer mainsession.Close() // clean up when we're done
}
