package dataFactory

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"k8s.io/klog/v2"
)

var connStr string
var ctx context.Context
var Dbpool *pgxpool.Pool
var DBPassWordPath string

// var DBRootPW stringe

func CreateConnection() {
	var err error
	content, err := ioutil.ReadFile(DBPassWordPath)
	if err != nil {
		klog.Errorln(err)
		return
	}
	dbRootPW := string(content)

	connStr = "postgres://postgres:{DB_ROOT_PW}@timescaledb.hypercloud5-system.svc.cluster.local:5432/postgres"
	// 치환
	connStr = strings.Replace(connStr, "{DB_ROOT_PW}", dbRootPW, -1)
	ctx = context.Background()
	// connConfig, err := pgx.ParseConfig(connString)
	// if err != nil {
	// 	return nil, err
	// }

	// config := &Config{
	// 	ConnConfig:           connConfig,
	// 	createdByParseConfig: true,
	// }
	Dbpool, err = pgxpool.Connect(ctx, connStr)
	if err != nil {
		klog.Errorf("Unable to connect to database: %v\n", err)
		panic(err)
	}
}
