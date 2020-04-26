package gormdb

import (
	"database/sql"
	"fmt"
	"log"
	"path"
	"runtime"

	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/storage"

	"github.com/golang-migrate/migrate"
	migratepg "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = path.Dir(b)
)

type Storage struct {
	*gorm.DB
	settings *storage.Settings

	models map[string]func() Model
}

func New(s *storage.Settings) *Storage {
	return &Storage{
		settings: s,
		models:   make(map[string]func() Model),
	}
}

func (s *Storage) Open(connectFunc func(driver, url string) (*sql.DB, error)) error {
	sslmode := "verify-full"
	if s.settings.SkipSSLValidation {
		sslmode = "disable"
	}
	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		s.settings.Username,
		s.settings.Password,
		s.settings.Host,
		s.settings.Port,
		s.settings.Database,
		sslmode)

	connection, err := connectFunc("postgres", dbURI)
	if err != nil {
		return err
	}

	db, err := gorm.Open("postgres", connection)
	if err != nil {
		return err
	}
	s.DB = db
	s.configure()

	s.registerModels(model.TransactionObjectType, func() Model { return &Transaction{} })
	s.registerModels(model.MerchantType, func() Model { return &Merchant{} })

	if err := s.migrate(dbURI); err != nil {
		return err
	}

	log.Printf("Storage connection opened")

	return nil
}

func (s *Storage) Create(object model.Object) (model.Object, error) {
	dbModelBlueprint, found := s.models[object.GetType()]
	if !found {
		return nil, fmt.Errorf("no such model found %s", object.GetType())
	}
	dbModel, err := dbModelBlueprint().FromObject(object)
	if err != nil {
		return nil, err
	}
	err = s.DB.Create(dbModel).Error
	if err != nil {
		// TODO: Wrap storage error
		return nil, err
	}

	return dbModel.ToObject(), nil
}

func (s *Storage) Save(object model.Object) error {
	dbModelBlueprint, found := s.models[object.GetType()]
	if !found {
		return fmt.Errorf("no such model found %s", object.GetType())
	}
	dbModel, err := dbModelBlueprint().FromObject(object)
	if err != nil {
		return err
	}
	return s.DB.Save(dbModel).Error
}

func (s *Storage) Get(typee string, id string) (model.Object, error) {
	dbModelBlueprint, found := s.models[typee]
	if !found {
		return nil, fmt.Errorf("no such model found %s", typee)
	}
	dbModel := dbModelBlueprint()
	result := s.Where("uuid = ?", id).Find(dbModel)
	if result.RecordNotFound() {
		return nil, storage.ErrNotFound
	}
	return dbModel.ToObject(), result.Error
}

func (s *Storage) Count(typee string, condition string, args ...interface{}) (int, error) {
	dbModelBlueprint, found := s.models[typee]
	if !found {
		return 0, fmt.Errorf("no such model found %s", typee)
	}
	dbModel := dbModelBlueprint()
	var count int
	result := s.New().Model(dbModel).Where(condition, args...).Count(&count)
	return count, result.Error
}

func (s *Storage) DeleteAll(typee string) error {
	dbModelBlueprint, found := s.models[typee]
	if !found {
		return fmt.Errorf("no such model found %s", typee)
	}
	dbModel := dbModelBlueprint()
	return s.Delete(dbModel).Error
}

func (s *Storage) Transaction(f func(s storage.Storage) error) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		return f(&Storage{
			models: s.models,
			DB:     tx,
		})
	})
}

func (s *Storage) registerModels(typee string, modelProvider func() Model) {
	s.models[typee] = modelProvider
}

func (s *Storage) migrate(dbURI string) error {
	driver, err := migratepg.WithInstance(s.DB.DB(), &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("could not initialize driver: %s", err)
	}
	// TODO: Introduce configuration for migrations path
	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s/migrations", basepath), "postgres", driver)
	if err != nil {
		return fmt.Errorf("could not initialize migrate: %s", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not execute migrations: %s", err)
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		for _, m := range s.models {
			tx.AutoMigrate(m())
		}
		for _, m := range s.models {
			model := m()
			if err := model.InitSQL(tx); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (s *Storage) configure() {
	s.DB.DB().SetMaxIdleConns(10)
	s.DB.DB().SetMaxOpenConns(100)
}

func (s *Storage) Close() {
	s.DB.Close()
}
