package client

import (
	"encoding/json"
	"time"

	"github.com/spf13/cobra"
	"github.com/zgsm-ai/smc/cmd/common"
	"github.com/zgsm-ai/smc/internal/env"
	"github.com/zgsm-ai/smc/internal/utils"
)

type Log struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ClientID    string    `json:"client_id" gorm:"index;not null"`
	UserID      string    `json:"user_id" gorm:"index"`
	FileName    string    `json:"file_name" gorm:"index;not null"`
	FirstLineNo int64     `json:"first_line_no"`
	LastLineNo  int64     `json:"end_line_no"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Paginated struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type LogsResponse struct {
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Data    []Log     `json:"data"`
	Paging  Paginated `json:"paging"`
}

func listLogs() error {
	ss := common.Session

	args := make(utils.Json)
	if optLogsClientID != "" {
		args["client_id"] = optLogsClientID
	}
	data, err := ss.Get("/client-manager/api/v1/logs", args)
	if err != nil {
		return err
	}
	var res LogsResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	return nil
}

func downloadLogs() error {
	return nil
}

var logsCmd = &cobra.Command{
	Use:   "logs [client-id | -c client-id] [-f file-name] [-o output-directory]",
	Short: "view logs",
	Long:  `view logs`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			optLogsClientID = args[0]
		}
		var err error
		if err = common.InitCommonEnv(); err != nil {
			return err
		}
		common.Session = utils.NewSession(env.BaseUrl)
		if optLogsClientID != "" {
			downloadLogs()
		} else {
			listLogs()
		}
		return nil
	},
}

const logsExample = `  # logs package
  smc client logs
  smc client logs -c xxx`

var optLogsClientID string
var optLogsFileName string
var optOutputDir string

func init() {
	clientCmd.AddCommand(logsCmd)
	logsCmd.Flags().SortFlags = false
	logsCmd.Example = logsExample

	logsCmd.Flags().StringVarP(&optLogsClientID, "client", "c", "", "client id")
	logsCmd.Flags().StringVarP(&optLogsFileName, "file", "f", "", "log file name")
	logsCmd.Flags().StringVarP(&optOutputDir, "output", "o", "./output", "output directory")
}
