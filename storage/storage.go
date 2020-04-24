package storage

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pankrator/payment/model"

	migratepg "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

var ErrNotFound error = errors.New("not found in storage")

type Storage struct {
	*gorm.DB
	settings *Settings

	models map[string]func() Model
}

type Settings struct {
	Host              string
	Port              string
	Database          string
	Username          string
	Password          string
	SkipSSLValidation bool
}

func DefaultSettings() *Settings {
	return &Settings{
		Host:              "127.0.0.1",
		Port:              "5432",
		Database:          "payment",
		Username:          "payment",
		Password:          "payment",
		SkipSSLValidation: true,
	}
}

func New(s *Settings) *Storage {
	return &Storage{
		settings: s,
		models:   make(map[string]func() Model),
	}
}

func (s *Storage) Open() error {
	sslmode := "require"
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

	db, err := gorm.Open("postgres", dbURI)

	if err != nil {
		return err
	}
	s.DB = db
	s.configure()
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
	dbModel := dbModelBlueprint().FromObject(object)
	err := s.DB.Create(dbModel).Error
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
	dbModel := dbModelBlueprint().FromObject(object)
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
		return nil, ErrNotFound
	}
	return dbModel.ToObject(), result.Error
}

func (s *Storage) Transaction(f func(s *Storage) error) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		return f(&Storage{
			models: s.models,
			DB:     tx,
		})
	})
}

func (s *Storage) RegisterModels(typee string, modelProvider func() Model) {
	s.models[typee] = modelProvider
}

func (s *Storage) migrate(dbURI string) error {
	driver, err := migratepg.WithInstance(s.DB.DB(), &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("could not initialize driver: %s", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://storage/migrations", "postgres", driver)
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
