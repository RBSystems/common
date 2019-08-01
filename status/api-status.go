package status

// Power is a base evaluator struct.
type Power struct {
	Power string `json:"power"`
}

// Blanked is a base evaluator struct.
type Blanked struct {
	Blanked bool `json:"blanked"`
}

// Mute is a base evaluator struct.
type Mute struct {
	Muted bool `json:"muted"`
}

// Input is a base evaluator struct.
type Input struct {
	Input string `json:"input,omitempty"`
}

// Volume is a base evaluator struct.
type Volume struct {
	Volume int `json:"volume"`
}

// Battery is a base evaluator struct.
type Battery struct {
	Battery int `json:"battery"`
}

// ActiveInput is a base struct for active inputs
type ActiveInput struct {
	ActiveInput string `json:"active_input,omitempty"`
}

// Error .
type Error struct {
	Error string `json:"error"`
}
