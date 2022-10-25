package mail

import (
	"log"
	"net/smtp"
	"text/template"
	"fmt"
	"bytes"
)

var templates = template.Must(template.ParseFiles("./templates/email-temp.html"))

func SendEmail(username, token , to_email string){
	//todo: use .env files to hide credentials.
	auth := smtp.PlainAuth("","smartblogger119@gmail.com","xlicbslexuzeixcr","smtp.gmail.com")

	to := []string{to_email}

    var body bytes.Buffer

    mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
    body.Write([]byte(fmt.Sprintf("Subject: Confirm your Email!\n%s\n\n", mimeHeaders)))

    templates.Execute(&body, struct {
     Username   string
     Token      string
    }{
     Username:  username,
     Token:     token,
    })

	err := smtp.SendMail("smtp.gmail.com:587",auth,"smartblogger",to,body.Bytes())
	 if err != nil{
		log.Fatal("Unable to send email:",err)
	 }
}
