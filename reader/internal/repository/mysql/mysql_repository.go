package mysql

import (
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

func (m *MysqlRepository) UpsertIsVespaUpdated(conditions []model.UpsertCondition) error {
	queries := make([]string, len(conditions))
	for i := range conditions {
		value := fmt.Sprintf("(%s,%s,%s,%T)", conditions[i].SpotId, conditions[i].UpdatedAt, conditions[i].VespaUpdatedAt, conditions[i].IsVespaUpdated)
		queries[i] = value
	}
	values := strings.Join(queries, ",")

	_, err := m.client.Query("INSERT INTO update_process (spot_id, updated_at, vespa_updated_at, is_vespa_updated) VALUES ? AS new ON DUPLICATE KEY UPDATE spot_id = new.spot_id, updated_at = new.updated_at, vespa_updated_at = new.vespa_updated_at, is_vespa_updated = new.is_vespa_updated", values)
	if err != nil {
		return err
	}

	return nil

}

func (m *MysqlRepository) GetSpotIdsToUpdate() ([]string, error) {
	rows, err := m.client.Query("SELECT spot_id FROM update_process WHERE is_vespa_updated = false")
	if err != nil {
		return nil, err
	}

	resultSpotIds := make([]string, 0)
	for rows.Next() {
		var spotId string
		err := rows.Scan(&spotId)
		if err != nil {
			return nil, err
		}
		resultSpotIds = append(resultSpotIds, spotId)
	}

	return resultSpotIds, nil
}
