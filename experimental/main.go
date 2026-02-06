package main

import( "fmt";"net/http" )


func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello, world!</h1>")
}



func main() {
		mux := http.NewServeMux()
	//mux.HandleFunc("/", handler)
	mux.Handle("/",http.FileServer(http.Dir(".")))
		//Server := http.Server{Handler: handler, Addr: ":8080"}
	fs := http.FileServer(http.Dir("../assets/"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
		//http.Handle("/assets/",http.FileServer(http.Dir(".")))

	http.ListenAndServe(":8080",mux)
	
	
}