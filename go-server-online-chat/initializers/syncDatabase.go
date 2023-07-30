package initializers

import "go-server-online-chat/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
}
