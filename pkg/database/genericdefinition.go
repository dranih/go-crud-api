package database

import (
	"database/sql"
	"fmt"
	"strings"
)

type GenericDefinition struct {
	pdo           *LazyPdo
	driver        string
	database      string
	typeConverter *TypeConverter
	reflection    *GenericReflection
}

func NewGenericDefinition(pdo *LazyPdo, driver, database string, tables map[string]bool) *GenericDefinition {
	return &GenericDefinition{pdo, driver, database, NewTypeConverter(driver), NewGenericReflection(pdo, driver, database, tables)}
}

func (gd *GenericDefinition) quote(identifier string) string {
	switch gd.driver {
	case "mysql":
		return "`" + strings.Replace(identifier, `"`, ``, -1) + "`"
	default:
		return `"` + strings.Replace(identifier, `"`, ``, -1) + `"`
	}
}

func (gd *GenericDefinition) GetColumnType(column *ReflectedColumn, update bool) string {
	if gd.driver == "pgsql" && !update && column.GetPk() && gd.canAutoIncrement(column) {
		return "serial"
	}
	columnType := gd.typeConverter.FromJdbc(column.GetType())
	size := ""
	if column.HasPrecision() && column.HasScale() {
		size = fmt.Sprintf("(%d,%d)", column.GetPrecision(), column.GetScale())
	} else if column.HasPrecision() {
		size = fmt.Sprintf("(%d)", column.GetPrecision())
	} else if column.HasLength() {
		size = fmt.Sprintf("(%d)", column.GetLength())
	}
	null := gd.getColumnNullType(column, update)
	auto := gd.getColumnAutoIncrement(column, update)
	return fmt.Sprintf("%s%s%s%s", columnType, size, null, auto)
}

func (gd *GenericDefinition) getPrimaryKey(tableName string) string {
	pks := gd.reflection.GetTablePrimaryKeys(tableName)
	if len(pks) == 1 {
		return pks[0]
	}
	return ""
}

func (gd *GenericDefinition) canAutoIncrement(column *ReflectedColumn) bool {
	return column.GetType() == "integer" || column.GetType() == "bigint"
}

func (gd *GenericDefinition) getColumnAutoIncrement(column *ReflectedColumn, update bool) string {
	if !gd.canAutoIncrement(column) {
		return ""
	}
	switch gd.driver {
	case "mysql":
		if column.GetPk() {
			return " AUTO_INCREMENT"
		} else {
			return ""
		}
	case "pgsql":
	case "sqlsrv":
		if column.GetPk() {
			return " IDENTITY(1,1)"
		} else {
			return ""
		}
	case "sqlite":
		if column.GetPk() {
			return " AUTOINCREMENT"
		} else {
			return ""
		}
	}
	return ""
}

func (gd *GenericDefinition) getColumnNullType(column *ReflectedColumn, update bool) string {
	if gd.driver == "pgsql" && update {
		return ""
	}
	if column.GetNullable() {
		return " NULL"
	}
	return " NOT NULL"
}

func (gd *GenericDefinition) getTableRenameSQL(tableName, newTableName string) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(newTableName)
	switch gd.driver {
	case "mysql":
		return fmt.Sprintf("RENAME TABLE %s TO %s", p1, p2)
	case "pgsql":
		return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", p1, p2)
	case "sqlsrv":
		return fmt.Sprintf("EXEC sp_rename %s, %s", p1, p2)
	case "sqlite":
		return fmt.Sprintf("ALTER TABLE %s RENAME TO %s", p1, p2)
	}
	return ""
}

func (gd *GenericDefinition) getColumnRenameSQL(tableName, columnName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(columnName)
	p3 := gd.quote(newColumn.GetName())

	switch gd.driver {
	case "mysql":
		p4 := gd.GetColumnType(newColumn, true)
		return fmt.Sprintf("ALTER TABLE %s CHANGE %s %s %s", p1, p2, p3, p4)
	case "pgsql":
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", p1, p2, p3)
	case "sqlsrv":
		p4 := gd.quote(tableName + `.` + columnName)
		return fmt.Sprintf("EXEC sp_rename %s, %s, 'COLUMN'", p4, p3)
	case "sqlite":
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", p1, p2, p3)
	}
	return ""
}

