// Package models implements models for this example
package models

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Dropper struct {
	// OwnerID    uuid.UUID `db:"owner_id"`
	gorm.Model `json:"-"`

	// Allow read and create of field SerialID
	SerialID   uuid.UUID `gorm:"<-;index;default:gen_random_uuid();" json:"serial_id"`
	Active     bool      `json:"active"`
	MachineURL string    `gorm:"<-;default:null;uniqueIndex" json:"machine_url"`
	Name       string    `json:"name"`

	// A dropper has many Schedules
	DispenseSchedules []DispenseSchedule `gorm:"constraint:OnDelete:SET NULL;" json:"schedules"`

	Sections []DropperSection `gorm:"constraint:OnDelete:SET NULL;" json:"sections"`
}

// DispenseSchedule Stores dropper medicine dispense schedule
type DispenseSchedule struct {
	gorm.Model `json:"-"`

	// Dropper foreign key
	DropperID uint `gorm:"uniqueIndex:uniqueSchedule" json:"-"`

	Name        string        `gorm:"uniqueIndex:uniqueSchedule" json:"name"`
	Active      bool          `gorm:"default:true;" json:"active"`
	Description string        `json:"description"`
	StartDate   time.Time     `json:"start_date"`
	EndDate     time.Time     `json:"end_date"`
	Interval    time.Duration `gorm:"default:6;" json:"interval"`
}

type ScheduledPills struct {
	gorm.Model `json:"-"`

	// DispenseSchedule foreign key
	DispenseScheduleID uint `json:"-"`

	Pills []Pill `json:"pills"`

	Dispensed     bool      `gorm:"default:false;" json:"dispensed"`
	DispensedTime time.Time ` json:"dispensed_time"`
}

type Pill struct {
	gorm.Model `json:"-"`

	// DispenseSchedule foreign key
	ScheduledPillsID uint `json:"-"`

	Name  string `gorm:"name" json:"name"`
	Count uint   `gorm:"count" json:"count"`
}

type DropperSection struct {
	gorm.Model `json:"-"`

	DropperID uint
	Positions []Position `json:"positions"`

	Section         string `json:"section"`
	CurrentPosition uint   `json:"current_position"`
	Empty           bool   `json:"is_empty"`
}

type Position struct {
	gorm.Model `json:"-"`

	DropperSectionID uint `json:"-"`

	Position uint   `json:"position"`
	PillName string `json:"pill_name"`
	Empty    bool   `json:"is_empty"`
}

/*
	Name        string        `gorm:"unique" json:"name"`
	Active      bool          `gorm:"default:true;" json:"active"`
	Description string        `json:"description"`
	StartDate   time.Time     `json:"start_date"`
	EndDate     time.Time     `json:"end_date"`
	Interval
*/

// TODO - Scan all dropper sections, to find which have the pills necessary to fulfil the request

func (d *Dropper) CreateDispenseSchedule(
	db *gorm.DB,
	active bool,
	name, descricao string,
	start, end time.Time,
	interval time.Duration,
) (err error) {
	schedule := DispenseSchedule{
		DropperID:   d.ID,
		Name:        name,
		Active:      active,
		Description: descricao,
		StartDate:   start,
		EndDate:     end,
		Interval:    interval,
	}

	// Create if not exists
	err = db.FirstOrCreate(DispenseSchedule{Name: name}, &schedule).Error
	if err != nil {
		log.Printf("Erro inesperado ao criar drop schedule: %s", err.Error())
		return err
	}

	return
}

