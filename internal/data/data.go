package data

import (
	"context"
	"fmt"
	"time"

	"xinyuan_tech/relayer-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Data 数据访问层
type Data struct {
	db       *gorm.DB
	redis    *redis.Client
	rocketmq RocketMQProducer
}

// NewData 创建数据访问层
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, rocketmq RocketMQProducer) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if rdb != nil {
			if err := rdb.Close(); err != nil {
				log.NewHelper(logger).Error(err)
			}
		}
		if rocketmq != nil {
			if err := rocketmq.Close(); err != nil {
				log.NewHelper(logger).Error(err)
			}
		}
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	return &Data{
		db:       db,
		redis:    rdb,
		rocketmq: rocketmq,
	}, cleanup, nil
}

// NewDB 创建数据库连接
func NewDB(c *conf.Data, logger log.Logger) (*gorm.DB, func(), error) {
	logHelper := log.NewHelper(logger)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&Transaction{},
		&Builder{},
		&BuilderFee{},
		&Operator{},
	); err != nil {
		return nil, nil, err
	}

	logHelper.Info("Database connected and migrated")

	cleanup := func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	return db, cleanup, nil
}

// NewRedis 创建 Redis 客户端
func NewRedis(c *conf.Data, logger log.Logger) (*redis.Client, func(), error) {
	logHelper := log.NewHelper(logger)

	if c.Redis == nil || c.Redis.Addr == "" {
		logHelper.Warn("Redis not configured, skipping")
		return nil, func() {}, nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Network:      c.Redis.Network,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logHelper.Info("Redis connected")

	cleanup := func() {
		if rdb != nil {
			rdb.Close()
		}
	}

	return rdb, cleanup, nil
}

// DB 获取数据库连接
func (d *Data) DB() *gorm.DB {
	return d.db
}

// Redis 获取 Redis 客户端
func (d *Data) Redis() *redis.Client {
	return d.redis
}


