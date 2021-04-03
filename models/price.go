package models

type Price struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Amount string `json:"amount"`
	Time   string `json:"time"`
	Round  string `json:"round" gorm:"unique"`
}
