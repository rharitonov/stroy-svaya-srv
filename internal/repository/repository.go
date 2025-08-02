package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"stroy-svaya/internal/model"
	"time"

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
				start_date,
				fact_pile_head,
				recorded_by
			)
			VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query,
		rec.PileFieldId,
		rec.PileNumber,
		rec.ProjectId,
		rec.StartDate.Format(time.DateOnly),
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
					args = append(args, date.Format(time.DateOnly))
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

func (r *SQLiteRepository) GetCompanyName() (string, error) {
	var company_name string
	query := "select name from company_info where code = ''"
	err := r.db.QueryRow(query).Scan(&company_name)
	if err != nil {
		return "", err
	}
	return company_name, nil
}

func (r *SQLiteRepository) GetProject(projectId int64) (*model.Project, error) {
	p := new(model.Project)
	query := "select id, code, name, address, parent_project_id, start_date, end_date, status " +
		"from project where id = ?"
	err := r.db.QueryRow(query, projectId).Scan(
		&p.Id,
		&p.Code,
		&p.Name,
		&p.Address,
		&p.ParentProjectId,
		&p.StartDate,
		&p.EndDate,
		&p.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return p, nil
}

func (r *SQLiteRepository) GetPile(filter model.PileFilter) (*model.PileDrivingRecordLine, error) {
	p := &model.PileDrivingRecordLine{}
	p.ProjectId = filter.ProjectId
	p.PileFieldId = filter.PileFieldId
	p.PileNumber = *filter.PileNumber

	var (
		num, sdate, rec_by string
		fph                int
		cr_at, upd_at      sql.NullString
		err                error
	)
	query := `select 1 from pile_in_field pif
        inner join pile_driving_record pdr 
            on pdr.pile_field_id = pif.pile_field_id 
            and pdr.pile_number = pif.pile_number 
        inner join pile_field pf 
            on pf.id = pif.pile_field_id 
        where pf.project_id = ?
            and pf.id = ?
            and pif.pile_number = ?
        order by cast(pif.pile_number as int);`
	err = r.db.QueryRow(query, filter.ProjectId, filter.PileFieldId, *filter.PileNumber).Scan(&num)
	if err != nil {
		if err == sql.ErrNoRows {
			p.Status = 10
			return p, nil
		}
		return nil, err
	}

	query = `select 
            pif.pile_number,
            date(pdr.start_date),
            pdr.fact_pile_head,
            pdr.recorded_by,
            strftime('%Y-%m-%d %H:%M:%S', pdr.created_at),
            strftime('%Y-%m-%d %H:%M:%S', pdr.updated_at)
        from pile_in_field pif
        left join pile_driving_record pdr 
            on pdr.pile_field_id = pif.pile_field_id 
            and pdr.pile_number = pif.pile_number 
        inner join pile_field pf 
            on pf.id = pif.pile_field_id 
        where pf.project_id = ?
            and pf.id = ?
            and pif.pile_number = ?
        order by cast(pif.pile_number as int);`

	rows, err := r.db.Query(query,
		filter.ProjectId,
		filter.PileFieldId,
		*filter.PileNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("pile number %s does not exist", *filter.PileNumber)
	}
	if err := rows.Scan(&num, &sdate, &fph, &rec_by, &cr_at, &upd_at); err != nil {
		return nil, err
	}

	dd, err := time.Parse(time.DateOnly, sdate)
	if err == nil {
		p.StartDate = dd
	}

	p.FactPileHead = fph
	p.RecordedBy = rec_by

	if cr_at.Valid {
		dd, err := time.Parse(time.DateTime, cr_at.String)
		if err == nil {
			p.CreatedAt = dd
		}
	}

	if upd_at.Valid {
		dd, err := time.Parse(time.DateTime, upd_at.String)
		if err == nil {
			p.UpdatedAt = dd
		}
	}

	p.Status = 20
	return p, nil
}

func (r *SQLiteRepository) GetPdrEquip(projectId int64) ([]model.Equip, error) {
	var equips []model.Equip
	query := `SELECT e.code, e.description, e.type, e.unit_type, e.unit_weight, e.unit_power
		FROM equip e
		INNER JOIN (	
			SELECT equip_code FROM pile_driving_record WHERE project_id = ? GROUP BY equip_code
		) pe on pe.equip_code = e.code`
	rows, err := r.db.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var e model.Equip
		if err := rows.Scan(&e.Code, &e.Description, &e.Type, &e.UnitType, &e.UnitWeight, &e.UnitPower); err != nil {
			return nil, err
		}
		equips = append(equips, e)
	}
	return equips, nil
}

func (r *SQLiteRepository) InsertOrUpdatePdrPile(rec *model.PileDrivingRecordLine) error {
	filter := model.PileFilter{}
	filter.ProjectId = rec.ProjectId
	filter.PileFieldId = rec.PileFieldId
	filter.PileNumber = &rec.PileNumber
	p, err := r.GetPile(filter)
	if err != nil {
		return err
	}
	query := ""
	switch p.Status {
	case 10:
		query = `INSERT INTO pile_driving_record (
				pile_field_id,
				pile_number,
				project_id,
				start_date,
				fact_pile_head,
				recorded_by
			)
			VALUES (?, ?, ?, ?, ?, ?)`
		_, err = r.db.Exec(query,
			rec.PileFieldId,
			rec.PileNumber,
			rec.ProjectId,
			rec.StartDate.Format(time.DateOnly),
			rec.FactPileHead,
			rec.RecordedBy,
		)

	case 20:
		query = `UPDATE pile_driving_record SET
				start_date = ?,
				fact_pile_head = ?,
				recorded_by = ?
			WHERE project_id = ? and pile_field_id = ? and pile_number = ?`
		_, err = r.db.Exec(query,
			rec.StartDate.Format(time.DateOnly),
			rec.FactPileHead,
			rec.RecordedBy,
			rec.ProjectId,
			rec.PileFieldId,
			rec.PileNumber)
		if err != nil { // DBG
			log.Println(err)
		}
	default:
		return fmt.Errorf("insert or update pile error")
	}

	return err
}

func (r *SQLiteRepository) GetUserSetup(tgChatID int64) (*model.User, error) {
	u := new(model.User)
	query := `select 
	    code, 
    	first_name,
    	last_name, 
    	surname, 
    	initials,
    	tg_user_id,
    	email
		from user_setup 
		where tg_user_id = ?;`
	err := r.db.QueryRow(query, tgChatID).Scan(
		&u.Code,
		&u.FirstName,
		&u.LastName,
		&u.Surname,
		&u.Initials,
		&u.TgUserId,
		&u.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
