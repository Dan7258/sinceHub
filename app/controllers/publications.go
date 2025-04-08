package controllers

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"io"
	"net/http"
	"os"
	"sinceHub/app/middleware"
	"sinceHub/app/models"
	"strconv"
	"strings"
	"time"
)

type Publications struct {
	*revel.Controller
}

type DeleteAuthorFromPublication struct {
	IDPublication uint64 `json:"id_publication"`
	IDAuthor      uint64 `json:"id_author"`
}

func (p Publications) CreatePublication() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	pub := new(models.Publications)
	pub.Title = p.Params.Get("title")
	pub.Abstract = p.Params.Get("abstract")
	pub.OwnerID = userID
	validate := validator.New()
	err = validate.Struct(pub)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	fileHeader, ok := p.Params.Files["file"]
	if !ok || len(fileHeader) == 0 {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": "No file uploaded"})
	}

	file, err := fileHeader[0].Open()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось открыть файл"})
	}
	defer file.Close()

	randomNumber, _ := Profiles{}.GenerateRandomNumber()
	filePath := fmt.Sprintf("public/uploads/%d_%s_%s", userID, randomNumber, fileHeader[0].Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сохранить файл"})
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Ошибка при сохранении файла"})
	}
	pub.FileLink = filePath

	unique := make(map[uint64]interface{}, 0)
	rawTagIDs := p.Params.Values["tags[]"]
	tagIDs := make([]uint64, 0)
	for _, ID := range rawTagIDs {
		tagID, _ := strconv.ParseUint(ID, 10, 64)
		_, ok := unique[tagID]
		if !ok {
			unique[tagID] = nil
			tagIDs = append(tagIDs, tagID)
		}
	}
	rawCoauthors := p.Params.Values["coauthors[]"]
	unique = make(map[uint64]interface{}, 0)
	coauthorIDs := make([]uint64, 0)
	coauthorIDs = append(coauthorIDs, userID)
	unique[userID] = nil
	for _, ID := range rawCoauthors {
		coauthorID, _ := strconv.ParseUint(ID, 10, 64)
		_, ok := unique[coauthorID]
		if !ok {
			unique[coauthorID] = nil
			coauthorIDs = append(coauthorIDs, coauthorID)
		}
	}
	err = models.CreatePublication(pub, tagIDs, coauthorIDs)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	//p.Response.Status = http.StatusCreated

	return p.Redirect("/profile")
}

//func fillSliceUniqueIDs(data []string, res []uint64) revel.Result {
//	unique := make(map[uint64]interface{}, 0)
//
//}

func (p Publications) DeleteAuthorFromPublication() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	dafp := new(DeleteAuthorFromPublication)
	p.Params.BindJSON(&dafp)
	fmt.Println(dafp)

	err = models.DeleteProfileFromPublication(dafp.IDPublication, dafp.IDAuthor)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) GetPublicationData(id uint64) revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	pub, err := models.GetPublicationByID(id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.RenderJSON(pub)
}