func (gd *GenericDefinition) getColumnRetypeSQL(tableName, columnName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(columnName)
	p3 := gd.quote(newColumn.GetName())
	p4 := gd.GetColumnType(newColumn, true)

	switch gd.driver {
	case "mysql":
		return fmt.Sprintf("ALTER TABLE %s CHANGE %s %s %s", p1, p2, p3, p4)
	case "pgsql":
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", p1, p3, p4)
	case "sqlsrv":
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s", p1, p3, p4)
	}
	return ""
}

func (gd *GenericDefinition) getSetColumnNullableSQL(tableName, columnName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(columnName)
	p3 := gd.quote(newColumn.GetName())
	p4 := gd.GetColumnType(newColumn, true)

	switch gd.driver {
	case "mysql":
		return fmt.Sprintf("ALTER TABLE %s CHANGE %s %s %s", p1, p2, p3, p4)
	case "pgsql":
		p5 := "SET NOT NULL"
		if newColumn.GetNullable() {
			p5 = "DROP NOT NULL"
		}
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s", p1, p3, p5)
	case "sqlsrv":
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s", p1, p2, p4)
	}
	return ""
}

func (gd *GenericDefinition) getSetColumnPkConstraintSQL(tableName, columnName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(columnName)
	p3 := gd.quote(tableName + "_pkey")

	switch gd.driver {
	case "mysql":
		p4 := "DROP PRIMARY KEY"
		if newColumn.GetPk() {
			p4 = fmt.Sprintf("ADD PRIMARY KEY (%s)", p2)
		}
		return fmt.Sprintf("ALTER TABLE %s %s", p1, p4)
	case "pgsql", "sqlsrv":
		p4 := fmt.Sprintf("DROP CONSTRAINT %s", p3)
		if newColumn.GetPk() {
			p4 = fmt.Sprintf("ADD CONSTRAINT %s PRIMARY KEY (%s)", p3, p2)
		}
		return fmt.Sprintf("ALTER TABLE %s %s", p1, p4)
	}
	return ""
}

func (gd *GenericDefinition) getSetColumnPkSequenceSQL(tableName, columnName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(columnName)
	p3 := gd.quote(tableName + "_" + columnName + "_seq")

	switch gd.driver {
	case "mysql":
		return "select 1"
	case "pgsql":
		if newColumn.GetPk() {
			return fmt.Sprintf("CREATE SEQUENCE %s OWNED BY %s.%s", p3, p1, p2)
		}
		return fmt.Sprintf("DROP SEQUENCE %s", p3)
	case "sqlsrv":
		if newColumn.GetPk() {
			return fmt.Sprintf("CREATE SEQUENCE %s", p3)
		}
		return fmt.Sprintf("DROP SEQUENCE %s", p3)
	}
	return ""
}

func (gd *GenericDefinition) getSetColumnPkSequenceStartSQL(tableName, columnName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(columnName)

	switch gd.driver {
	case "mysql":
		return "select 1"
	case "pgsql":
		p3 := "'" + tableName + "_" + columnName + "_seq" + "'"
		return fmt.Sprintf("SELECT setval(%s, (SELECT max(%s)+1 FROM %s))", p3, p2, p1)
	case "sqlsrv":
		p3 := gd.quote(tableName + "_" + columnName + "_seq")
		p4Map, err := gd.pdo.Query(nil, fmt.Sprintf("SELECT max(%s)+1 FROM %s", p2, p1))
		if err != nil {
			for _, p4Val := range p4Map[0] {
				if p4, ok := p4Val.(string); ok {
					return fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH %s", p3, p4)
				}
			}
		}
	}
	return ""
}

