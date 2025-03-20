package middleware

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/revel/revel"
	"log"
	"os"
	"time"
)

type Middleware struct {
	*revel.Controller
}

var secretKey []byte

func Init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}
	secretKey = []byte(os.Getenv("SECRET_KEY"))
	if secretKey == nil {
		log.Fatal("SECRET_KEY environment variable not set")
		return
	}
}

func GenerateJWT(userID uint64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["userID"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil

}

func ValidateJWT(request *revel.Request, cookieName string) (uint64, error) {
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		return 0, err
	}
	token, err := jwt.Parse(cookie.GetValue(), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error in parsing the token")
		}
		return secretKey, nil
	})
	if err != nil {
		return 0, err
	}
	exp := token.Claims.(jwt.MapClaims)["exp"].(float64)
	UID := uint64(token.Claims.(jwt.MapClaims)["userID"].(float64))

	if time.Until(time.Unix(int64(exp), 0)) < time.Minute*30 {
		GenerateJWT(UID)
	}

	return UID, nil
}
