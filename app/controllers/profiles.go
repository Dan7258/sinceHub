package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"scinceHub/app/middleware"
	"scinceHub/app/models"
	"scinceHub/app/smtp"
	"time"
)

type Profiles struct {
	*revel.Controller
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
	randomNumber, err := GenerateRandomNumber()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сгенерировать код подтверждения"})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)
	SetVerificationEmailCode(profile.Login, verificationCode, 5*time.Minute)

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
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	profile := new(models.Profiles)
	err = p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	_, ok := GetChangePasswordCode(sUserID)
	if ok {
		p.Response.Status = http.StatusConflict
		return p.RenderJSON(map[string]string{"error": "Завершите смену пароля!"})
	}

	if models.ThsProfilesIsExist(profile.Login) {
		p.Response.Status = http.StatusConflict
		return p.RenderJSON(map[string]string{"error": "Пользователь с таким email уже существует"})
	}

	randomNumber, err := GenerateRandomNumber()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сгенерировать код подтверждения"})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)
	SetChangeEmailCode(sUserID, verificationCode, 5*time.Minute)

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
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	var vprofile = new(models.VerifyProfile)
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
	changeEmailCode, ok := GetChangeEmailCode(sUserID)
	if !ok {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": "Код подтверждения не найден или истек его срок. Пожалуйста, запросите новый код."})
	}

	if changeEmailCode != vprofile.Code {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Неверный код подтверждения"})
	}

	err = models.UpdateProfileByID(userID, &vprofile.Profile)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	DeleteChangeEmailCode(sUserID)
	_ = models.DeleteDataFromRedis(sUserID)
	return p.Redirect("/settings")
}

func (p Profiles) StopChangeEmail() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	DeleteChangeEmailCode(sUserID)
	return p.Redirect("/settings")
}

func (p Profiles) SendVerificationCodeForChangePassword() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	profile := new(models.Profiles)
	err = p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		revel.AppLog.Error(err.Error())
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	randomNumber, err := GenerateRandomNumber()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сгенерировать код подтверждения"})
	}

	verificationCode := fmt.Sprintf("%06d", randomNumber)
	SetChangePasswordCode(sUserID, verificationCode, 5*time.Minute)

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
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	var vprofile = new(models.VerifyProfile)
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
	changePasswordCode, ok := GetChangePasswordCode(sUserID)
	if !ok {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": "Код подтверждения не найден или истек его срок. Пожалуйста, запросите новый код."})
	}
	if changePasswordCode != vprofile.Code {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Неверный код подтверждения"})
	}
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
	DeleteChangePasswordCode(sUserID)
	_ = models.DeleteDataFromRedis(sUserID)
	return p.Redirect("/settings")
}

func (p Profiles) StopChangePassword() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)

	DeleteChangePasswordCode(sUserID)
	return p.Redirect("/settings")
}

func (p Profiles) VerifyAndCreateUser() revel.Result {
	var vprofile = new(models.VerifyProfile)
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
	VerificationEmailCode, ok := GetVerificationEmailCode(vprofile.Profile.Login)
	if !ok {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": "Код подтверждения не найден или не действителен. Пожалуйста, запросите новый код."})
	}

	if VerificationEmailCode != vprofile.Code {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderJSON(map[string]string{"error": "Неверный код подтверждения"})
	}

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
	middleware.SetCookieData(p.Controller, "auth_token", token, false)

	p.Response.Status = http.StatusFound
	return p.Redirect("/profile")
}

func (p Profiles) Logout() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err == nil {
		sUserID := fmt.Sprintf("%d", userID)
		_ = models.DeleteDataFromRedis(sUserID)
	}
	middleware.SetCookieData(p.Controller, "auth_token", "", true)
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
	profile := new(models.Profiles)
	sUserID := fmt.Sprintf("%d", userID)
	data, err := models.GetDataFromRedis(sUserID)
	if data != nil && err == nil {
		err = json.Unmarshal(data, profile)
		if err == nil {
			revel.AppLog.Debug("Данные профиля получили с redis")
			return p.RenderJSON(profile)
		}
	}
	profile, err = models.GetUserProfile(userID)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	data, err = json.Marshal(profile)
	if err == nil {
		_ = models.SetDataInRedis(sUserID, data, time.Hour)
	}

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

func (p Profiles) DeleteProfileByID() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	err = models.DeleteProfileByID(userID)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	middleware.SetCookieData(p.Controller, "auth_token", "", true)
	_ = models.DeleteDataFromRedis(sUserID)
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) DeleteProfileByLogin(login string) revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	err = models.DeleteProfileByLogin(login)
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
	sUserID := fmt.Sprintf("%d", userID)
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
	_ = models.DeleteDataFromRedis(sUserID)
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

func (p Profiles) AddSubscriberToProfile(id uint64) revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	err = models.AddSubscriberToProfile(userID, id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	_ = models.DeleteDataFromRedis(sUserID)
	_ = models.DeleteDataFromRedis(fmt.Sprintf("%d", id))
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) DeleteSubscriberFromProfile(id uint64) revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	sUserID := fmt.Sprintf("%d", userID)
	err = models.DeleteSubscriberFromProfile(userID, id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	_ = models.DeleteDataFromRedis(sUserID)
	_ = models.DeleteDataFromRedis(fmt.Sprintf("%d", id))
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}
