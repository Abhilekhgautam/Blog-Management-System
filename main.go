package main

import(

   "net/http"
   "html/template"
   "github.com/gorilla/mux"
    "fmt"
	"log"
	"database/sql"
	"github.com/go-sql-driver/mysql"
    "math/rand"
	"time"
)

var db *sql.DB

type Blog struct{

  ID int64 `json:"id"`
  Title string `json:"title"`
  Author string `json:"author"`
  Description string `json:description`
  Published_date string `json:"published_date"`


}

type Secret struct{

	username string `json:"username"`

	password string   `json: "password"`


}

//global template variable

var templates = template.Must(template.ParseFiles("./templates/home.html", "./templates/addnew.html",
"./templates/adminhome.html", "./templates/edit.html","./templates/login.html"))

// connect to the database

func connectDB() {
   
       // TODO: add environment variable for user and passwd
       cfg := mysql.Config{
          User:                 "abhilekh",
          Passwd:               "abhilekh",
          Net:                  "tcp",
          Addr:                 "127.0.0.1:3306",
          DBName:               "blog",
          AllowNativePasswords:  true,
      }
      // Get a database handle.
      var err error
      db, err = sql.Open("mysql", cfg.FormatDSN())
      if err != nil {
          log.Fatal(err)
      }
  
      pingErr := db.Ping()
      if pingErr != nil {
          log.Fatal(pingErr)
      }
      fmt.Println("Connected to the database successfully")
  }

// loads all the blog (for client)

func getHome(w http.ResponseWriter, r *http.Request){

	 connectDB();
   
       rows, err := db.Query("select * from blogpost order by id desc")
       if err != nil{
   
           log.Fatal("Error while retrieving data: ",err)
   
       }
   
   
      defer rows.Close()
  
  
      var blogs []Blog
      for rows.Next(){
  
      var blog Blog
  
         if err := rows.Scan( &blog.ID, &blog.Title, &blog.Description, &blog.Author, &blog.Published_date); err != nil{
  
             log.Fatal("Unable to scan into slice",err)
  
         }
  
         // add element to blogs slice
        blogs = append(blogs, blog)
  
     }

	 fmt.Println(blogs)

	 err = templates.ExecuteTemplate(w, "home.html" , blogs)

	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)

	}
}

// loads form to add new blogs

func addBlogs(w http.ResponseWriter, r *http.Request){
 
	cookie , cookierr := r.Cookie("lemme-explain-cookie")


	if cookierr != nil{

		if cookierr == http.ErrNoCookie{
            fmt.Println("No cookie found")
            http.Redirect(w, r, "/login", 302)
            return;
		} else {

            log.Fatal("Some cookie error")

		}

	}


    connectDB();

	_, err := db.Query("select * from session where password = ?", cookie.Value)


	
	if err != nil{

     http.Redirect(w, r, "/login", 302)

     return;
	}


   err = templates.ExecuteTemplate(w, "addnew.html", nil)


   if err != nil{

     http.Error(w, err.Error(), http.StatusInternalServerError)

   }

}


// post blogs to the db

func postBlogs(w http.ResponseWriter, r *http.Request){

	fmt.Println("accepted a post request")

    err := r.ParseForm()

	if err !=nil{

		log.Fatal("Unable to parse Form")

	}

	title            := r.PostForm.Get("title")
	description      := r.PostForm.Get("description")
	author           := r.PostForm.Get("author")
	clickbait        := r.PostForm.Get("clickbait")
	published_date   := "Jun 27, 2022"

	//check if all the input fields are filled

	if len(title) == 0 || len(description) == 0 || len(author) == 0 || len(clickbait) == 0 || len(published_date) == 0{


	}

	// check if clickbait if maximum of 255 characters

	if len(clickbait) >= 255{



	}

    // if everything is validated insert into database

	 connectDB()
  
      // insert blog to db
      // execute sql query
       _, err = db.Exec("Insert into blogpost ( title, description, author, published_date) values(?,?,?,?)", title, description, author, published_date)
  
      if err != nil{
  
         log.Fatal("Unable to insert data to database",err);
  
      } else{


         http.Redirect(w, r, "/", http.StatusFound)

	  }
}

