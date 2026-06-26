// Package openai_compat provides OpenAI-compatible upstream capability helpers.
package openai_compat

// AccountResponsesSupport describes an OpenAI API key upstream's effective
// Responses API support.
type AccountResponsesSupport int

const (
	ResponsesSupportUnknown AccountResponsesSupport = iota
	ResponsesSupportYes
	ResponsesSupportNo
)

// ResponsesSupportMode is the account-level override for Responses routing.
type ResponsesSupportMode string

const (
	ResponsesSupportModeAuto                 ResponsesSupportMode = "auto"
	ResponsesSupportModeForceResponses       ResponsesSupportMode = "force_responses"
	ResponsesSupportModeForceChatCompletions ResponsesSupportMode = "force_chat_completions"
)

const (
	ExtraKeyResponsesMode      = "openai_responses_mode"
	ExtraKeyResponsesSupported = "openai_responses_supported"
)

// NormalizeResponsesSupportMode normalizes invalid or missing modes to auto.
func NormalizeResponsesSupportMode(mode string) ResponsesSupportMode {
	switch ResponsesSupportMode(mode) {
	case ResponsesSupportModeForceResponses:
		return ResponsesSupportModeForceResponses
	case ResponsesSupportModeForceChatCompletions:
		return ResponsesSupportModeForceChatCompletions
	default:
		return ResponsesSupportModeAuto
	}
}

// ResolveResponsesSupport reads the manual override and probe result from
// account extra data.
func ResolveResponsesSupport(extra map[string]any) AccountResponsesSupport {
	if extra == nil {
		return ResponsesSupportUnknown
	}
	if mode, ok := extra[ExtraKeyResponsesMode].(string); ok {
		switch NormalizeResponsesSupportMode(mode) {
		case ResponsesSupportModeForceResponses:
			return ResponsesSupportYes
		case ResponsesSupportModeForceChatCompletions:
			return ResponsesSupportNo
		}
	}
	supported, ok := extra[ExtraKeyResponsesSupported].(bool)
	if !ok {
		return ResponsesSupportUnknown
	}
	if supported {
		return ResponsesSupportYes
	}
	return ResponsesSupportNo
}

// ShouldUseResponsesAPI preserves old behavior for unknown accounts and only
// switches to raw Chat Completions when support is explicitly false.
func ShouldUseResponsesAPI(extra map[string]any) bool {
	return ResolveResponsesSupport(extra) != ResponsesSupportNo
}
