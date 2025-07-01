package main

import (
	"database/sql"
	"fmt"
	"log"
	"stroy-svaya/internal/config"

	_ "modernc.org/sqlite"
)

type TestRec struct {
	project_id          int
	pile_field_id       int
	pile_number         string
	pile_no             int
	pile_type           string
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
	t.pile_type = "С140.40-11.1"
	t.design_pile_head = 9900
	t.design_pile_tip = -4100
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
	tr := NewTextRec(20, 20)

	db, err = sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		panic(err.Error())
	}
	if err := db.Ping(); err != nil {
		panic(err.Error())
	}

	query := `INSERT INTO project (code, name, address, parent_project_id, start_date, end_date)
		VALUES(?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query,
		"99/2025-АМЦ-3-КЖ01",
		"Многоквартирный жилой дом со встроенным подземным гаражом",
		"г. Санкт-Петеребург, муниципальный округ Финляндский округ, Полюстровский проспект, участок 31",
		0,
		"2025-05-01",
		"2025-10-31")
	if err != nil {
		panic(err.Error())
	}
	id2, err := result.LastInsertId()
	if err != nil {
		panic(err.Error())
	}
	tr.project_id = int(id2)

	// pile_field
	query = "INSERT INTO pile_field (project_id, name, drawing_number) VALUES (?, ?, ?)"
	result, err = db.Exec(query, tr.project_id, "Секция 1", "Чертеж01")
	if err != nil {
		panic("Oops insert pile_field")
	}
	id2, err = result.LastInsertId()
	if err != nil {
		panic("Oops Get pile_field Id")
	}
	tr.pile_field_id = int(id2)

	// pile_in_field
	query = `INSERT INTO  pile_in_field (
    	pile_field_id,
    	pile_number,
    	pile_type,
    	x_coord,
    	y_coord,
    	design_pile_head,
    	design_pile_tip
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	for _, x := range tr.pile_x_coord_points {
		for _, y := range tr.pile_y_coord_points {
			tr.GetNextPileNumber()
			result, err = db.Exec(query,
				tr.pile_field_id,
				tr.pile_number,
				tr.pile_type,
				fmt.Sprintf("%dа", x),
				fmt.Sprintf("%dб", y),
				tr.design_pile_head,
				tr.design_pile_tip,
			)
			if err != nil {
				panic("Oops pile_in_field")
			}
		}
	}

	log.Print("Test data was created")
}
