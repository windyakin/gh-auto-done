package github

type Notification struct {
	ID         string                 `json:"id"`
	Unread     bool                   `json:"unread"`
	Reason     string                 `json:"reason"`
	UpdatedAt  string                 `json:"updated_at"`
	Subject    NotificationSubject    `json:"subject"`
	Repository NotificationRepository `json:"repository"`
}

type NotificationSubject struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

type NotificationRepository struct {
	FullName string `json:"full_name"`
}

type SubjectState struct {
	State string `json:"state"`
}
