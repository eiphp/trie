package trie

type Node struct {
	key        string
	pattern    string
	handle     Handler
	depth      int
	children   map[string]*Node
	isPattern  bool
	middleware []Middleware
}

func NewNode(key string, depth int) *Node {
	return &Node{
		key:      key,
		depth:    depth,
		children: make(map[string]*Node),
	}
}
