package access

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type Store struct {
	DefaultPolicy    string              `json:"default_policy"`
	AdminAllowAll    bool                `json:"admin_allow_all"`
	PublicPlugins    []string            `json:"public_plugins"`
	AdminOnlyPlugins []string            `json:"admin_only_plugins"`
	Groups           map[string][]string `json:"groups"`
	Users            map[string][]string `json:"users"`
}

type Manager struct {
	path string
	mu   sync.RWMutex
	data Store
}

func New(path string) (*Manager, error) {
	m := &Manager{path: path}
	if err := m.Load(); err != nil {
		return nil, err
	}
	return m, nil
}

func DefaultStore() Store {
	return Store{
		DefaultPolicy:    "deny",
		AdminAllowAll:    true,
		PublicPlugins:    []string{"help"},
		AdminOnlyPlugins: []string{"pluginctl"},
		Groups:           map[string][]string{},
		Users:            map[string][]string{},
	}
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		m.data = DefaultStore()
		return m.saveLocked()
	}

	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	store := DefaultStore()
	if err := json.Unmarshal(data, &store); err != nil {
		return err
	}
	if store.DefaultPolicy == "" {
		store.DefaultPolicy = "deny"
	}
	if store.Groups == nil {
		store.Groups = map[string][]string{}
	}
	if store.Users == nil {
		store.Users = map[string][]string{}
	}
	m.data = store
	return nil
}

func (m *Manager) CanUse(pluginName string, userID, groupID int64, messageType string, isAdmin bool) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if isAdmin && m.data.AdminAllowAll {
		return true
	}
	if contains(m.data.AdminOnlyPlugins, pluginName) {
		return false
	}
	if contains(m.data.PublicPlugins, pluginName) {
		return true
	}
	if messageType == "group" {
		return contains(m.data.Groups[strconv.FormatInt(groupID, 10)], pluginName)
	}
	if contains(m.data.Users[strconv.FormatInt(userID, 10)], pluginName) {
		return true
	}
	return m.data.DefaultPolicy == "allow"
}

func (m *Manager) Allow(scope string, id int64, pluginName string) error {
	return m.update(scope, id, pluginName, true)
}

func (m *Manager) Deny(scope string, id int64, pluginName string) error {
	return m.update(scope, id, pluginName, false)
}

func (m *Manager) List(scope string, id int64) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if scope == "group" {
		return append([]string{}, m.data.Groups[strconv.FormatInt(id, 10)]...)
	}
	if scope == "user" || scope == "private" {
		return append([]string{}, m.data.Users[strconv.FormatInt(id, 10)]...)
	}
	return nil
}

func (m *Manager) update(scope string, id int64, pluginName string, allow bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := strconv.FormatInt(id, 10)
	var target map[string][]string
	switch scope {
	case "group":
		target = m.data.Groups
	case "user", "private":
		target = m.data.Users
	default:
		return fmt.Errorf("invalid scope %q", scope)
	}

	if allow {
		if !contains(target[key], pluginName) {
			target[key] = append(target[key], pluginName)
		}
	} else {
		target[key] = remove(target[key], pluginName)
	}
	return m.saveLocked()
}

func (m *Manager) saveLocked() error {
	data, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0644)
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func remove(values []string, target string) []string {
	out := values[:0]
	for _, v := range values {
		if v != target {
			out = append(out, v)
		}
	}
	return out
}
