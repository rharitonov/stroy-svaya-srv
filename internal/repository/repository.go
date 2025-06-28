package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"stroy-svaya/internal/model"
	"time"

	_ "modernc.org/sqlite"
)

const (
	dateFormat = "2006-01-02"
)

type Repository interface {
	InsertPileDrivingRecordLine(rec *model.PileDrivingRecordLine) error
	Close() error
}

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &SQLiteRepository{db: db}, nil

}

func (r *SQLiteRepository) InsertPileDrivingRecordLine(rec *model.PileDrivingRecordLine) error {
	query := `INSERT INTO pile_driving_record (
				pile_field_id,
				pile_number,
				project_id,
				start_date,
				fact_pile_head,
				recorded_by
			)
			VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query,
		rec.PileFieldId,
		rec.PileNumber,
		rec.ProjectId,
		rec.StartDate.Format(dateFormat),
		rec.FactPileHead,
		rec.RecordedBy,
	)
	return err

}

func (r *SQLiteRepository) GetPileDrivingRecord(projectId int) ([]model.PileDrivingRecordLine, error) {
	var lines []model.PileDrivingRecordLine
	query := `SELECT 
				pile_number,
				start_date,
				fact_pile_head,
				recorded_by
			FROM pile_driving_record
			WHERE project_id = ?`
	rows, err := r.db.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ln model.PileDrivingRecordLine
		if err := rows.Scan(
			&ln.PileNumber,
			&ln.StartDate,
			&ln.FactPileHead,
			&ln.RecordedBy,
		); err != nil {
			return nil, err
		}
		lines = append(lines, ln)
	}
	return lines, nil
}

func (r *SQLiteRepository) GetPilesToDriving(projectId int) ([]string, error) {
	var pileNos []string
	query := `select pif.pile_number from pile_in_field pif
		left join pile_driving_record pdr 
			on pdr.pile_field_id = pif.pile_field_id 
			and pdr.pile_number = pif.pile_number 
		inner join pile_field pf 
			on pf.id = pif.pile_field_id 
		where pdr.pile_number is null 
			and pf.project_id = ?
		order by cast(pif.pile_number as int);`
	rows, err := r.db.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pileNo string
		if err := rows.Scan(&pileNo); err != nil {
			return nil, err
		}
		pileNos = append(pileNos, pileNo)
	}
	return pileNos, nil
}

func (r *SQLiteRepository) GetPiles(filter model.PileFilter) ([]string, error) {
	var pileNos []string
	var args []any
	var query string
	switch filter.Status {
	case 10:
		query = `select pif.pile_number from pile_in_field pif
			left join pile_driving_record pdr 
				on pdr.pile_field_id = pif.pile_field_id 
				and pdr.pile_number = pif.pile_number 
			inner join pile_field pf 
				on pf.id = pif.pile_field_id 
			where pdr.pile_number is null 
				and pf.project_id = ?
			order by cast(pif.pile_number as int);`
		args = append(args, filter.ProjectId)
	case 30:
		query = `select pif.pile_number from pile_in_field pif
			inner join pile_field pf 
				on pf.id = pif.pile_field_id 
			where pf.project_id = ?
			order by cast(pif.pile_number as int);`
		args = append(args, filter.ProjectId)
	case 20:
		query = `select pif.pile_number from pile_in_field pif
			inner join pile_driving_record pdr 
				on pdr.pile_field_id = pif.pile_field_id 
				and pdr.pile_number = pif.pile_number 
			inner join pile_field pf 
				on pf.id = pif.pile_field_id 
			%s
			order by cast(pif.pile_number as int);`
		var fields map[string]any
		var where string
		jsonData, err := json.Marshal(filter)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(jsonData), &fields); err != nil {
			return nil, err
		}

		for field, value := range fields {
			//log.Println(field, value) // DBG
			if field != "status" && field != "recorded_by" {
				if where == "" {
					where = "where "
				} else {
					where = where + " and "
				}
				isDate := false
				var v any = value
				switch v.(type) {
				case string:
					_, err := time.Parse(time.RFC3339, value.(string))
					if err == nil {
						isDate = true
					}
				default:
					isDate = false
				}
				if isDate {
					date, _ := time.Parse(time.RFC3339, value.(string))
					where = fmt.Sprintf("%spdr.%s = ?", where, field)
					args = append(args, date.Format(dateFormat))
				} else {
					where = fmt.Sprintf("%spdr.%s = ?", where, field)
					args = append(args, value)
				}
			}
		}
		query = fmt.Sprintf(query, where)
	}
	//log.Println(query) //DBG
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pileNo string
		if err := rows.Scan(&pileNo); err != nil {
			return nil, err
		}
		pileNos = append(pileNos, pileNo)
	}
	return pileNos, nil
}

func (r *SQLiteRepository) GetUserFullNameInitialFormat(tgChatId int64) (string, error) {
	var lastName, initials string
	query := `select last_name, initials 
		from user_setup 
		where tg_user_id = ?
		limit 1;`
	rows, err := r.db.Query(query, tgChatId)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&lastName, &initials); err != nil {
			return "", err
		}
	}
	if lastName == "" {
		return "", nil
	}
	return fmt.Sprintf("%s %s", lastName, initials), nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
