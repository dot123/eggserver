package config

import (
	"encoding/json"
	"fmt"
	"github.com/koding/multiconfig"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	C    = new(Config)
	once sync.Once
)

// MustLoad Load conf file (toml/json/yaml)
func MustLoad(paths ...string) {
	log.Println("load", paths)

	once.Do(func() {
		loaders := []multiconfig.Loader{
			&multiconfig.TagLoader{},
			&multiconfig.EnvironmentLoader{},
		}

		for _, path := range paths {
			if strings.HasSuffix(path, "toml") {
				loaders = append(loaders, &multiconfig.TOMLLoader{Path: path})
			}
			if strings.HasSuffix(path, "json") {
				loaders = append(loaders, &multiconfig.JSONLoader{Path: path})
			}
			if strings.HasSuffix(path, "yaml") {
				loaders = append(loaders, &multiconfig.YAMLLoader{Path: path})
			}
		}

		m := multiconfig.DefaultLoader{
			Loader:    multiconfig.MultiLoader(loaders...),
			Validator: multiconfig.MultiValidator(&multiconfig.RequiredValidator{}),
		}
		m.MustLoad(C)
	})
}

func PrintWithJSON() {
	if C.PrintConfig {
		b, err := json.MarshalIndent(C, "", " ")
		if err != nil {
			log.Print("[CONFIG] JSON marshal error: " + err.Error())
			return
		}
		log.Print(string(b) + "\n")
	}
}

type Config struct {
	RunMode       string
	PrintConfig   bool
	Swagger       bool
	GM            bool
	DataKey       byte
	WalletAddress string
	JWTAuth       JWTAuth
	RateLimiter   RateLimiter
	CORS          CORS
	Gorm          Gorm
	MySQL         MySQL
	RedisBackend  RedisBackend
	Log           Log
}

func (c *Config) IsDebugMode() bool {
	return c.RunMode == "debug"
}

type JWTAuth struct {
	Key        string
	Expired    time.Duration
	UseSession bool
}

type RateLimiter struct {
	Enable bool
	Count  int
}

type CORS struct {
	Enable           bool
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type Gorm struct {
	Debug           bool
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	TablePrefix     string
}

type MySQL struct {
	Host       string
	Port       int
	User       string
	Password   string
	DBName     string
	Parameters string
}

func (a MySQL) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", a.User, a.Password, a.Host, a.Port, a.DBName, a.Parameters)
}

type RedisBackend struct {
	Addrs           []string
	DB              int
	MaxRetries      int
	Username        string
	Password        string
	PoolSize        int
	MinIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

type Log struct {
	Level         int
	Format        string
	Output        string
	OutputFile    string
	RotationCount int
	RotationTime  int
}
