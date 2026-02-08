package main

import( "net/http"; "sync/atomic"; "fmt"; "encoding/json";"log"; "strings";"github.com/joho/godotenv";"os";"database/sql";"github.com/johannesalke/goserver/internal/database")
import("github.com/google/uuid";"time"; "github.com/johannesalke/goserver/internal/auth")
import _ "github.com/lib/pq" //"github.com/alexedwards/argon2id";


type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	PLATFORM string
	SECRET_STRING string
	POLKA_KEY string
}
type Chirp struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Body     string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}
type User struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Email     string    `json:"email"`
	IsChirpyRed bool	`json:"is_chirpy_red"`
	//HashedPassword string `json:"hashed_password"`
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
	user, err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w,400,fmt.Sprintf("Error: %v",err))
		return
	}
	if cfg.PLATFORM != "dev" {
		respondWithError(w,403,fmt.Sprintf("Error: %v","That action is not permitted on this platform."))
		return
	}
	cfg.fileserverHits.Store(0)
	fmt.Fprintln(w, "Database and page hits reset.")
	fmt.Println(user)
	//return w,r

	//
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
	//fmt.Println(string(dat))
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

func (cfg apiConfig) create_user(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	
	decoder := json.NewDecoder(r.Body)
	
	var params parameters
	err1 := decoder.Decode(&params)
	//chirp_body := chirp.Body
	fmt.Printf("Email: %v \n",params.Email)
	fmt.Printf("Pwd: %v \n",params.Password)
	
	if err1 != nil {
		w.Header().Set("Content-Type", "application/json")		
		w.WriteHeader(400)		
		fmt.Fprintln(w, `{"error":"Something went wrong."}`)
		return
	}
	hash,err2 := auth.HashPassword(params.Password)
	if err2 != nil {
		respondWithError(w,400,fmt.Sprintf("Error: %v",err2))
		return
	}


	
	UserParams := database.CreateUserParams{
		Email: params.Email,
		HashedPassword: hash,
	}

	user, err := cfg.db.CreateUser(r.Context(), UserParams)
	fmt.Println(user.ID)
	if err != nil {

		respondWithError(w,400,fmt.Sprintf("Error: %v",err))
		return
	}

	out := User{
        ID:        user.ID,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        Email:     user.Email,
		IsChirpyRed: user.IsChirpyRed,
		//HashedPassword: user.HashedPassword,
    } //Turn uppercase keys into lowercase ones, so that the test program properly recognizes them.

	respondWithJSON(w, 201,out)

}



func (cfg apiConfig) update_user(w http.ResponseWriter, r *http.Request) {
	type User_Update_Req struct{
		Email string `json:"email"`
		Password string `json:"password"`
		//UserID uuid.UUID `json:"user_id"`
	}
	
	decoder := json.NewDecoder(r.Body)
	
	var update_req User_Update_Req
	err := decoder.Decode(&update_req)
	//chirp_body := chirp.Body
	fmt.Printf("Pwd: %v \n",update_req)
		
	if err != nil {	respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }

	req_token,err := auth.GetBearerToken(r.Header)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }
	
	validated_uuid,err := auth.ValidateJWT(req_token, cfg.SECRET_STRING)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }


	hashed_password,err := auth.HashPassword(update_req.Password)
	if err != nil {	respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }

	type UpdateUserParams struct {
	ID             uuid.UUID
	HashedPassword string
	Email          string
	}

	updateObject := database.UpdateUserParams{ID: validated_uuid,HashedPassword: hashed_password,Email:  update_req.Email}
	

	updatedUser,err := cfg.db.UpdateUser(r.Context(),updateObject)
	if err != nil {	respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }

	out := User{
        ID:        updatedUser.ID,
        CreatedAt: updatedUser.CreatedAt,
        UpdatedAt: updatedUser.UpdatedAt,
        Email:     updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
		//HashedPassword: user.HashedPassword, This is secret information that should stay inside the database!
    }
	respondWithJSON(w,200,out)

}


func (cfg apiConfig) upgrade_user_red(w http.ResponseWriter, r *http.Request){
	
	type User_Update_Req struct{
		Event string `json:"event"`
		Data struct {
			UserID string `json:"user_id"`
		}
		//UserID uuid.UUID `json:"user_id"`
	}
	header_key,err := auth.GetAPIKey(r.Header)
	if err != nil || header_key != cfg.POLKA_KEY {respondWithError(w,401,fmt.Sprintf("Error: Invalid API Key | %v",err)); return }
	
	decoder := json.NewDecoder(r.Body)
	
	var update_req User_Update_Req
	err = decoder.Decode(&update_req)
	//chirp_body := chirp.Body		
	if err != nil {	respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }
	if update_req.Event != "user.upgraded" {	respondWithError(w,204,fmt.Sprintf("Error: %v",err)); return }
	parsed_uuid,_ :=	uuid.Parse(update_req.Data.UserID)

	updated_user,err := cfg.db.MakeUserRed(r.Context(),parsed_uuid)
	if err != nil {	respondWithError(w,404,fmt.Sprintf("Error: %v",err)); return }

	respondWithJSON(w,204,updated_user)
}







