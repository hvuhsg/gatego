package pathgraph

import "strings"

const IncRate = 0.01

// PathVertex represents a vertex in the graph
type PathVertex struct {
	Path   string
	Weight float64
}

// PathGraph represents a weighted directed graph of navigation paths
type PathGraph struct {
	// Map of source path to map of destination paths and their weights
	adjacencyList map[string]map[string]*PathVertex
}

// NewPathGraph creates a new instance of PathGraph
func NewPathGraph() *PathGraph {
	return &PathGraph{
		adjacencyList: make(map[string]map[string]*PathVertex),
	}
}

// AddJump adds or updates a path transition in the graph
func (g *PathGraph) AddJump(sourcePath, destPath string) float64 {
	sourcePath = normalizePath(sourcePath)
	destPath = normalizePath(destPath)

	// Initialize source path if it doesn't exist
	if _, exists := g.adjacencyList[sourcePath]; !exists {
		g.adjacencyList[sourcePath] = make(map[string]*PathVertex)
	}

	// Get or create destination node
	vertex, exists := g.adjacencyList[sourcePath][destPath]
	if !exists {
		vertex = &PathVertex{
			Path:   destPath,
			Weight: 0,
		}
		g.adjacencyList[sourcePath][destPath] = vertex
	}

	// Increment weight
	vertex.Weight++

	return vertex.Weight - 1 // The original weight (before the jump)
}

// GetDestinations returns all destinations and their weights for a given source path
func (g *PathGraph) GetDestinations(sourcePath string) map[string]float64 {
	sourcePath = normalizePath(sourcePath)

	result := make(map[string]float64)

	if vertexs, exists := g.adjacencyList[sourcePath]; exists {
		for path, vertex := range vertexs {
			result[path] = vertex.Weight
		}
	}

	return result
}

// GetAllPaths returns all unique paths in the graph
func (g *PathGraph) GetAllPaths() []string {
	pathSet := make(map[string]bool)

	// Add all source paths
	for sourcePath := range g.adjacencyList {
		pathSet[sourcePath] = true

		// Add all destination paths
		for destPath := range g.adjacencyList[sourcePath] {
			pathSet[destPath] = true
		}
	}

	// Convert set to slice
	paths := make([]string, 0, len(pathSet))
	for path := range pathSet {
		paths = append(paths, path)
	}

	return paths
}

func normalizePath(path string) string {
	if path[0] != '/' {
		path = "/" + path
	}

	path = strings.ToLower(path)

	return path
}
