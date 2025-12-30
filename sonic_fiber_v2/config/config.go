package config

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	App      AppConfig      `mapstructure:"app"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"`     // mysql, postgres, sqlite
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	FilePath string `mapstructure:"filepath"` // for sqlite
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}

type AppConfig struct {
	Mode        string `mapstructure:"mode"`
	WorkDir     string `mapstructure:"work_dir"`
	UploadDir   string `mapstructure:"upload_dir"`
	TemplateDir string `mapstructure:"template_dir"`
	ThemeDir    string `mapstructure:"theme_dir"`
}

func NewConfig() *Config {
	var configFile string
	flag.StringVar(&configFile, "config", "", "config file path")
	flag.Parse()

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigType("yaml")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("./conf/")
		viper.SetConfigName("config")
	}

	// 设置默认值
	setDefaults()

	conf := &Config{}
	if err := viper.ReadInConfig(); err != nil {
		// 如果没有配置文件，使用默认配置
		zap.L().Warn("No config file found, using defaults")
	} else {
		if err := viper.Unmarshal(conf); err != nil {
			panic(err)
		}
	}

	// 初始化工作目录
	if conf.App.WorkDir == "" {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		conf.App.WorkDir, _ = filepath.Abs(pwd)
	} else {
		workDir, err := filepath.Abs(conf.App.WorkDir)
		if err != nil {
			panic(err)
		}
		conf.App.WorkDir = workDir
	}

	// 初始化子目录
	conf.normalizeDirectories()

	return conf
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")
	
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.filepath", "sonic.db")
	
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")
	
	viper.SetDefault("app.mode", "development")
	viper.SetDefault("app.upload_dir", "./uploads")
	viper.SetDefault("app.template_dir", "./resources/template")
	viper.SetDefault("app.theme_dir", "./resources/template/theme")
}

func (c *Config) normalizeDirectories() {
	// 确保上传目录存在
	if c.App.UploadDir == "" {
		c.App.UploadDir = filepath.Join(c.App.WorkDir, "uploads")
	}
	
	// 创建必要的目录
	dirs := []string{c.App.UploadDir, c.App.TemplateDir, c.App.ThemeDir}
	for _, dir := range dirs {
		if dir != "" {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				os.MkdirAll(dir, os.ModePerm)
			}
		}
	}
}

func (c *Config) IsDev() bool {
	return c.App.Mode == "development"
}

func (c *Config) GetDSN() string {
	switch c.Database.Type {
	case "mysql":
		return c.Database.Username + ":" + c.Database.Password + "@tcp(" + c.Database.Host + ":" + c.Database.Port + ")/" + c.Database.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	case "postgres":
		return "host=" + c.Database.Host + " port=" + c.Database.Port + " user=" + c.Database.Username + " password=" + c.Database.Password + " dbname=" + c.Database.DBName + " sslmode=disable"
	case "sqlite":
		return c.Database.FilePath
	default:
		return c.Database.FilePath
	}
}
