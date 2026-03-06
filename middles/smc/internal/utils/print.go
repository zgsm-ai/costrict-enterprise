package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/iancoleman/orderedmap"
	"github.com/jedib0t/go-pretty/v6/table"
	"gopkg.in/yaml.v3"
)

/**
 * Format and print ordered map data as table
 * @param dataList slice of ordered maps to display
 * @return error if data is empty or formatting fails
 */
func PrintFormat(dataList []*orderedmap.OrderedMap) error {
	if len(dataList) == 0 {
		return fmt.Errorf("data list is empty")
	}
	// Get all keys
	keys := dataList[0].Keys()

	// Get table header
	header := make(table.Row, 0)
	for _, v := range keys {
		header = append(header, v)
	}

	// Get table contents
	rows := make([]table.Row, 0)
	for _, data := range dataList {
		row := make(table.Row, 0)
		for _, key := range keys {
			var value, _ = data.Get(key)
			// Check if value is slice type
			if reflect.TypeOf(value).Kind() == reflect.Slice {
				// Iterate through each slice element
				var jsonList []string
				for j := 0; j < reflect.ValueOf(value).Len(); j++ {
					elem := reflect.ValueOf(value).Index(j).Interface()
					// Check if element is struct type
					if reflect.TypeOf(elem).Kind() == reflect.Struct {
						// Convert struct to JSON string
						jsonBytes, err := json.Marshal(elem)
						if err != nil {
							return err
						}
						jsonList = append(jsonList, string(jsonBytes))
					} else {
						jsonList = append(jsonList, fmt.Sprintf("%v", elem))
					}
				}
				row = append(row, strings.Join(jsonList, ","))
			} else {
				row = append(row, value)
			}
		}
		rows = append(rows, row)
	}

	// Format and print data
	tt := table.NewWriter()
	tt.SetOutputMirror(os.Stdout)
	tt.AppendHeader(header)
	tt.AppendRows(rows)
	tt.Style().Options.DrawBorder = false
	tt.Style().Options.SeparateColumns = false
	tt.Style().Options.SeparateFooter = false
	tt.Style().Options.SeparateHeader = false
	tt.Style().Options.SeparateRows = false
	tt.Render()
	return nil
}

/**
 * Convert struct to ordered map
 * @param s struct to convert
 * @return *orderedmap.OrderedMap converted map
 * @return error if input is not struct type
 */
func StructToOrderedMap(s interface{}) (*orderedmap.OrderedMap, error) {
	values := reflect.ValueOf(s)
	fields := reflect.TypeOf(s)
	kind := fields.Kind()
	if kind != reflect.Struct {
		return nil, fmt.Errorf("parameter %s is not struct type", s)
	}
	m := orderedmap.New()

	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)
		value := values.Field(i).Interface()
		m.Set(field.Name, value)
	}
	return m, nil
}

/**
 * Print single struct as formatted table using ordered map
 * @param s struct to print
 * @return error if conversion or printing fails
 */
func PrintFormatByOrderMap(s interface{}) error {
	recordMap, err := StructToOrderedMap(s)
	if err != nil {
		return err
	}
	var dataList []*orderedmap.OrderedMap
	dataList = append(dataList, recordMap)
	err = PrintFormat(dataList)
	return nil
}

/**
 * Print array of items using callback to format each item
 * @param arr array of items to print
 * @param callback function to convert item to ordered map
 * @return error if conversion or printing fails
 */
func PrintArray(arr []interface{}, callback func(s interface{}) (*orderedmap.OrderedMap, error)) error {
	var dataList []*orderedmap.OrderedMap
	if len(arr) == 0 {
		return nil
	}
	for _, v := range arr {
		om, err := callback(v)
		if err != nil {
			return err
		}
		dataList = append(dataList, om)
	}
	return PrintFormat(dataList)
}

/**
 * Print data structure in YAML format
 * @param s data to print
 * @return error if YAML marshaling fails
 */
func PrintYaml(s interface{}) error {
	// Convert to YAML format and print
	yamlBytes, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	fmt.Println(string(yamlBytes))
	return nil
}

/**
 * Calculate time duration between two formatted times
 * Uses "2006-01-02 15:04:05" format if none specified
 */
func FormatDuration(layout string, startTime string, endTime string) (string, error) {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	startTimeParse, err := time.Parse(layout, startTime)
	if err != nil {
		return "", err
	}
	var endTimeParse time.Time
	if endTime == "" {
		endTimeParse = time.Now().Local()
	} else {
		endTimeParse, err = time.Parse(layout, endTime)
		if err != nil {
			return "", err
		}
	}

	duration := endTimeParse.Sub(startTimeParse)
	if duration < 0 {
		duration = 0
	}
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours), nil
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes), nil
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds), nil
	} else {
		return fmt.Sprintf("%ds", seconds), nil
	}
}
