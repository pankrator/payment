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

type User struct {
	Name     string
	Email    string
	Password string
}

type UserReader struct {
	Users    []User
	settings *Settings
}

func NewReader(settings *Settings) *UserReader {
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
			Name:     result[0],
			Email:    result[1],
			Password: result[2],
		})
	}

	return nil
}
