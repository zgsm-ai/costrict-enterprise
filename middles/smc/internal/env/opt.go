package env

import (
	"fmt"
	"regexp"
	"strconv"
)

/**
 *	Environment variable value, accepts values set via env vars and .env files
 */
type OptValue interface {
	Set(val string) error //Modify value
	String() string       //Get current value
	EnableModify() bool   //Allow user modification?
}

/**
 *	String type environment variable value
 */
type StringOptValue struct {
	ptr *string
	reg *regexp.Regexp
}

/**
 *	Integer type environment variable value
 */
type IntOptValue struct {
	ptr *int
}

/**
 *	Boolean type environment variable value
 */
type BoolOptValue struct {
	ptr *bool
}

/**
 *	Read-only environment variable value
 */
type ReadOnlyOptValue struct {
	ptr *string
}

/**
 *	Create a new string type option value
 */
func NewString(ptr *string) *StringOptValue {
	v := &StringOptValue{}
	v.ptr = ptr
	v.reg = nil
	return v
}

/**
 *	Create a string option value that accepts specific input
 */
func NewLimitedString(ptr *string, reg *regexp.Regexp) *StringOptValue {
	v := &StringOptValue{}
	v.ptr = ptr
	v.reg = reg
	return v
}

/**
 *	Create a new read-only string type option value
 */
func NewReadOnly(ptr *string) *ReadOnlyOptValue {
	v := &ReadOnlyOptValue{}
	v.ptr = ptr
	return v
}

/**
 *	Create a new integer type option value
 */
func NewInt(ptr *int) *IntOptValue {
	v := &IntOptValue{}
	v.ptr = ptr
	return v
}

/**
 *	Create a new BOOL type option value
 */
func NewBool(ptr *bool) *BoolOptValue {
	v := &BoolOptValue{}
	v.ptr = ptr
	return v
}

/**
 *	Set string type option value
 */
func (v StringOptValue) Set(val string) error {
	if v.reg != nil && !v.reg.MatchString(val) {
		return fmt.Errorf("invalid value: %s, expect: %v", val, v.reg)
	}
	*v.ptr = val
	return nil
}

/**
 *	Get option value as string
 */
func (v StringOptValue) String() string {
	return *v.ptr
}

/**
 *	Whether to allow user modification
 */
func (v StringOptValue) EnableModify() bool {
	return true
}

/**
 *	Set read-only string type option value
 */
func (v ReadOnlyOptValue) Set(val string) error {
	*v.ptr = val
	return nil
}

/**
 *	Get read-only option value as string
 */
func (v ReadOnlyOptValue) String() string {
	return *v.ptr
}

/**
 *	Whether to allow user modification
 */
func (v ReadOnlyOptValue) EnableModify() bool {
	return false
}

/**
 *	Set integer type option value
 */
func (v IntOptValue) Set(val string) error {
	iv, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	*v.ptr = iv
	return nil
}

/**
 *	Convert integer type option value to string form
 */
func (v IntOptValue) String() string {
	return fmt.Sprintf("%d", *v.ptr)
}

/**
 *	Whether to allow user modification
 */
func (v IntOptValue) EnableModify() bool {
	return true
}

/**
 *	Get BOOL type option value
 */
func (v BoolOptValue) Set(val string) error {
	if val == "ENABLE" || val == "enable" || val == "TRUE" || val == "true" {
		*v.ptr = true
	} else if val == "DISABLE" || val == "disable" || val == "FALSE" || val == "false" {
		*v.ptr = false
	} else {
		return fmt.Errorf("invalid bool value: %s", val)
	}
	return nil
}

/**
 *	Convert BOOL type option value to string form
 */
func (v BoolOptValue) String() string {
	if *v.ptr {
		return "ENABLE"
	} else {
		return "DISABLE"
	}
}

/**
 *	Whether to allow user modification
 */
func (v BoolOptValue) EnableModify() bool {
	return true
}
