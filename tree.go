package trie

import (
	"strings"
)

type Tree struct {
	root   *Node
	routes map[string]*Node
}

func NewTree() *Tree {
	return &Tree{
		root:   NewNode("/", 1),
		routes: make(map[string]*Node),
	}
}

func (t *Tree) Add(pattern string, handle Handler, middleware ...Middleware) {
	var currentNode = t.root
	if pattern != currentNode.key {
		pattern = strings.TrimPrefix(pattern, "/")
		res := strings.Split(pattern, "/")
		for _, key := range res {
			node, ok := currentNode.children[key]
			if !ok {
				node = NewNode(key, currentNode.depth+1)
				if len(middleware) > 0 {
					node.middleware = append(node.middleware, middleware...)
				}
				currentNode.children[key] = node
			}
			currentNode = node
		}
	}
	if len(middleware) > 0 && currentNode.depth == 1 {
		currentNode.middleware = append(currentNode.middleware, middleware...)
	}
	currentNode.handle = handle
	currentNode.isPattern = true
	currentNode.pattern = pattern
}

func (t *Tree) Find(pattern string, isRegex bool) (nodes []*Node) {
	var (
		node      = t.root
		nodeQueue []*Node
	)
	if pattern == node.pattern {
		nodes = append(nodes, node)
		return
	}
	if !isRegex {
		pattern = strings.TrimPrefix(pattern, "/")
	}
	res := strings.Split(pattern, "/")
	for _, key := range res {
		child, ok := node.children[key]
		if !ok && isRegex {
			break
		}
		if !ok && !isRegex {
			return
		}
		if pattern == child.pattern && !isRegex {
			nodes = append(nodes, child)
			return
		}
		node = child
	}
	nodeQueue = append(nodeQueue, node)
	for len(nodeQueue) > 0 {
		var queueTemp []*Node
		for _, n := range nodeQueue {
			if n.isPattern {
				nodes = append(nodes, n)
			}
			for _, childNode := range n.children {
				queueTemp = append(queueTemp, childNode)
			}
		}
		nodeQueue = queueTemp
	}
	return
}
