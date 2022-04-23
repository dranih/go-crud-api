package record

type Tree struct {
	branches map[string]*Tree
	values   []interface{}
}

const WILDCARD = `*`

func NewPathTree(tree *Tree) *Tree {
	if tree != nil {
		return tree
	} else {
		return NewTree()
	}
}

func NewTree() *Tree {
	return &Tree{map[string]*Tree{}, []interface{}{}}
}

func (t *Tree) GetKeys() []string {
	keys := []string{}
	for key := range t.branches {
		keys = append(keys, key)
	}
	return keys
}

func (t *Tree) GetValues() []interface{} {
	return t.values
}

func (t *Tree) Get(key string) *Tree {
	if _, exists := t.branches[key]; !exists {
		return nil
	} else {
		return NewPathTree(t.branches[key])
	}
}

func (t *Tree) Put(path []string, value interface{}) {
	tree := t
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
