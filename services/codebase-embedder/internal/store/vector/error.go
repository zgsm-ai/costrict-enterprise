package vector

import (
	"errors"
	"fmt"

	"github.com/weaviate/weaviate/entities/models"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

var ErrInvalidCodebasePath = errors.New("invalid codebasePath")
var ErrInvalidClientId = errors.New("invalid clientId")
var ErrEmptyResponse = errors.New("response is empty")
var ErrInvalidResponse = errors.New("response is invalid")

// CheckBatchErrors 检查批量操作的结果并返回第一个错误
func CheckBatchErrors(responses []models.ObjectsGetResponse) error {
	for _, resp := range responses {
		// 检查是否有错误
		if resp.Result != nil && resp.Result.Errors != nil && len(resp.Result.Errors.Error) > 0 {
			for _, errItem := range resp.Result.Errors.Error {
				if errItem != nil && errItem.Message != types.EmptyString {
					// 获取对象 ID（如果可用）
					var objID string
					if resp.ID != types.EmptyString {
						objID = resp.ID.String()
					}

					// 只返回第一个错误
					if objID != types.EmptyString {
						return fmt.Errorf("object %s: %s", objID, errItem.Message)
					}
					return fmt.Errorf("batch error: %s", errItem.Message)
				}
			}
		}
	}
	return nil
}

// CheckGraphQLResponseError 检查 GraphQL 响应中是否包含错误
// 如果有错误，返回第一个错误信息
func CheckGraphQLResponseError(resp *models.GraphQLResponse) error {
	if resp == nil || len(resp.Errors) == 0 {
		return nil
	}

	// 获取第一个错误
	firstError := resp.Errors[0]
	if firstError == nil || firstError.Message == "" {
		return fmt.Errorf("graphql error (no message provided)")
	}

	// 返回错误信息，包含路径信息（如果有）
	if len(firstError.Path) > 0 {
		return fmt.Errorf("graphql error at path %v: %s", firstError.Path, firstError.Message)
	}

	return fmt.Errorf("graphql error: %s", firstError.Message)
}

// CheckBatchDeleteErrors 检查批量删除响应中是否有失败的对象
// 如果有失败的对象，返回第一个错误信息
func CheckBatchDeleteErrors(resp *models.BatchDeleteResponse) error {
	if resp == nil || resp.Results == nil {
		return fmt.Errorf("invalid batch delete response")
	}

	// 如果没有失败的对象，返回 nil
	if resp.Results.Failed == 0 {
		return nil
	}

	// 查找第一个失败的对象
	for _, obj := range resp.Results.Objects {
		if obj == nil || obj.Status == nil {
			continue
		}

		// 如果对象状态为 "FAILED"
		if *obj.Status == models.BatchDeleteResponseResultsObjectsItems0StatusFAILED && obj.Errors != nil && len(obj.Errors.Error) > 0 {
			// 获取第一个错误
			firstError := obj.Errors.Error[0]
			if firstError != nil && firstError.Message != types.EmptyString {
				return fmt.Errorf("batch delete failed for object %s: %s", obj.ID, firstError.Message)
			}
		}
	}

	// 如果没有找到具体的错误信息，但有失败的对象
	return fmt.Errorf("batch delete failed: %d objects failed", resp.Results.Failed)
}
