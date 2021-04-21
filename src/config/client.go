package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type NodeRule struct {
	SeederTime  int     `json:"seeder_time"`
	SeederRatio float64 `json:"seeder_ratio"`
}

type Node struct {
	Source  string   `json:"source"`
	Passkey string   `json:"passkey"`
	Limit   int      `json:"limit"`
	Enable  bool     `json:"enable"`
	Rule    NodeRule `json:"rule"`
}

type ServerRule struct {
	ConcurrentDownload int  `json:"concurrent_download"`
	DiskThreshold      int  `json:"disk_threshold"`
	DiskOverCommit     bool `json:"disk_overcommit"`
	MaxSpeed           int  `json:"max_speed"`
	MinAliveTime       int  `json:"min_alivetime"`
	MaxAliveTime       int  `json:"max_alivetime"`
	MinTaskSize        int  `json:"min_tasksize"`
	MaxTaskSize        int  `json:"max_tasksize"`
	MaxDiskLatency     int  `json:"max_disklatency"`
}

type Server struct {
	Endpoint string     `json:"endpoint"`
	Username string     `json:"username"`
	Password string     `json:"password"`
	Remark   string     `json:"remark"`
	Enable   bool       `json:"enable"`
	Rule     ServerRule `json:"rule"`
}

type Config struct {
	// DB Confoig
	Db string `json:"dbserver"`
	// PT Datasource
	Node []Node `json:"node"`
	// QB Server
	Server []Server `json:"server"`
}

func GetConfig() (Config, error) {
	var cfg Config
	configFile := GetConfigFilePath()
	// Open our jsonFile
	jsonFile, err := os.Open(configFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		return cfg, errors.New("failed to open config!")
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, err
}

func GetConfigFilePath() string {
	if workingDir, err := os.Getwd(); err == nil {
		configFile := filepath.Join(workingDir, "config.json")
		if fileExists(configFile) {
			return configFile
		}
	}

	if fileExists("/etc/goseeder.conf") {
		return "/etc/goseeder.conf"
	}

	return ""
}

func fileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}
