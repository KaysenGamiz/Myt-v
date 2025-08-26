package db

type Movie struct {
	ID       uint   `gorm:"primarykey"`
	Path     string `gorm:"uniqueIndex"`
	Title    string
	Duration float64 // seconds
	CodecV   string
	CodecA   string
	Width    int
	Height   int
}
