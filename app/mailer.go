package main

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/mailjet/mailjet-apiv3-go"
)

//Uses Mailjet api to parse template and send to the user. modeled after Simple Mail Transfer Protocol (SMTP) script that was originally designed.
//SMPT does not work with App Engine though because ports 587, 25, etc. are all blocked and cannot be unblocked. An API is needed.

type Request struct {
	from    string
	to      []string
	subject string
	body    string
}

const (
	MIME = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

func NewRequest(to []string, subject string) *Request {
	return &Request{
		to:      to,
		subject: subject,
	}
}

func (r *Request) parseTemplate(fileName string, data interface{}) error {
	t, err := template.ParseFiles(fileName)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, data); err != nil {
		return err
	}
	r.body = buffer.String()
	return nil
}

func (r *Request) sendMail() bool {
	to := r.to[0]
	body := r.body
	publicKey := getConfig.Mail_Public_Key
	secretKey := getConfig.Mail_Private_Key

	mj := mailjet.NewMailjetClient(publicKey, secretKey)

	param := &mailjet.InfoSendMail{
		FromEmail: getConfig.Email,
		FromName:  getConfig.Sender_Name,
		Recipients: []mailjet.Recipient{
			mailjet.Recipient{
				Email: to,
			},
		},
		Subject:  r.subject,
		HTMLPart: body,
	}
	res, err := mj.SendMail(param)
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		fmt.Println("Email Sent")
		fmt.Println(res)
		return true
	}
	//if err := smtp.SendMail(SMTP, smtp.PlainAuth("", getConfig.Email, getConfig.Password, getConfig.Server), getConfig.Email, r.to, []byte(body)); err != nil {
}

func (r *Request) Send(templateName string, items interface{}) {
	err := r.parseTemplate(templateName, items)
	if err != nil {
		Log.Fatal(err)
	}
	if ok := r.sendMail(); ok {
		Log.Printf("Email has been sent to %s\n", r.to)
	} else {
		Log.Printf("Failed to send the email to %s\n", r.to)
	}
}
