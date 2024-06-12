package main


import(
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"math/rand"
	"strconv"
	"github.com/gorilla/mux"

)

type Movie struct {
	ID string `json:"id"`
	Isbn string `json:"isbn"`
	Title string `json:"title"`
	Director *Director `json:"director"`


}

type Director struct {
	Firstname string `json:"firstname"`
	Lastname string `json:"Lastname"`
}

var movies []Movie

func main(){
	// create a new router
	r := mux.NewRouter() 
	// hard coded movies  data
	movies =append(	movies,Movie(ID:"1",Isbn:"448743",Title:"Movie One",Director:&Director{Firstname:"John",Lastname:"Doe"}))
	movies =append(movies,Movie(ID:"2",Isbn:"448744",Title:"Movie Two",Director:&Director{Firstname:"Steve",Lastname:"Smith"}))
	movies =append(movies,Movie(ID:"3",Isbn:"448745",Title:"Movie Three",Director:&Director{Firstname:"Jane",Lastname:"Doe"}))
	movies =append(movies,Movie(ID:"4",Isbn:"448746",Title:"Movie Four",Director:&Director{Firstname:"Mike",Lastname:"Smith"}))


	// get movies funtion 
	func getMovies(w http.ResponseWriter, r *http.Request){
		w.Hader().Set("Content-Type","application/json"
		json.NewEncoder(w).Encode(movies)
	}

	// delete movie function
	func deleteMovie(w http.ResponseWriter, r *http.Request){
		w.Hader().Set("Content-Type","application/json"
		params :=mux.Vars(r)
		// loop through movies and find the movie to delete
		for index, item := range movies{

			if item.ID == params["id"]{
				movies = append(movies[:index],movies[index+1:]...) // remove the movie from the slice
			}
		}
	}


	// route handles 
	r.HandleFunc("/movies", getMovies).Methods("GET")
	r.HandleFunc("/movies/{id}",getMovies).Methods("GET")
	r.HandleFunc("/movies",createMovie).Method("POST")
	r.HandleFunc("/movies/{id}",updateMovie).Methods("PUT")
	r.HandleFunc*("/movies(id)",deleteMovie).Methods("DELETE")


	// start the server

	fmt.Printf("Server is running on port 8000\n")
	Log.fatal(http.ListenAndServe(":8000",r))

}
