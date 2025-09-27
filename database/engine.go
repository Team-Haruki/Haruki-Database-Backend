package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type HarukiDatabaseEngine struct {
	db *gorm.DB
}

func NewDatabaseEngine(dsn string) (*HarukiDatabaseEngine, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &HarukiDatabaseEngine{db: db}, nil
}

func (e *HarukiDatabaseEngine) InitEngine(models ...interface{}) error {
	return e.db.AutoMigrate(models...)
}

func (e *HarukiDatabaseEngine) ShutdownEngine() error {
	sqlDB, err := e.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (e *HarukiDatabaseEngine) Select(out interface{}, query interface{}, args ...interface{}) error {
	return e.db.Where(query, args...).Find(out).Error
}

func (e *HarukiDatabaseEngine) SelectOne(out interface{}, query interface{}, args ...interface{}) error {
	return e.db.Where(query, args...).First(out).Error
}

func (e *HarukiDatabaseEngine) SelectWithJoin(out interface{}, join string, query interface{}, args ...interface{}) error {
	return e.db.Joins(join).Where(query, args...).Find(out).Error
}

func (e *HarukiDatabaseEngine) Delete(model interface{}, query interface{}, args ...interface{}) (int64, error) {
	result := e.db.Where(query, args...).Delete(model)
	return result.RowsAffected, result.Error
}

func (e *HarukiDatabaseEngine) Add(instance interface{}) error {
	return e.db.Create(instance).Error
}
