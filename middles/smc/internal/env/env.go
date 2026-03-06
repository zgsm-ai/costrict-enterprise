package env

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

/**
 * Get the path to save COOKIE
 */
func ConfigPath(fname string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), fname)
	} else if runtime.GOOS == "linux" {
		u, err := user.Current()
		if err != nil {
			return filepath.Clean(fname)
		}
		return filepath.Join(u.HomeDir, fname)
	}
	return filepath.Clean(fname)
}

/**
 * Definition for predefined environment variables
 */
type EnvDef struct {
	Name    string
	Alias   string
	Comment string
	Default string
	Value   OptValue
}

/**
 * Predefined environment variable table
 */
type Envs struct {
	name2defs   map[string]*EnvDef
	alias2defs  map[string]*EnvDef
	envSlice    []*EnvDef
	envOnChange func() error
}

/**
 * Create a new table
 */
func NewEnvs() *Envs {
	envs := &Envs{}
	envs.name2defs = make(map[string]*EnvDef)
	envs.alias2defs = make(map[string]*EnvDef)
	return envs
}

/**
 * Register an environment variable
 */
func (envs *Envs) Register(name, alias, comment, defVal string, val OptValue) error {
	_, ok := envs.name2defs[name]
	if ok {
		return fmt.Errorf("env name '%s' already exist", name)
	}
	_, ok = envs.alias2defs[alias]
	if ok {
		return fmt.Errorf("env alias '%s' already exist", alias)
	}
	env := &EnvDef{}
	env.Name = name
	env.Alias = alias
	env.Comment = comment
	env.Default = defVal
	env.Value = val
	envs.name2defs[name] = env
	envs.alias2defs[alias] = env
	envs.envSlice = append(envs.envSlice, env)
	return nil
}

/**
 * Load environment variables from file and set to smc environment
 * ENV variables priority: .env file > system environment > default value
 * Values in these three locations will be kept consistent
 */
func (envs *Envs) Load(fname string) error {
	items := map[string]string{}
	// If ENV config file exists, load and set to system environment
	if _, err := os.Stat(fname); err == nil {
		bytes, err := os.ReadFile(fname)
		if err != nil {
			log.Println(err)
		}
		if err = json.Unmarshal(bytes, &items); err != nil {
			log.Println(err)
		}
	}
	// Sync all ENV variables from .env file to system environment
	for k, v := range items {
		if err := os.Setenv(k, v); err != nil {
			log.Println(err)
		}
	}
	// Check all predefined ENV variables: if exists in system env, read it;
	// otherwise use default value and sync to system environment
	for _, env := range envs.name2defs {
		val := os.Getenv(env.Name)
		if val != "" {
			if err := env.Value.Set(val); err != nil {
				log.Println(err)
			}
		} else {
			if err := env.Value.Set(env.Default); err != nil {
				log.Println(err)
			}
			if err := os.Setenv(env.Name, env.Default); err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}

/**
 * Write modified environment variables back to config file
 */
func (envs *Envs) Save(fname string) error {
	kvs := map[string]string{}
	for k, v := range envs.name2defs {
		kvs[k] = v.Value.String()
	}
	bytes, err := json.MarshalIndent(kvs, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(fname)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fname, bytes, 0664)
}

/**
 * Check if the variable can be modified by user
 */
func (envs *Envs) EnableModify(key string) error {
	env, ok := envs.name2defs[key]
	if !ok {
		return fmt.Errorf("%s not support", key)
	}
	if !env.Value.EnableModify() {
		return fmt.Errorf("%s is readonly", key)
	}
	return nil
}

/**
 * Get predefined environment variable value in smc
 */
func (envs *Envs) GetEnv(key string) string {
	env, ok := envs.name2defs[key]
	if !ok {
		log.Panicf("%s not support\n", key)
		return ""
	}
	return env.Value.String()
}

/**
 * Set environment variable value
 */
func (envs *Envs) SetEnv(key, value string) error {
	env, ok := envs.name2defs[key]
	if !ok {
		log.Panicf("%s not support\n", key)
		return fmt.Errorf("unsupported environment variable:%s", key)
	}
	if err := env.Value.Set(value); err != nil {
		return err
	}
	if err := os.Setenv(key, value); err != nil {
		return err
	}
	if envs.envOnChange != nil {
		return envs.envOnChange()
	}
	return nil
}

/**
 * Get predefined environment variable value in smc
 */
func (envs *Envs) GetEnvByAlias(alias string) string {
	env, ok := envs.alias2defs[alias]
	if !ok {
		log.Panicf("alias '%s' not support\n", alias)
		return ""
	}
	return env.Value.String()
}

/**
 * Set environment variable value
 */
func (envs *Envs) SetEnvByAlias(alias, value string) error {
	env, ok := envs.alias2defs[alias]
	if !ok {
		log.Panicf("alias '%s' not support\n", alias)
		return fmt.Errorf("unsupported environment variable alias:%s", alias)
	}
	if err := env.Value.Set(value); err != nil {
		return err
	}
	if err := os.Setenv(env.Name, value); err != nil {
		return err
	}
	if envs.envOnChange != nil {
		return envs.envOnChange()
	}
	return nil
}

/**
 * Set callback function when environment variables are modified
 */
func (envs *Envs) SetOnChange(onChange func() error) {
	envs.envOnChange = onChange
}

/**
 * Iterate through all registered environment variables
 */
func (envs *Envs) VisitAll(doEnv func(name, alias, comment, defVal string, value OptValue) error) error {
	for _, env := range envs.envSlice {
		if err := doEnv(env.Name, env.Alias, env.Comment, env.Default, env.Value); err != nil {
			return err
		}
	}
	return nil
}
