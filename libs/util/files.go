package util

import (
	"compress/gzip"
	"encoding/gob"
	"os"
)

// SaveObjectToFile write gob+gzip to file
func SaveObjectToFile(path string, value interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	zip := gzip.NewWriter(file)
	defer zip.Close()
	gob.Register(value)
	encoder := gob.NewEncoder(zip)
	return encoder.Encode(value)
}

// LoadObjectFromFile read gob+gzip from file
func LoadObjectFromFile(path string, value interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zip, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer zip.Close()

	gob.Register(value)
	decoder := gob.NewDecoder(zip)
	return decoder.Decode(value)
}
