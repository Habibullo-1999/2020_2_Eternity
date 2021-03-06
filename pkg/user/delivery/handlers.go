package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/go-park-mail-ru/2020_2_Eternity/api"
	"github.com/go-park-mail-ru/2020_2_Eternity/configs/config"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/domain"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/jwthelper"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/user"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/utils"
	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Handler struct {
	uc user.IUsecase
	p  *bluemonday.Policy
}

func NewHandler(uc user.IUsecase, p *bluemonday.Policy) *Handler {
	return &Handler{
		uc: uc,
		p:  p,
	}
}

func (h *Handler) SignUp(c *gin.Context) {
	profile := api.SignUp{}
	if err := c.BindJSON(&profile); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "[BindJSON]: " + err.Error()})
		return
	}
	if err := utils.ValidProfile(profile); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "[ValidProfile]: " + err.Error()})
		return
	}

	profile.Description = h.p.Sanitize(profile.Description)

	hash, err := bcrypt.GenerateFromPassword([]byte(profile.Password), config.Conf.Token.Value)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: "[GenerateFromPassword]: " + err.Error()})
		return
	}

	profile.Password = string(hash)

	u, err := h.uc.CreateUser(&profile)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, utils.Error{Error: "[CreateUser]: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) Login(c *gin.Context) {
	form := api.Login{}

	if err := c.BindJSON(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	u, err := h.uc.GetUserByNameWithFollowers(form.Username)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "invalid username"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(form.Password)); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "bad password"})
		return
	}

	ss, err := jwthelper.CreateJwtToken(u.ID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: "cannot create token"})
		return
	}
	sr, err := utils.RandomUuid()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: "cannot generate value"})
		return
	}
	t, err := jwthelper.CreateCsrfToken(u.ID, sr, time.Now().Add(45*time.Minute))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: "cannot create csrf token"})
		return
	}

	cookie := http.Cookie{
		Name:     config.Conf.Token.CookieName,
		Value:    ss,
		Expires:  time.Now().Add(45 * time.Minute),
		HttpOnly: true,
		Path:     "/",
	}
	c.Header("X-CSRF-TOKEN", t)
	http.SetCookie(c.Writer, &cookie)
	c.JSON(http.StatusOK, u)
}

func (h *Handler) Logout(c *gin.Context) {
	ss, err := c.Cookie(config.Conf.Token.CookieName)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cookie := http.Cookie{
		Name:     config.Conf.Token.CookieName,
		Value:    ss,
		Expires:  time.Now().Add(-24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(c.Writer, &cookie)
	c.JSON(http.StatusOK, "success")
}

func (h *Handler) UpdateUser(c *gin.Context) {
	claimsId, ok := jwthelper.GetClaims(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.Error{Error: "invalid token"})
		return
	}

	profile := api.UpdateUser{}
	if err := c.BindJSON(&profile); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	if err := utils.ValidUpdate(profile); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	profile.Description = h.p.Sanitize(profile.Description)

	u, err := h.uc.UpdateUser(claimsId, &profile)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, utils.Error{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handler) UpdatePassword(c *gin.Context) {
	claimsId, ok := jwthelper.GetClaims(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.Error{Error: "invalid token"})
		return
	}

	psswds := api.UpdatePassword{}
	if err := c.BindJSON(&psswds); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	if err := utils.ValidPasswords(psswds); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	u, err := h.uc.GetUser(claimsId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "user doesnt exist"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(psswds.OldPassword)); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "bad password"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(psswds.NewPassword), config.Conf.Token.Value)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	if err := h.uc.UpdatePassword(claimsId, string(hash)); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) GetProfile(c *gin.Context) {
	claimsId, ok := jwthelper.GetClaims(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.Error{Error: "can't get key"})
		return
	}

	u, err := h.uc.GetUser(claimsId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "user not found"})
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handler) SetAvatar(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "Form error"})
		return
	}

	claimsId, ok := jwthelper.GetClaims(c)

	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.Error{Error: "invalid token"})
		return
	}

	root, err := os.Getwd()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: "server env error"})
		return
	}

	filename, err := utils.RandomUuid()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: "Random cant generate UUID"})
		return
	}

	path := root + config.Conf.Web.Static.DirAvt + filename

	if err := c.SaveUploadedFile(file, path); err != nil {
		config.Lg("handlers", "SetAvatar").Error(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: "Upload error"})
		return
	}

	if err := h.uc.UpdateAvatar(claimsId, utils.GenerateUrlAvatar(filename)); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
		return
	}

	c.JSON(http.StatusOK, "success")
}