func (p Publications) DeletePublication() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	pub := new(models.Publications)
	err = p.Params.BindJSON(pub)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	if pub.OwnerID != userID {
		p.Response.Status = http.StatusForbidden
		return p.RenderJSON(map[string]string{"error": "Вы не можете редактировать чужие публикации!"})
	}

	err = models.DeletePublicationByID(pub.ID)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = os.Remove(pub.FileLink)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) UpdatePublication() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	pub := new(models.Publications)
	pubID := p.Params.Get("publication_id")
	pub.ID, err = strconv.ParseUint(pubID, 10, 64)
	pub.Title = p.Params.Get("title")
	pub.Abstract = p.Params.Get("abstract")
	ownerIDStr := p.Params.Get("owner_id")
	pub.OwnerID, err = strconv.ParseUint(ownerIDStr, 10, 64)
	pub.FileLink = p.Params.Get("fileLink")

	validate := validator.New()
	err = validate.Struct(pub)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	if pub.OwnerID != userID {
		p.Response.Status = http.StatusForbidden
		return p.RenderJSON(map[string]string{"error": "Вы не можете редактировать чужие публикации!"})
	}

	fileHeader, ok := p.Params.Files["file"]
	if ok && len(fileHeader) != 0 {

		err = os.Remove(pub.FileLink)
		if err != nil {
			p.Response.Status = http.StatusInternalServerError
			return p.RenderJSON(map[string]string{"error": err.Error()})
		}
		file, err := fileHeader[0].Open()
		if err != nil {
			p.Response.Status = http.StatusInternalServerError
			return p.RenderJSON(map[string]string{"error": "Не удалось открыть файл"})
		}
		defer file.Close()
		randomNumber, _ := Profiles{}.GenerateRandomNumber()
		filePath := fmt.Sprintf("public/uploads/%d_%s_%s", userID, randomNumber, fileHeader[0].Filename)
		dst, err := os.Create(filePath)
		if err != nil {
			p.Response.Status = http.StatusInternalServerError
			return p.RenderJSON(map[string]string{"error": "Не удалось сохранить файл"})
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			p.Response.Status = http.StatusInternalServerError
			return p.RenderJSON(map[string]string{"error": "Ошибка при сохранении файла"})
		}
		pub.FileLink = filePath
	}

	unique := make(map[uint64]interface{}, 0)
	rawTagIDs := p.Params.Values["tags[]"]
	tagIDs := make([]uint64, 0)
	for _, ID := range rawTagIDs {
		tagID, _ := strconv.ParseUint(ID, 10, 64)
		_, ok := unique[tagID]
		if !ok {
			unique[tagID] = nil
			tagIDs = append(tagIDs, tagID)
		}
	}
	rawCoauthors := p.Params.Values["coauthors[]"]
	unique = make(map[uint64]interface{}, 0)
	coauthorIDs := make([]uint64, 0)
	coauthorIDs = append(coauthorIDs, userID)
	unique[userID] = nil
	for _, ID := range rawCoauthors {
		coauthorID, _ := strconv.ParseUint(ID, 10, 64)
		_, ok := unique[coauthorID]
		if !ok {
			unique[coauthorID] = nil
			coauthorIDs = append(coauthorIDs, coauthorID)
		}
	}
	err = models.UpdatePublication(pub, tagIDs, coauthorIDs)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) GetPublicationsData() revel.Result {
	pub, err := models.GetAllPublications()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(pub)
}

func (p Publications) AddTagsToPublication(id uint64) revel.Result {
	var tagIDs []uint64
	err := p.Params.BindJSON(&tagIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddTagsToPublication(id, tagIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) DeleteTagsFromPublication(id uint64) revel.Result {
	var tagIDs []uint64
	err := p.Params.BindJSON(&tagIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.DeleteTagsFromPublication(id, tagIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) AddProfilesToPublication(id uint64) revel.Result {
	var profileIDs []uint64
	err := p.Params.BindJSON(&profileIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddProfilesToPublication(id, profileIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) GetFileWithPublicationList() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	filters := new(models.PublicationFiltres)
	err = p.Params.BindJSON(filters)
	if err != nil {
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	filename, err := GetFileWithPublicationList(userID, *filters)
	if err != nil {
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	file, err := os.Open(filename)
	if err != nil {
		return p.RenderJSON(map[string]string{"error": "Ошибка при открытии файла"})
	}
	defer file.Close()

	writer := p.Response.GetWriter()
	httpWriter := writer.(http.ResponseWriter)

	array := strings.Split(filename, "/")
	name := array[len(array)-1]
	fmt.Println(name)
	httpWriter.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	httpWriter.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")

	_, err = io.Copy(httpWriter, file)
	if err != nil {
		return p.RenderJSON(map[string]string{"error": "Ошибка при отправке файла"})
	}

	go func() {
		time.Sleep(2 * time.Second)
		_ = os.Remove(filename)
	}()

	return nil
}
