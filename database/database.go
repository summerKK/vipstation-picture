package database

import (
	"database/sql"
	"vipstation-picture/config"
	"log"
	"vipstation-picture/base"

	_ "github.com/go-sql-driver/mysql"
	"time"
	"fmt"
	"strings"
)

type Database struct {
	Connection *sql.DB
	RowCache   base.ImgCache

	UpdateChan chan map[string]string
}

func NewDatabase() *Database {

	database := &Database{
		RowCache:   base.NewCache(),
		UpdateChan: make(chan map[string]string, 10),
	}
	database.connection()
	return database
}

func (database *Database) connection() {
	config := config.Config.Vps1
	//sql.Open("mysql","user:password@tcp(127.0.0.1:3306)/hello")
	connection, err := sql.Open("mysql", config.DbUsername + ":" + config.DbPassword+
		"@tcp("+ config.DbHost+ ":"+ config.DbPort+ ")/"+ config.DbDatabase)

	if err != nil {
		log.Fatal("数据库连接失败!", err)
	}

	database.Connection = connection
}

func (database *Database) GetProducts() {

	sql := `SELECT media_gallery,sku FROM lux_products WHERE crawler = "API_VIPSTATION" AND media_gallery like "http://%"`

	rows, err := database.Connection.Query(sql)

	if err != nil {
		log.Fatalf("获取数据出问题! error:%v", err)
	}

	defer rows.Close()

	var media_gallery, sku string
	for rows.Next() {
		rows.Scan(&media_gallery, &sku)

		database.RowCache.Put(map[string]string{"sku": sku, "imgs": media_gallery})
	}

	database.ticker()
}

func (database *Database) ticker() {
	go func() {
		for {
			fmt.Println(database.RowCache.Summary())
			time.Sleep(20 * time.Second)
		}
	}()
}

func (database *Database) ReceiveMediaGallery() {
	go func() {
		for {
			imgs := <-database.UpdateChan
			database.updateMediaGallery(imgs["img"], imgs["sku"])
		}
	}()
}

func (database *Database) updateMediaGallery(imgs string, sku string) {

	sql := `UPDATE lux_products SET media_gallery = ?,audit = 1,thumbnail = ?,image = ?,updated_at = ? WHERE sku = ?`

	stmt, err := database.Connection.Prepare(sql)
	if err != nil {
		log.Fatalf("更新数据出错!请检查!,sql:%s\nerror:%s", sql, err)
	}
	defer stmt.Close()
	thumbnail := strings.Split(imgs, ";")[0]
	image := thumbnail
	updatedAt := time.Now().Format("2006-01-02 15:04:05")
	stmt.Query(imgs, thumbnail, image, updatedAt, sku)
	log.Printf("sku:%s 更新完成!",sku)
}
