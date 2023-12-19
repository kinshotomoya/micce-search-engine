package mysql

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type MysqlRepository struct {
	client *sql.DB
}

func NewMysqlRepository(viper *viper.Viper) (*MysqlRepository, error) {
	conf := mysql.Config{
		User:   viper.GetString("MYSQL_DB_USER"),
		Passwd: viper.GetString("MYSQL_DB_PASSWORD"),
		Net:    "tcp",
		Addr:   viper.GetString("MYSQL_DB_URL"),
		DBName: viper.GetString("MYSQL_DB_NAME"),
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
