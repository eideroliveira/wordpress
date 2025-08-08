package wordpress

	
type Entity struct {
	ID            int64 `gorm:"polymorphic:Model;" json:"id"`
	ImportVersion int64 `json:"import_version"`
}
