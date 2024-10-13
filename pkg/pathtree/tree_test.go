package pathtree

import (
	"testing"
)

func TestTrieInsertAndSearch(t *testing.T) {
	trie := NewTrie[string]()

	// Test cases
	testCases := []struct {
		insert   string
		search   string
		expected string
	}{
		{"/", "/", "/"},
		{"/api", "/api", "/api"},
		{"/api/users", "/api/users", "/api/users"},
		{"/api/users/123", "/api/users/123", "/api/users/123"},
		{"/api/users/123", "/api/users/456", "/api/users"},
		{"/api/posts", "/api/posts/1", "/api/posts"},
		{"/blog", "/blog/2023/05/01", "/blog"},
	}

	// Insert test data
	for _, tc := range testCases {
		trie.Insert(tc.insert, tc.insert)
	}

	// Test searches
	for _, tc := range testCases {
		path, data := trie.Search(tc.search)
		if path != tc.expected {
			t.Errorf("Search(%s) = %s, expected %s", tc.search, path, tc.expected)
		}
		if data != tc.expected {
			t.Errorf("Search(%s) data = %s, expected %s", tc.search, data, tc.expected)
		}
	}
}

func TestTrieRootInsert(t *testing.T) {
	trie := NewTrie[string]()
	trie.Insert("/", "root")

	path, data := trie.Search("/")
	if path != "/" {
		t.Errorf("Search('/') path = %s, expected '/'", path)
	}
	if data != "root" {
		t.Errorf("Search('/') data = %s, expected 'root'", data)
	}
}

func TestTrieEmptySearch(t *testing.T) {
	trie := NewTrie[string]()
	trie.Insert("/api", "api")

	path, data := trie.Search("")
	if path != "" {
		t.Errorf("Search('') path = %s, expected ''", path)
	}
	if data != "" {
		t.Errorf("Search('') data = %s, expected ''", data)
	}
}

func TestTrieNonExistentPath(t *testing.T) {
	trie := NewTrie[string]()
	trie.Insert("/api/users", "users")

	path, data := trie.Search("/api/posts")
	if path != "" {
		t.Errorf("Search('/api/posts') path = %s, expected ''", path)
	}
	if data != "" {
		t.Errorf("Search('/api/posts') data = %s, expected ''", data)
	}
}

func TestTrieWithIntData(t *testing.T) {
	trie := NewTrie[int]()
	trie.Insert("/api", 1)
	trie.Insert("/api/users", 2)

	path, data := trie.Search("/api/users/123")
	if path != "/api/users" {
		t.Errorf("Search('/api/users/123') path = %s, expected '/api/users'", path)
	}
	if data != 2 {
		t.Errorf("Search('/api/users/123') data = %d, expected 2", data)
	}
}
