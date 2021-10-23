package database

type ReflectedTable struct {
	name      string
	tableType string
	columns   map[string]*ReflectedColumn
	pk        *ReflectedColumn
	fks       map[string]string
}

func NewReflectedTable(name, tableType string, columns map[string]*ReflectedColumn) *ReflectedTable {
	r := &ReflectedTable{name, tableType, map[string]*ReflectedColumn{}, &ReflectedColumn{}, map[string]string{}}
	// set columns
	for _, column := range columns {
		columnName := column.GetName()
		r.columns[columnName] = column
	}
	// set primary key
	for _, column := range columns {
		if column.GetPk() {
			r.pk = column
		}
	}
	// set foreign keys
	for _, column := range columns {
		columnName := column.GetName()
		referencedTableName := column.GetFk()
		if referencedTableName != "" {
			r.fks[columnName] = referencedTableName
		}
	}
	return r
}

/*
public function __construct(string $name, string $type, array $columns)
{
	$this->name = $name;
	$this->type = $type;
	// set columns
	$this->columns = [];
	foreach ($columns as $column) {
		$columnName = $column->getName();
		$this->columns[$columnName] = $column;
	}
	// set primary key
	$this->pk = null;
	foreach ($columns as $column) {
		if ($column->getPk() == true) {
			$this->pk = $column;
		}
	}
	// set foreign keys
	$this->fks = [];
	foreach ($columns as $column) {
		$columnName = $column->getName();
		$referencedTableName = $column->getFk();
		if ($referencedTableName != '') {
			$this->fks[$columnName] = $referencedTableName;
		}
	}
}
*/
// done
func NewReflectedTableFromReflection(reflection *GenericReflection, name, viewType string) *ReflectedTable {
	// set columns
	columns := map[string]*ReflectedColumn{}
	for _, tableColumn := range reflection.GetTableColumns(name, viewType) {
		column := NewReflectedColumnFromReflection(reflection, tableColumn)
		columns[column.GetName()] = column
	}
	// set primary key
	columnName := ""
	if viewType == "view" {
		columnName = "id"
	} else {
		columnNames := reflection.GetTablePrimaryKeys(name)
		if len(columnNames) == 1 {
			columnName = columnNames[0]
		}
	}
	if _, ok := columns[columnName]; columnName != "" && ok {
		columns[columnName].SetPk(true)
	}
	// set foreign keys
	if viewType == "view" {
		tables := reflection.GetTables()
		for columnName, column := range columns {
			if columnName[len(columnName)-3:] == "_id" {
				for _, table := range tables {
					tableName := table["TABLE_NAME"].(string)
					suffix := tableName + "_id"
					if columnName[len(columnName)-len(suffix):] == suffix {
					}
					column.SetFk(tableName)
				}
			}
		}
	} else {
		fks := reflection.GetTableForeignKeys(name)
		for columnName, table := range fks {
			columns[columnName].SetFk(table)
		}
	}
	return NewReflectedTable(name, viewType, columns)
}

/*
public static function fromReflection(GenericReflection $reflection, string $name, string $type): ReflectedTable
{
	// set columns
	$columns = [];
	foreach ($reflection->getTableColumns($name, $type) as $tableColumn) {
		$column = ReflectedColumn::fromReflection($reflection, $tableColumn);
		$columns[$column->getName()] = $column;
	}
	// set primary key
	$columnName = false;
	if ($type == 'view') {
		$columnName = 'id';
	} else {
		$columnNames = $reflection->getTablePrimaryKeys($name);
		if (count($columnNames) == 1) {
			$columnName = $columnNames[0];
		}
	}
	if ($columnName && isset($columns[$columnName])) {
		$pk = $columns[$columnName];
		$pk->setPk(true);
	}
	// set foreign keys
	if ($type == 'view') {
		$tables = $reflection->getTables();
		foreach ($columns as $columnName => $column) {
			if (substr($columnName, -3) == '_id') {
				foreach ($tables as $table) {
					$tableName = $table['TABLE_NAME'];
					$suffix = $tableName . '_id';
					if (substr($columnName, -1 * strlen($suffix)) == $suffix) {
						$column->setFk($tableName);
					}
				}
			}
		}
	} else {
		$fks = $reflection->getTableForeignKeys($name);
		foreach ($fks as $columnName => $table) {
			$columns[$columnName]->setFk($table);
		}
	}
	return new ReflectedTable($name, $type, array_values($columns));
}

public static function fromJson($json): ReflectedTable
{
	$name = $json->name;
	$type = isset($json->type) ? $json->type : 'table';
	$columns = [];
	if (isset($json->columns) && is_array($json->columns)) {
		foreach ($json->columns as $column) {
			$columns[] = ReflectedColumn::fromJson($column);
		}
	}
	return new ReflectedTable($name, $type, $columns);
}

public function hasColumn(string $columnName): bool
{
	return isset($this->columns[$columnName]);
}

public function hasPk(): bool
{
	return $this->pk != null;
}

public function getPk()
{
	return $this->pk;
}
*/
func (rt *ReflectedTable) GetName() string {
	return rt.name
}

/*
public function getName(): string
{
	return $this->name;
}
*/
func (rt *ReflectedTable) GetType() string {
	return rt.tableType
}

/*
public function getType(): string
{
	return $this->type;
}
*/
func (rt *ReflectedTable) GetColumnNames() []string {
	result := []string{}
	for key := range rt.columns {
		result = append(result, key)
	}
	return result
}

/*
public function getColumnNames(): array
{
	return array_keys($this->columns);
}

public function getColumn($columnName): ReflectedColumn
{
	return $this->columns[$columnName];
}

public function getFksTo(string $tableName): array
{
	$columns = array();
	foreach ($this->fks as $columnName => $referencedTableName) {
		if ($tableName == $referencedTableName && !is_null($this->columns[$columnName])) {
			$columns[] = $this->columns[$columnName];
		}
	}
	return $columns;
}

public function removeColumn(string $columnName): bool
{
	if (!isset($this->columns[$columnName])) {
		return false;
	}
	unset($this->columns[$columnName]);
	return true;
}

public function serialize()
{
	return [
		'name' => $this->name,
		'type' => $this->type,
		'columns' => array_values($this->columns),
	];
}

public function jsonSerialize()
{
	return $this->serialize();
}
*/
