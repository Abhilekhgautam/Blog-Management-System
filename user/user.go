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
	"projectOne/blog/Services"
	"projectOne/blog/db"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var templates = template.Must(template.ParseFiles("./templates/login.html","./templates/signup.html","./templates/chose-title.html","./templates/email-prompt.html","./templates/new-password.html"))

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

//todo use env variable for the key
var Store = sessions.NewCookieStore([]byte("VqM5LNd8BhUFnI4P8oIWEqf6oiPIASM4TV7RilpGtgUkKmnyIiKPnEC43aut"))

// this denotes the secret datas related to a user.
type Secret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Verified bool   `json:"verified"`
	Title    string  `json:"title"`
}

// this denotes the actual user information
type User struct{
	UserId   int64   `json:"userid"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
    BlogTitle string `json:"blogTitle"`
	Verified    bool  `json:"verified"`
}
// Get signup form for user.
// this is a get request from the user.
func GetSignup(w http.ResponseWriter, r *http.Request){
	fmt.Println("Trying to get signup page for you....")
	session, err := Store.Get(r, "smart-blogger-cookie")
	if err != nil{
		log.Fatal("Unable to retrieve session:",err)
	}
	fmt.Println("Trying to get signup page for you....")

	if session.IsNew{
		fmt.Println("New Session....")
		err := templates.ExecuteTemplate(w,"signup.html", nil)
		if err != nil{
			log.Fatal("Unable to render provided template")
		}
	} else{
		var username string
		fmt.Println("Trying to get signup page for you....")
		if len(session.Values) != 4{
			fmt.Println("len(session.Values) =",len(session.Values))
			fmt.Println("session.Values =", session.Values["value"].(string))
		}
	    err = db.Db.QueryRow("select username from session where password = ?", session.Values["value"].(string)).Scan(&username)
	    if err != nil{
		    if err == sql.ErrNoRows{
			    err = templates.ExecuteTemplate(w, "signup.html", nil)
			        if err != nil {
				        log.Fatal("Unable to render provided template")
			           }
		            } else{
						log.Fatal("Something went wrong:", err)
					}
	    } else{
			blogTitle := session.Values["blogTitle"].(string)
			http.Redirect(w,r, "/" + blogTitle + "/admin/", http.StatusFound)
		}
	}
}

// accepts post request for creating a new user.
// todo: parse old data in case of failure.
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
	if err := row.Scan(&username); err != nil{
		if err == sql.ErrNoRows{
			row := db.Db.QueryRow("Select email from emails where email = ?", email)
			if err := row.Scan(&email); err != nil{
				if err == sql.ErrNoRows{
					hashedPass := HashPassword(password)
					_, err = db.Db.Exec("Insert into User( username, email, password, blogTitle, verified) values(?,?,?,?,false)", username, email, hashedPass,"")
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
						  mail.SendEmail(username, generated_token, email,"confirmation")
						  err := templates.ExecuteTemplate(w,"signup.html","A confirmation email has been sent")
						  if err != nil{
							log.Fatal("Unable to render the signup template")
						  }
						  _, err = db.Db.Exec("Insert into Token(token,date_generated, date_expires, date_used, user_id) values(?,?,?,0,?)",generated_token,current_token, current_token + 172800,user_id)
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
		err := templates.ExecuteTemplate(w,"signup.html","this usename is already taken")
				if err != nil{
				  log.Fatal("Unable to render the signup template")
				}
	}
}

// GetLogin - this is a get request, renders login form for the user.
func GetLogin(w http.ResponseWriter, r *http.Request) {
	
	session, err := Store.Get(r,"smart-blogger-cookie")
	if err != nil{
		log.Fatal("Unable to decode available session info", err)
	}
	if session.IsNew{
		fmt.Println("New session")
		err = templates.ExecuteTemplate(w, "login.html", nil)
		if err != nil {
			log.Fatal("Unable to render provided template")
		}
		return
	}
	db.ConnectDB()
	var username string
	err = db.Db.QueryRow("select username from session where password = ?", session.Values["value"].(string)).Scan(&username)
	if err != nil{
		if err == sql.ErrNoRows{
			err = templates.ExecuteTemplate(w, "login.html", nil)
		    if err != nil {
			  log.Fatal("Unable to render provided template")
		    }
			return
		}
	}
	// user is already logged in so take him to admin home page
    // but check if title is empty..
	title := session.Values["blogTitle"].(string)
	if len(title) < 2{
        err = templates.ExecuteTemplate(w, "chose-title.html", struct {
			Username string 
			Msg string
			}{
			Username: username,
			Msg: "",
			})
		if err != nil {
			log.Fatal("Unable to render provided template:",err)
		   }
		return  
	}
	path := "/" + title + "/admin/"
	http.Redirect(w, r, path, http.StatusFound)
}

//Login - this is a post request, logs in the user
func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal("Unable to Parse Form")
	}
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")
	//todo: check password against hashed password.
	// hashed_password := HashPassword(password)
	db.ConnectDB()
	row := db.Db.QueryRow("Select user_id, Verified, blogTitle, password from User where username = ?", username)
	var hashed_password string
	var Verified bool
	var blogTitle string
	var user_id int64
	if err := row.Scan(&user_id, &Verified, &blogTitle, &hashed_password); err != nil {
		if err == sql.ErrNoRows {
			// return invalid username or password error
			errMsg := "Invalid username."
			err = templates.ExecuteTemplate(w, "login.html", errMsg)
			if err != nil {
				log.Fatal("Unable to render provided template")
			}
			return
		} else {
			log.Fatal("Unable to Scan:", err)
		}
	} 
	err = bcrypt.CompareHashAndPassword([]byte(hashed_password), []byte(password))
	if err != nil{
		errMsg := "Invalid password."
			err = templates.ExecuteTemplate(w, "login.html", errMsg)
			if err != nil {
				log.Fatal("Unable to render provided template")
			}
		} else {
			if !Verified{
				//todo: prompt for verification
				errMsg := "You need to verify first."
			err = templates.ExecuteTemplate(w, "login.html", errMsg)
			if err != nil {
				log.Fatal("Unable to render provided template")
			}
			return
			}
			randString := RandStr(60)
		    session, err := Store.Get(r, "smart-blogger-cookie")
			if err == nil{
				if session.IsNew{
					fmt.Println("post login new session")
					session.Values["value"] = randString
					session.Values["blogTitle"] = strings.Replace(blogTitle," ", "-", -1)
					session.Values["username"] = username
					err = sessions.Save(r, w)
					if err != nil{
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					_, err := db.Db.Exec("Insert into session values(?, ?)", username, randString)
					if err != nil{
						log.Fatal("Something went wrong:",err)
					}
					if len(session.Values["blogTitle"].(string)) < 2{
						err = templates.ExecuteTemplate(w, "chose-title.html", struct {
							Username string 
							Msg string
							}{
							Username: username,
							Msg: "",
							})
						if err != nil {
							log.Fatal("Unable to render provided template:",err)
						   }
						return  
					}
					http.Redirect(w, r, "/" + session.Values["blogTitle"].(string) + "/admin/", http.StatusFound)
				} else{
					var username string
					err = db.Db.QueryRow("select username from session where password = ?", session.Values["value"].(string)).Scan(&username)
					if err != nil{
					  if err == sql.ErrNoRows{
						fmt.Println("NO such sessions found:",err)
						fmt.Println(session.Values["value"].(string))
						session.Options.MaxAge = -1
						err := sessions.Save(r, w)
						if err != nil{
							log.Fatal(err)
						}
					   http.Redirect(w, r, "/login", http.StatusFound)
						}
					} else{
						if len(session.Values["blogTitle"].(string)) < 2{
							err = templates.ExecuteTemplate(w, "chose-title.html", struct {
								Username string 
								Msg string
								}{
								Username: username,
								Msg: "",
								})
							if err != nil {
								log.Fatal("Unable to render provided template:",err)
							   }
							return  
						}
						http.Redirect(w,r, "/" + session.Values["blogTitle"].(string) + "/admin/", http.StatusFound)
					}
				}
			} else{
				log.Fatal(err)
			}
	}
}

// LogOut ... logs out the user.
func LogOut(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "smart-blogger-cookie")
	if err != nil{
		log.Fatal("Unable to get session value:", err)
	}
	db.ConnectDB()
	_, err = db.Db.Query("delete from session where password = ?", session.Values["value"].(string))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	// setting Maxage < 0 deletes the session..
	session.Options.MaxAge = -1
	err = sessions.Save(r, w)
	if err != nil{
		log.Fatal(err)
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

// prompts user to set title..
func SetTitle(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil {
		log.Fatal("Unable to Parse Form")
	}
	session, err := Store.Get(r, "smart-blogger-cookie")
    if err != nil{
		log.Fatal(err)
	}
	if session.IsNew{
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	var username string
	err = db.Db.QueryRow("select username from session where password = ?", session.Values["value"].(string)).Scan(&username)
	if err != nil{
		if err == sql.ErrNoRows{
		http.Redirect(w, r, "/login", http.StatusFound)
		} else{
			log.Fatal(err)
		}
	} else{
		title := r.PostForm.Get("title")
		_, err = db.Db.Exec("Update User set blogTitle = ? where username = ?", title, username)
		if err != nil {
			log.Fatal("Unable to update with given data")
		} else {
			blogTitle := strings.Replace(title, " ", "-", -1)
			session.Values["blogTitle"] = blogTitle
			err = session.Save(r, w)
			if err != nil{
				log.Fatal("Could not save session value")
			}
			http.Redirect(w, r, "/" + blogTitle + "/admin/", http.StatusFound)
		}
	}
}

// prompt to enter new password for user....
func GetChangePassword(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil{
		log.Fatal("Unable to parse data")
	}
	token := r.Form.Get("token")
	
	err = templates.ExecuteTemplate(w,"new-password.html", struct{
		Token string
		Msg   string
	}{
		Token: token,
		Msg: " ",
	})
	if err != nil{
		log.Fatal("Unable to render provided template")
	}

}
// change password changes password ...
func ChangePassword(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil {
		log.Fatal("Unable to Parse Form")
	}
	token := r.Form.Get("token")

	pass := r.PostForm.Get("password")
	current_time := time.Now().Unix()

	var user_id int
	var date_generated int64
	var date_expires  int64
	var date_used     int64
	row := db.Db.QueryRow("Select user_id, date_generated, date_expires, date_used from Token where token = ?", token)
	if err := row.Scan(&user_id, &date_generated, &date_expires, &date_used); err != nil{
       if err == sql.ErrNoRows{
		//todo: no such token provide a link to signup..
		http.Redirect(w, r, "/signup", http.StatusFound)
		fmt.Println("No such rows..")
	   } else {
		log.Fatal("Something went wrong:", err)
	   }
	}

   //reuse of the token...
	if (date_used != 0){
       http.Redirect(w,r, "/signup", http.StatusFound)
	}
	// use of expired token...
	if(date_expires < current_time){
		//todo: inform about the expired token and prompt for re confirmation..
		fmt.Println("Token expired..")

	} else{
		_, err = db.Db.Exec("Update Token set date_used = ? where token=?",current_time, token)
	    if err != nil {
		  log.Fatal("Unable to update with given data")
	     }
         hashed_password := HashPassword(pass)
		_, err = db.Db.Exec("Update User set password = ? where user_id=?", hashed_password, user_id)
	    if err != nil {
		  log.Fatal("Unable to update with given data")
	     } else {
		   http.Redirect(w, r, "/login", http.StatusFound)
	      }
	}
}

// prompts user to enter email...
func GetForgot(w http.ResponseWriter, r *http.Request){
	err := templates.ExecuteTemplate(w,"email-prompt.html", struct{
		Link string
		Msg  string
	}{
		Link: "forgot",
		Msg : "",
	})
		if err != nil{
			log.Fatal("Unable to render provided template:",err)
		}
}

// post request for change in password...
func ForgotPassword(w http.ResponseWriter, r *http.Request){
	err := r.ParseForm()
	if err != nil {
		log.Fatal("Unable to Parse Form")
	}
	email := r.PostForm.Get("email")
	fmt.Println("Email Address:",email)
    var user_id int64
	db.ConnectDB()
	fmt.Println("Select user_id from User where email = ",email)
	if err := db.Db.QueryRow("Select user_id from User where email = ?", email).Scan(&user_id); err != nil{
		if err == sql.ErrNoRows{
			//.. no such email found.
			fmt.Println("NO such email address found")
			err = templates.ExecuteTemplate(w,"email-prompt.html", struct{
				Link string
				Msg  string
			}{
				Link: "forgot",
				Msg : "No such email address found",
			})
				if err != nil{
					log.Fatal("Unable to render provided template")
				}
		} else{
			log.Fatal("Unable to scan data")
		}
	}
    fmt.Println("The Id for the given email is:",user_id)
	generated_token := RandStr(24)
	current_token := time.Now().Unix()

	_, err = db.Db.Exec("Insert into Token(token,date_generated, date_expires, date_used, user_id) values(?,?,?,0,?)",generated_token,current_token, current_token + 172800,user_id)
	if err != nil{
		log.Fatal("Unable to insert token data into database:", err)
	}

	mail.SendEmail("",generated_token,email,"forgot")
	err = templates.ExecuteTemplate(w,"email-prompt.html", struct{
		Link string
		Msg  string
	}{
		Link: "forgot",
		Msg : "A confirmation email has been sent.",
	})
		if err != nil{
			log.Fatal("Unable to render provided template")
		}
}
