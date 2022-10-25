package user

/*

This package proviedes facilities related to the user (admin) of the site.


*/

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"projectOne/blog/db"
	"time"
	"golang.org/x/crypto/bcrypt"
	"projectOne/blog/Services"
)

var templates = template.Must(template.ParseFiles("./templates/login.html","./templates/signup.html"))

var characters = []rune("abcdefghijklmnopqrstuvwxyz123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func HashPassword(password string) string{
    hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil{
		log.Fatal(err)
	}
		return string(hashedPass)
}

// RandStr ... generates random string
func RandStr(length int) string {
	rand.Seed(time.Now().UnixNano())
	str := make([]rune, length)
	for i := range str {
		str[i] = characters[rand.Intn(len(characters))]
	}
	return string(str)
}

// this denotes the secret datas related to a user.
type Secret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// this denotes the actual user information
type User struct{
	UserId   int64   `json:"userid"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
    BlogTitle string `json:"blogTitle"`
}
// Get signup form for user.
// this is a get request from the user.
func GetSignup(w http.ResponseWriter, r *http.Request){
	err := templates.ExecuteTemplate(w,"signup.html", nil)
	if err != nil{
		log.Fatal("Unable to render provided template")
	}
}

// accepts post request for creating a new user.
func Signup(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil{
		log.Fatal("Unable to parse the form data")
	}
	username := r.PostForm.Get("username")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	if len(username) == 0 || len(email) == 0 || len(password) == 0 {
		errMsg := "All the fields are required"
		err := templates.ExecuteTemplate(w,"signup.html", errMsg)
		if err != nil{
			log.Fatal("Unable to render signup template")
		}
	}

	db.ConnectDB()
	defer db.Db.Close()
	row := db.Db.QueryRow("Select username from Users where username = ?",username)
	if err := row.Scan(); err != nil{
		if err == sql.ErrNoRows{
			row := db.Db.QueryRow("Select email from emails where email = ?", email)
			if err := row.Scan(); err != nil{
				if err == sql.ErrNoRows{
					hashedPass := HashPassword(password)
					_, err = db.Db.Exec("Insert into User( username, email, password, blogTitle) values(?,?,?,NULL)", username, email, hashedPass)
					if err != nil {
						 log.Fatal("Unable to insert data to database", err)
					} else {
						  generated_token := RandStr(24)
						  current_token := time.Now().Unix()
						  user_id := 0
						  row := db.Db.QueryRow("Select user_id from User where username = ?",username)
						  if row.Scan(&user_id); err != nil{
							log.Fatal("Unable to scan data")
						  }
						  mail.SendEmail(username, generated_token, email)
						  err := templates.ExecuteTemplate(w,"signup.html","A confirmation email has been sent")
						  if err != nil{
							log.Fatal("Unable to render the signup template")
						  }
						  _, err = db.Db.Exec("Insert into Token(token,date_generated, date_expires, date_used, user_id) values(?,?,?,NULL,?)",generated_token,current_token, current_token + 172800,user_id)
						  if err != nil{
							  log.Fatal("Unable to insert token data into database:", err)
						  }
					 } 
				} else{
					log.Fatal("Something went wrong:", err)
				}
			} else{
				err := templates.ExecuteTemplate(w,"signup.html","this email already exists")
				if err != nil{
				  log.Fatal("Unable to render the signup template")
				}
			}
		} else{
			err := templates.ExecuteTemplate(w,"signup.html","this email already exists")
				if err != nil{
				  log.Fatal("Unable to render the signup template")
				}
		}
	} else{
		err := templates.ExecuteTemplate(w,"signup.html","this usename is already takaen")
				if err != nil{
				  log.Fatal("Unable to render the signup template")
				}
	}
}

// GetLogin - this is a get request, renders login form for the user.
func GetLogin(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("lemme-explain-cookie")
	if err != nil || cookie == nil {
		err = templates.ExecuteTemplate(w, "login.html", nil)
		if err != nil {
			log.Fatal("Unable to render provided template")
		}
		return
	}
	db.ConnectDB()
	_, err = db.Db.Query("select * from session where password = ?", cookie.Value)
	if err != nil {
		err = templates.ExecuteTemplate(w, "login.html", nil)
		if err != nil {
			log.Fatal("Unable to render provided template")
		}
	}
	// user is already logged in so take him to admin home page
	http.Redirect(w, r, "/admin", 302)
}

//Login - this is a post request, logs in the user
func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal("Unable to Parse Form")
	}
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	db.ConnectDB()
	row := db.Db.QueryRow("Select * from admin where username = ? and password = ?", username, password)
	var sec Secret
	if err := row.Scan(&sec.Username, &sec.Password); err != nil {
		if err == sql.ErrNoRows {
			// return invalid username or password error
			errMsg := "Invalid username or Password."
			err = templates.ExecuteTemplate(w, "login.html", errMsg)
			if err != nil {
				log.Fatal("Unable to render provided template")
			}
		} else {
			log.Fatal("Unable to Scan")
		}
	} else {
		//checks if session already exists
		cookie, err := r.Cookie("lemme-explain-cookie")
		if err != nil || cookie == nil {
			rand.Seed(time.Now().UnixNano())
			// store random string value
			randString := RandStr(60)
			fmt.Println("Cookie value:", randString)

			http.SetCookie(w, &http.Cookie{
				Name:  "lemme-explain-cookie",
				Value: randString,
			})

			//creates and store a new session
			_, err := db.Db.Exec("Insert into session values(?, ?)", username, randString)
			if err != nil {
				// because there is already a username present in the db
				// update session value if the username already exist
				_, err = db.Db.Exec("Update session set password = ? where username = ?", randString, username)
				if err != nil {
					log.Fatal("Unable to update session")
				}
			}
			http.Redirect(w, r, "/admin", 302)
			return
		}

		// store random string value
		randString := RandStr(60)
		fmt.Println("Cookie value:", randString)

		http.SetCookie(w, &http.Cookie{
			Name:  "lemme-explain-cookie",
			Value: randString,
		})
		_, err = db.Db.Exec("Update  session set password = ? where username = ?", randString, username)
		if err != nil {
			log.Fatal("Unable to update session")
		}
		// redirect to the admin home page
		fmt.Println("Login was successfull")
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
		return
	}
}

// LogOut ... logs out the user.
func LogOut(w http.ResponseWriter, r *http.Request) {
	cookie, cookierr := r.Cookie("lemme-explain-cookie")
	if cookierr != nil {
		if cookierr == http.ErrNoCookie {
			fmt.Println("No cookie found")
			http.Redirect(w, r, "/login", 302)
			return
		} else {
			log.Fatal("Some cookie error")
		}
	}
	db.ConnectDB()
	_, err := db.Db.Query("select * from session where password = ?", cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	// setting a new cookie value with the previous cookie name deletes previous cookie
	cookie = &http.Cookie{

		Name:   "lemme-explain-cookie",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/login", 302)
}
