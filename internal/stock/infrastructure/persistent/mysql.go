package persistent

import (
	"context"
	"fmt"
	_ "github.com/biggestfatboy/gorder-v2/common/config"
	"github.com/biggestfatboy/gorder-v2/common/logging"
	"github.com/biggestfatboy/gorder-v2/stock/infrastructure/persistent/builder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type MySQL struct {
	db *gorm.DB
}

type StockModel struct {
	Id        int64     `gorm:"column:id"`
	ProductID string    `gorm:"column:product_id"`
	Quantity  int32     `gorm:"column:quantity"`
	Version   int64     `gorm:"column:version"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func NewMySQL() *MySQL {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		viper.GetString("mysql.dbname"),
	)
	logrus.Debugf("dsn = %s", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Panicf("connect to mysql fail: %v", err)
	}
	db.Callback().Create().Before("gorm:create").Register("set_create_time", func(d *gorm.DB) {
		d.Statement.SetColumn("create_time", time.Now().Format(time.DateTime))
	})
	return &MySQL{db: db}
}

func NewMySQLWithDB(db *gorm.DB) *MySQL {
	return &MySQL{db: db}
}

func (s *StockModel) BeforeCreate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	s.CreatedAt = time.Now()
	tx.Statement.SetColumn("created_at", time.Now().Format(time.DateTime))
	tx.Statement.SetColumn("updated_at", time.Now().Format(time.DateTime))
	return nil
}

func (s StockModel) TableName() string {
	return "o_stock"
}

func (d *MySQL) UseTransaction(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return d.db
	}
	return tx
}

func (d MySQL) StartTransaction(f func(tx *gorm.DB) error) error {
	return d.db.Transaction(f)
}

func (d MySQL) GetStockByID(ctx context.Context, query *builder.Stock) (result *StockModel, err error) {

	_, deferLog := logging.WhenMySQL(ctx, "GetStockByID", query)
	defer deferLog(result, &err)
	err = query.Fill(d.db.WithContext(ctx)).First(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d MySQL) BatchGetStockByID(ctx context.Context, query *builder.Stock) (result []StockModel, err error) {
	_, deferLog := logging.WhenMySQL(ctx, "BatchGetStockByID", query)
	defer deferLog(result, &err)
	var results []StockModel
	err = query.Fill(d.db.WithContext(ctx)).Find(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}

func (d MySQL) Update(ctx context.Context, tx *gorm.DB, cond *builder.Stock, update map[string]any) (err error) {
	var returning StockModel
	_, deferLog := logging.WhenMySQL(ctx, "BatchUpdateStock", cond)
	defer deferLog(returning, &err)

	res := cond.Fill(d.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{})).Updates(update)
	return res.Error
}

func (d MySQL) Create(ctx context.Context, tx *gorm.DB, create *StockModel) (err error) {
	var returning StockModel
	_, deferLog := logging.WhenMySQL(ctx, "Create", create)
	defer deferLog(returning, &err)
	err = d.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(&create).Error
	return err
}
