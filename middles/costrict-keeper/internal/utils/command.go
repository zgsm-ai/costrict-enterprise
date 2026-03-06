package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

func GetCommandLine(command string, args []string, data interface{}) (string, []string, error) {
	cmdTemplate, err := template.New("command").Parse(command)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse command template: %w", err)
	}

	var cmdBuf bytes.Buffer
	if err := cmdTemplate.Execute(&cmdBuf, data); err != nil {
		return "", nil, fmt.Errorf("failed to execute command template: %w", err)
	}

	// 处理Args模板
	var processedArgs []string
	for _, arg := range args {
		argTemplate, err := template.New("arg").Parse(arg)
		if err != nil {
			return "", nil, fmt.Errorf("failed to parse arg template '%s': %w", arg, err)
		}

		var argBuf bytes.Buffer
		if err := argTemplate.Execute(&argBuf, data); err != nil {
			return "", nil, fmt.Errorf("failed to execute arg template '%s': %w", arg, err)
		}

		processedArgs = append(processedArgs, strings.TrimSpace(argBuf.String()))
	}

	return cmdBuf.String(), processedArgs, nil
}
