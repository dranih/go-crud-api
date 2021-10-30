package database

import "log"

type PathTree struct {
	tree *tree
}
type tree struct {
	branches map[string]*tree
	values   []interface{ Condition }
}

const WILDCARD = `*`

func NewPathTree(tree *tree) *PathTree {
	pathTree := &PathTree{}
	if tree != nil {
		pathTree.tree = tree
	} else {
		pathTree.tree = NewTree()
	}

	return pathTree
}

func NewTree() *tree {
	return &tree{map[string]*tree{}, []interface{ Condition }{}}
}

func (t *tree) GetKeys() []string {
	branches := t.branches
	keys := []string{}
	for key := range branches {
		keys = append(keys, key)
	}
	return keys
}

func (t *tree) GetValues() []interface{ Condition } {
	return t.values
}

func (t *tree) Get(key string) *PathTree {
	if _, exists := t.branches[key]; !exists {
		return nil
	} else {
		return NewPathTree(t.branches[key])
	}
}

/*
public function get(string $key): PathTree
{
	if (!isset($this->tree->branches->$key)) {
		return null;
	}
	return new PathTree($this->tree->branches->$key);
}
*/
func (pt *PathTree) Put(path []string, value interface{ Condition }) {
	tree := pt.tree
	log.Printf("--- Tree1 : %v\n", tree)
	log.Printf("--- PathTree1 ; %v\n", pt.tree)
	for _, key := range path {
		if key == `` {
			key = `0`
		}
		if _, exists := pt.tree.branches[key]; !exists {
			log.Printf("Branches ; %v\n", pt.tree.branches)
			log.Printf("Key ; %v\n", key)
			pt.tree.branches[key] = NewTree()
		}
		tree = pt.tree.branches[key]
	}
	tree.values = append(tree.values, value)
	log.Printf("--- Tree2 : %v\n", tree)
	log.Printf("--- PathTree2 ; %v\n", pt)
}

/*
public function put(array $path, $value)
{
	$tree = &$this->tree;
	foreach ($path as $key) {
		if (!isset($tree->branches->$key)) {
			$tree->branches->$key = $this->newTree();
		}
		$tree = &$tree->branches->$key;
	}
	$tree->values[] = $value;
}

public function match(array $path): array
{
	$star = self::WILDCARD;
	$tree = &$this->tree;
	foreach ($path as $key) {
		if (isset($tree->branches->$key)) {
			$tree = &$tree->branches->$key;
		} elseif (isset($tree->branches->$star)) {
			$tree = &$tree->branches->$star;
		} else {
			return [];
		}
	}
	return $tree->values;
}

public static function fromJson($tree): PathTree
{
	return new PathTree($tree);
}

public function jsonSerialize()
{
	return $this->tree;
}
*/
