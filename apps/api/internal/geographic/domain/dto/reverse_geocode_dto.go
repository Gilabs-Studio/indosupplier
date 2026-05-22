package dto

// ReverseGeocodeRequest represents the request for reverse geocoding
type ReverseGeocodeRequest struct {
	Latitude  float64 `form:"lat" binding:"required,min=-90,max=90"`
	Longitude float64 `form:"lng" binding:"required,min=-180,max=180"`
}

// ReverseGeocodeResult represents a single level of resolved administrative boundary
type ReverseGeocodeResult struct {
	ProvinceID   string `json:"province_id"`
	ProvinceName string `json:"province_name"`
	CityID       string `json:"city_id"`
	CityName     string `json:"city_name"`
	CityType     string `json:"city_type"`
	DistrictID   string `json:"district_id"`
	DistrictName string `json:"district_name"`
}
