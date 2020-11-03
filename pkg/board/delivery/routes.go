package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/go-park-mail-ru/2020_2_Eternity/internal/app/database"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/auth"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/board/repository"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/board/usecase"
)

func AddBoardRoutes(r *gin.Engine, db database.IDbConn) {
	rep := repository.NewRepo(db)
	uc := usecase.NewUsecase(rep)
	handler := NewHandler(uc)

	r.GET("/board/:id", handler.GetBoard)
	r.GET("/boards/:username", handler.GetAllBoardsbyUser)

	authorized := r.Group("/", auth.AuthCheck())
	{
		authorized.POST("/board", handler.CreateBoard)
	}
}