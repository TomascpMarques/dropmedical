package models

import (
	"log"
	"testing"

	"github.com/TomascpMarques/dropmedical/database"
	"github.com/joho/godotenv"
)

func TestSetupDatabase(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Failed to read the environment variables: %s\n", err)
	}

	db, _ := database.NewPostgresConnection()
	MigrateAll(db)
}

func TestCreateDropper(t *testing.T) {
	db, _ := database.NewPostgresConnection()

	dropper := NewDropper("SupaOne", "")
	_, err := dropper.Create(db)
	if err != nil {
		t.Fatalf("Failed to create a dropper")
	}
}

func TestReloadDropperSection(t *testing.T) {
	db, _ := database.NewPostgresConnection()

	dropper := NewDropper("SupaTwo", "")
	id, err := dropper.Create(db)
	if err != nil {
		t.Fatalf("Failed to create a dropper")
	}
	log.Printf("ID: %d", id)

	// Dropper with pills should be created
	pills := PillList{
		"Ibuprofen": 2,
		"Aspirin":   3,
		"Plan B":    3,
	}
	_, err = dropper.CreateDropperSection(db, "NewOne", pills)
	if err != nil {
		t.Fatalf("Failed to create a dropper section that has 9 pills: %s", err.Error())
	}

	err = dropper.ReloadSection(db, 1, "TEST", 1)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	err = dropper.ReloadSection(db, 1, "TEST II", 1)
	if err == nil {
		t.Fatal("Error: Falha ao recarregar secção, devia falhar aqui.")
	}
}

func TestCreateSection(t *testing.T) {
	db, _ := database.NewPostgresConnection()

	dropper := NewDropper("SupaThree", "")
	_, err := dropper.Create(db)
	if err != nil {
		t.Fatalf("Failed to create a dropper")
	}

	// Dropper with pills should be created
	pills := PillList{
		"Ibuprofen": 2,
		"Aspirin":   3,
		"Plan B":    4,
	}
	_, err = dropper.CreateDropperSection(db, "NewOne", pills)
	if err != nil {
		t.Fatalf("Failed to create a dropper section that has 9 pills: %s", err.Error())
	}
	var dS DropperSection
	err = db.Model(&DropperSection{}).Preload("Positions").First(&dS).Error
	if err != nil {
		t.Fatalf("Failed to get the wanted data from the Droppers: %s", err.Error())
	}

	// Dropper with no pills should also be created
	i, err := dropper.CreateDropperSection(db, "NewTwo", nil)
	if err != nil {
		t.Fatalf("Failed to create a dropper section that has 0 pills: %s", err.Error())
	}
	if i == 0 {
		t.Fatalf("Failed to create a section with no pills")
	}

	// Dropper cant have a section with more than 9 total pills
	pills = PillList{
		"Ibuprofen": 3,
		"Aspirin":   3,
		"Plan B":    4,
	}
	_, err = dropper.CreateDropperSection(db, "NewThree", pills)
	if err == nil {
		t.Fatalf("Failed to not create a dropper section that has 10 pills")
	}
}
