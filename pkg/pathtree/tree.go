package pathtree

import (
	"strings"
)

type TrieNode[T any] struct {
	children map[string]*TrieNode[T]
	isEnd    bool
	value    string
	data     T
}

type Trie[T any] struct {
	root *TrieNode[T]
}

func NewTrie[T any]() *Trie[T] {
	return &Trie[T]{
		root: &TrieNode[T]{
			children: make(map[string]*TrieNode[T]),
		},
	}
}

func (t *Trie[T]) Insert(path string, data T) {
	if path == "/" {
		t.root.value = path
		t.root.data = data
		t.root.isEnd = true
		return
	}

	node := t.root
	parts := strings.Split(strings.Trim(path, "/"), "/")

	for _, part := range parts {
		if _, exists := node.children[part]; !exists {
			node.children[part] = &TrieNode[T]{
				children: make(map[string]*TrieNode[T]),
			}
		}
		node = node.children[part]
	}

	node.isEnd = true
	node.value = path
	node.data = data
}

func (t *Trie[T]) Search(path string) (string, T) {
	node := t.root
	parts := strings.Split(strings.Trim(path, "/"), "/")
	lastMatch := node.value
	var lastData T = node.data

	for _, part := range parts {
		if child, exists := node.children[part]; exists {
			node = child
			if node.isEnd {
				lastMatch = node.value
				lastData = node.data
			}
		} else {
			break
		}
	}

	return lastMatch, lastData
}
