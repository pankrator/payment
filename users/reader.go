package users

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

type Settings struct {
	FileName     string `mapstructure:"file_name"`
	FileLocation string `mapstructure:"file_location"`
}

func (s *Settings) Keys() []string {
	return []string{
		"file_name",
		"file_location",
	}
}

type UserType string

var (
	Merchant UserType = "merchant"
	Admin    UserType = "admin"
)

type User struct {
	Type        UserType
	Name        string
	Email       string
	Password    string
	Status      string
	Description string
}

type UserReader struct {
	Users    []User
	settings *Settings
}

func NewCSVReader(settings *Settings) *UserReader {
	return &UserReader{
		settings: settings,
	}
}

func (ur *UserReader) Load() error {
	file, err := os.Open(filepath.Join(ur.settings.FileLocation, ur.settings.FileName))
	if err != nil {
		return fmt.Errorf("could not read file: %s", err)
	}
	reader := csv.NewReader(file)
	results, err := reader.ReadAll()
	if err != nil {
		return err
	}
	ur.Users = make([]User, 0, len(results))
	for _, result := range results {
		ur.Users = append(ur.Users, User{
			Type:        UserType(result[0]),
			Name:        result[1],
			Email:       result[2],
			Password:    result[3],
			Description: result[4],
			Status:      result[5],
		})
	}

	return nil
}
