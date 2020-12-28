package room

// MemberEvent is payload for room leave/join event
type MemberEvent struct {
	// type leave or join
	Type string `json:"type"`
	// member(s) that leave or join
	Members []string `json:"members"`
}
