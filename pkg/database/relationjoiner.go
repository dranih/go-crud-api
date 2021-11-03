package database

import "strings"

type RelationJoiner struct {
	reflection *ReflectionService
	ordering   *OrderingInfo
	columns    *ColumnIncluder
}

func NewRelationJoiner(reflection *ReflectionService, columns *ColumnIncluder) *RelationJoiner {
	return &RelationJoiner{reflection, &OrderingInfo{}, columns}
}

func (rj *RelationJoiner) AddMandatoryColumns(table *ReflectedTable, params *map[string][]string) {
	_, exists1 := (*params)["join"]
	_, exists2 := (*params)["include"]
	if !exists1 || !exists2 {
		return
	}
	(*params)["mandatory"] = []string{}
	for _, tableNames := range (*params)["join"] {
		t1 := table
		for _, tableName := range strings.Split(tableNames, ",") {
			if !rj.reflection.HasTable(tableName) {
				continue
			}
			t2 := rj.reflection.GetTable(tableName)
			fks1 := t1.GetFksTo(t2.GetName())
			t3 := rj.hasAndBelongsToMany(t1, t2)
			if t3 != nil && len(fks1) > 0 {
				(*params)["mandatory"] = append((*params)["mandatory"], t2.GetName()+"."+t2.GetPk().GetName())
			}
			for _, fk := range fks1 {
				(*params)["mandatory"] = append((*params)["mandatory"], t1.GetName()+"."+fk.GetName())
			}
			fks2 := t2.GetFksTo(t1.GetName())
			if t3 != nil && len(fks2) > 0 {
				(*params)["mandatory"] = append((*params)["mandatory"], t1.GetName()+"."+t1.GetPk().GetName())
			}
			for _, fk := range fks2 {
				(*params)["mandatory"] = append((*params)["mandatory"], t2.GetName()+"."+fk.GetName())
			}
			t1 = t2
		}
	}
}

/*
	private function getJoinsAsPathTree(array $params): PathTree
	{
		$joins = new PathTree();
		if (isset($params['join'])) {
			foreach ($params['join'] as $tableNames) {
				$path = array();
				foreach (explode(',', $tableNames) as $tableName) {
					if (!$this->reflection->hasTable($tableName)) {
						continue;
					}
					$t = $this->reflection->getTable($tableName);
					if ($t != null) {
						$path[] = $t->getName();
					}
				}
				$joins->put($path, true);
			}
		}
		return $joins;
	}

	public function addJoins(ReflectedTable $table, array &$records, array $params, GenericDB $db)
	{
		$joins = $this->getJoinsAsPathTree($params);
		$this->addJoinsForTables($table, $joins, $records, $params, $db);
	}
*/
func (rj *RelationJoiner) hasAndBelongsToMany(t1, t2 *ReflectedTable) *ReflectedTable {
	for _, tableName := range rj.reflection.GetTableNames() {
		t3 := rj.reflection.GetTable(tableName)
		if len(t3.GetFksTo(t1.GetName())) > 0 && len(t3.GetFksTo(t2.GetName())) > 0 {
			return t3
		}
	}
	return nil
}

