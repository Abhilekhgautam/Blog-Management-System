package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"projectOne/blog/models"
	"projectOne/blog/user"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", blog.GetHome)
	r.HandleFunc("/add", blog.AddBlogs)
	r.HandleFunc("/edit/{id}", blog.EditBlogs)
	r.HandleFunc("/admin", blog.GetAdminHome)
	r.HandleFunc("/login", user.GetLogin)
	r.HandleFunc("/logout", user.LogOut)
	r.HandleFunc("/delete/{id}", blog.GetDelete).Methods("GET")

	r.HandleFunc("/update/{id}", blog.UpdateBlogs).Methods("POST")
	r.HandleFunc("/postblog", blog.PostBlogs).Methods("POST")
	r.HandleFunc("/signin", user.Login).Methods("POST")

	r.HandleFunc("/delete/{id}", blog.DeleteBlog).Methods("POST")
	// serve static files
	fs := http.FileServer(http.Dir("./static/css/"))
	r.PathPrefix("/static/css/").Handler(http.StripPrefix("/static/css", fs))
	log.Fatal(http.ListenAndServe(":8080", r))

}
