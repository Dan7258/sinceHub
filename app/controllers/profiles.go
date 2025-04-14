package controllers

import (
	"crypto/rand"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"net/http"
	"scinceHub/app/middleware"
	"scinceHub/app/models"
	"scinceHub/app/smtp"
	"time"
)

type Profiles struct {
	*revel.Controller
}

type VerificationEmail struct {
	Code      string
	ExpiresAt time.Time
}

type ChangeEmail struct {
	Code      string
	ExpiresAt time.Time
}

type ChangePassword struct {
	Code      string
	ExpiresAt time.Time
}

type VerifyProfile struct {
	Profile models.Profiles
	Code    string `json:"code"`
}

var verificationEmailCodes = make(map[string]VerificationEmail)

var changeEmailCodes = make(map[uint64]ChangeEmail)

var changePasswordCodes = make(map[uint64]ChangePassword)

func (p Profiles) GenerateRandomNumber() (*big.Int, error) {
	max := big.NewInt(100000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		fmt.Println("Ошибка при генерации случайного числа:", err)
		return nil, err
	}
	return randomNumber, nil
}

func (p Profiles) SendVerificationCodeForRegister() revel.Result {
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
	randomNumber, err := p.GenerateRandomNumber()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сгенерировать код подтверждения"})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)

	verificationEmailCodes[profile.Login] = VerificationEmail{
		Code:      verificationCode,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	err = smtp.SendMessage(profile.Login, "Подтверждение почты", verificationCode)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	revel.AppLog.Info("message sent")
	return p.RenderJSON(map[string]string{"message": "Код подтверждения отправлен"})
}

func (p Profiles) SendVerificationCodeForChangeEmail() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	profile := new(models.Profiles)
	err = p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	if _, ok := changePasswordCodes[userID]; ok {
		p.Response.Status = http.StatusConflict
		return p.RenderJSON(map[string]string{"error": "Завершите смену пароля!"})
	}

	if models.ThsProfilesIsExist(profile.Login) {
		p.Response.Status = http.StatusConflict
		return p.RenderJSON(map[string]string{"error": "Пользователь с таким email уже существует"})
	}

	randomNumber, err := p.GenerateRandomNumber()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сгенерировать код подтверждения"})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)

	changeEmailCodes[userID] = ChangeEmail{
		Code:      verificationCode,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	err = smtp.SendMessage(profile.Login, "Смена почты", verificationCode)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	revel.AppLog.Info("message sent")
	return p.RenderJSON(map[string]string{"message": "Код смены почты"})
}

func (p Profiles) VerifyAndChangeEmail() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	var vprofile = new(VerifyProfile)
	err = p.Params.BindJSON(vprofile)
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
	changeEmail, ok := changeEmailCodes[userID]
	if !ok {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": "Код подтверждения не найден. Пожалуйста, запросите новый код."})
	}
	if time.Now().After(changeEmail.ExpiresAt) {
		delete(changeEmailCodes, userID)
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Срок действия кода истек. Пожалуйста, запросите новый код."})
	}
	if changeEmail.Code != vprofile.Code {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Неверный код подтверждения"})
	}
	delete(changeEmailCodes, userID)

	err = models.UpdateProfileByID(userID, &vprofile.Profile)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.Redirect("/settings")
}

func (p Profiles) StopChangeEmail() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	var vprofile = new(VerifyProfile)
	err = p.Params.BindJSON(vprofile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": "Неверный запрос"})
	}
	validate := validator.New()
	err = validate.Struct(vprofile.Profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	delete(changeEmailCodes, userID)
	return p.Redirect("/settings")
}

func (p Profiles) SendVerificationCodeForChangePassword() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	profile := new(models.Profiles)
	err = p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	randomNumber, err := p.GenerateRandomNumber()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сгенерировать код подтверждения"})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)

	changePasswordCodes[userID] = ChangePassword{
		Code:      verificationCode,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	err = smtp.SendMessage(profile.Login, "Смена пароля", verificationCode)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	revel.AppLog.Info("message sent")
	return p.RenderJSON(map[string]string{"message": "Код смены пароля"})
}

func (p Profiles) VerifyAndChangePassword() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	var vprofile = new(VerifyProfile)
	err = p.Params.BindJSON(vprofile)
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
	changePassword, ok := changePasswordCodes[userID]
	if !ok {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": "Код подтверждения не найден. Пожалуйста, запросите новый код."})
	}
	if time.Now().After(changePassword.ExpiresAt) {
		delete(changePasswordCodes, userID)
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Срок действия кода истек. Пожалуйста, запросите новый код."})
	}
	if changePassword.Code != vprofile.Code {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Неверный код подтверждения"})
	}
	delete(changePasswordCodes, userID)
	hashPassword, err := middleware.HashPassword(vprofile.Profile.Password)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	vprofile.Profile.Password = hashPassword
	err = models.UpdateProfileByID(userID, &vprofile.Profile)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.Redirect("/settings")
}

func (p Profiles) StopChangePassword() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	var vprofile = new(VerifyProfile)
	err = p.Params.BindJSON(vprofile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": "Неверный запрос"})
	}
	validate := validator.New()
	err = validate.Struct(vprofile.Profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	delete(changePasswordCodes, userID)
	return p.Redirect("/settings")
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
	VerificationEmail, ok := verificationEmailCodes[vprofile.Profile.Login]
	if !ok {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": "Код подтверждения не найден. Пожалуйста, запросите новый код."})
	}
	if time.Now().After(VerificationEmail.ExpiresAt) {
		delete(verificationEmailCodes, vprofile.Profile.Login)
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Срок действия кода истек. Пожалуйста, запросите новый код."})
	}

	if VerificationEmail.Code != vprofile.Code {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Неверный код подтверждения"})
	}
	delete(verificationEmailCodes, vprofile.Profile.Login)
	hashPassword, err := middleware.HashPassword(vprofile.Profile.Password)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	vprofile.Profile.Password = hashPassword
	err = models.CreateProfile(&vprofile.Profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.Redirect("/login")
}

func (p Profiles) Login(login, password string) revel.Result {
	user, err := models.GetProfileLoginData(login)
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderTemplate("login.html")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
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
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	profile, err := models.GetProfileByID(id)
	if err != nil {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	profileWithSubscribitionStatus := new(models.ProfileWithSubscribitionStatus)
	profileWithSubscribitionStatus.Profile = *profile
	profileWithSubscribitionStatus.Isubscribed = models.CheckMySubscribesForProfile(userID, id)
	profileWithSubscribitionStatus.IsSubscribed = models.CheckMySubscribesForProfile(id, userID)
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profileWithSubscribitionStatus)
}

func (p Profiles) GetUserData() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	profile, _ := models.GetUserProfile(userID)

	return p.RenderJSON(profile)
}

func (p Profiles) GetUsersDataForCreatePublication() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	profile, _ := models.GetAllProfileIDAndNames()

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

func (p Profiles) UpdateProfileByID() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	profile := new(models.Profiles)
	err = p.Params.BindJSON(profile)
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

	err = models.UpdateProfileByID(userID, profile)

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
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	profiles, err := models.GetAllProfiles()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profiles)
}

func (p Profiles) GetMySubscribersList() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	profiles, err := models.GetMySubscribersList(userID)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profiles)
}

func (p Profiles) GetMySubscribesList() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	profiles, err := models.GetMySubscribesList(userID)
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

func (p Profiles) AddSubscriberToProfile(id uint64) revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}

	err = models.AddSubscriberToProfile(userID, id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) DeleteSubscriberFromProfile(id uint64) revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}

	err = models.DeleteSubscriberFromProfile(userID, id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}