/*
	private function hasAndBelongsToMany(ReflectedTable $t1, ReflectedTable $t2)
	{
		foreach ($this->reflection->getTableNames() as $tableName) {
			$t3 = $this->reflection->getTable($tableName);
			if (count($t3->getFksTo($t1->getName())) > 0 && count($t3->getFksTo($t2->getName())) > 0) {
				return $t3;
			}
		}
		return null;
	}

	private function addJoinsForTables(ReflectedTable $t1, PathTree $joins, array &$records, array $params, GenericDB $db)
	{
		foreach ($joins->getKeys() as $t2Name) {
			$t2 = $this->reflection->getTable($t2Name);

			$belongsTo = count($t1->getFksTo($t2->getName())) > 0;
			$hasMany = count($t2->getFksTo($t1->getName())) > 0;
			if (!$belongsTo && !$hasMany) {
				$t3 = $this->hasAndBelongsToMany($t1, $t2);
			} else {
				$t3 = null;
			}
			$hasAndBelongsToMany = ($t3 != null);

			$newRecords = array();
			$fkValues = null;
			$pkValues = null;
			$habtmValues = null;

			if ($belongsTo) {
				$fkValues = $this->getFkEmptyValues($t1, $t2, $records);
				$this->addFkRecords($t2, $fkValues, $params, $db, $newRecords);
			}
			if ($hasMany) {
				$pkValues = $this->getPkEmptyValues($t1, $records);
				$this->addPkRecords($t1, $t2, $pkValues, $params, $db, $newRecords);
			}
			if ($hasAndBelongsToMany) {
				$habtmValues = $this->getHabtmEmptyValues($t1, $t2, $t3, $db, $records);
				$this->addFkRecords($t2, $habtmValues->fkValues, $params, $db, $newRecords);
			}

			$this->addJoinsForTables($t2, $joins->get($t2Name), $newRecords, $params, $db);

			if ($fkValues != null) {
				$this->fillFkValues($t2, $newRecords, $fkValues);
				$this->setFkValues($t1, $t2, $records, $fkValues);
			}
			if ($pkValues != null) {
				$this->fillPkValues($t1, $t2, $newRecords, $pkValues);
				$this->setPkValues($t1, $t2, $records, $pkValues);
			}
			if ($habtmValues != null) {
				$this->fillFkValues($t2, $newRecords, $habtmValues->fkValues);
				$this->setHabtmValues($t1, $t2, $records, $habtmValues);
			}
		}
	}

	private function getFkEmptyValues(ReflectedTable $t1, ReflectedTable $t2, array $records): array
	{
		$fkValues = array();
		$fks = $t1->getFksTo($t2->getName());
		foreach ($fks as $fk) {
			$fkName = $fk->getName();
			foreach ($records as $record) {
				if (isset($record[$fkName])) {
					$fkValue = $record[$fkName];
					$fkValues[$fkValue] = null;
				}
			}
		}
		return $fkValues;
	}

	private function addFkRecords(ReflectedTable $t2, array $fkValues, array $params, GenericDB $db, array &$records)
	{
		$columnNames = $this->columns->getNames($t2, false, $params);
		$fkIds = array_keys($fkValues);

		foreach ($db->selectMultiple($t2, $columnNames, $fkIds) as $record) {
			$records[] = $record;
		}
	}

	private function fillFkValues(ReflectedTable $t2, array $fkRecords, array &$fkValues)
	{
		$pkName = $t2->getPk()->getName();
		foreach ($fkRecords as $fkRecord) {
			$pkValue = $fkRecord[$pkName];
			$fkValues[$pkValue] = $fkRecord;
		}
	}

	private function setFkValues(ReflectedTable $t1, ReflectedTable $t2, array &$records, array $fkValues)
	{
		$fks = $t1->getFksTo($t2->getName());
		foreach ($fks as $fk) {
			$fkName = $fk->getName();
			foreach ($records as $i => $record) {
				if (isset($record[$fkName])) {
					$key = $record[$fkName];
					$records[$i][$fkName] = $fkValues[$key];
				}
			}
		}
	}

	private function getPkEmptyValues(ReflectedTable $t1, array $records): array
	{
		$pkValues = array();
		$pkName = $t1->getPk()->getName();
		foreach ($records as $record) {
			$key = $record[$pkName];
			$pkValues[$key] = array();
		}
		return $pkValues;
	}

	private function addPkRecords(ReflectedTable $t1, ReflectedTable $t2, array $pkValues, array $params, GenericDB $db, array &$records)
	{
		$fks = $t2->getFksTo($t1->getName());
		$columnNames = $this->columns->getNames($t2, false, $params);
		$pkValueKeys = implode(',', array_keys($pkValues));
		$conditions = array();
		foreach ($fks as $fk) {
			$conditions[] = new ColumnCondition($fk, 'in', $pkValueKeys);
		}
		$condition = OrCondition::fromArray($conditions);
		$columnOrdering = array();
		$limit = VariableStore::get("joinLimits.maxRecords") ?: -1;
		if ($limit != -1) {
			$columnOrdering = $this->ordering->getDefaultColumnOrdering($t2);
		}
		foreach ($db->selectAll($t2, $columnNames, $condition, $columnOrdering, 0, $limit) as $record) {
			$records[] = $record;
		}
	}

	private function fillPkValues(ReflectedTable $t1, ReflectedTable $t2, array $pkRecords, array &$pkValues)
	{
		$fks = $t2->getFksTo($t1->getName());
		foreach ($fks as $fk) {
			$fkName = $fk->getName();
			foreach ($pkRecords as $pkRecord) {
				$key = $pkRecord[$fkName];
				if (isset($pkValues[$key])) {
					$pkValues[$key][] = $pkRecord;
				}
			}
		}
	}

	private function setPkValues(ReflectedTable $t1, ReflectedTable $t2, array &$records, array $pkValues)
	{
		$pkName = $t1->getPk()->getName();
		$t2Name = $t2->getName();

		foreach ($records as $i => $record) {
			$key = $record[$pkName];
			$records[$i][$t2Name] = $pkValues[$key];
		}
	}

	private function getHabtmEmptyValues(ReflectedTable $t1, ReflectedTable $t2, ReflectedTable $t3, GenericDB $db, array $records): HabtmValues
	{
		$pkValues = $this->getPkEmptyValues($t1, $records);
		$fkValues = array();

		$fk1 = $t3->getFksTo($t1->getName())[0];
		$fk2 = $t3->getFksTo($t2->getName())[0];

		$fk1Name = $fk1->getName();
		$fk2Name = $fk2->getName();

		$columnNames = array($fk1Name, $fk2Name);

		$pkIds = implode(',', array_keys($pkValues));
		$condition = new ColumnCondition($t3->getColumn($fk1Name), 'in', $pkIds);
		$columnOrdering = array();

		$limit = VariableStore::get("joinLimits.maxRecords") ?: -1;
		if ($limit != -1) {
			$columnOrdering = $this->ordering->getDefaultColumnOrdering($t3);
		}
		$records = $db->selectAll($t3, $columnNames, $condition, $columnOrdering, 0, $limit);
		foreach ($records as $record) {
			$val1 = $record[$fk1Name];
			$val2 = $record[$fk2Name];
			$pkValues[$val1][] = $val2;
			$fkValues[$val2] = null;
		}

		return new HabtmValues($pkValues, $fkValues);
	}

	private function setHabtmValues(ReflectedTable $t1, ReflectedTable $t2, array &$records, HabtmValues $habtmValues)
	{
		$pkName = $t1->getPk()->getName();
		$t2Name = $t2->getName();
		foreach ($records as $i => $record) {
			$key = $record[$pkName];
			$val = array();
			$fks = $habtmValues->pkValues[$key];
			foreach ($fks as $fk) {
				$val[] = $habtmValues->fkValues[$fk];
			}
			$records[$i][$t2Name] = $val;
		}
	}
*/
