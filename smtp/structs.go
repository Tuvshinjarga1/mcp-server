package smtp

type EmailInput struct {
	Email    string   `json:"email"`
	MultiBcc []string `json:"multi_bcc"`
	Subtitle string   `json:"subtitle"`
	Template string   `json:"template"`
}
