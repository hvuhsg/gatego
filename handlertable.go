package gatego

import (
	"context"
	"net/http"
	"strings"

	"github.com/hvuhsg/gatego/config"
	"github.com/hvuhsg/gatego/pkg/pathtree"
)

type HandlerTable map[string]*pathtree.Trie[http.Handler]

func cleanDomain(domain string) string {
	return removePort(strings.ToLower(domain))
}

func NewHandlersTable(ctx context.Context, useOtel bool, servicesConfig []config.Service) (HandlerTable, error) {
	servers := make(map[string]*pathtree.Trie[http.Handler])

	for _, service := range servicesConfig {
		servicePathTree := pathtree.NewTrie[http.Handler]()

		cleanedDomain := cleanDomain(service.Domain)

		servers[cleanedDomain] = servicePathTree

		for _, path := range service.Paths {
			handler, err := NewHandler(ctx, useOtel, service, path)
			if err != nil {
				return nil, err
			}

			cleanPath := strings.ToLower(path.Path)
			servicePathTree.Insert(cleanPath, handler)
		}
	}

	return servers, nil
}

func (table HandlerTable) GetHandler(domain string, path string) http.Handler {
	cleanedDomain := cleanDomain(domain)

	pathTree, ok := table[cleanedDomain]
	if !ok {
		return nil
	}

	endpoint, server := pathTree.Search(path)
	if len(endpoint) == 0 {
		return nil
	}

	return server
}

func removePort(addr string) string {
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}
