package constants

import "crypto/aes"

// DBIndexField database default constants
const (
	DBIndexField = "id"
)

// AES
const (
	BlockSize = aes.BlockSize
	IV        = "TRYTOCN314402233"
)

// MaxPageLimit You can find out through the following link
// https://stackoverflow.com/questions/25265465/why-github-api-gives-me-a-lower-number-stars-of-a-repo
const (
	GitHubStarBaseURL = "https://api.github.com/repos"
	DefaultPageSize   = 100 // The maximum value cannot exceed 100
	MaxPageLimit      = 400 // Maximum number of pages to fetch
)

// login status constants
const (
	LoginStatusLoggedIn      = "logged_in"      // out -> in
	LoginStatusLoggedOut     = "logged_out"     // Initial state
	LoginStatusLoggedOffline = "logged_offline" // in -> offline
)

// Binding account related
const (
	LoginSuccessPath       = "/login/success"
	LoginCallbackURI       = "/oidc-auth/api/v1/plugin/login/callback"
	WebLoginCallbackURI    = "/oidc-auth/api/v1/manager/login/callback"
	BindAccountBindURI     = "/credit/manager/"
	BindAccountCallbackURI = "/oidc-auth/api/v1/manager/bind/account/callback"
)

// Casdoor certification related
const (
	CasdoorAuthURI         = "/login/oauth/authorize"
	CasdoorTokenURI        = "/api/login/oauth/access_token"
	CasdoorRefreshTokenURI = "/api/login/oauth/refresh_token"
	CasdoorMergeURI        = "/api/identity/merge"
)

// Invite code related constants
const (
	InviteCodeLength    = 4
	InviteCodeChars     = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	InviteCodeSeparator = "__inviter_code=" // separator for inviter code in state parameter
)

// Quota service related constants
const (
	QuotaMergeURI = "/quota-manager/api/v1/quota/merge"
)
