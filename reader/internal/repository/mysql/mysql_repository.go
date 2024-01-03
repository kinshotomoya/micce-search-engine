package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"os"
	"reader/internal/repository/mysql/model"
	"strings"
)

type MysqlRepository struct {
	client *sql.DB
}

func NewMysqlRepository() (*MysqlRepository, error) {
	conf := mysql.Config{
		User:   os.Getenv("MYSQL_DB_USER"),
		Passwd: os.Getenv("MYSQL_DB_PASSWORD"),
		Net:    "tcp",
		Addr:   os.Getenv("MYSQL_DB_URL"),
		DBName: os.Getenv("MYSQL_DB_NAME"),
	}

	db, err := sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &MysqlRepository{
		client: db,
	}, nil

}

func (m *MysqlRepository) UpdateIndexStatus(timeoutCtx context.Context, conditions []model.UpsertCondition) error {
	tx, err := m.client.BeginTx(timeoutCtx, nil)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if err != nil {
		return err
	}
	queries := make([]string, len(conditions))
	for i := range conditions {
		value := fmt.Sprintf("('%s','%s', '%s')", conditions[i].SpotId, conditions[i].UpdatedAt, conditions[i].IndexStatus)
		queries[i] = value
	}
	values := strings.Join(queries, ",")

	q := fmt.Sprintf("INSERT INTO update_process (spot_id, updated_at, index_status) VALUES %s AS new ON DUPLICATE KEY UPDATE spot_id = new.spot_id, updated_at = new.updated_at, index_status = new.index_status", values)
	_, err = tx.ExecContext(timeoutCtx, q)
	if err != nil {
		return err
	}

	return nil
}

// UpsertIsVespaUpdatedAndGetSpotIdsToUpdate NOTE: upsertとselectを１つのトランザクションの中で行う
func (m *MysqlRepository) UpsertIsVespaUpdatedAndGetSpotIdsToUpdate(timeoutCtx context.Context, conditions []model.UpsertCondition) ([]string, error) {
	tx, err := m.client.BeginTx(timeoutCtx, nil)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if err != nil {
		return nil, err
	}
	queries := make([]string, len(conditions))
	for i := range conditions {
		value := fmt.Sprintf("('%s','%s',%t, '%s')", conditions[i].SpotId, conditions[i].UpdatedAt, conditions[i].IsVespaUpdated, conditions[i].IndexStatus)
		queries[i] = value
	}
	values := strings.Join(queries, ",")

	q := fmt.Sprintf("INSERT INTO update_process (spot_id, updated_at, is_vespa_updated, index_status) VALUES %s AS new ON DUPLICATE KEY UPDATE spot_id = new.spot_id, updated_at = new.updated_at, is_vespa_updated = new.is_vespa_updated, index_status = new.index_status", values)
	_, err = tx.ExecContext(timeoutCtx, q)
	if err != nil {
		return nil, err
	}

	// NOTE: firestoreのin operatorのlimitが30件なのでlimit 30にする
	rows, err := tx.QueryContext(timeoutCtx, "SELECT spot_id FROM update_process WHERE index_status = 'READY' limit 30")
	if err != nil {
		return nil, err
	}

	resultSpotIds := make([]string, 0)
	for rows.Next() {
		var spotId string
		err = rows.Scan(&spotId)
		if err != nil {
			return nil, err
		}
		resultSpotIds = append(resultSpotIds, spotId)
	}

	return resultSpotIds, nil
}

func (m *MysqlRepository) GetIndexStatusReady(timeoutCtx context.Context) ([]string, error) {
	rows, err := m.client.QueryContext(timeoutCtx, "SELECT spot_id FROM update_process WHERE index_status = 'READY'")
	if err != nil {
		return nil, err
	}

	resultSpotIds := make([]string, 0)
	for rows.Next() {
		var spotId string
		err = rows.Scan(&spotId)
		if err != nil {
			return nil, err
		}
		resultSpotIds = append(resultSpotIds, spotId)
	}

	return resultSpotIds, nil
}
