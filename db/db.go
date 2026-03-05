package db

import (
	"fmt"

	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	//"gorm.io/gorm/logger"

	m "golang-rest-api/models"
)

const (
	// DBHostEnv is the database host environment variable name
	DBHostEnv string = "DB_HOST"
	// DBPortEnv is the database port environment variable name
	DBPortEnv string = "DB_PORT"
	// DBNameEnv is the database name environment variable name
	DBNameEnv string = "DB_NAME"
	// DBUserEnv is the database user environment variable name
	DBUserEnv string = "DB_USER"
	// DBPasswordEnv is the database user password environment variable name
	DBPasswordEnv string = "DB_PASSWORD"

	DBSSLMode string = "DB_SSLMODE"
	// DBSearchPathEnv is the database search path environment variable name
	//DBSearchPathEnv string = "DB_SEARCH_PATH"
)

// PostgreSQLDSN is a PostgreSQL datasource name
type PostgreSQLDSN struct {
	Host       string
	Port       string
	DBName     string
	//SearchPath string
	User       string
	Password   string
	SSLMode		string	
}


func (db PostgreSQLDSN) ConnectionString() string{
	s :=fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=%s", db.Host, db.Port, db.User, db.Password, db.DBName, db.SSLMode)

	return s
}

type BDData struct{
	*gorm.DB
}

func Connection(dbdata PostgreSQLDSN, lgr zerolog.Logger) (*BDData, error){
	db, err := gorm.Open(postgres.Open(dbdata.ConnectionString()), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	lgr.Info().Msgf("sql database opened for %s on port %s", dbdata.Host, dbdata.Port)

	return  &BDData{db}, err
}


//NOT IN USE
/*
func (dbData *BDData) AutomigrateTables(tables []interface{}, lgr zerolog.Logger) error{
	err := dbData.DB.AutoMigrate(tables...)
	
	if err != nil {
		lgr.Info().Str("error", err.Error()).Msg("Failed to AutoMigrate to DB")
		return err
	}

	return nil
}
*/

func (dbData *BDData) AutomigrateTables(lgr zerolog.Logger) error{
	var tables = []interface{}{
		&m.Audiobook{},
		&m.Tag{},
		&m.Language{},
		&m.GenderVoice{},
	}
	
	err := dbData.AutoMigrate(tables...)
	
	if err != nil {
		lgr.Info().Str("error", err.Error()).Msg("Failed to AutoMigrate to DB")
		return err
	}

	return nil
}