package record

import "github.com/dranih/go-crud-api/pkg/database"

type PathTree struct {
	tree *tree
}
type tree struct {
	branches map[string]*tree
	values   []interface{ database.Condition }
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
	return &tree{map[string]*tree{}, []interface{ database.Condition }{}}
}

func (t *tree) GetKeys() []string {
	keys := []string{}
	for key := range t.branches {
		keys = append(keys, key)
	}
	return keys
}

func (t *tree) GetValues() []interface{ database.Condition } {
	return t.values
}

func (t *tree) Get(key string) *PathTree {
	if _, exists := t.branches[key]; !exists {
		return nil
	} else {
		return NewPathTree(t.branches[key])
	}
}

func (pt *PathTree) Put(path []string, value interface{ database.Condition }) {
	tree := pt.tree
	for _, key := range path {
		if key == `` {
			key = `0`
		}
		if _, exists := tree.branches[key]; !exists {
			tree.branches[key] = NewTree()
		}
		tree = tree.branches[key]
	}
	tree.values = append(tree.values, value)
}

/*
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
