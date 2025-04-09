package middleware

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
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
		revel.AppLog.Error("SECRET_KEY environment variable not set")
		return
	}
	secretKey = []byte(os.Getenv("SECRET_KEY"))
	if len(secretKey) == 0 {
		revel.AppLog.Error("SECRET_KEY environment variable not set")
		return
	}
}

func GenerateJWT(userID uint64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
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
	userID := uint64(token.Claims.(jwt.MapClaims)["sub"].(float64))

	if time.Until(time.Unix(int64(exp), 0)) < time.Minute*30 {
		GenerateJWT(userID)
	}

	return userID, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
