package roomrules

// Decision describes whether a room rule allows or rejects an action.
type Decision struct {
	Allowed bool
	Code    string
	Message string
}

// Allow returns a neutral allowed decision.
func Allow() Decision {
	return Decision{Allowed: true}
}

// Reject returns a neutral rejected decision with a code and message.
func Reject(code string, message string) Decision {
	return Decision{
		Allowed: false,
		Code:    code,
		Message: message,
	}
}
