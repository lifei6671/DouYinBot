package main

import (
	"flag"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin"
	"github.com/lifei6671/douyinbot/admin/models"
	"log"
	"path/filepath"
)

var (
	port       = ":9080"
	configFile = "./admin/conf/app.conf"
	dataPath   = "./data/douyinbot.db"
)

func main() {
	flag.StringVar(&port, "port", port, "Listening address and port.")
	flag.StringVar(&configFile, "config-file", configFile, "config file path.")
	flag.StringVar(&dataPath, "data-file", dataPath, "database file path.")
	flag.Parse()
	if port == "" {
		port = ":9080"
	}
	if work, err := filepath.Abs("admin"); err == nil {
		if configFile == "" {
			configFile = filepath.Join(work, "/conf/app.conf")
		}
		web.WorkPath = work
	}
	if err := models.Init(dataPath); err != nil {
		panic(err)
	}
	if err := admin.Run(port, configFile); err != nil {
		panic(err)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
