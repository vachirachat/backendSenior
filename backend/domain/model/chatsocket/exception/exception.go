package exception

// Event represent error (exception) event (ex. unauthorized)
type Event struct {
	// Reason is the reason of error
	Reason string `json:"reason"`
	// Data is detail of the error
	Data interface{} `json:"data"`
}
