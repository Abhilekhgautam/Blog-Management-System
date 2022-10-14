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
)

var templates = template.Must(template.ParseFiles("./templates/login.html"))

var characters = []rune("abcdefghijklmnopqrstuvwxyz123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandStr ... generates random string for cookie value.
func RandStr() string {
	str := make([]rune, 60)
	for i := range str {
		str[i] = characters[rand.Intn(len(characters))]
	}
	return string(str)
}

type Secret struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
			randString := RandStr()
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
		rand.Seed(time.Now().UnixNano())
		// store random string value
		randString := RandStr()
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
