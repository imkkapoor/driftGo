package stytch

type WebhookEvent struct {
	Action       string     `json:"action"`
	EventID      string     `json:"event_id"`
	StytchUserID string     `json:"id"`
	Source       string     `json:"source"`
	Timestamp    string     `json:"timestamp"`
	User         *UserEvent `json:"user,omitempty"`
}

type UserEvent struct {
	CreatedAt    string        `json:"created_at"`
	Emails       []Email       `json:"emails"`
	Name         Name          `json:"name"`
	PhoneNumbers []interface{} `json:"phone_numbers"`
	Providers    []interface{} `json:"providers"`
	Status       string        `json:"status"`
}

type Email struct {
	Email    string `json:"email"`
	EmailID  string `json:"email_id"`
	Verified bool   `json:"verified"`
}

type Name struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
}
