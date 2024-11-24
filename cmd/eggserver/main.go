package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	_ "eggServer/docs"
	"eggServer/internal/config"
	cfg "eggServer/internal/gamedata"
	"eggServer/internal/handler"
	"eggServer/internal/logic"
	"eggServer/internal/middleware"
	"eggServer/internal/models"
	"eggServer/pkg/redisbackend"
	"eggServer/pkg/utils"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	mysqlDriver "github.com/go-sql-driver/mysql"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"path/filepath"
	"time"
)

// https://github.com/swaggo/swag/blob/master/README_zh-CN.md

// @title          gin-gorm-admin API
// @version        1.0
// @description    This is a game management background. you can use the api key `ApiKeyAuth` to test the authorization filters.
// @termsOfService https://github.com

// @contact.name  conjurer
// @contact.url   https:/github.com/dot123
// @contact.email conjurer888888@gmail.com

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     127.0.0.1:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in                         header
// @name                       Authorization

// Usage: go build -ldflags "-X main.VERSION=x.x.x"
var VERSION = "1.0.0"

var (
	l            *logrus.Entry
	ginLambda    *ginadapter.GinLambdaV2
	gormDB       *gorm.DB
	tables       *cfg.Tables
	redisBackend *redisbackend.RedisBackend
	version      = os.Getenv("version")
)

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	fmt.Printf("path: %s\n", req.RawPath)
	resp, err := ginLambda.ProxyWithContext(ctx, req)
	return resp, err
}

func main() {
	utils.SetConsoleTitle("eggserver")
	if version == "" {
		version = "debug"
	}
	log.Println("egg start...")
	log.Println("version:", version)

	config.MustLoad(fmt.Sprintf("./data/conf/game%s.toml", version))

	config.PrintWithJSON()
	initLogger()
	initGameData()
	initRedisBackend()
	initGorm()
	initLogic()
	initGin()
}

func initLogger() {
	plog := logrus.New()

	c := config.C.Log
	plog.SetLevel(logrus.Level(c.Level))

	plog.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05"})
	if c.Format == "json" {
		plog.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	}

	if c.Output != "" {
		switch c.Output {
		case "stdout":
			plog.SetOutput(os.Stdout)
		case "stderr":
			plog.SetOutput(os.Stdout)
		case "file":
			if name := c.OutputFile; name != "" {
				_ = os.MkdirAll(filepath.Dir(name), 0777)

				r, err := rotatelogs.New(name+".%Y-%m-%d",
					rotatelogs.WithLinkName(name),
					rotatelogs.WithRotationTime(time.Duration(c.RotationTime)*time.Second),
					rotatelogs.WithRotationCount(uint(c.RotationCount)))
				if err != nil {
					return
				}
				plog.SetOutput(r)
			}
		}
	}

	l = plog.WithFields(logrus.Fields{
		"source": "eggserver",
	})
}

func loader(file string) ([]map[string]interface{}, error) {
	if bytes, err := os.ReadFile("./data/static/" + file + ".json"); err != nil {
		return nil, err
	} else {
		jsonData := make([]map[string]interface{}, 0)
		if err = json.Unmarshal(bytes, &jsonData); err != nil {
			return nil, err
		}
		return jsonData, nil
	}
}

func initGameData() {
	log.Println("initGameData")
	if tb, err := cfg.NewTables(loader); err != nil {
		log.Fatalln(err.Error())
	} else {
		tables = tb
	}
}

// 创建数据库
func createDatabaseWithMySQL(cfg *mysqlDriver.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", cfg.User, cfg.Passwd, cfg.Addr)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET = `utf8mb4`;", cfg.DBName)
	_, err = db.Exec(query)
	return err
}

func initGorm() {
	log.Println("initGorm")

	dns := config.C.MySQL.DSN()

	mysqlCFG, err := mysqlDriver.ParseDSN(dns)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	// 创建数据库
	if err := createDatabaseWithMySQL(mysqlCFG); err != nil {
		log.Fatalln(err.Error())
		return
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,          // Slow SQL threshold
			LogLevel:                  logger.Warn,          // Log level
			IgnoreRecordNotFoundError: false,                // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,                // Don't include params in the SQL log
			Colorful:                  version != "release", // Disable color
		},
	)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dns,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   config.C.Gorm.TablePrefix,
			SingularTable: true,
		},
		Logger: newLogger,
	})

	if config.C.Gorm.Debug {
		db = db.Debug()
	}

	gormDB = db

	// 同步数据库表
	if err := models.AutoMigrate(db); err != nil {
		log.Fatalln(err.Error())
		return
	}
	gormConfig := config.C.Gorm
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(gormConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(gormConfig.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(gormConfig.ConnMaxIdleTime * time.Second)
	sqlDB.SetConnMaxLifetime(gormConfig.ConnMaxLifetime * time.Second)
}

func closeDB() {
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalln(err.Error())
	}

	if err := sqlDB.Close(); err != nil {
		log.Fatalln(err.Error())
	}
}

func initLogic() {
	log.Println("initLogic")
	logic.Init(tables)
}

func initGin() {
	log.Println("initGin")

	gin.SetMode(config.C.RunMode)

	r := gin.Default()
	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatalln(err.Error())
		return
	}

	if err := register(r); err != nil {
		log.Fatalln(err.Error())
		return
	}

	if version == "debug" {
		// 在本地启动 HTTP 服务器，监听指定端口
		if err := r.Run(":8080"); err != nil {
			log.Fatalln(err.Error())
			return
		}
	}

	ginLambda = ginadapter.NewV2(r)
	lambda.Start(Handler)
}

func initRedisBackend() {
	log.Println("initRedisBackend")

	var tlsConfig *tls.Config
	if version == "release" {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	redisBackendConfig := config.C.RedisBackend
	redisBackend = redisbackend.NewRedisBackend(&redis.UniversalOptions{
		Addrs:           redisBackendConfig.Addrs,
		Username:        redisBackendConfig.Username,
		Password:        redisBackendConfig.Password,
		DB:              redisBackendConfig.DB,
		MaxRetries:      redisBackendConfig.MaxRetries,
		DialTimeout:     redisBackendConfig.DialTimeout * time.Second,     // 连接超时时间
		ReadTimeout:     redisBackendConfig.ReadTimeout * time.Second,     // 读取超时时间
		WriteTimeout:    redisBackendConfig.WriteTimeout * time.Second,    // 写入超时时间
		ConnMaxIdleTime: redisBackendConfig.ConnMaxIdleTime * time.Second, // 连接池中连接的最大闲置时间
		ConnMaxLifetime: redisBackendConfig.ConnMaxLifetime * time.Second, // 连接最大生命周期
		PoolSize:        redisBackendConfig.PoolSize,                      // 连接池大小
		MinIdleConns:    redisBackendConfig.MinIdleConns,                  // 最小空闲连接数
		TLSConfig:       tlsConfig,
	})
}

// 注册路由
func register(app *gin.Engine) error {
	app.Use(middleware.Cors())
	app.Use(middleware.DB(gormDB))
	app.Use(middleware.RB(redisBackend))
	app.Use(middleware.Trace())

	// Swagger
	if config.C.Swagger {
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		log.Println("visit http://127.0.0.1:8080/swagger/index.html")
	}

	handler.RegisterHandlers(app, redisBackend.Client(), l)

	return nil
}
