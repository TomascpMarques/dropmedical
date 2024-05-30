// Package http_api joins  the routes for this example
package http_api

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/TomascpMarques/dropmedical/models"
)

// SetupRoutesGroup groups the API routes into a Engine route group with the path prefix of `/api`
func SetupRoutesGroup(router *gin.Engine, db *gorm.DB, ch *chan models.MqttActionRequest) {
	// Logger
	router.Use(gin.LoggerWithFormatter(apiLogger))

	// Recovery from panics inside middlewares and handlers
	router.Use(gin.Recovery())

	// Routes
	api := router.Group("/api")
	// ------------------------
	api.POST("/dropper", func(ctx *gin.Context) { registerDropperPOST(ctx, db) })
	api.POST("/dropper/section", func(ctx *gin.Context) { registerDropperSectionPOST(ctx, db) })
	api.POST("/dropper/section/reload", func(ctx *gin.Context) { reloadDropperSectionPOST(ctx, db, ch) })
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
	DropperID string `form:"id" query:"id"`
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
		SerialID: uuid.MustParse(queryParams.DropperID),
		Active:   false,
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
	DropperID string `form:"dropper" query:"dropper"`
}

type createDispenseScheduleBody struct {
	Name        string         `json:"name"        form:"name"`
	Active      bool           `json:"active"      form:"active"`
	Description string         `json:"description" form:"description"`
	StartDate   time.Time      `json:"start_date"  form:"start_date"`
	EndDate     time.Time      `json:"end_date"    form:"end_date"`
	Interval    time.Duration  `json:"interval"    form:"interval"`
	Pills       map[string]int `json:"pills"       form:"pills"`
}

func createDropperDispenseSchedulePOST(c *gin.Context, db *gorm.DB) {
	var query createDispenseScheduleQuery
	var new_schedule createDispenseScheduleBody

	if err := c.ShouldBindQuery(&query); err != nil {
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

	if err := c.ShouldBindJSON(&new_schedule); err != nil {
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
	err := db.Find(&dropper, "serial_id = ?", query.DropperID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("O dropper pedido <%s> não existe\n", query.DropperID)
		c.JSON(
			400,
			returnMessage(
				"erro",
				"dados fornecidos são invalidos ou mal-formados",
			),
		)
		return
	} else if errors.Is(err, gorm.ErrForeignKeyViolated) {
		log.Printf("Erro de chave estrangeira: %s\n", err.Error())
		c.JSON(
			400,
			returnMessage(
				"erro",
				"Ay Caramba!",
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
		db,
		new_schedule.Active,
		new_schedule.Name,
		new_schedule.Description,
		new_schedule.StartDate,
		new_schedule.EndDate,
		new_schedule.Interval,
		new_schedule.Pills,
	); err != nil {
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

/* type dispensePills struct {} */

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
	Dropper  uuid.UUID `form:"dropper_id" json:"dropper_id"`
	Section  uint      `form:"section_pos" json:"section_pos"`
	PillName string    `form:"pill_name" json:"pill_name"`
	Quantity uint      `form:"pill_quantity" json:"pill_quantity"`
}

func reloadDropperSectionPOST(c *gin.Context, db *gorm.DB, ch *chan models.MqttActionRequest) {
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
	err := db.Find(&dropper, "serial_id = ?", reloadSectionAction.Dropper).Error
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
		log.Printf("Erro ao recarregar secção, erro: %+e\n", err)
		c.JSON(
			500,
			returnMessage(
				"erro",
				"Erro interno tente mais tarde",
			),
		)
		return
	}

	*ch <- models.MqttActionRequest{
		Topic: fmt.Sprintf("angle%d", reloadSectionAction.Section),
		Value: []byte(fmt.Sprintf("0,%d", reloadSectionAction.Section)),
	}

	c.JSON(200, gin.H{
		"status": "sucesso",
		"razao":  "Secção carregada",
	})
}

type newDropperSection struct {
	Dropper uuid.UUID       `form:"dropper_id" json:"dropper_id"`
	Name    string          `form:"name" json:"name"`
	Pills   models.PillList `form:"pills" json:"pills"`
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
	err := db.Find(&dropper, "serial_id = ?", newSection.Dropper).Error
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

	err := db.Create(&dropper).Error

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		log.Println("Tentativa de criar um dropper que já existe")
		c.JSON(409, gin.H{
			"status": "erro",
			"razao":  "Este dropper já existe",
		})
		return
	} else if err != nil {
		log.Printf("Erro inesperado: %+v", err)
		c.JSON(500, gin.H{
			"status": "erro",
			"razao":  "Ocorreu um erro inesperado",
		})
		return
	}

	c.JSON(200, dropper)
}
