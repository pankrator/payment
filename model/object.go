package model

type Object interface {
	GetType() string
	Validate() error
}
