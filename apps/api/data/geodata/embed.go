package geodata

import _ "embed"

// IndonesiaProvinces contains the simplified GeoJSON for Indonesian provinces (old, 32 province polygons)
//
//go:embed indonesia-provinces.json
var IndonesiaProvinces []byte

// IndonesiaProvincesSimple contains village-level GeoJSON with province/city/district metadata
// Properties: WADMPR (province), WADMKK (city/regency), WADMKC (district), WADMKD (village)
// Codes: KDPPUM (province BPS), KDPKAB (city BPS), KDCPUM (district BPS)
//
//go:embed indonesia-provinces-simple.geojson
var IndonesiaProvincesSimple []byte