// ReloadSection recebe um ponteiro gorm.DB, uma secção, o nome do comprimido e a sua quantidade
// e recarrega se possivel esse comprimido na secção definida da máquina escolhida
func (dp *Dropper) ReloadSection(db *gorm.DB, section uint, pillName string, count uint) error {
	if count > 9 {
		return errors.New("demasiados comprimidos fornecidos")
	} else if count < 1 {
		return errors.New("poucos comprimidos fornecidos")
	}

	if section < 1 || section > 9 {
		return errors.New("posição de secção fora do intervalo permitido")
	}
	// Reformat for section to be between 0 and 8, and not 1 - 9
	section -= 1

	err := db.Model(&Dropper{}).First(nil, "id", dp.ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("dropper não encontrado")
	} else if err != nil {
		log.Printf("Erro inesperado: %s", err.Error())
		return errors.New("erro inesperado encontrado")
	}
	dp.reloadDropperData(db)

	if len(dp.Sections[section].Positions) > 8 {
		return errors.New("secção cheia")
	}

	dp.Sections[section].Positions = append(dp.Sections[section].Positions, Position{
		PillName: pillName,
		Position: uint(len(dp.Sections[section].Positions) + 1),
		Empty:    false,
	})
	db.Save(dp)
	// Update the runtime instance of the dropper
	db.Preload("Sections.Positions").Find(dp, "id", dp.ID)

	return nil
}

func NewSectionPosition(pillName string, position uint, empty bool) Position {
	return Position{
		PillName: pillName,
		Empty:    empty,
		Position: position,
	}
}

type PillList map[string]int

func (dp *Dropper) CreateDropperSection(db *gorm.DB, name string, pills PillList) (uint, error) {
	var newSection DropperSection
	var pillCount int = 0

	err := db.First(&Dropper{}, dp.ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errors.New("nenhum dropper encontrado para o id fornecido")
	}

	if len(pills) == 0 {
		newSection = DropperSection{
			DropperID: dp.ID,
			Section:   name,
			Empty:     true,
			Positions: nil,
		}

		err := db.Create(&newSection).Error

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return 0, errors.New("esta secção já existe")
		} else if err != nil {
			log.Printf("Erro inesperado: %s", err.Error())
			return 0, errors.New("erro inesperado")
		}

		return newSection.ID, nil
	}

	// If each section has 1 pill per position, we cant have more than 9 pills in the list
	if len(pills) > 9 {
		return 0, errors.New("to many pills")
	}

	// Any configuration of pills is accepted, as long its not larger than 9
	// So we can have 5 ibuprofens and 4 aspirins, and so on
	for pill := range pills {
		pillCount += int(pills[pill])
		if pillCount > 9 {
			return 0, errors.New("defined to many pills")
		}
	}

	// log.Printf("PillCount: %d", pillCount)
	newSection = DropperSection{
		DropperID:       dp.ID,
		Section:         name,
		Empty:           false,
		CurrentPosition: 1,
		Positions:       make([]Position, pillCount),
	}

	// Building and adding each dropper section positions pill's
	sectionPosition := 1
	for pillName, count := range pills {
		for range count {
			newSection.Positions[sectionPosition-1] = NewSectionPosition(pillName, uint(sectionPosition), false)
			sectionPosition += 1
		}
	}

	err = db.Create(&newSection).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return 0, errors.New("esta secção e/ou os seus comprimidos já foram definidos")
	} else if err != nil {
		log.Printf("Erro inesperado: %s\n", err.Error())
		return 0, errors.New("erro inesperado")
	}

	dp.Sections = append(dp.Sections, newSection)
	db.Save(dp)

	dp.reloadDropperData(db)

	return newSection.ID, nil
}

// MigrateAll runs all migrations for the models defined in this folder
func MigrateAll(db *gorm.DB) {
	err := db.AutoMigrate(&Dropper{}, &DispenseSchedule{}, &DropperSection{}, &Position{}, &ScheduledPills{}, &Pill{})
	if err != nil {
		log.Fatalf("Failed to migrate gorm models: %s", err.Error())
	}
}

// NewDropper creates a new dropper struct instance
func NewDropper(name string, machine_url string) *Dropper {
	return &Dropper{
		Name:       name,
		Active:     false,
		MachineURL: machine_url,
	}
}

func (d *Dropper) reloadDropperData(db *gorm.DB) {
	db.Preload("Sections.Positions").Find(d, "id", d.ID)
}

func (d *Dropper) Create(db *gorm.DB) (uint, error) {
	result := db.Create(d)

	return d.ID, result.Error
}
