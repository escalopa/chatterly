package domain

type (
	User struct {
		ID       string `json:"id" bson:"_id"`
		Name     string `json:"name" bson:"name"`
		Email    string `json:"email" bson:"email"`
		Avatar   string `json:"avatar" bson:"avatar"`
		Provider string `json:"provider" bson:"provider"`
		Username string `json:"username" bson:"username"`
	}

	UserTokenPayload struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}

	ChatTokenPayload struct {
		UserID    string `json:"user_id"`
		SessionID string `json:"session_id"`
	}

	Token struct {
		Access  string
		Refresh string
	}
)
