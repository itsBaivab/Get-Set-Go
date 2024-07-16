package main

import (
	"fmt"
	"log"
	"net/http"
)

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	fmt.Fprintf(w, "POST request successful\n")
	name := r.FormValue("name")
	email := r.FormValue("email")
	message := r.FormValue("message")
	fmt.Fprintf(w, "Name = %s\n", name)
	fmt.Fprintf(w, "Email = %s\n", email)
	fmt.Fprintf(w, "Message = %s\n", message)
}
func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	fileserver := http.FileServer(http.Dir("./static")) // autometically pics up the index.html file
	http.Handle("/", fileserver)
	http.Handle("/form", http.HandlerFunc(formHandler))
	http.Handle("/hello",http.HandlerFunc( helloHandler))

	fmt.Println("Server is running on port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}// nil is the default server mux // this will create the server

}
