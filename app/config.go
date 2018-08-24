package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Sender_Name                 string
	Mail_Public_Key             string
	Mail_Private_Key            string
	Email                       string
	Recipient                   string
	Recipient_Name              string
	Subject                     string
	MaxResults                  int64
	Threshold                   int
	Expire                      int32
	Seconds_Between_Api         int
	Count_Between_Check         int
	Service_Account_Email       string
	Service_Account_Private_Key string
	Service_Account_Scopes      []string
	Token_Refresh_Count         int
	Mongo_URI                   string
	Mongo_Hosts                 string
	Mongo_Database              string
	Mongo_Username              string
	Mongo_Password              string
	Count_Between_Log_Reset     int
	Database_Name               string
	User_Data_Collection        string
	Suspect_Data_Collection     string
	Log_Data_Collection         string
	Admin_Email                 string
	Mongo_Connect               string
	REST_Port                   string
	Frontend_Port               int
	Backend_Port                string
	REST_Handler                string
	Whitelisted_Emails_Regex    string
}

func (c *Config) Read() {
	if _, err := toml.DecodeFile("config/config.toml", &c); err != nil {
		Log.Fatalf("problem with config file", err)
		panic(err)
	}
}