func editBlogs(w http.ResponseWriter, r *http.Request){
     
	fmt.Println("Trying to get edit blog")
	cookie , cookierr := r.Cookie("lemme-explain-cookie")


	if cookierr != nil{

		if cookierr == http.ErrNoCookie{
            fmt.Println("No cookie found")
            http.Redirect(w, r, "/login", 302)
            return;
		} else {

            log.Fatal("Some cookie error")

		}

	}


    connectDB();

	_, err := db.Query("select * from session where password = ?", cookie.Value)

	
	if err != nil{

     http.Redirect(w, r, "/login", 302)

     return;
	}


    id := r.URL.Path[len("/edit/"):]

	connectDB()
  
      row := db.QueryRow("select * from blogpost where id = ?",id)
  
      var blog Blog
  
      if err := row.Scan(&blog.ID, &blog.Title, &blog.Description, &blog.Author, &blog.Published_date); err != nil {
          if err == sql.ErrNoRows {
              log.Fatal("No such row found with given id:",id)
          } else {
  
             log.Fatal("Unable to scan ")
          }
      } else{

         err := templates.ExecuteTemplate(w,"edit.html",blog)

		 if err != nil{

           log.Fatal("Unable to render Template")

		 }



	  }


}

func updateBlogs(w http.ResponseWriter, r *http.Request){

	fmt.Println("recieved a update request")

	//TODO check if admin is logged in

    cookie , err := r.Cookie("lemme-explain-cookie")

	if err != nil || cookie == nil{

        http.Redirect(w,r, "/login", 302)
		return;

	}

    connectDB();

	_, err = db.Query("select * from session where value = ?", cookie.Value)

	if err != nil{

     http.Redirect(w, r, "/login", http.StatusMovedPermanently)

     return;
	}


	id := r.URL.Path[len("/update/"):]

	fmt.Println("Update id :", id)

	 err = r.ParseForm()

	 if err != nil{

        log.Fatal("Unable to parse form data")

	 }

	 title            := r.PostForm.Get("title")
	 description      := r.PostForm.Get("description")
	 author           := r.PostForm.Get("author")
	 //clickbait        := r.PostForm.Get("clickbait")

	 if len(title) == 0 || len(description) == 0 || len(author) == 0{

       // TODO:set error message of all fields required


	 }

	 //check if clickbait is atmost 255 characters long

	 //if len(clickbait) > 255{

        // TODO:set err message

//	 }//


	  _, err = db.Exec("Update blogpost set description = ? where id = ? ",description, id)
   
     if err != nil{

       log.Fatal("Unable to update with given data")

   } else {

       fmt.Println("Data updated successfully")
      // everything is fine

	  http.Redirect(w, r, "/admin", http.StatusFound)

   }
}

func getLogin(w http.ResponseWriter, r *http.Request){

    cookie , err := r.Cookie("lemme-explain-cookie")

	if err != nil || cookie == nil{

	   err = templates.ExecuteTemplate(w, "login.html", nil)

       if err != nil{

         log.Fatal("Unable to render provided template")

       } 

	   return;

	}

	fmt.Println(cookie.Value)
    connectDB();

	_, err = db.Query("select * from session where password = ?", cookie.Value)

	if err != nil{
   
       err = templates.ExecuteTemplate(w, "login.html", nil)

       if err != nil{

         log.Fatal("Unable to render provided template")

       }  
	}

	// user is already logged in show take him to admin home page

	http.Redirect(w, r, "/admin", 302)
 }    



func getAdminHome(w http.ResponseWriter, r *http.Request){

 // TODO: check if admin is logged in

    fmt.Println("Trying to load admin home page")
    cookie , err := r.Cookie("lemme-explain-cookie")

	if err != nil || cookie == nil{

		fmt.Println("Redirecting cookie might be nil")
        http.Redirect(w, r, "/login", 302)
		return;

	}


	fmt.Println(cookie.Value)
    connectDB();

	_, err = db.Query("select * from session where password = ?", cookie.Value)

	// if cookie value doesnot match, redirect to login..
	if err != nil{
     
	 fmt.Println("redirecting ...")
     http.Redirect(w, r, "/login", http.StatusMovedPermanently)

    return;
	}

	// if everything is fine proceed towards loading the home page

       rows, err := db.Query("select * from blogpost order by id desc")
       if err != nil{
   
           log.Fatal("Error while retrieving data: ",err)
   
       }
   
   
      defer rows.Close()
  
  
      var blogs []Blog
      for rows.Next(){
  
      var blog Blog
  
         if err = rows.Scan( &blog.ID, &blog.Title, &blog.Description, &blog.Author, &blog.Published_date); err != nil{
  
             log.Fatal("Unable to scan into slice",err)
  
         }
  
         // add element to blogs slice
        blogs = append(blogs, blog)
  
     }

     err = templates.ExecuteTemplate(w, "adminhome.html", blogs)

   if err != nil{

    log.Fatal("Unable to render provided template")

   }

}

