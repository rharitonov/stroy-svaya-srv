package main

import (
	"database/sql"
	"fmt"
	"log"
	"stroy-svaya/internal/config"
	"time"

	_ "modernc.org/sqlite"
)

type TestRec struct {
	project_id          int
	pile_field_id       int
	pile_number         string
	pile_no             int
	pile_code           string
	design_pile_head    float32
	design_pile_tip     float32
	pile_x_coord_points []int
	pile_y_coord_points []int
}

func NewTextRec(max_x_coord int, max_y_coord int) *TestRec {
	var t TestRec
	t.pile_x_coord_points = make([]int, max_x_coord)
	for i := 0; i < max_x_coord; i++ {
		t.pile_x_coord_points[i] = i + 1
	}
	t.pile_y_coord_points = make([]int, max_y_coord)
	for i := 0; i < max_y_coord; i++ {
		t.pile_y_coord_points[i] = i + 1
	}
	t.pile_code = "С90.30-3"
	t.design_pile_head = 11680
	t.design_pile_tip = 0
	return &t
}

func (t *TestRec) GetNextPileNumber() {
	t.pile_no += 1
	t.pile_number = fmt.Sprintf("%d", t.pile_no)
}

func main() {
	cfg := config.Load()
	var db *sql.DB
	var err error
	var result sql.Result
	tr := NewTextRec(1, 251)

	db, err = sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		panic(err.Error())
	}
	if err := db.Ping(); err != nil {
		panic(err.Error())
	}

	// item
	query := "INSERT INTO item (code, description, weight) VALUES (?, ?, ?), (?, ?, ?)"
	_, err = db.Exec(query, "С140.40-11.1", "Свая ж/б С140.40-11.1", 5600, "С90.30-3", "Свая ж/б С90.30-3", 2030)
	if err != nil {
		panic(fmt.Errorf("item: %v", err))
	}

	// equip
	query = "INSERT INTO equip (code, description, unit_type, unit_model, unit_weight, unit_power)" +
		" VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)"
	_, err = db.Exec(query,
		"ЮНТТАН-25", "Юнттан PM25HD", "Гидравлический", "HHK7", 7000, 84,
		"ЮНТТАН-20", "Юнттан PM20LC", "Гидравлический", "HHK5AL", 5000, 60)
	if err != nil {
		panic(fmt.Errorf("equip: %v", err))
	}

	// project
	sd := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
	ed := time.Date(2025, 8, 31, 0, 0, 0, 0, time.UTC)
	query = `INSERT INTO project (code, name, address, parent_project_id, start_date, end_date)
		VALUES(?, ?, ?, ?, ?, ?)`
	result, err = db.Exec(query,
		"ЧЕРЕП",
		"Цех котонизации",
		"г. Череповец",
		0,
		sd.Format(time.DateOnly),
		ed.Format(time.DateOnly))
	if err != nil {
		panic(fmt.Errorf("project: %v", err))
	}
	id2, err := result.LastInsertId()
	if err != nil {
		panic(fmt.Errorf("project: %v", err))
	}
	tr.project_id = int(id2)

	// pile_field
	query = "INSERT INTO pile_field (project_id, name, drawing_number) VALUES (?, ?, ?)"
	result, err = db.Exec(query, tr.project_id, "", "")
	if err != nil {
		panic(fmt.Errorf("pile_field: %v", err))
	}
	id2, err = result.LastInsertId()
	if err != nil {
		panic(fmt.Errorf("pile_field: %v", err))
	}
	tr.pile_field_id = int(id2)

	// pile_in_field
	query = `INSERT INTO pile_in_field (
    	pile_field_id,
    	pile_number,
    	pile_code,
    	x_coord,
    	y_coord,
    	design_pile_head,
    	design_pile_tip
		) VALUES (?, ?, ?, ?, ?, ?, ?)`
	for _, x := range tr.pile_x_coord_points {
		for _, y := range tr.pile_y_coord_points {
			tr.GetNextPileNumber()
			_, err = db.Exec(query,
				tr.pile_field_id,
				tr.pile_number,
				tr.pile_code,
				fmt.Sprintf("%dа", x),
				fmt.Sprintf("%dб", y),
				tr.design_pile_head,
				tr.design_pile_tip,
			)
			if err != nil {
				panic(fmt.Errorf("pile_in_field: %v", err))
			}
		}
	}

	// company_info
	query = "INSERT INTO company_info (code, name) VALUES (?, ?)"
	_, err = db.Exec(query, "", "ООО \"Строй Свая\"")
	if err != nil {
		panic(fmt.Errorf("company_info: %v", err))
	}

	log.Print("Test data was created")
}
