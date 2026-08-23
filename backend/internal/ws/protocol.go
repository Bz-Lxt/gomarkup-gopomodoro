package ws

type Inbound struct {
	Type         string `json:"type"`
	ResumeToken  string `json:"resume_token,omitempty"`
	MilestoneID  string `json:"milestone_id,omitempty"`
}

type Outbound struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

const (
	TypeHello      = "hello"
	TypeSubscribe  = "subscribe"
	TypePing       = "ping"
	TypePong       = "pong"
	TypeSession    = "session.state"
	TypeTick       = "session.tick"
	TypeBurndown   = "burndown.update"
	TypeGrace      = "grace"
	TypeError      = "error"
)
