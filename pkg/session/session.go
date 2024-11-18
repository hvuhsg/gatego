package session

type Session interface {
	GetItems(string) (map[string]any, error)
	SetItems(key string, items map[string]any) error
}
