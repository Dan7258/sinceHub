package controllers

import (
	"crypto/rand"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"math/big"
	"net/http"
	"sinceHub/app/middleware"
	"sinceHub/app/models"
	"sinceHub/app/smtp"
	"time"
)

type Profiles struct {
	*revel.Controller
}

type VerificationInfo struct {
	Code      string
	ExpiresAt time.Time
}

type VerifyProfile struct {
	Profile models.Profiles
	Code    string `json:"code"`
}

var verificationCodes = make(map[string]VerificationInfo)

//func (p Profiles) CreateProfile() revel.Result {
//	profile := new(models.Profiles)
//	err := p.Params.BindJSON(profile)
//	if err != nil {
//		p.Response.Status = http.StatusBadRequest
//		revel.AppLog.Error(err.Error())
//		return p.RenderJSON(map[string]string{"error": err.Error()})
//	}
//	validate := validator.New()
//	err = validate.Struct(profile)
//	if err != nil {
//		p.Response.Status = http.StatusUnprocessableEntity
//		revel.AppLog.Error(err.Error())
//		return p.RenderJSON(map[string]string{"error": err.Error()})
//	}
//
//	err = smtp.SendMessage(profile.Login, "hello")
//	if err != nil {
//		p.Response.Status = http.StatusUnprocessableEntity
//		revel.AppLog.Error(err.Error())
//		return p.RenderJSON(map[string]string{"error": err.Error()})
//	}
//	revel.AppLog.Info("message sent")
//
//	err = models.CreateProfile(profile)
//
//	if err != nil {
//		p.Response.Status = http.StatusInternalServerError
//		revel.AppLog.Error(err.Error())
//		return p.RenderJSON(map[string]string{"error": err.Error()})
//	}
//
//	return p.Redirect("/login")
//}

func (p Profiles) ShowRegisterPage() revel.Result {
	return p.RenderTemplate("register.html")
}

func (p Profiles) generateRandomNumber() (*big.Int, error) {
	max := big.NewInt(100000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		fmt.Println("Ошибка при генерации случайного числа:", err)
		return nil, err
	}
	return randomNumber, nil
}

func (p Profiles) SendVerificationCode() revel.Result {
	profile := new(models.Profiles)
	err := p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	if models.ThsProfilesIsExist(profile.Login) {
		p.Response.Status = http.StatusConflict
		return p.RenderJSON(map[string]string{"error": "Пользователь с таким email уже существует"})
	}
	randomNumber, err := p.generateRandomNumber()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сгенерировать код подтверждения"})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)

	verificationCodes[profile.Login] = VerificationInfo{
		Code:      verificationCode,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	err = smtp.SendMessage(profile.Login, verificationCode)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	revel.AppLog.Info("message sent")
	return p.RenderJSON(map[string]string{"message": "Код подтверждения отправлен"})
}

func (p Profiles) VerifyAndCreateUser() revel.Result {
	var vprofile = new(VerifyProfile)
	err := p.Params.BindJSON(vprofile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": "Неверный запрос"})
	}
	validate := validator.New()
	err = validate.Struct(vprofile.Profile)
	if err != nil || vprofile.Code == "" {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	verificationInfo, ok := verificationCodes[vprofile.Profile.Login]
	if !ok {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": "Код подтверждения не найден. Пожалуйста, запросите новый код."})
	}
	if time.Now().After(verificationInfo.ExpiresAt) {
		delete(verificationCodes, vprofile.Profile.Login)
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Срок действия кода истек. Пожалуйста, запросите новый код."})
	}

	if verificationInfo.Code != vprofile.Code {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Неверный код подтверждения"})
	}
	delete(verificationCodes, vprofile.Profile.Login)

	err = models.CreateProfile(&vprofile.Profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.Redirect("/login")
}

func (p Profiles) showRegisterVerifyPage(profile *models.Profiles, code *big.Int) error {
	err := smtp.SendMessage(profile.Login, "hello")
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		revel.AppLog.Error(err.Error())
		return err
	}
	return nil
}

func (p Profiles) Login(login, password string) revel.Result {
	user, err := models.GetProfileLoginData(login)
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderTemplate("login.html")
	}
	if user.Password != password {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderTemplate("login.html")
	}
	token, err := middleware.GenerateJWT(user.ID)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderText("Ошибка генерации токена")
	}
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	p.SetCookie(cookie)

	p.Response.Status = http.StatusFound
	return p.Redirect("/profile")
}

func (p Profiles) ShowLoginPage() revel.Result {
	return p.RenderTemplate("login.html")
}

func (p Profiles) ShowSettingsPage() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	return p.RenderTemplate("settings.html")
}

func (p Profiles) Logout() revel.Result {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	p.SetCookie(cookie)
	return p.Redirect("/login")
}

func (p Profiles) GetProfileByID(id uint64) revel.Result {
	profile, err := models.GetProfileByID(id)
	if err != nil {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profile)
}
func (p Profiles) ShowUserProfile() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}

	return p.RenderTemplate("profile.html")
}
func (p Profiles) GetUserData() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	profile, _ := models.GetUserProfile(userID)
	fmt.Println(profile)
	return p.RenderJSON(profile)
}

func (p Profiles) DeleteProfileByID(id uint64) revel.Result {
	err := models.DeleteProfileByID(id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) DeleteProfileByLogin(login string) revel.Result {
	err := models.DeleteProfileByLogin(login)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) UpdateProfileByID(id uint64) revel.Result {
	profile := new(models.Profiles)
	err := p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	validate := validator.New()
	err = validate.Struct(profile)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.UpdateProfileByID(id, profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) UpdateProfileByLogin(login string) revel.Result {
	profile := new(models.Profiles)
	err := p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	validate := validator.New()
	err = validate.Struct(profile)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.UpdateProfileByLogin(login, profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) GetAllProfiles() revel.Result {
	profiles, err := models.GetAllProfiles()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profiles)
}

func (p Profiles) AddPublicationsToProfile(id uint64) revel.Result {
	var pubIDs []uint64
	err := p.Params.BindJSON(&pubIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddPublicationsToProfile(id, pubIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) DeletePublicationsFromProfile(id uint64) revel.Result {
	var pubIDs []uint64
	err := p.Params.BindJSON(&pubIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.DeletePublicationsFromProfile(id, pubIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) AddSubscribersToProfile(id uint64) revel.Result {
	var subIDs []uint64
	err := p.Params.BindJSON(&subIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddSubscribersToProfile(id, subIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) DeleteSubscribersFromProfile(id uint64) revel.Result {
	var subIDs []uint64
	err := p.Params.BindJSON(&subIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.DeleteSubscribersFromProfile(id, subIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}
