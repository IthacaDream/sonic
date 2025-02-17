package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/util"
)

func NewConfig() *Config {
	var configFile string
	flag.StringVar(&configFile, "config", "", "")
	flag.Parse()
	viper.SetConfigType("yaml")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("./conf/")
		viper.SetConfigName("config.yaml")
	}

	conf := &Config{}
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(conf); err != nil {
		panic(err)
	}

	if conf.Sonic.WorkDir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			panic(errors.Wrap(err, "init config: get current dir"))
		}
		conf.Sonic.WorkDir, _ = filepath.Abs(pwd)
	} else {
		workDir, err := filepath.Abs(conf.Sonic.WorkDir)
		if err != nil {
			panic(err)
		}
		conf.Sonic.WorkDir = workDir
	}
	normalizeDir := func(path *string, subDir string) {
		if *path == "" {
			*path = filepath.Join(conf.Sonic.WorkDir, subDir)
		} else {
			temp, err := filepath.Abs(*path)
			if err != nil {
				panic(err)
			}
			*path = temp
		}
	}
	normalizeDir(&conf.Sonic.LogDir, "log")
	normalizeDir(&conf.Sonic.TemplateDir, "resources/template")
	normalizeDir(&conf.Sonic.AdminResourcesDir, "resources/admin")
	normalizeDir(&conf.Sonic.UploadDir, consts.SonicUploadDir)
	normalizeDir(&conf.Sonic.ThemeDir, "resources/template/theme")
	if conf.SQLite3 != nil && conf.SQLite3.Enable {
		normalizeDir(&conf.SQLite3.File, "sonic.db")
	}
	if !util.FileIsExisted(conf.Sonic.TemplateDir) {
		panic("template dir: " + conf.Sonic.TemplateDir + " not exist")
	}
	if !util.FileIsExisted(conf.Sonic.AdminResourcesDir) {
		panic("AdminResourcesDir: " + conf.Sonic.AdminResourcesDir + "not exist")
	}
	if !util.FileIsExisted(conf.Sonic.ThemeDir) {
		panic("theme dir: " + conf.Sonic.ThemeDir + " not exist")
	}

	initDirectory(conf)
	mode = conf.Sonic.Mode
	return conf
}

func initDirectory(conf *Config) {
	mkdirFunc := func(dir string, err error) error {
		if err == nil {
			if _, err = os.Stat(dir); os.IsNotExist(err) {
				err = os.MkdirAll(dir, os.ModePerm)
			}
		}
		return err
	}
	err := mkdirFunc(conf.Sonic.LogDir, nil)
	err = mkdirFunc(conf.Sonic.UploadDir, err)
	if err != nil {
		panic(fmt.Errorf("initDirectory err=%v", err))
	}
}

var mode string

func IsDev() bool {
	return mode == "development"
}
