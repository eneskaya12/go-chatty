package initializers

import (
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
)

var DB *gorm.DB

func ConnectToDb() {
	var err error

	errEnv := godotenv.Load()
	if errEnv != nil {
	  log.Fatal("Error loading .env file")
	}

	mysqluser := os.Getenv("mysqluser")
	mysqlpass := os.Getenv("mysqlpass")
	mysqlhost := os.Getenv("mysqlhost")
	mysqldb := os.Getenv("mysqldb")

	dsn := mysqluser + ":" + mysqlpass + "@tcp(" + mysqlhost + ":" + "3306" + ")/" + mysqldb
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database")
	}
}
