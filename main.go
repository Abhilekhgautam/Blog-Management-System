package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"projectOne/blog/models"
	"projectOne/blog/user"
	"projectOne/blog/Services"
)

func main() {

	r := mux.NewRouter()
    r.HandleFunc("/",blog.GetHome)
	r.HandleFunc("/view/{id}", blog.ViewBlog)
	r.HandleFunc("/{title}/admin/add", blog.AddBlogs)
	r.HandleFunc("/{title}/admin/edit/{id}", blog.EditBlogs)
	r.HandleFunc("/{title}/admin/", blog.GetAdminHome)
	r.HandleFunc("/signup", user.GetSignup)
	r.HandleFunc("/create", user.Signup).Methods("POST")
	r.HandleFunc("/login", user.GetLogin)
	r.HandleFunc("/logout", user.LogOut)
	r.HandleFunc("/forgotpassword", user.GetForgot)
	r.HandleFunc("/forgot", user.ForgotPassword)

	r.HandleFunc("/{title}/admin/delete/{id}", blog.GetDelete).Methods("GET")
	r.HandleFunc("/confirm_email",mail.ConfirmEmail).Queries("token","{name}")
	r.HandleFunc("/forgot_password",user.GetChangePassword).Queries("token","{name}")
	r.HandleFunc("/changepassword",user.ChangePassword).Queries("token","{name}")

	r.HandleFunc("/updatetitle", user.SetTitle).Methods("POST")

	r.HandleFunc("/update/{id}", blog.UpdateBlogs).Methods("POST")
	r.HandleFunc("/postblog", blog.PostBlogs).Methods("POST")
	r.HandleFunc("/signin", user.Login).Methods("POST")
	r.HandleFunc("/{title}/", blog.GetBlogHome)


	r.HandleFunc("/delete/{id}", blog.DeleteBlog).Methods("POST")
	// serve static files
	fs := http.FileServer(http.Dir("./static/css/"))
	r.PathPrefix("/static/css/").Handler(http.StripPrefix("/static/css", fs))

	log.Fatal(http.ListenAndServe(":8080", r))

}
