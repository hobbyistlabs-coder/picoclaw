package api

import (
	"time"

	"jane/pkg/auth"
	"jane/pkg/config"
	"jane/pkg/providers"
)

const (
	oauthProviderOpenAI            = "openai"
	oauthProviderAnthropic         = "anthropic"
	oauthProviderGoogleAntigravity = "google-antigravity"

	oauthMethodBrowser    = "browser"
	oauthMethodDeviceCode = "device_code"
	oauthMethodToken      = "token"

	oauthFlowPending = "pending"
	oauthFlowSuccess = "success"
	oauthFlowError   = "error"
	oauthFlowExpired = "expired"
)

const (
	oauthBrowserFlowTTL    = 10 * time.Minute
	oauthDeviceCodeFlowTTL = 15 * time.Minute
	oauthTerminalFlowGC    = 30 * time.Minute
)

var oauthProviderOrder = []string{
	oauthProviderOpenAI,
	oauthProviderAnthropic,
	oauthProviderGoogleAntigravity,
}

var oauthProviderMethods = map[string][]string{
	oauthProviderOpenAI:            {oauthMethodBrowser, oauthMethodDeviceCode, oauthMethodToken},
	oauthProviderAnthropic:         {oauthMethodToken},
	oauthProviderGoogleAntigravity: {oauthMethodBrowser},
}

var oauthProviderLabels = map[string]string{
	oauthProviderOpenAI:            "OpenAI",
	oauthProviderAnthropic:         "Anthropic",
	oauthProviderGoogleAntigravity: "Google Antigravity",
}

var (
	oauthNow                      = time.Now
	oauthGeneratePKCE             = auth.GeneratePKCE
	oauthGenerateState            = auth.GenerateState
	oauthBuildAuthorizeURL        = auth.BuildAuthorizeURL
	oauthRequestDeviceCode        = auth.RequestDeviceCode
	oauthPollDeviceCodeOnce       = auth.PollDeviceCodeOnce
	oauthExchangeCodeForTokens    = auth.ExchangeCodeForTokens
	oauthGetCredential            = auth.GetCredential
	oauthSetCredential            = auth.SetCredential
	oauthDeleteCredential         = auth.DeleteCredential
	oauthLoadConfig               = config.LoadConfig
	oauthSaveConfig               = config.SaveConfig
	oauthFetchAntigravityProject  = providers.FetchAntigravityProjectID
	oauthFetchGoogleUserEmailFunc = fetchGoogleUserEmail
)
