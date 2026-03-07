package main

import "fmt"

// UserService 用户服务
type UserService struct {
	users map[string]string
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]string),
	}
}

// AddUser 添加用户
func (s *UserService) AddUser(id, name string) {
	s.users[id] = name
}

// GetUser 获取用户
func (s *UserService) GetUser(id string) (string, bool) {
	name, exists := s.users[id]
	return name, exists
}

// ListUsers 列出所有用户
func (s *UserService) ListUsers() map[string]string {
	return s.users
}

func main() {
	service := NewUserService()
	service.AddUser("1", "Alice")
	service.AddUser("2", "Bob")

	if name, exists := service.GetUser("1"); exists {
		fmt.Printf("User found: %s\n", name)
	}

	users := service.ListUsers()
	for id, name := range users {
		fmt.Printf("ID: %s, Name: %s\n", id, name)
	}
}
