package repo

import "time"

type Node struct {
	ID         int       `gorm:"column:id"`
	IPAddress  string    `gorm:"column:ip_address"`
	UserName   string    `gorm:"column:user_name"`
	Password   string    `gorm:"column:password"`
	CreateTime time.Time `gorm:"column:create_time"`
}
