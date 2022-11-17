package mail

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"projectOne/blog/db"
	"text/template"
	"time"

	"github.com/joho/godotenv"
)

var templates = template.Must(template.ParseFiles("./templates/email-temp.html","./templates/login.html"))

//todo: use .env files to hide credentials.
func SendEmail(username, token , to_email, purpose string){
	godotenv.Load()
	password := os.Getenv("APP_PASSWORD")
	auth := smtp.PlainAuth("","smartblogger119@gmail.com", password, "smtp.gmail.com")

	to := []string{to_email}

    var body bytes.Buffer

    mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
    body.Write([]byte(fmt.Sprintf("Subject: Confirm your Email!\n%s\n\n", mimeHeaders)))
    
	var msg string
	var link string
	if purpose == "confirmation"{
		msg = "We recieved a registration request for Smart Blogger with this email.Click the given link to confirm:"
		link = "confirm_email"

	} else if purpose == "forgot"{
		msg = "We came to know that you have forgotton you password for Smart Blogger.Click this link to confirm"
		link = "forgot_password"
	}

    templates.Execute(&body, struct {
     Username   string
     Token      string
	 Msg        string
	 Link       string
    }{
     Username:  username,
     Token:     token,
	 Msg: msg,
	 Link: link,
    })

	err := smtp.SendMail("smtp.gmail.com:587",auth,"smartblogger",to,body.Bytes())
	 if err != nil{
		log.Fatal("Unable to send email:",err)
	 }
}

func ConfirmEmail(w http.ResponseWriter, r *http.Request){
	fmt.Println("Called ConfirmEmail")
	err := r.ParseForm()
	if err != nil{
		log.Fatal("Unable to parse data")
	}
	token := r.Form.Get("token")
	fmt.Println(token)
	db.ConnectDB()
	current_time := time.Now().Unix()
    user_id := 0
	var date_generated int64
	var date_expires  int64
	var date_used     int64
	row := db.Db.QueryRow("Select user_id, date_generated, date_expires, date_used from Token where token = ?", token)
	if err := row.Scan(&user_id, &date_generated, &date_expires, &date_used); err != nil{
       if err == sql.ErrNoRows{
		//todo: no such token, provide a link to signup..
		errMsg := "Invalid Token Used."
		err := templates.ExecuteTemplate(w,"signup.html", struct{
			Color string
			Msg string 
			Username string
			Email string
			Password string
		}{
			"red",
			errMsg,
			"",
			"",
			"",
		})
		if err != nil{
			log.Fatal("Unable to render signup template")
		}
		fmt.Println("No such rows..")
	   } else {
		log.Fatal("Something went wrong:", err)
	   }
	}

   //reuse of the token...
	if (date_used != 0){
		errMsg := "The token is already used."
		err := templates.ExecuteTemplate(w,"signup.html", struct{
			Color string
			Msg string 
			Username string
			Email string
			Password string
		}{
			"red",
			errMsg,
			"",
			"",
			"",
		})
		if err != nil{
			log.Fatal("Unable to render signup template")
		}
	}
	// use of expired token...
	if(date_expires < current_time){
		// delete the unverified user after the token has expired ..
		_, err = db.Db.Exec("delete from User where user_id = ?", user_id)
	    if err != nil {
		 log.Fatal("Unable to delete the blog")
	    }
		errMsg := "The token has already expired. Sign up again."
		err := templates.ExecuteTemplate(w,"signup.html", struct{
			Color string
			Msg string 
			Username string
			Email string
			Password string
		}{
			"red",
			errMsg,
			"",
			"",
			"",
		})
		if err != nil{
			log.Fatal("Unable to render signup template")
		}
		fmt.Println("Token expired..")

	} else{
		var title string
		var username string
		if err := db.Db.QueryRow("select username, blogTitle from User where user_id = ?", user_id).Scan(&username, &title); err != nil{
			if err == sql.ErrNoRows{
				http.Redirect(w, r, "/signup", http.StatusFound)
			}
		}
		_, err = db.Db.Exec("Update Token set date_used = ? where token=?",current_time, token)
	    if err != nil {
		  log.Fatal("Unable to update with given data")
	     }

		_, err = db.Db.Exec("Update User set Verified = true where user_id=?",user_id)
	    if err != nil {
		  log.Fatal("Unable to update with given data")
	     } else {
			msg := "sucessfully verfied,login now"
			err = templates.ExecuteTemplate(w, "login.html", struct{
				Msg string
				Username string
				Password string

			}{
              msg,
			  "",
			  "",
			})
			if err != nil {
				log.Fatal("Unable to render provided template")
			}
	      }
	}
}

