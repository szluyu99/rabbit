package rabbit

import (
	"io"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDatabase(driver, dsn string, logWrite io.Writer) *gorm.DB {
	if driver == "" {
		driver = GetEnv(ENV_DB_DRIVER)
	}
	if dsn == "" {
		dsn = GetEnv(ENV_DSN)
	}

	var l logger.Interface
	if logWrite == nil {
		logWrite = os.Stdout
	}

	l = logger.New(
		log.New(logWrite, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Warn, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)

	cfg := &gorm.Config{
		Logger:                 l,
		SkipDefaultTransaction: true,
	}

	db, err := CreateDatabaseInstance(driver, dsn, cfg)
	if err != nil {
		panic(err)
	}

	return db
}

func MakeMigrates(db *gorm.DB, insts ...any) error {
	return db.AutoMigrate(insts...)
}

func CreateDatabaseInstance(driver, dsn string, cfg *gorm.Config) (*gorm.DB, error) {
	if driver == "mysql" {
		return gorm.Open(mysql.Open(dsn), cfg)
	}
	if dsn == "" {
		dsn = "file::memory:"
	}
	return gorm.Open(sqlite.Open(dsn), cfg)
}
