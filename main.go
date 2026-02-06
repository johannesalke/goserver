package main

import( "net/http"; "sync/atomic"; "fmt"; "encoding/json";"log"; "strings")
import _ "github.com/lib/pq"


type apiConfig struct {
	fileserverHits atomic.Int32
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
		
	
	page := fmt.Sprintf(`
	<html>
  	<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  	</body>
	</html>`,cfg.fileserverHits.Load())

	fmt.Fprintln(w,  page) 
	//return w,r
}

func (cfg *apiConfig) metricsReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	fmt.Fprintln(w, "Number of hits reset")
	//return w,r
}


func respondWithError(w http.ResponseWriter, code int, msg string){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)		
	err_msg := fmt.Sprintf(`{"error":"%s"}`,msg)
	fmt.Fprintln(w, err_msg)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(payload)
	dat, err := json.Marshal(payload)
	if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
	}
	fmt.Println(string(dat))
	w.WriteHeader(code)
	fmt.Fprintln(w, string(dat))
	//w.Write(dat) 
}


func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		
	w.WriteHeader(200)

	w.Write([]byte("OK"))
	//fmt.Fprintln(w, "OK") This does the same thing!
}

func validate_chirp(w http.ResponseWriter, r *http.Request) {
	type valid_chirp struct {
		Body string `json:"Body"`
	}

	decoder := json.NewDecoder(r.Body)
	
	var chirp valid_chirp
	err := decoder.Decode(&chirp)
	chirp_body := chirp.Body
	fmt.Printf("Call: %v \n",chirp_body)
	
	if err != nil {
		w.Header().Set("Content-Type", "application/json")		
		w.WriteHeader(400)		
		fmt.Fprintln(w, `{"error":"Something went wrong."}`)
		return
	}

	if len(chirp_body)>140 {
		w.Header().Set("Content-Type", "application/json")
		
		w.WriteHeader(400)

		
		fmt.Fprintln(w, `{"error":"This chirp is too long.}`)
		return
	}

    if true {
	type resp struct{
		//valid bool `json:valid`
		Cleaned_body string `json:"cleaned_body"`
	}

	badwords := [3]string{"kerfuffle","sharbert","fornax"}
	split_chirp := strings.Split(chirp_body," ")
	for i,word := range split_chirp {
		for _,badword := range badwords{
			if strings.ToLower(word) == badword{
				split_chirp[i] = "****"
			}
		}


	}
	cleaned := strings.Join(split_chirp," ")
	fmt.Println(cleaned)

	response := resp{Cleaned_body: cleaned}
	//fmt.Println(response)
	respondWithJSON(w, 200, response)
	/*
	w.Header().Set("Content-Type", "application/json")
		
	w.WriteHeader(200)

		
	fmt.Fprintln(w, `{"valid":true}`)
	*/
	return

	
	}
	


	

	
}


func main() {
	mux := http.NewServeMux()
	cfg := apiConfig{}

	server := http.Server{Handler: mux, Addr: ":8080"}
	fs := cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))
	mux.Handle("/app/",http.StripPrefix("/app/", fs))

	


	

	
	mux.HandleFunc("GET /admin/metrics", cfg.metrics)
	mux.HandleFunc("POST  /admin/reset", cfg.metricsReset)
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/validate_chirp", validate_chirp)

	server.ListenAndServe()
	
}


