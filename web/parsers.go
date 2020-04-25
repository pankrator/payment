package web

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"log"
	"strings"
)

type Parser interface {
	Unmarshal([]byte, interface{}) error
}

var parsers map[string]Parser = make(map[string]Parser)

func GetParser(contentType string) (Parser, bool) {
	parser, found := parsers[contentType]
	if found {
		return parser, true
	}
	for typee, parser := range parsers {
		if strings.Contains(contentType, typee) {
			return parser, true
		}
	}
	return nil, false
}

func RegisterParser(contentType string, parser Parser) {
	parsers[contentType] = parser
}

type XMLParser struct{}

func (xp *XMLParser) Unmarshal(data []byte, object interface{}) error {
	err := xml.Unmarshal(data, object)
	if err != nil {
		log.Printf("Could not parse XML: %s", err)
		return errors.New("Could not parse XML")
	}
	return nil
}

type JSONParser struct{}

func (jp *JSONParser) Unmarshal(data []byte, object interface{}) error {
	err := json.Unmarshal(data, object)
	if err != nil {
		log.Printf("Could not parse JSON: %s", err)
		return errors.New("Could not parse JSON")
	}
	return nil
}
