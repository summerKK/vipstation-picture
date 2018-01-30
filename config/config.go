package config

import (
	"github.com/burntsushi/toml"
	"log"
)

type database struct {
	DbUsername string `toml:"DB_USERNAME"`
	DbPassword string `toml:"DB_PASSWORD"`
	DbHost     string `toml:"DB_HOST"`
	DbPort     string `toml:"DB_PORT"`
	DbDatabase string `toml:"DB_DATABASE"`
}

type saveDir struct {
	Dir string `toml:"IMG_SAVE_DIR"`
}

type Localconfig struct {
	Vps1 database `toml:"VPS1"`
	Vps2 database `toml:"VPS2"`
	SaveDir string `toml:"IMG_SAVE_DIR"`
}

var (
	Config *Localconfig
)

func init() {
	if _, err := toml.DecodeFile("config.toml", &Config); err != nil {
		log.Fatal("解析配置文件失败!")
	}
}