func (gd *GenericDefinition) getSetColumnPkDefaultSQL(tableName, columnName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(columnName)

	switch gd.driver {
	case "mysql":
		p3 := gd.quote(newColumn.GetName())
		p4 := gd.GetColumnType(newColumn, true)
		return fmt.Sprintf("ALTER TABLE %s CHANGE %s %s %s", p1, p2, p3, p4)
	case "pgsql":
		var p4 string
		if newColumn.GetPk() {
			p3 := "'" + tableName + "_" + columnName + "_seq" + "'"
			p4 = fmt.Sprintf("SET DEFAULT nextval(%s)", p3)
		} else {
			p4 = "DROP DEFAULT"
		}
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s", p1, p2, p4)
	case "sqlsrv":
		p3 := gd.quote(tableName + "_" + columnName + "_seq")
		p4 := gd.quote(tableName + "_" + columnName + "_def")
		if newColumn.GetPk() {
			return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s DEFAULT NEXT VALUE FOR %s FOR %s", p1, p4, p3, p2)
		} else {
			return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", p1, p4)
		}
	}
	return ""
}

func (gd *GenericDefinition) getAddColumnFkConstraintSQL(tableName, columnName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(columnName)
	p3 := gd.quote(tableName + "_" + columnName + "_fkey")
	p4 := gd.quote(newColumn.GetFk())
	p5 := gd.quote(gd.getPrimaryKey(newColumn.GetFk()))

	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)", p1, p3, p2, p4, p5)
}

func (gd *GenericDefinition) getRemoveColumnFkConstraintSQL(tableName, columnName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(tableName + "_" + columnName + "_fkey")

	switch gd.driver {
	case "mysql":
		return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", p1, p2)
	case "pgsql", "sqlsrv":
		return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", p1, p2)
	}
	return ""
}

func (gd *GenericDefinition) getAddTableSQL(newTable *ReflectedTable) string {
	tableName := newTable.GetName()
	p1 := gd.quote(tableName)
	fields := []string{}
	constraints := []string{}
	pkColumn := gd.getPrimaryKey(tableName)
	for _, columnName := range newTable.GetColumnNames() {
		newColumn := newTable.GetColumn(columnName)
		f1 := gd.quote(columnName)
		f2 := gd.GetColumnType(newColumn, false)
		f3 := gd.quote(tableName + "_" + columnName + "_fkey")
		f4 := gd.quote(newColumn.GetFk())
		f5 := gd.quote(gd.getPrimaryKey(newColumn.GetFk()))
		f6 := gd.quote(tableName + "_" + pkColumn + "_pkey")
		if gd.driver == "sqlite" {
			if newColumn.GetPk() {
				f2 = strings.Replace(f2, "NULL", "NULL PRIMARY KEY", -1)
			}
			fields = append(fields, fmt.Sprintf("%s %s", f1, f2))
			if newColumn.GetFk() != "" {
				constraints = append(constraints, fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)", f1, f4, f5))
			}
		} else {
			fields = append(fields, fmt.Sprintf("%s %s", f1, f2))
			if newColumn.GetPk() {
				constraints = append(constraints, fmt.Sprintf("CONSTRAINT %s PRIMARY KEY (%s)", f6, f1))

			}
			if newColumn.GetFk() != "" {
				constraints = append(constraints, fmt.Sprintf("CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)", f3, f1, f4, f5))
			}
		}
	}
	p2 := strings.Join(append(fields, constraints...), ",")
	return fmt.Sprintf("CREATE TABLE %s (%s);", p1, p2)
}

func (gd *GenericDefinition) getAddColumnSQL(tableName string, newColumn *ReflectedColumn) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(newColumn.GetName())
	p3 := gd.GetColumnType(newColumn, false)

	switch gd.driver {
	case "mysql", "pgsql":
		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", p1, p2, p3)
	case "sqlsrv":
		return fmt.Sprintf("ALTER TABLE %s ADD %s %s", p1, p2, p3)
	case "sqlite":
		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", p1, p2, p3)
	}
	return ""
}

