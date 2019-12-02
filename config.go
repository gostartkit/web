package web

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// ServerConfig struct
type ServerConfig struct {
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

// DatabaseHostConfig struct
type DatabaseHostConfig struct {
	Host string
	Port int
}

// DatabaseCluster struct
type DatabaseCluster struct {
	Driver    string
	Database  string
	Username  string
	Password  string
	Charset   string
	Collation string
	Write     *DatabaseHostConfig
	Read      *[]DatabaseHostConfig
}

// Config struct
type Config struct {
	Server   *ServerConfig
	Database *DatabaseCluster
}

// ReadJSON read json to data
func ReadJSON(data interface{}, filename string) error {

	if !filepath.IsAbs(filename) {
		dir, err := os.Getwd()

		if err != nil {
			return err
		}

		filename = filepath.Join(dir, filename)
	}

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, data); err != nil {
		return err
	}

	return nil
}

// WriteJSON write data to json
func WriteJSON(data interface{}, filename string, force bool) error {

	if !filepath.IsAbs(filename) {
		dir, err := os.Getwd()

		if err != nil {
			return err
		}

		filename = filepath.Join(dir, filename)
	}

	if force || !exists(filename) {

		b, err := json.MarshalIndent(data, "", "  ")

		if err != nil {
			return err
		}

		return ioutil.WriteFile(filename, b, 0600)
	}

	return os.ErrExist
}
