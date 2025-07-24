package api

// Team represents the structure of a team
type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TeamMember represents a user within a team context
type TeamMember struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
