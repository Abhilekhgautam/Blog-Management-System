package blog

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"projectOne/blog/db"
	"time"
)

type Blog struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Description   string `json:"description"`
	PublishedDate string `json:"published_date"`
	Clickbait     string `json:"clickbait"`
}

var templates = template.Must(template.ParseFiles("./templates/home.html", "./templates/addnew.html",
	"./templates/adminhome.html", "./templates/edit.html", "./templates/delete.html"))

// GetHome - loads all the blog (for client)
func GetHome(w http.ResponseWriter, r *http.Request) {
	db.ConnectDB()
	rows, err := db.Db.Query("select * from blogpost order by id desc")
	if err != nil {
		log.Fatal("Error while retrieving data: ", err)
	}
	defer rows.Close()
	var blogs []Blog
	for rows.Next() {
		var blog Blog
		if err := rows.Scan(&blog.ID, &blog.Title, &blog.Description, &blog.Author, &blog.PublishedDate, &blog.Clickbait); err != nil {
			log.Fatal("Unable to scan into slice", err)
		}
		// add element to blogs slice
		blogs = append(blogs, blog)
	}
	fmt.Println(blogs)
	err = templates.ExecuteTemplate(w, "home.html", blogs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AddBlogs - this is a get request, renders a template with form field to add new blogs.
func AddBlogs(w http.ResponseWriter, r *http.Request) {
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
	err = templates.ExecuteTemplate(w, "addnew.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// PostBlogs - post blogs to the db
func PostBlogs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("accepted a post request")
	err := r.ParseForm()
	if err != nil {
		log.Fatal("Unable to parse Form")
	}
	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	author := r.PostForm.Get("author")
	clickbait := r.PostForm.Get("clickbait")
	publishedDate := time.Now().Format("2006-Jan-02")

	//check if all the input fields are filled
	if len(title) == 0 || len(description) == 0 || len(author) == 0 || len(clickbait) == 0 {
	}
	// check if clickbait is maximum of 255 characters
	if len(clickbait) >= 255 {
	}
	// if everything is validated insert into database
	db.ConnectDB()
	// insert blog to db
	// execute sql query
	_, err = db.Db.Exec("Insert into blogpost ( title, description, author, published_date, clickbait) values(?,?,?,?,?)", title, description, author, publishedDate, clickbait)
	if err != nil {
		log.Fatal("Unable to insert data to database", err)
	} else {
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

// EditBlogs - this is a get request, renders input fields to edit blogs
func EditBlogs(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Trying to get edit blog")
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
	id := r.URL.Path[len("/edit/"):]

	row := db.Db.QueryRow("select * from blogpost where id = ?", id)
	var blog Blog
	if err := row.Scan(&blog.ID, &blog.Title, &blog.Description, &blog.Author, &blog.PublishedDate, &blog.Clickbait); err != nil {
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

// UpdateBlogs - this is a post request, adds the blog to the database.
func UpdateBlogs(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("lemme-explain-cookie")
	if err != nil || cookie == nil {
		http.Redirect(w, r, "/login", 302)
		return
	}
	db.ConnectDB()
	_, err = db.Db.Query("select * from session where password = ?", cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	id := r.URL.Path[len("/update/"):]
	fmt.Println("Update id :", id)
	err = r.ParseForm()
	if err != nil {
		log.Fatal("Unable to parse form data")

	}

	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")
	author := r.PostForm.Get("author")
	clickbait := r.PostForm.Get("clickbait")

	if len(title) == 0 || len(description) == 0 || len(author) == 0 || len(clickbait) == 0 {

		// TODO:set error message of all fields required
	}
	//check if clickbait is at most 255 characters long
	//if len(clickbait) > 255{
	// TODO:set err message
	//	 }
	_, err = db.Db.Exec("Update blogpost set title = ? ,description = ?, clickbait = ?, author = ? where id = ? ", title, description, clickbait, author, id)
	if err != nil {
		log.Fatal("Unable to update with given data")
	} else {
		http.Redirect(w, r, "/admin", 302)
	}
}

//GetAdminHome - this is a get request, loads the admin home page
func GetAdminHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Trying to load admin home page")
	cookie, err := r.Cookie("lemme-explain-cookie")
	if err != nil || cookie == nil {
		fmt.Println("Redirecting cookie might be nil")
		http.Redirect(w, r, "/login", 302)
		return
	}
	db.ConnectDB()
	_, err = db.Db.Query("select * from session where password = ?", cookie.Value)
	// if cookie value does not match, redirect to login..
	if err != nil {
		fmt.Println("redirecting ...")
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		return
	}
	// if everything is fine proceed towards loading the home page
	rows, err := db.Db.Query("select * from blogpost order by id desc")
	if err != nil {
		log.Fatal("Error while retrieving data: ", err)
	}
	defer rows.Close()
	var blogs []Blog
	for rows.Next() {
		var blog Blog
		if err = rows.Scan(&blog.ID, &blog.Title, &blog.Description, &blog.Author, &blog.PublishedDate, &blog.Clickbait); err != nil {
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
	fmt.Println("Getting you to delete home page")
	_, cookierr := r.Cookie("lemme-explain-cookie")
	if cookierr != nil {
		if cookierr == http.ErrNoCookie {
			fmt.Println("No cookie found")
			http.Redirect(w, r, "/login", 302)
			return
		} else {
			log.Fatal("Some cookie error")
		}
	}
	id := r.URL.Path[len("/delete/"):]
	fmt.Println("will render delete file soon")
	err := templates.ExecuteTemplate(w, "delete.html", id)
	if err != nil {
		log.Fatal("Unable to render provided template")
	}
}

// DeleteBlog - this is a delete request, if everything goes well this deletes the blog from db.
func DeleteBlog(w http.ResponseWriter, r *http.Request) {
	_, cookierr := r.Cookie("lemme-explain-cookie")
	if cookierr != nil {
		if cookierr == http.ErrNoCookie {
			fmt.Println("No cookie found")
			http.Redirect(w, r, "/login", 302)
			return
		} else {
			log.Fatal("Some cookie error")
		}
	}
	id := r.URL.Path[len("/delete/"):]
	_, err := db.Db.Exec("delete from blogpost where id = ?", id)
	if err != nil {
		log.Fatal("Unable to delete the blog")
	}
	http.Redirect(w, r, "/admin", 302)
	return
}
