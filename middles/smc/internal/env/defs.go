package env

import (
	"regexp"
)

/**
 * Basic execution environment for smc command family (common components of smc command parameters)
 */
var (
	BaseUrl       string //Base URL
	MachineId     string //Machine ID
	PluginVersion string //vscode plugin version
	VscodeVersion string //vscode version
	TaskdAddr     string //Login server address
	PromptAddr    string //AI-Prompt-Shell service address
	Username      string //Login username
	Password      string //Login password
	RedisAddr     string //REDIS address
	RedisPwd      string //REDIS password
	RedisDb       int    //REDIS DB
	RedisTimeout  int    //REDIS record TTL
	Cookie        string //AIP platform login cookie
	Callback      string //Callback URL to receive notifications
	Listen        string //Local server listening for callbacks
	Logfile       string //Log file
	Debug         string //Debug level(Off,Err,Dbg), controls output verbosity
	SkipSSL       bool   //skip ssl verify:InsecureSkipVerify
)

/**
 *	Default environment variable registry
 */
var defEnvs *Envs

/**
 * Register smc predefined environment variables
 */
func InitEnvs() {
	if defEnvs != nil {
		return
	}
	defEnvs = NewEnvs()
	debugExp := regexp.MustCompile(`(?i)^Off|Err|Dbg$`)
	defEnvs.Register("SMC_DEBUG", "envDebug",
		"Debug switch: Off,Err,Dbg", "Err", NewLimitedString(&Debug, debugExp))
	defEnvs.Register("SMC_LOGFILE", "envLogfile",
		"Log file path, empty writes to stderr, +xx.log writes to both stderr and file", "", NewString(&Logfile))
	defEnvs.Register("SMC_TASKD_ADDR", "taskdAddr",
		"SMC platform address", "http://localhost:8080", NewString(&TaskdAddr))
	defEnvs.Register("SMC_USER", "user",
		"SMC login username", "", NewString(&Username))
	defEnvs.Register("SMC_REDIS_ADDR", "redisAddr",
		"REDIS server address", "localhost:6379", NewString(&RedisAddr))
	defEnvs.Register("SMC_REDIS_PWD", "redisPwd",
		"REDIS password", "", NewString(&RedisPwd))
	defEnvs.Register("SMC_REDIS_DB", "redisDb",
		"REDIS database", "0", NewInt(&RedisDb))
	defEnvs.Register("SMC_REDIS_TIMEOUT", "timeout",
		"REDIS timeout (hours)", "8760", NewInt(&RedisTimeout))
	defEnvs.Register("SMC_PROMPT_ADDR", "promptAddr",
		"AI-Prompt-Shell service address", "http://localhost:8080", NewString(&PromptAddr))
	defEnvs.Register("SMC_COOKIE", "cookie",
		"Login cookie", "", NewString(&Cookie))
	defEnvs.Register("SMC_CALLBACK", "callback",
		"Callback URL for task notifications", "http://localhost:8888/callback", NewString(&Callback))
	defEnvs.Register("SMC_LISTEN", "listen",
		"Listening address", ":8888", NewString(&Listen))
	defEnvs.Register("SMC_BASE_URL", "baseUrl",
		"Costrict cloud base url", "https://zgsm.sangfor.com", NewString(&BaseUrl))
	defEnvs.Register("SMC_MACHINE_ID", "machineId",
		"Costrict machine ID", "", NewString(&MachineId))
	defEnvs.Register("SMC_PLUGIN_VER", "pluginVersion",
		"Costrict plugin version", "", NewString(&PluginVersion))
	defEnvs.Register("SMC_VSCODE_VER", "vscodeVersion",
		"Costrict vscode version", "", NewString(&VscodeVersion))
	defEnvs.Register("SMC_SKIP_SSL", "skipSsl",
		"Skip SSL verification", "false", NewBool(&SkipSSL))

	defEnvs.Load(ConfigPath(".smc/smc.env"))
	defEnvs.SetOnChange(func() error {
		return defEnvs.Save(ConfigPath(".smc/smc.env"))
	})
}

/**
 *	Check if variable key allows user editing
 */
func EnableModify(key string) error {
	return defEnvs.EnableModify(key)
}

/**
 *	Get smc predefined environment variable value
 */
func GetEnv(key string) string {
	return defEnvs.GetEnv(key)
}

/**
 *	Set environment variable value
 */
func SetEnv(key, value string) error {
	return defEnvs.SetEnv(key, value)
}

/**
 *	Get environment variable value by alias
 */
func GetEnvByAlias(alias string) string {
	return defEnvs.GetEnvByAlias(alias)
}

/**
 *	Set environment variable value by alias
 */
func SetEnvByAlias(alias, value string) error {
	return defEnvs.SetEnvByAlias(alias, value)
}

/**
 *	Iterate all registered environment variables
 */
func VisitAll(doEnv func(name, alias, comment, defVal string, value OptValue) error) error {
	return defEnvs.VisitAll(doEnv)
}
