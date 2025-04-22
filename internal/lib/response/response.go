package response

type Response struct {
	UserID      string `json:"user_id"`
	AccessToken string `json:"access_token"`
}

func New(guid, accessToken string) *Response {
	return &Response{
		UserID:      guid,
		AccessToken: accessToken,
	}
}