func (h *Handler) GetAvatar(c *gin.Context) {
	root, err := os.Getwd()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: "server env error"})
		return
	}
	filename := c.Param("file")

	path := root + config.Conf.Web.Static.DirAvt + filename

	data, err := ioutil.ReadFile(path)
	if err != nil {
		config.Lg("user", "Get Avatar").Error(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "Error filename"})
		return
	}
	c.Header("Content-Type", "image/jpeg")
	c.Header("Content-Length", strconv.Itoa(len(data)))
	_, err = c.Writer.Write(data)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, utils.Error{Error: "Error write to response"})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) Follow(c *gin.Context) {
	followingUser, followedUser, status, err := h.prepareFollow(c)
	if err != nil {
		c.AbortWithStatusJSON(status, err)
		return
	}
	if followingUser == followedUser {
		c.AbortWithStatusJSON(status, utils.Error{Error: "Cannot self follow"})
		return
	}
	if err := h.uc.Follow(followingUser, followedUser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "Already Followed"})
		return
	}
	c.Status(http.StatusOK)
	c.Set("status", 200)
	c.Set("follow_id", followedUser)

	c.Set(domain.NotificationKey, domain.NoteFollow{
		FollowerId: followingUser,
		UserId:     followedUser,
	})
}

func (h *Handler) Unfollow(c *gin.Context) {
	followingUser, followedUser, status, err := h.prepareFollow(c)
	if err != nil {
		c.AbortWithStatusJSON(status, err)
		return
	}
	if err := h.uc.UnFollow(followingUser, followedUser); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "Already Followed"})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) prepareFollow(c *gin.Context) (int, int, int, *utils.Error) {
	claimsId, ok := jwthelper.GetClaims(c)

	if !ok {
		return -1, -1, http.StatusUnauthorized, &utils.Error{Error: "invalid token"}
	}

	followed := api.UserAct{}
	if err := c.BindJSON(&followed); err != nil {
		return -1, -1, http.StatusBadRequest, &utils.Error{Error: "bad json"}
	}

	if err := utils.ValidUsername(followed); err != nil {
		return -1, -1, http.StatusBadRequest, &utils.Error{Error: "invalid username"}
	}

	u, err := h.uc.GetUserByName(followed.Username)
	if err != nil {
		return -1, -1, http.StatusBadRequest, &utils.Error{Error: "User not found"}
	}

	return claimsId, u.ID, http.StatusOK, nil
}

func (h *Handler) GetUserPage(c *gin.Context) {
	u, err := h.uc.GetUserByNameWithFollowers(h.p.Sanitize(c.Param("username")))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "not found user"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) GetFollowers(c *gin.Context) {
	username := h.p.Sanitize(c.Param("username"))
	users, err := h.uc.GetFollowers(username)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "Cannot show followers"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetFollowing(c *gin.Context) {
	username := h.p.Sanitize(c.Param("username"))
	users, err := h.uc.GetFollowing(username)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.Error{Error: "Cannot show following"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) IsFollowing(c *gin.Context) {
	claimsId, ok := jwthelper.GetClaims(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, utils.Error{Error: "can't get key"})
		return
	}
	username := h.p.Sanitize(c.Param("username"))

	f := api.IsFollowing{
		Following: false,
	}

	if err := h.uc.IsFollowing(claimsId, username); err != nil {
		c.JSON(http.StatusOK, f)
		return
	}
	f.Following = true
	c.JSON(http.StatusOK, f)
}