func (gd *GenericDefinition) getRemoveTableSQL(tableName string) string {
	p1 := gd.quote(tableName)

	switch gd.driver {
	case "mysql", "pgsql":
		return fmt.Sprintf("DROP TABLE %s CASCADE", p1)
	case "sqlsrv":
		return fmt.Sprintf("DROP TABLE %s", p1)
	case "sqlite":
		return fmt.Sprintf("DROP TABLE %s", p1)
	}
	return ""
}

func (gd *GenericDefinition) getRemoveColumnSQL(tableName, columnName string) string {
	p1 := gd.quote(tableName)
	p2 := gd.quote(columnName)

	switch gd.driver {
	case "mysql", "pgsql":
		return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s CASCADE", p1, p2)
	case "sqlsrv":
		return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", p1, p2)
	case "sqlite":
		return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", p1, p2)
	}
	return ""
}

func (gd *GenericDefinition) RenameTable(tableName, newTableName string) error {
	sql := gd.getTableRenameSQL(tableName, newTableName)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) RenameColumn(tableName, columnName string, newColumn *ReflectedColumn) error {
	sql := gd.getColumnRenameSQL(tableName, columnName, newColumn)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) RetypeColumn(tableName, columnName string, newColumn *ReflectedColumn) error {
	sql := gd.getColumnRetypeSQL(tableName, columnName, newColumn)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) SetColumnNullable(tableName, columnName string, newColumn *ReflectedColumn) error {
	sql := gd.getSetColumnNullableSQL(tableName, columnName, newColumn)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) AddColumnPrimaryKey(tableName, columnName string, newColumn *ReflectedColumn) error {
	sql := gd.getSetColumnPkConstraintSQL(tableName, columnName, newColumn)
	if _, err := gd.exec(sql); err != nil {
		return err
	}
	if gd.canAutoIncrement(newColumn) {
		sql = gd.getSetColumnPkSequenceSQL(tableName, columnName, newColumn)
		if _, err := gd.exec(sql); err != nil {
			return err
		}
		sql = gd.getSetColumnPkSequenceStartSQL(tableName, columnName, newColumn)
		if _, err := gd.exec(sql); err != nil {
			return err
		}
		sql = gd.getSetColumnPkDefaultSQL(tableName, columnName, newColumn)
		if _, err := gd.exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (gd *GenericDefinition) RemoveColumnPrimaryKey(tableName, columnName string, newColumn *ReflectedColumn) error {

	if gd.canAutoIncrement(newColumn) {
		sql := gd.getSetColumnPkDefaultSQL(tableName, columnName, newColumn)
		if _, err := gd.exec(sql); err != nil {
			return err
		}
		sql = gd.getSetColumnPkSequenceSQL(tableName, columnName, newColumn)
		if _, err := gd.exec(sql); err != nil {
			return err
		}
	}
	sql := gd.getSetColumnPkConstraintSQL(tableName, columnName, newColumn)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) AddColumnForeignKey(tableName, columnName string, newColumn *ReflectedColumn) error {
	sql := gd.getAddColumnFkConstraintSQL(tableName, columnName, newColumn)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) RemoveColumnForeignKey(tableName, columnName string, newColumn *ReflectedColumn) error {
	sql := gd.getRemoveColumnFkConstraintSQL(tableName, columnName, newColumn)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) AddTable(newTable *ReflectedTable) error {
	sql := gd.getAddTableSQL(newTable)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) AddColumn(tableName string, newColumn *ReflectedColumn) error {
	sql := gd.getAddColumnSQL(tableName, newColumn)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) RemoveTable(tableName string) error {
	sql := gd.getRemoveTableSQL(tableName)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) RemoveColumn(tableName, columnName string) error {
	sql := gd.getRemoveColumnSQL(tableName, columnName)
	_, err := gd.exec(sql)
	return err
}

func (gd *GenericDefinition) exec(sql string, parameters ...interface{}) (sql.Result, error) {
	res, err := gd.pdo.connect().Exec(sql, parameters...)
	if err != nil {
		return nil, err
	}
	return res, err
}
