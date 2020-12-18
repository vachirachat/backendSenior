package model

type Permission struct {
	Resource string   `json:"resource" bson:"resource"`
	Scopes   []string `json:"scopes" bson:"scopes"`
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type AccessDetails struct {
	AccessUuid string
	UserId     uint64
}

type JWTClaim struct {
	AccessUuid string `json:"access_uuid"`
	Authorized string `json:"authorized"`
	Exp        string `json:"exp"`
	Role       string `json:"role"`
	UserId     string `json:"user_id"`
}
