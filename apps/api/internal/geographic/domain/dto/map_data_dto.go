package dto

// MapDataRequest represents the request for geographic map data
type MapDataRequest struct {
	Level      string `form:"level" binding:"required,oneof=province city district"`
	ProvinceID string `form:"province_id" binding:"omitempty,uuid"`
	CityID     string `form:"city_id" binding:"omitempty,uuid"`
}

// GeoJSONFeature represents a single GeoJSON feature
type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   interface{}            `json:"geometry"`
}

// GeoJSONFeatureCollection represents a GeoJSON FeatureCollection
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}
