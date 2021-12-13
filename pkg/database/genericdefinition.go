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
	return `"` + strings.Replace(identifier, `"`, ``, -1) + `"`
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
	if column.getNullable() {
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

/*

   private function getColumnRenameSQL(string $tableName, string $columnName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($columnName);
       $p3 = $this->quote($newColumn->getName());

       switch ($this->driver) {
           case 'mysql':
               $p4 = $this->getColumnType($newColumn, true);
               return "ALTER TABLE $p1 CHANGE $p2 $p3 $p4";
           case 'pgsql':
               return "ALTER TABLE $p1 RENAME COLUMN $p2 TO $p3";
           case 'sqlsrv':
               $p4 = $this->quote($tableName . '.' . $columnName);
               return "EXEC sp_rename $p4, $p3, 'COLUMN'";
           case 'sqlite':
               return "ALTER TABLE $p1 RENAME COLUMN $p2 TO $p3";
       }
   }

   private function getColumnRetypeSQL(string $tableName, string $columnName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($columnName);
       $p3 = $this->quote($newColumn->getName());
       $p4 = $this->getColumnType($newColumn, true);

       switch ($this->driver) {
           case 'mysql':
               return "ALTER TABLE $p1 CHANGE $p2 $p3 $p4";
           case 'pgsql':
               return "ALTER TABLE $p1 ALTER COLUMN $p3 TYPE $p4";
           case 'sqlsrv':
               return "ALTER TABLE $p1 ALTER COLUMN $p3 $p4";
       }
   }

   private function getSetColumnNullableSQL(string $tableName, string $columnName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($columnName);
       $p3 = $this->quote($newColumn->getName());
       $p4 = $this->getColumnType($newColumn, true);

       switch ($this->driver) {
           case 'mysql':
               return "ALTER TABLE $p1 CHANGE $p2 $p3 $p4";
           case 'pgsql':
               $p5 = $newColumn->getNullable() ? 'DROP NOT NULL' : 'SET NOT NULL';
               return "ALTER TABLE $p1 ALTER COLUMN $p2 $p5";
           case 'sqlsrv':
               return "ALTER TABLE $p1 ALTER COLUMN $p2 $p4";
       }
   }

   private function getSetColumnPkConstraintSQL(string $tableName, string $columnName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($columnName);
       $p3 = $this->quote($tableName . '_pkey');

       switch ($this->driver) {
           case 'mysql':
               $p4 = $newColumn->getPk() ? "ADD PRIMARY KEY ($p2)" : 'DROP PRIMARY KEY';
               return "ALTER TABLE $p1 $p4";
           case 'pgsql':
           case 'sqlsrv':
               $p4 = $newColumn->getPk() ? "ADD CONSTRAINT $p3 PRIMARY KEY ($p2)" : "DROP CONSTRAINT $p3";
               return "ALTER TABLE $p1 $p4";
       }
   }

   private function getSetColumnPkSequenceSQL(string $tableName, string $columnName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($columnName);
       $p3 = $this->quote($tableName . '_' . $columnName . '_seq');

       switch ($this->driver) {
           case 'mysql':
               return "select 1";
           case 'pgsql':
               return $newColumn->getPk() ? "CREATE SEQUENCE $p3 OWNED BY $p1.$p2" : "DROP SEQUENCE $p3";
           case 'sqlsrv':
               return $newColumn->getPk() ? "CREATE SEQUENCE $p3" : "DROP SEQUENCE $p3";
       }
   }

   private function getSetColumnPkSequenceStartSQL(string $tableName, string $columnName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($columnName);

       switch ($this->driver) {
           case 'mysql':
               return "select 1";
           case 'pgsql':
               $p3 = $this->pdo->quote($tableName . '_' . $columnName . '_seq');
               return "SELECT setval($p3, (SELECT max($p2)+1 FROM $p1));";
           case 'sqlsrv':
               $p3 = $this->quote($tableName . '_' . $columnName . '_seq');
               $p4 = $this->pdo->query("SELECT max($p2)+1 FROM $p1")->fetchColumn();
               return "ALTER SEQUENCE $p3 RESTART WITH $p4";
       }
   }

   private function getSetColumnPkDefaultSQL(string $tableName, string $columnName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($columnName);

       switch ($this->driver) {
           case 'mysql':
               $p3 = $this->quote($newColumn->getName());
               $p4 = $this->getColumnType($newColumn, true);
               return "ALTER TABLE $p1 CHANGE $p2 $p3 $p4";
           case 'pgsql':
               if ($newColumn->getPk()) {
                   $p3 = $this->pdo->quote($tableName . '_' . $columnName . '_seq');
                   $p4 = "SET DEFAULT nextval($p3)";
               } else {
                   $p4 = 'DROP DEFAULT';
               }
               return "ALTER TABLE $p1 ALTER COLUMN $p2 $p4";
           case 'sqlsrv':
               $p3 = $this->quote($tableName . '_' . $columnName . '_seq');
               $p4 = $this->quote($tableName . '_' . $columnName . '_def');
               if ($newColumn->getPk()) {
                   return "ALTER TABLE $p1 ADD CONSTRAINT $p4 DEFAULT NEXT VALUE FOR $p3 FOR $p2";
               } else {
                   return "ALTER TABLE $p1 DROP CONSTRAINT $p4";
               }
       }
   }

   private function getAddColumnFkConstraintSQL(string $tableName, string $columnName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($columnName);
       $p3 = $this->quote($tableName . '_' . $columnName . '_fkey');
       $p4 = $this->quote($newColumn->getFk());
       $p5 = $this->quote($this->getPrimaryKey($newColumn->getFk()));

       return "ALTER TABLE $p1 ADD CONSTRAINT $p3 FOREIGN KEY ($p2) REFERENCES $p4 ($p5)";
   }

   private function getRemoveColumnFkConstraintSQL(string $tableName, string $columnName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($tableName . '_' . $columnName . '_fkey');

       switch ($this->driver) {
           case 'mysql':
               return "ALTER TABLE $p1 DROP FOREIGN KEY $p2";
           case 'pgsql':
           case 'sqlsrv':
               return "ALTER TABLE $p1 DROP CONSTRAINT $p2";
       }
   }

   private function getAddTableSQL(ReflectedTable $newTable): string
   {
       $tableName = $newTable->getName();
       $p1 = $this->quote($tableName);
       $fields = [];
       $constraints = [];
       foreach ($newTable->getColumnNames() as $columnName) {
           $pkColumn = $this->getPrimaryKey($tableName);
           $newColumn = $newTable->getColumn($columnName);
           $f1 = $this->quote($columnName);
           $f2 = $this->getColumnType($newColumn, false);
           $f3 = $this->quote($tableName . '_' . $columnName . '_fkey');
           $f4 = $this->quote($newColumn->getFk());
           $f5 = $this->quote($this->getPrimaryKey($newColumn->getFk()));
           $f6 = $this->quote($tableName . '_' . $pkColumn . '_pkey');
           if ($this->driver == 'sqlite') {
               if ($newColumn->getPk()) {
                   $f2 = str_replace('NULL', 'NULL PRIMARY KEY', $f2);
               }
               $fields[] = "$f1 $f2";
               if ($newColumn->getFk()) {
                   $constraints[] = "FOREIGN KEY ($f1) REFERENCES $f4 ($f5)";
               }
           } else {
               $fields[] = "$f1 $f2";
               if ($newColumn->getPk()) {
                   $constraints[] = "CONSTRAINT $f6 PRIMARY KEY ($f1)";
               }
               if ($newColumn->getFk()) {
                   $constraints[] = "CONSTRAINT $f3 FOREIGN KEY ($f1) REFERENCES $f4 ($f5)";
               }
           }
       }
       $p2 = implode(',', array_merge($fields, $constraints));

       return "CREATE TABLE $p1 ($p2);";
   }

   private function getAddColumnSQL(string $tableName, ReflectedColumn $newColumn): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($newColumn->getName());
       $p3 = $this->getColumnType($newColumn, false);

       switch ($this->driver) {
           case 'mysql':
           case 'pgsql':
               return "ALTER TABLE $p1 ADD COLUMN $p2 $p3";
           case 'sqlsrv':
               return "ALTER TABLE $p1 ADD $p2 $p3";
           case 'sqlite':
               return "ALTER TABLE $p1 ADD COLUMN $p2 $p3";
       }
   }

   private function getRemoveTableSQL(string $tableName): string
   {
       $p1 = $this->quote($tableName);

       switch ($this->driver) {
           case 'mysql':
           case 'pgsql':
               return "DROP TABLE $p1 CASCADE;";
           case 'sqlsrv':
               return "DROP TABLE $p1;";
           case 'sqlite':
               return "DROP TABLE $p1;";
       }
   }

   private function getRemoveColumnSQL(string $tableName, string $columnName): string
   {
       $p1 = $this->quote($tableName);
       $p2 = $this->quote($columnName);

       switch ($this->driver) {
           case 'mysql':
           case 'pgsql':
               return "ALTER TABLE $p1 DROP COLUMN $p2 CASCADE;";
           case 'sqlsrv':
               return "ALTER TABLE $p1 DROP COLUMN $p2;";
           case 'sqlite':
               return "ALTER TABLE $p1 DROP COLUMN $p2;";
       }
   }
*/
func (gd *GenericDefinition) RenameTable(tableName, newTableName string) error {
	sql := gd.getTableRenameSQL(tableName, newTableName)
	_, err := gd.exec(sql)
	return err
}

/*
   public function renameColumn(string $tableName, string $columnName, ReflectedColumn $newColumn)
   {
       $sql = $this->getColumnRenameSQL($tableName, $columnName, $newColumn);
       return $this->query($sql, []);
   }

   public function retypeColumn(string $tableName, string $columnName, ReflectedColumn $newColumn)
   {
       $sql = $this->getColumnRetypeSQL($tableName, $columnName, $newColumn);
       return $this->query($sql, []);
   }

   public function setColumnNullable(string $tableName, string $columnName, ReflectedColumn $newColumn)
   {
       $sql = $this->getSetColumnNullableSQL($tableName, $columnName, $newColumn);
       return $this->query($sql, []);
   }

   public function addColumnPrimaryKey(string $tableName, string $columnName, ReflectedColumn $newColumn)
   {
       $sql = $this->getSetColumnPkConstraintSQL($tableName, $columnName, $newColumn);
       $this->query($sql, []);
       if ($this->canAutoIncrement($newColumn)) {
           $sql = $this->getSetColumnPkSequenceSQL($tableName, $columnName, $newColumn);
           $this->query($sql, []);
           $sql = $this->getSetColumnPkSequenceStartSQL($tableName, $columnName, $newColumn);
           $this->query($sql, []);
           $sql = $this->getSetColumnPkDefaultSQL($tableName, $columnName, $newColumn);
           $this->query($sql, []);
       }
       return true;
   }

   public function removeColumnPrimaryKey(string $tableName, string $columnName, ReflectedColumn $newColumn)
   {
       if ($this->canAutoIncrement($newColumn)) {
           $sql = $this->getSetColumnPkDefaultSQL($tableName, $columnName, $newColumn);
           $this->query($sql, []);
           $sql = $this->getSetColumnPkSequenceSQL($tableName, $columnName, $newColumn);
           $this->query($sql, []);
       }
       $sql = $this->getSetColumnPkConstraintSQL($tableName, $columnName, $newColumn);
       $this->query($sql, []);
       return true;
   }

   public function addColumnForeignKey(string $tableName, string $columnName, ReflectedColumn $newColumn)
   {
       $sql = $this->getAddColumnFkConstraintSQL($tableName, $columnName, $newColumn);
       return $this->query($sql, []);
   }

   public function removeColumnForeignKey(string $tableName, string $columnName, ReflectedColumn $newColumn)
   {
       $sql = $this->getRemoveColumnFkConstraintSQL($tableName, $columnName, $newColumn);
       return $this->query($sql, []);
   }

   public function addTable(ReflectedTable $newTable)
   {
       $sql = $this->getAddTableSQL($newTable);
       return $this->query($sql, []);
   }

   public function addColumn(string $tableName, ReflectedColumn $newColumn)
   {
       $sql = $this->getAddColumnSQL($tableName, $newColumn);
       return $this->query($sql, []);
   }

   public function removeTable(string $tableName)
   {
       $sql = $this->getRemoveTableSQL($tableName);
       return $this->query($sql, []);
   }

   public function removeColumn(string $tableName, string $columnName)
   {
       $sql = $this->getRemoveColumnSQL($tableName, $columnName);
       return $this->query($sql, []);
   }

   private function query(string $sql, array $arguments): bool
   {
       $stmt = $this->pdo->prepare($sql);
       // echo "- $sql -- " . json_encode($arguments) . "\n";
       return $stmt->execute($arguments);
   }
*/
func (gd *GenericDefinition) exec(sql string, parameters ...interface{}) (sql.Result, error) {
	res, err := gd.pdo.connect().Exec(sql, parameters...)
	if err != nil {
		return nil, err
	}
	return res, err
}