//random string generating function

var characters = []rune("abcdefghijklmnopqrstuvwxyz123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr() string{

    str := make([]rune ,60)
	for i := range str{

        str[i] = characters[rand.Intn(len(characters))]

	}

	return string(str)
}

func login(w http.ResponseWriter, r *http.Request){

	err := r.ParseForm()

	if err != nil{

        log.Fatal("Unable to Parse Form")

	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	connectDB()

	row := db.QueryRow("Select * from admin where username = ? and password = ?", username, password)

	var sec Secret

	if err := row.Scan(&sec.username, &sec.password); err != nil{

        if err == sql.ErrNoRows {

           // return invalid username or password error

		} else{

          log.Fatal("Unable to Scan")

		}

	} else{

        // TODO:check if session already exists
        cookie , err := r.Cookie("lemme-explain-cookie")

	   if err != nil || cookie == nil{

		rand.Seed(time.Now().UnixNano())
		// store random string value
        randString := randStr()
		fmt.Println("Cookie value:", randString)

	   http.SetCookie(w, &http.Cookie{
		   Name  : "lemme-explain-cookie",
		   Value : randString,
	   })

	   _ , err := db.Exec("Insert into session values(?, ?)", username,randString)

	   
	   if err != nil{
          // because there is already a username present in the db
		  // update session value if the username already exist
           _ , err = db.Exec("Update session set password = ? where username = ?", randString, username)

	        if err != nil{

		       log.Fatal("Unable to update session")

	         }
	   }
 
          http.Redirect(w, r, "/admin", 302)
		  return;

	}

        
		rand.Seed(time.Now().UnixNano())
		// store random string value
        randString := randStr()
		fmt.Println("Cookie value:", randString)
       //TODO: set session 

	   http.SetCookie(w, &http.Cookie{
		   Name  : "lemme-explain-cookie",
		   Value : randString,
	   })

	   _ , err = db.Exec("Update  session set password = ? where username = ?", randString, username)

	   if err != nil{

		   log.Fatal("Unable to update session")

	   }

	   fmt.Println("Update User Session")
        // redirect to the admin home page
        fmt.Println("Login was successfull")
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
        return

	} 
}

//logs out the user

func logOut(w http.ResponseWriter, r *http.Request){

    cookie , cookierr := r.Cookie("lemme-explain-cookie")
	if cookierr != nil{

		if cookierr == http.ErrNoCookie{
            fmt.Println("No cookie found")
            http.Redirect(w, r, "/login", 302)
            return;
		} else {

            log.Fatal("Some cookie error")

		}

	}


    connectDB();

	_, err := db.Query("select * from session where password = ?", cookie.Value)

	
	if err != nil{

     http.Redirect(w, r, "/login", 302)

     return;
	}

     cookie = &http.Cookie{

        Name: "lemme-explain-cookie",
		Value: "",
		MaxAge: -1,

	 }

	 http.SetCookie(w, cookie)
	 http.Redirect(w,r,"/login", 302)



}

func main(){

   r := mux.NewRouter()

   r.HandleFunc("/",          getHome)
   r.HandleFunc("/add",       addBlogs)
   r.HandleFunc("/edit/{id}", editBlogs)
   r.HandleFunc("/admin",     getAdminHome)
   r.HandleFunc("/login",     getLogin)
   r.HandleFunc("/logout",    logOut)

   r.HandleFunc("/update/{id}", updateBlogs).Methods("POST")
   r.HandleFunc("/postblog", postBlogs).Methods("POST")
   r.HandleFunc("/signin",    login).Methods("POST")

   // serve static files
   fs := http.FileServer(http.Dir("./static/css/"))
   r.PathPrefix("/static/css/").Handler(http.StripPrefix("/static/css", fs))
   log.Fatal(http.ListenAndServe(":8080", r))

}
