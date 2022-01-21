package geojson

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dranih/go-crud-api/pkg/database"
)

type Service struct {
	reflection *database.ReflectionService
	records    *database.RecordService
}

func NewGeoJsonService(reflection *database.ReflectionService, records *database.RecordService) *Service {
	return &Service{reflection, records}
}

func (s *Service) HasTable(table string) bool {
	return s.reflection.HasTable(table)
}

func (s *Service) GetType(table string) string {
	return s.reflection.GetType(table)
}

func (s *Service) getGeometryColumnName(tableName string, params *map[string][]string) string {
	var geometryParam, geometryColumnName string
	if param, exists := (*params)["geometry"]; exists && len(param) > 0 {
		geometryParam = param[0]
	}
	table := s.reflection.GetTable(tableName)
	for _, columnName := range table.GetColumnNames() {
		if geometryParam != "" && geometryParam != columnName {
			continue
		}
		column := table.GetColumn(columnName)
		if column.IsGeometry() {
			geometryColumnName = columnName
			break
		}
	}
	if geometryColumnName != "" {
		(*params)["mandatory"] = append((*params)["mandatory"], fmt.Sprintf("%s.%s", tableName, geometryColumnName))
	}
	return geometryColumnName
}

func (s *Service) setBoudingBoxFilter(geometryColumnName string, params *map[string][]string) {
	if param, exists := (*params)["bbox"]; exists && len(param) > 0 {
		c := strings.Split(param[0], ",")
		if _, exists := (*params)["filter"]; !exists {
			(*params)["filter"] = []string{}
		}
		(*params)["filter"] = append((*params)["filter"], fmt.Sprintf("%s,sin,POLYGON((%s %s,%s %s,%s %s,%s %s,%s %s))", geometryColumnName, c[0], c[1], c[2], c[1], c[2], c[3], c[0], c[3], c[0], c[1]))
	}
	if param, exists := (*params)["tile"]; exists && len(param) > 0 {
		if zxy := strings.Split(param[0], ","); len(zxy) == 3 {
			c := []float64{}
			z, err1 := strconv.ParseFloat(zxy[0], 64)
			x, err2 := strconv.ParseFloat(zxy[1], 64)
			y, err3 := strconv.ParseFloat(zxy[2], 64)
			if err1 == nil && err2 == nil && err3 == nil {
				c = append(c, s.convertTileToLatLonOfUpperLeftCorner(z, x, y)...)
				c = append(c, s.convertTileToLatLonOfUpperLeftCorner(z, x+1, y+1)...)
				(*params)["filter"] = append((*params)["filter"], fmt.Sprintf("%s,sin,POLYGON((%f %f,%f %f,%f %f,%f %f,%f %f))", geometryColumnName, c[0], c[1], c[2], c[1], c[2], c[3], c[0], c[3], c[0], c[1]))
			}
		}
	}
}

func (s *Service) convertTileToLatLonOfUpperLeftCorner(z, x, y float64) []float64 {
	n := z * z
	lon := x/n*360 - 180
	lat := (180 / math.Pi) * math.Atan(math.Sinh(math.Pi*(1-2*y/n)))
	return []float64{lon, lat}
}

func (s *Service) convertRecordToFeature(record interface{}, primaryKeyColumnName, geometryColumnName string) *Feature {
	var id int
	if recordMap, ok := record.(map[string]interface{}); ok {
		if primaryKeyColumnName != "" {
			if v1, exists := recordMap[primaryKeyColumnName]; exists {
				if v, err := strconv.Atoi(fmt.Sprint(v1)); err == nil {
					id = v
					delete(recordMap, primaryKeyColumnName)
				}
			}
		}
		var geometry *Geometry
		if v, exists := recordMap[geometryColumnName]; exists {
			geometry = NewGeometryFromWkt(fmt.Sprint(v))
			delete(recordMap, geometryColumnName)
		}
		return &Feature{id, recordMap, geometry}
	} else {
		return nil
	}
}

func (s *Service) getPrimaryKeyColumnName(tableName string, params *map[string][]string) string {
	primaryKeyColumn := s.reflection.GetTable(tableName).GetPk()
	if primaryKeyColumn == nil {
		return ""
	}
	primaryKeyColumnName := primaryKeyColumn.GetName()
	(*params)["mandatory"] = append((*params)["mandatory"], fmt.Sprintf("%s.%s", tableName, primaryKeyColumnName))
	return primaryKeyColumnName
}

func (s *Service) List(tableName string, params map[string][]string) *FeatureCollection {
	geometryColumnName := s.getGeometryColumnName(tableName, &params)
	s.setBoudingBoxFilter(geometryColumnName, &params)
	primaryColumnName := s.getPrimaryKeyColumnName(tableName, &params)
	records := s.records.List(tableName, params)
	var features []*Feature
	for _, record := range records.GetRecords() {
		features = append(features, s.convertRecordToFeature(record, primaryColumnName, geometryColumnName))
	}
	return NewFeatureCollection(features, records.GetResults())
}

func (s *Service) Read(tableName, id string, params map[string][]string) *Feature {
	geometryColumnName := s.getGeometryColumnName(tableName, &params)
	primaryColumnName := s.getPrimaryKeyColumnName(tableName, &params)
	if record, err := s.records.Read(tableName, params, id); err == nil {
		return s.convertRecordToFeature(record, primaryColumnName, geometryColumnName)
	}
	return nil
}
