package repository

import (
	"database/sql"
	"fmt"
	"stroy-svaya/internal/model"

	_ "modernc.org/sqlite"
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
				start_time,
				fact_pile_head,
				recorded_by
			)
			VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query,
		rec.PileFieldId,
		rec.PileNumber,
		rec.ProjectId,
		rec.StartDate,
		rec.FactPileHead,
		rec.RecordedBy,
	)
	return err

}

func (r *SQLiteRepository) GetPileDrivingRecord(projectId int) ([]model.PileDrivingRecordLine, error) {
	var lines []model.PileDrivingRecordLine
	query := `SELECT 
				pile_number,
				start_time,
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
