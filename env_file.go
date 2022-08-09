package confserver

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func (conf *Configuration) dockerize() {
	_ = writeToFile("./config/default.yml", conf.defaultConfig(), yaml.Marshal)
	_ = writeToFile("./Dockerfile", conf.dockerfile(), func(v interface{}) ([]byte, error) {
		return v.([]byte), nil
	})
}

func (conf *Configuration) defaultConfig() map[string]string {
	m := map[string]string{}
	m["GOENV"] = "DEV"

	for _, envVar := range conf.defaultEnvVars.Values {
		if !envVar.Optional {
			m[envVar.Key(conf.Prefix())] = envVar.Value
		}
	}

	return m
}

func writeToFile(filename string, v interface{}, marshal func(v interface{}) ([]byte, error)) error {
	bytes, _ := marshal(v)
	dir := filepath.Dir(filename)
	if dir != "" {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(filename, bytes, os.ModePerm)
}
