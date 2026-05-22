package models

// TimeZone represents the timezone data from TimezoneDB
type TimeZone struct {
	ZoneName     string `gorm:"column:zone_name;type:varchar(35);index:idx_zone_name"`
	CountryCode  string `gorm:"column:country_code;type:char(2);index:idx_country_code"`
	Abbreviation string `gorm:"column:abbreviation;type:varchar(6)"`
	TimeStart    int64  `gorm:"column:time_start;type:decimal(11,0);index:idx_time_start"`
	GMTOffset    int    `gorm:"column:gmt_offset;type:int"`
	DST          string `gorm:"column:dst;type:char(1)"`
}

// TableName specifies the table name for TimeZone
func (TimeZone) TableName() string {
	return "time_zones"
}

// Country represents country data from TimezoneDB
type Country struct {
	CountryCode string `gorm:"column:country_code;type:char(2);index:idx_country_code"`
	CountryName string `gorm:"column:country_name;type:varchar(45)"`
}

// TableName specifies the table name for Country
func (Country) TableName() string {
	return "countries"
}

// TimezoneInfo holds processed timezone information for a location
type TimezoneInfo struct {
	ZoneName     string
	CountryCode  string
	CountryName  string
	Abbreviation string
	GMTOffset    int // in seconds
	IsDST        bool
}
