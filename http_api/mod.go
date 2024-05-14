// Package http_api joins  the routes for this example
package http_api

import (
	"errors"
	"fmt"
	"log"
	"time"

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
	api.POST("/dropper/schedule", func(ctx *gin.Context) { createDropperDispenseSchedulePOST(ctx, db) })

	api.GET("/dropper/section/pills", func(ctx *gin.Context) { dropperSectionPillsGET(ctx, db) })
	api.GET("/dropper/activate", func(ctx *gin.Context) { activateDropperGET(ctx, db) })
	api.GET("/dropper/dispense", func(ctx *gin.Context) { dropperDispensePillsGET(ctx, db) })
	// ------------------------

	health_check := router.Group("/health_check")
	// ------------------------
	health_check.GET("/up", func(ctx *gin.Context) { ctx.Status(200) })
	// ------------------------
}

type getSectionPillsQuery struct {
	Dropper uint `form:"dropper" json:"dropper"`
	Section uint `form:"section" json:"section"`
}

func dropperSectionPillsGET(ctx *gin.Context, db *gorm.DB) {
	var query getSectionPillsQuery

	if err := ctx.ShouldBindQuery(&query); err != nil {
		log.Printf("Tentativa de ativar o dropper falhada! <query> : %s \n", err.Error())
		ctx.JSON(
			400,
			returnMessage(
				"erro",
				"query string fornecida invalida ou mal-formada",
			),
		)
		return
	}

	var section models.DropperSection
	err := db.Model(models.DropperSection{DropperID: query.Dropper}).Preload("Positions").First(&section).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(
			404,
			returnMessage(
				"not found",
				"o dropper não foi encontrado",
			),
		)
		return
	} else if err != nil {
		log.Printf("Erro inesperado: %s", err.Error())
		ctx.JSON(
			500,
			returnMessage(
				"erro",
				"tente novamente mais tarde",
			),
		)
		return
	}

	ctx.JSON(
		200,
		section.Positions,
	)
}

type dropperActivationQuery struct {
	DropperID uint `form:"id"`
}

func activateDropperGET(ctx *gin.Context, db *gorm.DB) {
	var queryParams dropperActivationQuery

	if err := ctx.ShouldBindQuery(&queryParams); err != nil {
		log.Printf("Tentativa de ativar o dropper falhada! <query> : %s \n", err.Error())
		ctx.JSON(
			400,
			returnMessage(
				"erro",
				"query string fornecida invalida ou mal-formada",
			),
		)
		return
	}

	dropper := models.Dropper{
		Model:  gorm.Model{ID: queryParams.DropperID},
		Active: false,
	}

	err := db.Limit(1).Find(&dropper).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(
			404,
			returnMessage(
				"not found",
				"o dropper não foi encontrado",
			),
		)
		return
	} else if err != nil {
		log.Printf("Erro inesperado: %s", err.Error())
		ctx.JSON(
			500,
			returnMessage(
				"erro",
				"tente novamente mais tarde",
			),
		)
		return
	}

	dropper.Active = true
	db.Save(&dropper)

	ctx.JSON(
		200,
		returnMessage(
			"sucesso",
			"dropper ativado",
		),
	)
}

type createDispenseScheduleQuery struct {
	DropperID uint `query:"dropper"`
}

type createDispenseScheduleBody struct {
	Name        string        `json:"name"`
	Active      bool          `json:"active"`
	Description string        `json:"descriptiont"`
	StartDate   time.Time     `json:"start_date"`
	EndDate     time.Time     `json:"end_date"`
	Interval    time.Duration `json:"interval"`
}

func createDropperDispenseSchedulePOST(c *gin.Context, db *gorm.DB) {
	var requestQuery createDispenseScheduleQuery
	var newDispenseSchedule createDispenseScheduleBody

	if err := c.BindQuery(&requestQuery); err != nil {
		log.Printf("Tentativa de criar horario falhada! <query> : %s \n", err.Error())
		c.JSON(
			400,
			returnMessage(
				"erro",
				"dados fornecidos são invalidos ou mal-formados",
			),
		)
		return
	}

	if err := c.ShouldBind(&newDispenseSchedule); err != nil {
		log.Printf("Tentativa de criar horario falhada! <body> : %s \n", err.Error())
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
	err := db.Find(&dropper, requestQuery.DropperID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("O dropper pedido <%d> não existe\n", requestQuery.DropperID)
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

	if err := dropper.CreateDispenseSchedule(
		db, newDispenseSchedule.Active,
		newDispenseSchedule.Name,
		newDispenseSchedule.Description,
		newDispenseSchedule.StartDate,
		newDispenseSchedule.EndDate,
		newDispenseSchedule.Interval); err != nil {
		log.Println("O horário não foi criado")
		c.JSON(
			400,
			returnMessage(
				"erro",
				"dados fornecidos são invalidos ou mal-formados",
			),
		)
		return
	}
}

type dispensePills struct {
}

func dropperDispensePillsGET(ctx *gin.Context, db *gorm.DB) {
	panic("unimplemented")
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
	Dropper  uint   `form:"dropper_id" json:"dropper_id"`
	Section  uint   `form:"section_pos" json:"section_pos"`
	PillName string `form:"pill_name" json:"pill_name"`
	Quantity uint   `form:"pill_quantity" json:"pill_quantity"`
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
	Dropper uint            `form:"dropper_id"`
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

	section_id, err := dropper.CreateDropperSection(db, newSection.Name, newSection.Pills)
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
		gin.H{
			"status":    "sucesso",
			"id_seccao": section_id,
		},
	)
}

type newDropper struct {
	Name string `form:"name" json:"name" binding:"required"`
	// Active     bool   `form:"active" json:"active" binding:"required"`
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

	dropper := models.NewDropper(newDropper.Name, newDropper.MachineUrl)

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
