package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

type UserServiceImpl struct {
	mu    sync.RWMutex
	users []map[string]interface{}
	idSeq int64
}

func NewUserService() UserService {
	return &UserServiceImpl{
		users: []map[string]interface{}{
			{
				"id":       int64(1),
				"username": "admin",
				"password": "admin123", // 简化密码
				"nickname": "管理员",
				"email":    "admin@example.com",
				"avatar":   "",
			},
		},
		idSeq: 1,
	}
}

func (s *UserServiceImpl) GetByID(ctx context.Context, id int64) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user["id"] == id {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *UserServiceImpl) UpdateProfile(ctx context.Context, id int64, nickname, email, avatar string) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, user := range s.users {
		if user["id"] == id {
			if nickname != "" {
				s.users[i]["nickname"] = nickname
			}
			if email != "" {
				s.users[i]["email"] = email
			}
			if avatar != "" {
				s.users[i]["avatar"] = avatar
			}
			return s.users[i], nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *UserServiceImpl) UpdatePassword(ctx context.Context, id int64, oldPassword, newPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, user := range s.users {
		if user["id"] == id {
			currentPassword, _ := user["password"].(string)
			if currentPassword != oldPassword {
				return errors.New("old password incorrect")
			}
			s.users[i]["password"] = newPassword
			return nil
		}
	}
	return errors.New("user not found")
}

func (s *UserServiceImpl) Login(ctx context.Context, username, password string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user["username"] == username && user["password"] == password {
			// 简化token生成
			token := "token_" + username + "_" + time.Now().Format("20060102150405")
			return token, nil
		}
	}
	return "", errors.New("invalid credentials")
}

func (s *UserServiceImpl) Install(ctx context.Context, username, password, email, siteName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否已存在用户
	if len(s.users) > 0 {
		return errors.New("already installed")
	}

	s.idSeq++
	s.users = append(s.users, map[string]interface{}{
		"id":       s.idSeq,
		"username": username,
		"password": password,
		"nickname": username,
		"email":    email,
		"avatar":   "",
	})

	return nil
}

func (s *UserServiceImpl) VerifyToken(ctx context.Context, token string) (int64, error) {
	// 简化token验证
	if token == "" {
		return 0, errors.New("empty token")
	}
	// 返回固定用户ID
	return 1, nil
}
