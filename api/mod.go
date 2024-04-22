// Package api joins  the routes for this example
package api

import (
	"errors"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/TomascpMarques/dropmedical/models"
)

// SetupRoutesGroup groups the API routes into a Engine route group with the path prefix of `/api`
func SetupRoutesGroup(router *gin.Engine, db *gorm.DB) {
	// Logger
	router.Use(gin.LoggerWithFormatter(apiLogger))

	// Recovery from panics inside middlewares and handlers
	router.Use(gin.Recovery())

	// Routes
	api := router.Group("/api")
	// ------------------------
	api.POST("/dropper", func(ctx *gin.Context) { registerDropperPOST(ctx, db) })
	api.POST("/dropper/section", func(ctx *gin.Context) { registerDropperSectionPOST(ctx, db) })
	api.POST("/dropper/section/reload", func(ctx *gin.Context) { reloadDropperSectionPOST(ctx, db) })
	// ------------------------
}

func apiLogger(param gin.LogFormatterParams) string {
	return fmt.Sprintf(
		`[%s] (%s) %s %s %d`,
		param.TimeStamp.UTC(),
		param.ClientIP,
		param.Method,
		param.Path,
		param.StatusCode,
	)
}

func returnMessage(status, message string) gin.H {
	return gin.H{
		"status":  status,
		"message": message,
	}
}

type reloadDropperSection struct {
	Dropper  uint   `form:"dropper_id"`
	Section  uint   `form:"section_pos"`
	PillName string `form:"pill_name"`
	Quantity uint   `form:"pill_quantity"`
}

func reloadDropperSectionPOST(c *gin.Context, db *gorm.DB) {
	var reloadSectionAction reloadDropperSection

	if err := c.ShouldBind(&reloadSectionAction); err != nil {
		log.Printf("Tentativa de recarregar uma secção falhada!: %s \n", err.Error())
		c.JSON(
			400,
			returnMessage(
				"erro",
				"dados fornecidos são invalidos ou mal-formados",
			),
		)
		return
	}

	var dropper models.Dropper
	err := db.Find(&dropper, reloadSectionAction.Dropper).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("O dropper pedido <%d> não existe\n", reloadSectionAction.Dropper)
		c.JSON(
			400,
			returnMessage(
				"erro",
				"dados fornecidos são invalidos ou mal-formados",
			),
		)
		return
	} else if err != nil {
		log.Printf("Erro interno, base de dados: %s\n", err.Error())
		c.JSON(
			500,
			returnMessage(
				"erro",
				"Erro interno tente mais tarde",
			),
		)
		return
	}

	err = dropper.ReloadSection(
		db,
		reloadSectionAction.Section,
		reloadSectionAction.PillName,
		reloadSectionAction.Quantity,
	)
	if err != nil {
		log.Printf("Erro ao recarregar secção, erro: %s\n", err.Error())
		c.JSON(
			500,
			returnMessage(
				"erro",
				"Erro interno tente mais tarde",
			),
		)
		return
	}

	c.JSON(200, gin.H{
		"status": "sucesso",
		"razao":  "Secção carregada",
	})
}

type newDropperSection struct {
	Dropper uint            `form:"dropper"`
	Name    string          `form:"name"`
	Pills   models.PillList `form:"pills"`
}

func registerDropperSectionPOST(c *gin.Context, db *gorm.DB) {
	var newSection newDropperSection

	if err := c.ShouldBindJSON(&newSection); err != nil {
		log.Printf("Tentativa de criar dropper com request-body malformed\n")
		c.JSON(
			400,
			returnMessage(
				"erro",
				"dados inválidos enviados para criar dropper",
			),
		)
		return
	}

	var dropper models.Dropper
	err := db.Find(&dropper, "id", newSection.Dropper).Error
	if err != nil {
		c.JSON(
			400,
			returnMessage(
				"erro",
				err.Error(),
			),
		)
	}

	_, err = dropper.CreateDropperSection(db, newSection.Name, newSection.Pills)
	if err != nil {
		c.JSON(
			400,
			returnMessage(
				"erro",
				err.Error(),
			),
		)
	}

	c.JSON(
		201,
		returnMessage(
			"sucesso",
			"secção criada",
		),
	)
}

type newDropper struct {
	Name       string `form:"name" json:"name" binding:"required"`
	Active     bool   `form:"active" json:"active" binding:"required"`
	MachineUrl string `form:"machine_url" json:"machine_url"`
}

func registerDropperPOST(c *gin.Context, db *gorm.DB) {
	var newDropper newDropper

	if c.ShouldBind(&newDropper) != nil {
		c.JSON(400, gin.H{
			"status": "error",
			"reason": "Bad query parameters",
		})
		return
	}

	dropper := models.NewDropper(newDropper.Name, newDropper.Active, newDropper.MachineUrl)

	err := db.Select("ID").Create(&dropper).Error

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		log.Println("Tentativa de criar um dropper que já existe")
		c.JSON(409, gin.H{
			"status": "erro",
			"razao":  "Este dropper já existe",
		})
		return
	} else if err != nil {
		log.Printf("Erro inesperado: %s", err.Error())
		c.JSON(500, gin.H{
			"status": "erro",
			"razao":  "Ocorreu um erro inesperado",
		})
		return
	}

	c.JSON(200, dropper)
}
