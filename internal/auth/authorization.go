package auth

//import( "fmt"; "strings")
import(
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"github.com/google/uuid"
	"time"
	"strings"
	"fmt"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)

	return hash,err
}//Hashes a password with	argon2id.CreateHash



func CheckPasswordHash(password, hash string) (bool, error) {

	match, err := argon2id.ComparePasswordAndHash(password,hash)
	return match, err
}//Compares entered password and has with	argon2id.ComparePasswordAndHash


func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)

	claims := jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt:jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject: userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)

	signedToken,err := token.SignedString(signingKey)
	return signedToken,err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	
	claims := jwt.RegisteredClaims{}
	token,err := jwt.ParseWithClaims(
		tokenString, 
		&claims, 
		func(token *jwt.Token) (any, error) {
	return []byte(tokenSecret), nil 
	}) 
	if err != nil {return uuid.Nil,err}


	user_id,err := token.Claims.GetSubject()
	if err != nil {return uuid.Nil,err}


	user_uuid,err := uuid.Parse(user_id) 
	if err != nil {return uuid.Nil,err}

	return user_uuid,nil


}

func GetBearerToken(headers http.Header) (string, error) {
	auth_header,err := headers["Authorization"]
	fmt.Println(auth_header)
	if len(auth_header) == 0 {return "",errors.New("Error: No authorization token")}
	if auth_header[0] == "" || err != true {return "",errors.New("Error: Header does not exist")}
	TOKEN_STRING := strings.Split(auth_header[0]," ")[1]

	return TOKEN_STRING,nil

}



func MakeRefreshToken() (string, error) {

	token := make([]byte, 32)
	rand.Read(token)
	encodedToken := hex.EncodeToString(token)
	return encodedToken,nil

}

func GetAPIKey(headers http.Header) (string, error) {
	auth_header,err := headers["Authorization"]
	fmt.Println(auth_header)
	if len(auth_header) == 0 {return "",errors.New("Error: No API Key")}
	if auth_header[0] == "" || err != true {return "",errors.New("Error: No API Key")}
	api_key := strings.Split(auth_header[0]," ")[1]

	return api_key,nil

}