package perror

//Error Model
type Error struct {
	Code     int                    `json:"code"`
	Msg      string                 `json:"Message"`
	Trace    string                 `json:"trace,omitempty"`
	DebugMsg string                 `json:"debug_msg,omitempty"`
	Info     map[string]interface{} `json:"info,omitempty"`
}
