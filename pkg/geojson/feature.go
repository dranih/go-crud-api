package geojson

import "encoding/json"

type Feature struct {
	id         int
	properties map[string]interface{}
	geometry   *Geometry
}

func NewFeature(id int, properties map[string]interface{}, geometry *Geometry) *Feature {
	return &Feature{id, properties, geometry}
}

// json marshaling for struct Feature
func (f *Feature) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		map[string]interface{}{
			"type":       "Feature",
			"id":         f.id,
			"properties": f.properties,
			"geometry":   f.geometry,
		})
}

type FeatureCollection struct {
	features []*Feature
	results  int
}

func NewFeatureCollection(features []*Feature, results int) *FeatureCollection {
	return &FeatureCollection{features, results}
}

// json marshaling for struct FeatureCollection
func (fc *FeatureCollection) MarshalJSON() ([]byte, error) {
	if fc.results == -1 {
		return json.Marshal(
			map[string]interface{}{
				"type":     "FeatureCollection",
				"features": fc.features,
			})
	} else {
		return json.Marshal(
			map[string]interface{}{
				"type":     "FeatureCollection",
				"features": fc.features,
				"results":  fc.results,
			})
	}
}
