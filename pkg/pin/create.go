package pin

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-park-mail-ru/2020_2_Eternity/api"
	"github.com/go-park-mail-ru/2020_2_Eternity/configs/config"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/jwthelper"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/user"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func generateRelPath(filename string) string {
	fn := []rune(filename)

	depth := config.Conf.Web.Static.DirDepth
	dirNameLen := config.Conf.Web.Static.DirNameLength

	if depth*dirNameLen > len(fn) {
		depth = 2
		dirNameLen = 1
	}

	if depth <= 0 {
		depth = 2
	}

	if dirNameLen <= 0 {
		dirNameLen = 1
	}

	dirs := []string{}
	for i := 0; i < depth; i += dirNameLen {
		dirs = append(dirs, string(fn[i:i+dirNameLen]))
	}

	res := strings.Join(dirs, "/") + "/" + filename
	log.Print("[generateRelPath]: ", res)
	return res
}

func prepareFileStorage() (relPath string, err error) {
	u, err := utils.RandomUuid()
	if err != nil {
		return "", err
	}

	relPath = generateRelPath(u)
	path := config.Conf.Web.Static.DirImg + "/" + relPath
	err = os.MkdirAll(filepath.Dir(path), os.ModePerm|os.ModeDir)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

func CreatePin(c *gin.Context) {
	claims, ok := c.Get("info")
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, user.Error{"can't get key"})
		return
	}

	requester, ok := claims.(jwthelper.Claims)
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, user.Error{"can't lead claims"})
		return
	}

	file, err := c.FormFile("img") // config
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, user.Error{"[FormFile] :" + err.Error()})
		return
	}

	relPath, err := prepareFileStorage()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, user.Error{"[prepareFileStorage]: " + err.Error()})
		return
	}

	if err := c.SaveUploadedFile(file, config.Conf.Web.Static.DirImg+"/"+relPath); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, user.Error{"[SaveUploadedFile]: " + err.Error()})
		return
	}

	pinApi := api.CreatePinApi{}
	jsonStr := c.PostForm("data") // config
	log.Print(jsonStr)

	if err := json.Unmarshal([]byte(jsonStr), &pinApi); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, user.Error{"[Unmarshal]: " + err.Error()})
		return
	}

	pin := Pin{
		Title:   pinApi.Title,
		Content: pinApi.Content,
		ImgLink: config.Conf.Web.Static.UrlImg + "/" + relPath,
		UserId:  requester.Id,
	}

	if err := pin.CreatePin(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, user.Error{"[CreatePin]: " + err.Error()})
		return
	}

	log.Printf("pin{%v %v %v %v %v}", pin.Id, pin.Title, pin.Content, pin.ImgLink, pin.UserId)

	c.JSON(http.StatusOK, "")
}
