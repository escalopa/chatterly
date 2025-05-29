package auth

import (
	"fmt"

	"github.com/escalopa/chatterly/internal/domain"
	"golang.org/x/oauth2"
)

const (
	googleProvider = "google"
	yandexProvider = "yandex"
	gitlabProvider = "gitlab"
	githubProvider = "github"
)

type payload interface {
	ToUser() *domain.User
}

type provider struct {
	config   *oauth2.Config
	endpoint string
	payload  func() payload
}

type (
	googlePayload struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"picture"`
	}

	yandexPayload struct {
		Name   string `json:"real_name"`
		Email  string `json:"default_email"`
		Avatar string `json:"default_avatar_id"`
	}

	gitlabPayload struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"avatar_url"`
	}

	githubPayload struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"avatar_url"`
	}
)

func (p *googlePayload) ToUser() *domain.User {
	return &domain.User{
		Name:   p.Name,
		Email:  p.Email,
		Avatar: p.Avatar,
	}
}

func (p *yandexPayload) ToUser() *domain.User {
	return &domain.User{
		Name:   p.Name,
		Email:  p.Email,
		Avatar: fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", p.Avatar),
	}
}

func (p *gitlabPayload) ToUser() *domain.User {
	return &domain.User{
		Name:   p.Name,
		Email:  p.Email,
		Avatar: p.Avatar,
	}
}

func (p *githubPayload) ToUser() *domain.User {
	return &domain.User{
		Name:   p.Name,
		Email:  p.Email,
		Avatar: p.Avatar,
	}
}
