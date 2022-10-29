package blog

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"projectOne/blog/db"
	"strings"
	"time"
	"projectOne/blog/user"
)

type Blog struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Description   string `json:"description"`
	Clickbait     string `json:"clickbait"`
	PublishedDate string `json:"published_date"`
}

var templates = template.Must(template.ParseFiles("./templates/home.html", "./templates/addnew.html",
	"./templates/adminhome.html", "./templates/edit.html", "./templates/delete.html","./templates/view.html","./templates/chose-title.html"))

// GetHome - loads all the blog (for client)
func GetHome(w http.ResponseWriter, r *http.Request) {
	db.ConnectDB()
	title := r.URL.Path
	spaced_title := strings.Replace(title,"-"," ", -1)
	var username string
	if err := db.Db.QueryRow("Select username from User where blogTitle = ?", spaced_title).Scan(&username); err != nil{
      if err == sql.ErrNoRows{
		// todo: Send a 404...
	  } else{
		log.Fatal("Sth went wrong", err)
	  }
	}
	rows, err := db.Db.Query("select * from blogpost where author = ? order by id desc", username)
	if err != nil {
		log.Fatal("Error while retrieving data: ", err)
	}
	defer rows.Close()
	var blogs []Blog
	for rows.Next() {
		var blog Blog
		if err := rows.Scan(&blog.ID, &blog.Title, &blog.Author, &blog.Description, &blog.Clickbait, &blog.PublishedDate); err != nil {
			log.Fatal("Unable to scan into slice", err)
		}
		// adds element to blogs slice
		blogs = append(blogs, blog)
	}
	err = templates.ExecuteTemplate(w, "home.html", blogs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AddBlogs - this is a get request, renders a template with form field to add new blogs.
func AddBlogs(w http.ResponseWriter, r *http.Request) {
	session, err := user.Store.Get(r,"smart-blogger-cookie")
	if err != nil {
		log.Fatal("Unable to decode provided email value", err)
	} 
	if session.IsNew{
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	db.ConnectDB()
	var username string
	if err = db.Db.QueryRow("select username from session where password = ?", session.Values["Value"].(string)).Scan(&username); err != nil{
		if  err == sql.ErrNoRows{
			http.Redirect(w, r, "/login", http.StatusFound)
		} else{
			log.Fatal("Something went wrong:", err)
		}
	}

	err = templates.ExecuteTemplate(w, "addnew.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
// this is a post request
// PostBlogs - post blogs to the db
func PostBlogs(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal("Unable to parse Form")
	}
    
	session, err := user.Store.Get(r,"smart-blogger-cookie")
	if err != nil {
		log.Fatal("Unable to decode provided email value", err)
	} 
	if session.IsNew{
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	var username string
	if err := db.Db.QueryRow("Select username from session where password = ?", session.Values["Value"]).Scan(&username); err != nil{
      if err == sql.ErrNoRows{
		http.Redirect(w, r, "/login", http.StatusFound)
	  } else{
		log.Fatal("Sth went wrong", err)
	  }
	}
	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	author := username
	clickbait := r.PostForm.Get("clickbait")
	publishedDate := time.Now().Format("2006-Jan-02")

	//check if all the input fields are filled
	if len(title) == 0 || len(description) == 0 ||  len(clickbait) == 0 {
		errMsg := "All the input fields are required"
		err = templates.ExecuteTemplate(w, "addnew.html", errMsg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	// check if clickbait is maximum of 255 characters
	if len(clickbait) > 255 || len(title) > 255 || len(author) > 255 {
		errMsg := "Fields except description cannot be more than 255 characters"
		err = templates.ExecuteTemplate(w, "addnew.html", errMsg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	// if everything is validated insert into database
	db.ConnectDB()
	// insert blog to db
	// execute sql query
	_, err = db.Db.Exec("Insert into blogpost ( title, description, author, published_date, clickbait) values(?,?,?,?,?)", title, description, author, publishedDate, clickbait)
	if err != nil {
		log.Fatal("Unable to insert data to database", err)
	} else {
		http.Redirect(w, r, "/"+ session.Values["blogTitle"].(string) + "/admin", http.StatusFound)
	}
}

// EditBlogs - this is a get request, renders input fields to edit blogs
func EditBlogs(w http.ResponseWriter, r *http.Request) {

	session, err := user.Store.Get(r,"smart-blogger-cookie")
	if err != nil {
		log.Fatal("Unable to decode session value:", err)
	}
	if session.IsNew{
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	db.ConnectDB()
	var username string
	if err = db.Db.QueryRow("select username from session where password = ?", session.Values["value"].(string)).Scan(&username); err != nil{
		if err == sql.ErrNoRows{
			http.Redirect(w, r, "/login", http.StatusFound)
		} else{
			log.Fatal("Something went wrong:", err)
		}
	}
	path := "/" + session.Values["blogTitle"].(string) + "/edit/"
	id := r.URL.Path[len(path):]

	row := db.Db.QueryRow("select * from blogpost where id = ?", id)
	var blog Blog
	if err := row.Scan(&blog.ID, &blog.Title, &blog.Author, &blog.Description, &blog.Clickbait, &blog.PublishedDate); err != nil {
		if err == sql.ErrNoRows {
			log.Fatal("No such row found with given id:", id)
		} else {

			log.Fatal("Unable to scan ")
		}
	} else {
		err := templates.ExecuteTemplate(w, "edit.html", blog)
		if err != nil {
			log.Fatal("Unable to render Template")
		}
	}
}

// UpdateBlogs - this is a post request, updates the blog to the database.
func UpdateBlogs(w http.ResponseWriter, r *http.Request) {
	session, err := user.Store.Get(r,"smart-blogger-cookie")
	if err != nil{
       log.Fatal("something went wrong:", err)
	}
	if session.IsNew{
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	db.ConnectDB()
	_, err = db.Db.Query("select * from session where password = ?", session.Values["value"].(string))
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	id := r.URL.Path[len("/update/"):]
	err = r.ParseForm()
	if err != nil {
		log.Fatal("Unable to parse form data")

	}

	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	author := r.PostForm.Get("author")
	clickbait := r.PostForm.Get("clickbait")

	_, err = db.Db.Exec("Update blogpost set title = ? ,description = ?, clickbait = ?, author = ? where id = ? ", title, description, clickbait, author, id)
	if err != nil {
		log.Fatal("Unable to update with given data")
	} else {
		http.Redirect(w, r, "/" + session.Values["value"].(string) + "/admin", http.StatusFound)
	}
}

//GetAdminHome - this is a get request, loads the admin home page
func GetAdminHome(w http.ResponseWriter, r *http.Request) {
	session, err := user.Store.Get(r,"smart-blogger-cookie")
	if err != nil{
		log.Fatal("something went wrong:", err)
	}
	if session.IsNew{
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
	db.ConnectDB()
	var username string
	if err = db.Db.QueryRow("select username from session where password = ?", session.Values["value"]).Scan(&username); err != nil{
       if err == sql.ErrNoRows{
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	   } else{
		log.Fatal("something went wrong:", err)
	   }
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

	// if everything is fine proceed towards loading the home page
	rows, err := db.Db.Query("select * from blogpost where author = ? order by id desc", username)
	if err != nil {
		log.Fatal("Error while retrieving data: ", err)
	}
	defer rows.Close()
	var blogs []Blog
	for rows.Next() {
		var blog Blog
		if err = rows.Scan(&blog.ID, &blog.Title, &blog.Author, &blog.Description, &blog.Clickbait, &blog.PublishedDate); err != nil {
			log.Fatal("Unable to scan into slice", err)
		}
		// add element to blogs slice
		blogs = append(blogs, blog)
	}
	err = templates.ExecuteTemplate(w, "adminhome.html", blogs)
	if err != nil {
		log.Fatal("Unable to render provided template")
	}
}

// GetDelete - this is a get request, renders delete option for user confirmation.
func GetDelete(w http.ResponseWriter, r *http.Request) {
	session, err := user.Store.Get(r,"smart-blogger-cookie")
	if err!= nil {
		log.Fatal("Something went wrong:", err)
	}
	if session.IsNew{
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	db.ConnectDB()
	var username string
	if err = db.Db.QueryRow("select username from session where password = ?", session.Values["value"]).Scan(&username); err != nil{
       if err == sql.ErrNoRows{
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	   } else{
		log.Fatal("something went wrong:", err)
	   }
	}
	path := "/" + session.Values["blogTitle"].(string) + "/delete/"
	id := r.URL.Path[len(path):]
	err = templates.ExecuteTemplate(w, "delete.html", id)
	if err != nil {
		log.Fatal("Unable to render provided template")
	}
}

// DeleteBlog - this is a delete request, if everything goes well this deletes the blog from db.
func DeleteBlog(w http.ResponseWriter, r *http.Request) {
	session, err := user.Store.Get(r,"smart-blogger-cookie")
	if err != nil {
		log.Fatal("Something went wrong:", err)
	}
	if session.IsNew{
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	var username string
	if err = db.Db.QueryRow("select username from session where password = ?", session.Values["value"]).Scan(&username); err != nil{
       if err == sql.ErrNoRows{
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	   } else{
		log.Fatal("something went wrong:", err)
	   }
	}
	id := r.URL.Path[len("/delete/"):]
	_, err = db.Db.Exec("delete from blogpost where id = ?", id)
	if err != nil {
		log.Fatal("Unable to delete the blog")
	}
	http.Redirect(w, r, "/admin", http.StatusFound)
}

//viewBlog - this is a get request.
// views the content of blog
func ViewBlog(w http.ResponseWriter, r *http.Request){
	id := r.URL.Path[len("/view/"):]
	db.ConnectDB()
	rows:= db.Db.QueryRow("select * from blogpost where id = ?", id)
	var blog Blog

	if err := rows.Scan(&blog.ID, &blog.Title, &blog.Author, &blog.Description, &blog.Clickbait, &blog.PublishedDate); err != nil {
		log.Fatal("Unable to scan into slice", err)
	}
	
	err := templates.ExecuteTemplate(w, "view.html", blog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