func (cfg apiConfig) user_login(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Email string `json:"email"`
		Password string `json:"password"`
		ExpiresInSeconds *int64 `json:"expires_in_seconds"`
	}
	type Logged_In_User struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Email     string    `json:"email"`
	Token	  string  	`json:"token"`
	RefreshToken string `json:"refresh_token"`
	IsChirpyRed bool	`json:"is_chirpy_red"`
	//HashedPassword string `json:"hashed_password"`
	}

	decoder := json.NewDecoder(r.Body)
	
	var req LoginRequest
	err := decoder.Decode(&req)
	
	fmt.Printf("Email: %v \n",req.Email)
	if err != nil { respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }



	user,err := cfg.db.GetUserByEmail(r.Context(),req.Email)
	if err != nil { respondWithError(w,401,"Incorrect email or password"); return }
	
	hash_match, err := auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil || hash_match != true { respondWithError(w,401,"Incorrect email or password"); return }



	//var req LoginRequest


	var expiresSeconds int64
	if req.ExpiresInSeconds == nil || *req.ExpiresInSeconds > 60*60{
		expiresSeconds = 60 * 60
	} else {
    	expiresSeconds = *req.ExpiresInSeconds
	}

	signedToken,err := auth.MakeJWT(user.ID, cfg.SECRET_STRING, time.Duration(expiresSeconds)*time.Second)
	if err != nil { respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }

	refreshToken,err := auth.MakeRefreshToken()
	if err != nil { respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }


	refreshTokenParameters := database.InsertRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	}

	RefreshTokenResult,err := cfg.db.InsertRefreshToken(r.Context(),refreshTokenParameters)
	if err != nil { respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }

	if hash_match == true { 
		out := Logged_In_User{
        ID:        user.ID,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        Email:     user.Email,
		IsChirpyRed: user.IsChirpyRed,
		Token:	   signedToken,
		RefreshToken: RefreshTokenResult.Token,
		
		//HashedPassword: user.HashedPassword,
    	} //Turn uppercase keys into lowercase ones, so that the test program properly recognizes them.

		respondWithJSON(w, 200,out)

	}
}

func (cfg apiConfig) refresh_token(w http.ResponseWriter, r *http.Request) {

	refresh_token,err := auth.GetBearerToken(r.Header)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }

	refreshTokenObject,err := cfg.db.GetRefreshToken(r.Context(), refresh_token)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }
	if refreshTokenObject.ExpiresAt.Before(time.Now()) { respondWithError(w,401,"Error: Token is expired"); return }
	if  refreshTokenObject.RevokedAt.Valid { respondWithError(w,401,"Error: Token was revoked"); return }

	signedToken,err := auth.MakeJWT(refreshTokenObject.UserID, cfg.SECRET_STRING, time.Duration(3600)*time.Second)
	if err != nil { respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }

	response := struct{	Token string `json:"token"`}{ Token: signedToken, }
	respondWithJSON(w,200,response)
}

func (cfg apiConfig) revoke_token(w http.ResponseWriter, r *http.Request) {

	refresh_token,err := auth.GetBearerToken(r.Header)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }

	_,err = cfg.db.RevokeRefreshToken(r.Context(), refresh_token)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }
	
	respondWithJSON(w,204,nil)
}



func (cfg apiConfig) create_chirp(w http.ResponseWriter, r *http.Request) {
	type Chirp_Req struct{
		Body string `json:"body"`
		//UserID uuid.UUID `json:"user_id"`
	}
	
	decoder := json.NewDecoder(r.Body)
	
	var chirp_req Chirp_Req
	err1 := decoder.Decode(&chirp_req)
	//chirp_body := chirp.Body
	
	
	if err1 != nil {
		respondWithError(w,400,`{"error":"Something went wrong."}`)				
		return
	}
	
	

	//Get token from header and validate it.
	req_token,err := auth.GetBearerToken(r.Header)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }

	validated_uuid,err := auth.ValidateJWT(req_token, cfg.SECRET_STRING)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }

	//fmt.Printf("User_ID: %v \n",chirp_req2.UserID)
	fmt.Printf("Validated_ID: %v \n",validated_uuid)
	


	//If the token is valid, create the chirp and return its contents.
	chirp_req2 := database.CreateChirpParams{
		Body: chirp_req.Body,
		UserID: validated_uuid,		
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), chirp_req2)
	//fmt.Println(chirp)

	if err != nil {	respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }


	out := Chirp{
		ID:        chirp.ID,
        CreatedAt: chirp.CreatedAt,
        UpdatedAt: chirp.UpdatedAt,
        Body: chirp.Body,
		UserID: chirp.UserID,
	} //Turn uppercase keys into lowercase ones, so that the test program properly recognizes them.*/

	respondWithJSON(w, 201,out)

}

