package main

import (
	"flag"
	"github.com/beego/beego/v2/server/web"
	"github.com/lifei6671/douyinbot/admin"
	"log"
	"path/filepath"
)

var (
	port       = ":9080"
	configFile = "./admin/conf/app.conf"
)

func main() {
	flag.StringVar(&port, "port", ":9080", "Listening address and port.")
	flag.StringVar(&configFile, "config-file", configFile, "config file path.")
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
	admin.Run(port, configFile)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
