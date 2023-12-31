package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"indexer/internal/repository/mysql/model"
	"os"
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

func (m *MysqlRepository) UpsertIsVespaUpdated(timeoutCtx context.Context, conditions []model.UpsertCondition) error {
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
		value := fmt.Sprintf("('%s','%s',%t)", conditions[i].SpotId, conditions[i].VespaUpdatedAt, conditions[i].IsVespaUpdated)
		queries[i] = value
	}
	values := strings.Join(queries, ",")

	q := fmt.Sprintf("INSERT INTO update_process (spot_id, vespa_updated_at, is_vespa_updated) VALUES %s AS new ON DUPLICATE KEY UPDATE spot_id = new.spot_id, vespa_updated_at = new.vespa_updated_at, is_vespa_updated = new.is_vespa_updated", values)
	_, err = tx.ExecContext(timeoutCtx, q)
	if err != nil {
		return err
	}

	return nil

}
