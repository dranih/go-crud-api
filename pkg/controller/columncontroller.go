package controller

import "github.com/dranih/go-crud-api/pkg/database"

type ColumnController struct {
	responder  Responder
	reflection database.ReflectionService
	definition database.DefinitionService
}
