package geojson

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
)

type Geometry struct {
	geoType     string
	coordinates [][2]float64
}

func NewGeometry(geoType string, coordinates [][2]float64) *Geometry {
	return &Geometry{geoType, coordinates}
}

func NewGeometryFromWkt(wkt string) *Geometry {
	geoTypes := []string{
		"Point",
		"MultiPoint",
		"LineString",
		"MultiLineString",
		"Polygon",
		"MultiPolygon",
		//"GeometryCollection",
	}
	bracket := strings.Index(wkt, "(")
	geoType := strings.ToUpper(strings.TrimSpace(wkt[0:bracket]))
	supported := false
	for _, typeName := range geoTypes {
		if strings.ToUpper(typeName) == geoType {
			geoType = typeName
			supported = true
		}
	}
	if !supported {
		log.Printf("Geometry type not supported : %s", geoType)
		return nil
	}
	coordinates := wkt[bracket:]
	if !strings.HasSuffix(geoType, "Point") || (geoType == "MultiPoint" && coordinates[1:1] != "(") {
		re := regexp.MustCompile(`([0-9\-\.]+ )+([0-9\-\.]+)`)
		coordinates = re.ReplaceAllString(coordinates, "[${1}${2}]")
	}
	re := strings.NewReplacer("(", "[", ")", "]", ", ", ",", " ", ",")
	coordinates = re.Replace(coordinates)
	var coord [][2]float64
	if err := json.Unmarshal([]byte(coordinates), &coord); err != nil {
		log.Printf("Could not decode WKT %s : %s", coordinates, err)
		return nil
	} else {
		return &Geometry{geoType, coord}
	}
}

// json marshaling for struct Geometry
func (g *Geometry) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		map[string]interface{}{
			"type":        g.geoType,
			"coordinates": g.coordinates,
		})
}