func (cfg apiConfig) get_chirps(w http.ResponseWriter, r *http.Request) {




	string_id := r.URL.Query().Get("author_id")
	//DBchirps := []database.Chirp{}
	chirps := []Chirp{}


	if string_id != "" {
		uuid_id,err := uuid.Parse(string_id)
		if err != nil {	respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return}
		DBchirps, err := cfg.db.GetChirpsByAuthor(r.Context(),uuid_id)
		if err != nil {	respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }

		
		for _,dbChirp := range DBchirps{
		
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
		}
		
		respondWithJSON(w, http.StatusOK, chirps)


	} else {
		DBchirps, err := cfg.db.GetChirps(r.Context())
		if err != nil {	respondWithError(w,400,fmt.Sprintf("Error: %v",err)); return }
		
		for _,dbChirp := range DBchirps{
		
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
		}
		
		respondWithJSON(w, http.StatusOK, chirps)
	}
	
	
	

}

func (cfg apiConfig) get_single_chirp(w http.ResponseWriter, r *http.Request) {

	path_user_id :=	r.PathValue("chirpID")
	fmt.Printf("PathValue: %v \n",path_user_id)
	user_id,err := uuid.Parse(path_user_id)
	if err != nil {
		respondWithError(w,400,fmt.Sprintf("UUID Parsing error: %v",err))
		return
	}


	dbChirp, err := cfg.db.GetSingleChirp(r.Context(),user_id)
	if err != nil {
		respondWithError(w,404,fmt.Sprintf("Error: %v",err))
		return
	}
	
	chirp := Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
	}
	
		
	respondWithJSON(w, http.StatusOK, chirp)

}

func (cfg apiConfig) delete_chirp(w http.ResponseWriter, r *http.Request) {

	path_chirp_id :=	r.PathValue("chirpID")
	fmt.Printf("PathValue: %v \n",path_chirp_id)
	chirpID,err := uuid.Parse(path_chirp_id)
	if err != nil {
		respondWithError(w,400,fmt.Sprintf("UUID Parsing error: %v",err))
		return
	}


	req_token,err := auth.GetBearerToken(r.Header)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }

	token_uuid,err := auth.ValidateJWT(req_token, cfg.SECRET_STRING)
	if err != nil {	respondWithError(w,401,fmt.Sprintf("Error: %v",err)); return }

	 

	chirp,err := cfg.db.GetSingleChirp(r.Context(),chirpID)
	if err != nil {	respondWithError(w,404,fmt.Sprintf("Error: %v",err)); return }
	chirp_uuid := chirp.UserID

	if chirp_uuid != token_uuid { respondWithError(w,403,fmt.Sprintf("Error: %v","Cannot delete another user's chirps!")); return }

	_, err =cfg.db.DeleteChirp(r.Context(),chirpID)
	if err != nil {	respondWithError(w,403,fmt.Sprintf("Error: %v",err)); return }




	respondWithJSON(w,204,nil)
}



func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	PLATFORM := os.Getenv("PLATFORM")
	SECRET_STRING := os.Getenv("SECRET_STRING")
	POLKA_KEY := os.Getenv("POLKA_KEY")

	mux := http.NewServeMux()
	cfg := apiConfig{fileserverHits: atomic.Int32{}, db: dbQueries, PLATFORM: PLATFORM, SECRET_STRING: SECRET_STRING, POLKA_KEY: POLKA_KEY,}

	server := http.Server{Handler: mux, Addr: ":8080"}
	fs := cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))
	mux.Handle("/app/",http.StripPrefix("/app/", fs))

	


	

	
	mux.HandleFunc("GET /admin/metrics", cfg.metrics)
	mux.HandleFunc("POST  /admin/reset", cfg.metricsReset)
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/validate_chirp", validate_chirp)
	mux.HandleFunc("POST /api/users",cfg.create_user)
	mux.HandleFunc("PUT /api/users",cfg.update_user)
	mux.HandleFunc("POST /api/polka/webhooks",cfg.upgrade_user_red)
	
	

	mux.HandleFunc("POST /api/login",cfg.user_login)
	mux.HandleFunc("POST /api/refresh",cfg.refresh_token)
	mux.HandleFunc("POST /api/revoke",cfg.revoke_token)
	mux.HandleFunc("POST /api/chirps",cfg.create_chirp)
	mux.HandleFunc("GET /api/chirps",cfg.get_chirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}",cfg.get_single_chirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}",cfg.delete_chirp)

	server.ListenAndServe()
	
}


