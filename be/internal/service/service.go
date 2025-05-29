package service

import (
	"context"
	"errors"

	"github.com/escalopa/chatterly/internal/domain"
)

type (
	database interface {
		GetUser(ctx context.Context, userID string) (*domain.User, error)
		CreateUser(ctx context.Context, user *domain.User, provider string) (string, error)
		SetUsername(ctx context.Context, userID string, username string) error
	}

	oauthProvider interface {
		GetRedirectURL(provider string) (string, error)
		HandleCallback(ctx context.Context, provider string, code string) (*domain.User, error)
	}

	userTokenProvider interface {
		CreateToken(userID string, email string) (*domain.Token, error)
		VerifyToken(token string) (*domain.UserTokenPayload, error)
	}

	chatTokenProvider interface {
		CreateToken(userID string, sessionID string) (string, error)
		VerifyToken(token string) (*domain.ChatTokenPayload, error)
	}
)

type Service struct {
	db                database
	oauthProvider     oauthProvider
	userTokenProvider userTokenProvider
	chatTokenProvider chatTokenProvider
}

func New(
	db database,
	oauthProvider oauthProvider,
	userTokenProvider userTokenProvider,
	chatTokenProvider chatTokenProvider,
) *Service {
	return &Service{
		db:                db,
		oauthProvider:     oauthProvider,
		userTokenProvider: userTokenProvider,
		chatTokenProvider: chatTokenProvider,
	}
}

func (s *Service) GetOAuthRedirectURL(provider string) (string, error) {
	return s.oauthProvider.GetRedirectURL(provider)
}

func (s *Service) RegisterUser(ctx context.Context, provider string, code string) (*domain.Token, error) {
	user, err := s.oauthProvider.HandleCallback(ctx, provider, code)
	if err != nil {
		return nil, err
	}

	userID, err := s.db.CreateUser(ctx, user, provider)
	if err != nil {
		return nil, err
	}

	token, err := s.userTokenProvider.CreateToken(userID, user.Email)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Service) AuthenticateUser(ctx context.Context, token *domain.Token) (*domain.User, *domain.Token, error) {
	userID, token, err := s.verifyUserToken(token)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.db.GetUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

func (s *Service) verifyUserToken(token *domain.Token) (string, *domain.Token, error) {
	payload, err := s.userTokenProvider.VerifyToken(token.Access)
	if err != nil && !errors.Is(err, domain.ErrTokenExpired) {
		return "", nil, err
	}

	// refresh token if access token is expired
	if errors.Is(err, domain.ErrTokenExpired) {
		payload, err = s.userTokenProvider.VerifyToken(token.Refresh)
		if err != nil {
			return "", nil, err
		}

		token, err = s.userTokenProvider.CreateToken(payload.UserID, payload.Email)
		if err != nil {
			return "", nil, err
		}

		return payload.UserID, token, nil
	}

	return payload.UserID, nil, nil
}

func (s *Service) CreateRoomToken(userID string, sessionID string) (string, error) {
	return s.chatTokenProvider.CreateToken(userID, sessionID)
}

//func (s *Service) AuthenticateWS(ctx context.Context, token string, roomID string) (*domain.User, error) {
//	payload, err := s.roomTokenProvider.VerifyToken(token)
//	if err != nil {
//		return nil, err
//	}
//
//	if payload.RoomID != roomID {
//		return nil, domain.ErrRoomIDTokenMismatch
//	}
//
//	user, err := s.db.GetUser(ctx, payload.UserID)
//	if err != nil {
//		return nil, err
//	}
//
//	return user, nil
//}
//
//func (s *Service) HandleWS(user *domain.User, roomID string, conn *websocket.Conn) {
//	s.hub.Handle(user, roomID, conn)
//}
