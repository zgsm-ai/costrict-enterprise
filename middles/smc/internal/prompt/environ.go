package prompt

import (
	"time"

	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/rc"
)

/**
 * Get environment variable by ID
 * @param environId identifier of the environment variable
 * @return interface{} value of the environment variable
 * @return error if retrieval fails
 */
func GetEnviron(environId string) (interface{}, error) {
	var p interface{}
	err := rc.GetJSON(IDtoKey(environId, PREFIX_ENVIRONS), &p)
	if err != nil {
		return "", err
	}
	return p, nil
}

/**
 * List all environment variables
 * @return map[string]interface{} map of environment IDs to values
 * @return error if list operation fails
 */
func ListEnvirons() (map[string]interface{}, error) {
	keys, err := rc.KeysByPrefix(PREFIX_ENVIRONS)
	if err != nil {
		return nil, err
	}
	environs := make(map[string]interface{})
	for _, key := range keys {
		var p interface{}
		err := rc.GetJSON(key, &p)
		if err != nil {
			return nil, err
		}
		environs[KeyToID(key, PREFIX_ENVIRONS)] = p
	}
	return environs, nil
}

/**
 * Set/update an environment variable
 * @param environId identifier of the environment variable
 * @param environValue value to set
 * @return error if update fails
 */
func SetEnviron(environId string, environValue interface{}) error {
	return rc.SetJSON(IDtoKey(environId, PREFIX_ENVIRONS), environValue,
		time.Hour*time.Duration(env.RedisTimeout))
}
