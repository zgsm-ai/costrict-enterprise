package logic

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"io"
	"net/http"
	"os"
	"strings"

	"github.com/zgsm-ai/codebase-indexer/internal/dao/model"
	"github.com/zgsm-ai/codebase-indexer/internal/job"
	"github.com/zgsm-ai/codebase-indexer/internal/store/vector"
	"github.com/zgsm-ai/codebase-indexer/internal/tracer"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
	"github.com/zgsm-ai/codebase-indexer/pkg/utils"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zgsm-ai/codebase-indexer/internal/svc"
)

func extractFileOperations(metadata *types.SyncMetadata) map[string]string {
	operations := make(map[string]string)

	// 遍历FileList，该字段存储了文件路径到操作类型的映射
	for filePath, operation := range metadata.FileList {
		operations[filePath] = operation
	}

	return operations
}

type TaskLogic struct {
	logx.Logger
	ctx           context.Context
	svcCtx        *svc.ServiceContext
	syncMetadata  *types.SyncMetadata
	uploadedFiles map[string][]byte // 存储上传的文件内容
}

func NewTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskLogic {
	return &TaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		syncMetadata: &types.SyncMetadata{
			ClientId:      "",
			CodebasePath:  "",
			CodebaseName:  "",
			ExtraMetadata: make(map[string]types.MetadataValue),
			FileList:      make(map[string]string),
			FileListItems: []types.FileListItem{},
			Timestamp:     0,
		},
		uploadedFiles: make(map[string][]byte),
	}
}

func (l *TaskLogic) SubmitTask(req *types.IndexTaskRequest, r *http.Request) (resp *types.IndexTaskResponseData, err error) {
	startTime := time.Now()
	clientId := req.ClientId
	clientPath := req.CodebasePath
	codebaseName := req.CodebaseName
	uploadToken := req.UploadToken

	l.Logger.Infof("SubmitTask 开始执行 - RequestId: %s, ClientId: %s, CodebasePath: %s, CodebaseName: %s,uploadToken: %s",
		req.RequestId, clientId, clientPath, codebaseName, uploadToken)

	// 在函数结束时记录执行时间
	defer func() {
		duration := time.Since(startTime)
		l.Logger.Infof("SubmitTask 执行完成 - RequestId: %s, 总耗时: %v", req.RequestId, duration)
	}()

	// 验证uploadToken的有效性
	l.Logger.Infof("验证uploadToken开始 - RequestId: %s", req.RequestId)
	if err := l.validateUploadToken(uploadToken); err != nil {
		l.Logger.Errorf("验证uploadToken失败 - RequestId: %s, 错误: %v", req.RequestId, err)
		return nil, err
	}
	l.Logger.Infof("验证uploadToken成功 - RequestId: %s", req.RequestId)

	userUid := utils.ParseJWTUserInfo(r, l.svcCtx.Config.Auth.UserInfoHeader)
	l.Logger.Infof("解析用户信息完成 - RequestId: %s, UserUid: %s", req.RequestId, userUid)

	// 查找代码库记录，不存在则初始化
	l.Logger.Infof("开始初始化代码库 - RequestId: %s, ClientId: %s, CodebasePath: %s", req.RequestId, clientId, clientPath)
	codebase, err := l.initCodebaseIfNotExists(clientId, clientPath, userUid, codebaseName)
	if err != nil {
		l.Logger.Errorf("初始化代码库失败 - RequestId: %s, 错误: %v", req.RequestId, err)
		return nil, err
	}
	codebase.Name = clientId // 更新代码库名称
	codebase.Path = clientPath
	l.Logger.Infof("初始化代码库成功 - RequestId: %s, CodebaseId: %d", req.RequestId, codebase.ID)

	ctx := context.WithValue(l.ctx, tracer.Key, tracer.RequestTraceId(int(codebase.ID)))

	// 处理上传的ZIP文件
	l.Logger.Infof("开始处理上传的ZIP文件 - RequestId: %s", req.RequestId)
	files, fileCount, metadata, err := l.processUploadedZipFile(r)
	if err != nil {
		l.Logger.Errorf("处理ZIP文件失败 - RequestId: %s, 错误: %v", req.RequestId, err)
		return nil, err
	}
	l.Logger.Infof("处理ZIP文件成功 - RequestId: %s, 文件数量: %d", req.RequestId, fileCount)

	// 存储上传的文件内容，供重命名操作使用
	l.uploadedFiles = files

	// 遍历任务并分类
	var addTasks, deleteTasks, modifyTasks []string
	var renameTasks []types.FileListItem

	if l.syncMetadata != nil {
		// 处理格式一的FileList（map格式）
		for key, value := range l.syncMetadata.FileList {
			switch strings.ToLower(value) {
			case "add":
				addTasks = append(addTasks, key)
			case "delete":
				deleteTasks = append(deleteTasks, key)
			case "modify":
				modifyTasks = append(modifyTasks, key)
			default:
				l.Logger.Errorf("未知的操作类型 %s 对于文件 %s", value, key)
			}
		}

		// 处理格式二的FileListItems（数组格式）
		for _, item := range l.syncMetadata.FileListItems {
			switch strings.ToLower(item.Status) {
			case "add":
				addTasks = append(addTasks, item.Path)
			case "delete":
				deleteTasks = append(deleteTasks, item.Path)
			case "modify":
				modifyTasks = append(modifyTasks, item.Path)
			case "rename":
				renameTasks = append(renameTasks, item)
			default:
				l.Logger.Errorf("未知的操作类型 %s 对于文件 %s", item.Status, item.Path)
			}
		}
	}

	// 记录任务分类统计
	l.Logger.Infof("任务分类统计 - 添加: %d, 删除: %d, 修改: %d, 重命名: %d",
		len(addTasks), len(deleteTasks), len(modifyTasks), len(renameTasks))

	/////////////////////////////////////////////////////////////
	//初始化任务状态
	/////////////////////////////////////////////

	fileOperations := make(map[string]string)
	if metadata != nil {
		fileOperations = extractFileOperations(metadata)
	}

	l.Logger.Infof("任务分类统计 - 添加: %d, 删除: %d, 修改: %d, 重命名: %d", len(addTasks), len(deleteTasks), len(modifyTasks), len(renameTasks))

	// 状态修改为处理中
	l.svcCtx.StatusManager.UpdateFileStatus(ctx, req.RequestId,
		func(status *types.FileStatusResponseData) {
			status.Process = "processing"
			status.TotalProgress = 0
			var fileStatusItems []types.FileStatusItem

			// 添加任务
			// for path, _ := range files {
			// 	fileStatusItem := types.FileStatusItem{
			// 		Path:    path, // 使用当前处理的文件路径，而不是codebasePath
			// 		Status:  "processing",
			// 		Operate: fileOperations[path],
			// 	}
			// 	fileStatusItems = append(fileStatusItems, fileStatusItem)
			// }

			// 删除任务
			for path, op := range fileOperations {
				fileStatusItem := types.FileStatusItem{
					Path:    path, // 使用当前处理的文件路径，而不是codebasePath
					Status:  "processing",
					Operate: op,
				}
				fileStatusItems = append(fileStatusItems, fileStatusItem)
			}

			// 重命名任务
			for _, item := range renameTasks {
				fileStatusItem := types.FileStatusItem{
					Path:    item.TargetPath, // 目标路径
					Status:  "processing",
					Operate: "rename",
				}
				fileStatusItems = append(fileStatusItems, fileStatusItem)
			}

			status.FileList = fileStatusItems
			l.Logger.Infof("初始化状态： - RequestId: %s , %v", req.RequestId, status.FileList)
		})

	/////////////////////////////////////////////////////////////
	//执行任务
	/////////////////////////////////////////////

	// 如果有删除任务，从向量数据库中删除对应的文件
	if len(deleteTasks) > 0 {
		l.Logger.Infof("开始从向量数据库中删除 %d 个文件", len(deleteTasks))
		if err := l.deleteFilesFromVectorDB(ctx, codebase, deleteTasks); err != nil {
			l.Logger.Errorf("从向量数据库删除文件失败: %v", err)
			// 不返回错误，继续处理其他任务
		} else {
			l.Logger.Infof("成功从向量数据库中删除 %d 个文件", len(deleteTasks))
		}
	}

	l.Logger.Infof("任务分类统计 - 添加: %d, 删除: %d, 修改: %d, 重命名: %d",
		len(addTasks), len(deleteTasks), len(modifyTasks), len(renameTasks))

	// 执行重命名任务
	if len(renameTasks) > 0 {
		l.Logger.Infof("开始执行 %d 个重命名任务", len(renameTasks))
		if err := l.executeRenameTasks(ctx, clientId, req.CodebasePath, renameTasks); err != nil {
			l.Logger.Errorf("执行重命名任务失败: %v", err)
			// 不返回错误，继续处理其他任务
		} else {
			l.Logger.Infof("成功执行 %d 个重命名任务", len(renameTasks))
		}
	}

	// 更新代码库信息
	l.Logger.Infof("开始更新代码库信息 - RequestId: %s, CodebaseId: %d, 文件数量: %d", req.RequestId, codebase.ID, fileCount)
	if err := l.updateCodebaseInfo(codebase, fileCount, int64(req.FileTotals)); err != nil {
		l.Logger.Errorf("更新代码库信息失败 - RequestId: %s, 错误: %v", req.RequestId, err)
		return nil, err
	}
	l.Logger.Infof("更新代码库信息成功 - RequestId: %s", req.RequestId)

	// 提交索引任务
	l.Logger.Infof("开始提交索引任务 - RequestId: %s, 文件数量: %d", req.RequestId, len(files))

	// 检查文件处理个数是否为0，如果为0则标识完成状态，不提交submitIndexTask任务
	if len(files) == 0 {
		l.Logger.Infof("文件处理个数为0，直接标识完成状态 - RequestId: %s", req.RequestId)

		l.svcCtx.StatusManager.UpdateFileStatus(ctx, req.RequestId, func(status *types.FileStatusResponseData) {
			status.Process = "completed"
			status.TotalProgress = 100

			for i, _ := range status.FileList {
				status.FileList[i].Status = "completed"
			}

		})

		l.Logger.Infof("初始化文件处理状态为完成成功 - RequestId: %s", req.RequestId)
	} else {
		// 文件数量大于0，正常提交索引任务
		if err := l.submitIndexTask(ctx, codebase, clientId, req.RequestId, files, metadata); err != nil {
			l.Logger.Errorf("提交索引任务失败 - RequestId: %s, 错误: %v", req.RequestId, err)
			return nil, err
		}
		l.Logger.Infof("提交索引任务成功 - RequestId: %s", req.RequestId)
	}

	return &types.IndexTaskResponseData{TaskId: req.RequestId}, nil
}

// validateUploadToken 验证上传令牌的有效性
func (l *TaskLogic) validateUploadToken(uploadToken string) error {
	return nil
}

// processUploadedZipFile 处理上传的ZIP文件
func (l *TaskLogic) processUploadedZipFile(r *http.Request) (map[string][]byte, int, *types.SyncMetadata, error) {
	// 解析multipart表单
	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}
	defer r.MultipartForm.RemoveAll()

	// 从表单中获取ZIP文件
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get file from form: %w", err)
	}
	defer file.Close()

	// 验证文件是否为ZIP格式
	if !strings.HasSuffix(header.Filename, ".zip") {
		return nil, 0, nil, fmt.Errorf("uploaded file must be a ZIP file, got: %s", header.Filename)
	}

	// 处理ZIP文件内容
	return l.extractZipFiles(file)
}

// extractZipFiles 从ZIP文件中提取文件内容
func (l *TaskLogic) extractZipFiles(file io.Reader) (map[string][]byte, int, *types.SyncMetadata, error) {
	files := make(map[string][]byte)

	// 创建临时文件存储上传的ZIP
	tempFile, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // 清理临时文件

	tracer.WithTrace(l.ctx).Infof("extractZipFiles tempPath %s", tempPath)

	// 将上传的ZIP内容复制到临时文件
	_, err = io.Copy(tempFile, file)
	tempFile.Close() // 关闭文件以便后续读取
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to copy file to temp location: %w", err)
	}

	// 打开ZIP文件进行读取
	zipReader, err := zip.OpenReader(tempPath)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer zipReader.Close()

	// 检查是否存在.shenma_sync文件夹
	if !l.hasShenmaSyncFolder(zipReader) {
		return nil, 0, nil, fmt.Errorf("ZIP文件中必须包含.shenma_sync文件夹")
	}

	// 提取文件内容
	fileCount, err := l.extractFilesFromZip(zipReader, files)
	if err != nil {
		return nil, 0, nil, err
	}

	// 获取元数据
	metadata := l.getSyncMetadata()

	return files, fileCount, metadata, nil
}

// hasShenmaSyncFolder 检查ZIP中是否存在.shenma_sync文件夹
func (l *TaskLogic) hasShenmaSyncFolder(zipReader *zip.ReadCloser) bool {
	for _, zipFile := range zipReader.File {
		if strings.HasPrefix(zipFile.Name, ".shenma_sync/") {
			return true
		}
	}
	return false
}

// extractFilesFromZip 从ZIP中提取文件内容
func (l *TaskLogic) extractFilesFromZip(zipReader *zip.ReadCloser, files map[string][]byte) (int, error) {
	fileCount := 0
	shenmaSyncFiles := make(map[string][]byte)

	// 先找到控制源文件
	for _, zipFile := range zipReader.File {
		// 跳过目录
		if zipFile.FileInfo().IsDir() {
			continue
		}

		// 处理.shenma_sync文件夹中的文件
		if strings.HasPrefix(zipFile.Name, ".shenma_sync/") {
			if err := l.processShenmaSyncFile(zipFile, shenmaSyncFiles); err != nil {
				return 0, err
			}
			break
		}
	}

	// 遍历ZIP中的所有文件
	for _, zipFile := range zipReader.File {
		// 跳过目录
		if zipFile.FileInfo().IsDir() {
			continue
		}

		// 检查文件是否存在于ExtraMetadata中，如果不存在则忽略
		if l.syncMetadata != nil {
			// 将zipFile.Name中的Windows路径格式（反斜杠\）转换为Linux路径格式（正斜杠/）
			linuxPath := strings.ReplaceAll(zipFile.Name, "\\", "/")
			if _, exists := l.syncMetadata.FileList[linuxPath]; !exists {
				// l.Logger.Infof("文件 %s 不存在于syncMetadata.FileList中，跳过处理 %v", zipFile.Name, l.syncMetadata.FileList)
				continue
			} else {
				if l.syncMetadata.FileList[linuxPath] == "delete" {
					continue
				}
			}
		}

		// 处理普通文件
		fileCount++
		if err := l.processRegularFile(zipFile, files); err != nil {
			return 0, err
		}
	}

	// 打印.shenma_sync文件夹中的文件摘要
	l.Logger.Infof("共找到 %d 个.shenma_sync文件夹中的文件", len(shenmaSyncFiles))
	for fileName := range shenmaSyncFiles {
		l.Logger.Infof(" - %s", fileName)
	}

	return fileCount, nil
}

// processShenmaSyncFile 处理.shenma_sync文件夹中的文件
func (l *TaskLogic) processShenmaSyncFile(zipFile *zip.File, shenmaSyncFiles map[string][]byte) error {
	fileReader, err := zipFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open file %s in zip: %w", zipFile.Name, err)
	}

	content, err := io.ReadAll(fileReader)
	fileReader.Close()
	if err != nil {
		return fmt.Errorf("failed to read file %s in zip: %w", zipFile.Name, err)
	}

	shenmaSyncFiles[zipFile.Name] = content
	l.Logger.Infof("读取.shenma_sync文件夹中的文件: %s", zipFile.Name)
	l.Logger.Infof("文件内容:\n%s", string(content))

	// 解析JSON格式的.shenma_sync文件内容并提取fileList
	l.extractFileListFromShenmaSync(content, zipFile.Name)

	// 额外输出到控制台，确保用户能看到
	fmt.Printf("=== .shenma_sync文件内容 ===\n")
	fmt.Printf("文件名: %s\n", zipFile.Name)
	fmt.Printf("内容:\n%s\n", string(content))
	fmt.Printf("========================\n\n")

	return nil
}

// extractFileListFromShenmaSync 从.shenma_sync文件内容中提取fileList
func (l *TaskLogic) extractFileListFromShenmaSync(content []byte, fileName string) {
	// 解析JSON内容
	var metadata types.SyncMetadata
	metadata.FileList = make(map[string]string)
	metadata.ExtraMetadata = make(map[string]types.MetadataValue)

	// 先解析为通用类型，然后转换为MetadataValue
	var tempMetadata struct {
		ClientId      string                 `json:"clientId"`
		CodebasePath  string                 `json:"codebasePath"`
		CodebaseName  string                 `json:"codebaseName"`
		ExtraMetadata map[string]interface{} `json:"extraMetadata"`
		FileList      interface{}            `json:"fileList"` // 使用interface{}来兼容两种格式
		Timestamp     int64                  `json:"timestamp"`
	}

	if err := json.Unmarshal(content, &tempMetadata); err != nil {
		l.Logger.Errorf("解析.shenma_sync文件失败 %s: %v", fileName, err)
		return
	}

	// 转换ExtraMetadata的类型
	for key, value := range tempMetadata.ExtraMetadata {
		switch v := value.(type) {
		case string:
			metadata.ExtraMetadata[key] = types.NewStringMetadataValue(v)
		case float64:
			metadata.ExtraMetadata[key] = types.NewNumberMetadataValue(v)
		case bool:
			metadata.ExtraMetadata[key] = types.NewBoolMetadataValue(v)
		case []interface{}:
			// 处理数组类型
			if len(v) > 0 {
				switch v[0].(type) {
				case string:
					strSlice := make([]string, len(v))
					for i, elem := range v {
						strSlice[i] = elem.(string)
					}
					metadata.ExtraMetadata[key] = types.NewStringArrayMetadataValue(strSlice)
				case float64:
					numSlice := make([]float64, len(v))
					for i, elem := range v {
						numSlice[i] = elem.(float64)
					}
					metadata.ExtraMetadata[key] = types.NewNumberArrayMetadataValue(numSlice)
				}
			}
		}
	}

	metadata.ClientId = tempMetadata.ClientId
	metadata.CodebasePath = tempMetadata.CodebasePath
	metadata.CodebaseName = tempMetadata.CodebaseName
	metadata.Timestamp = tempMetadata.Timestamp

	// 处理FileList的两种格式
	switch fileList := tempMetadata.FileList.(type) {
	case map[string]interface{}:
		// 格式1: "fileList":{"pkg/codegraph/analyzer/package_classifier/cpp_classifier.go":"add"}
		l.Logger.Infof("检测到对象格式的fileList")
		for filePath, operation := range fileList {
			if opStr, ok := operation.(string); ok {
				metadata.FileList[filePath] = opStr
			} else {
				l.Logger.Errorf("fileList中文件 %s 的操作类型不是字符串: %v", filePath, operation)
			}
		}
	case []interface{}:
		// 格式2: "fileList":[{"path":"pkg/codegraph/proto/codegraphpb/types.pb.go","targetPath":"","hash":"1755050845505","status":"modify","requestId":""}]
		l.Logger.Infof("检测到数组格式的fileList")
		for _, item := range fileList {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if path, ok := itemMap["path"].(string); ok {
					// 优先使用status字段，如果没有则使用其他字段作为操作类型
					var operation string
					if status, ok := itemMap["status"].(string); ok {
						operation = status
					} else if operate, ok := itemMap["operate"].(string); ok {
						operation = operate
					} else {
						operation = "unknown"
						l.Logger.Errorf("fileList数组项中未找到status或operate字段: %v", itemMap)
					}

					// 创建FileListItem
					fileListItem := types.FileListItem{
						Path: path,
					}

					// 解析可选字段
					if targetPath, ok := itemMap["targetPath"].(string); ok {
						fileListItem.TargetPath = targetPath
					}
					if hash, ok := itemMap["hash"].(string); ok {
						fileListItem.Hash = hash
					}
					if status, ok := itemMap["status"].(string); ok {
						fileListItem.Status = status
					}
					if operate, ok := itemMap["operate"].(string); ok {
						fileListItem.Operate = operate
					}
					if requestId, ok := itemMap["requestId"].(string); ok {
						fileListItem.RequestId = requestId
					}

					// 添加到FileListItems数组
					metadata.FileListItems = append(metadata.FileListItems, fileListItem)

					// 同时兼容原有的FileList map格式（非rename操作）
					if operation != "rename" {
						metadata.FileList[path] = operation
					}
				} else {
					l.Logger.Errorf("fileList数组项中缺少path字段: %v", itemMap)
				}
			} else {
				l.Logger.Errorf("fileList数组项不是map类型: %v", item)
			}
		}
	default:
		l.Logger.Errorf("不支持的fileList格式: %T", tempMetadata.FileList)
		return
	}

	l.Logger.Infof("从 %s 中提取到 %d 个文件:", fileName, len(metadata.FileList))

	// 打印fileList中的文件
	for filePath, status := range metadata.FileList {
		l.Logger.Infof("  文件: %s, 状态: %s", filePath, status)
		fmt.Printf("  文件: %s, 状态: %s\n", filePath, status)
	}

	// 存储提取的元数据
	l.syncMetadata = &metadata
}

// processRegularFile 处理常规文件
func (l *TaskLogic) processRegularFile(zipFile *zip.File, files map[string][]byte) error {
	fileReader, err := zipFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open file %s in zip: %w", zipFile.Name, err)
	}

	content, err := io.ReadAll(fileReader)
	fileReader.Close()
	if err != nil {
		return fmt.Errorf("failed to read file %s in zip: %w", zipFile.Name, err)
	}

	// 存储文件内容到映射
	files[zipFile.Name] = content
	return nil
}

// updateCodebaseInfo 更新代码库信息
func (l *TaskLogic) updateCodebaseInfo(codebase *model.Codebase, fileCount int, fileTotals int64) error {
	// 更新codebase的file_count和total_size字段
	codebase.FileCount = int32(fileCount)
	codebase.TotalSize = fileTotals
	err := l.svcCtx.Querier.Codebase.WithContext(l.ctx).Save(codebase)
	if err != nil {
		return fmt.Errorf("failed to update codebase file count: %w", err)
	}

	l.Logger.Infof("Updated codebase %d with file_count: %d, total_size: %d", codebase.ID, fileCount, fileTotals)
	return nil
}

// submitIndexTask 提交索引任务
func (l *TaskLogic) submitIndexTask(ctx context.Context, codebase *model.Codebase, clientId, requestId string, files map[string][]byte, metadata *types.SyncMetadata) error {
	startTime := time.Now()
	l.Logger.Infof("开始创建索引任务 - RequestId: %s, CodebaseId: %d, 文件数量: %d", requestId, codebase.ID, len(files))

	task := &job.IndexTask{
		SvcCtx: l.svcCtx,
		Params: &job.IndexTaskParams{
			ClientId:     clientId,
			CodebaseID:   codebase.ID,
			CodebasePath: codebase.Path,
			CodebaseName: codebase.Name,
			RequestId:    requestId,
			Files:        files,
			Metadata:     metadata,
			TotalFiles:   len(files),
		},
	}

	runningTasks := l.svcCtx.TaskPool.Running()
	taskCapacity := l.svcCtx.TaskPool.Cap()
	l.Logger.Infof("任务池状态 - RequestId: %s, 正在运行任务: %d, 任务容量: %d", requestId, runningTasks, taskCapacity)

	// 使用任务池提交任务

	l.Logger.Infof("开始提交任务到任务池 - RequestId: %s, 超时时间: %v", requestId, l.svcCtx.Config.IndexTask.GraphTask.Timeout)
	err := l.svcCtx.TaskPool.Submit(func() {
		taskStartTime := time.Now()
		l.Logger.Infof("任务开始执行 - RequestId: %s", requestId)

		taskTimeout, cancelFunc := context.WithTimeout(context.Background(), l.svcCtx.Config.IndexTask.GraphTask.Timeout)
		traceCtx := context.WithValue(taskTimeout, tracer.Key, tracer.TaskTraceId(int(codebase.ID)))
		defer cancelFunc()

		task.Run(traceCtx)

		taskDuration := time.Since(taskStartTime)
		l.Logger.Infof("任务执行完成 - RequestId: %s, 任务执行耗时: %v", requestId, taskDuration)
	})

	submitDuration := time.Since(startTime)
	if err != nil {
		l.Logger.Errorf("提交任务到任务池失败 - RequestId: %s, 提交耗时: %v, 错误: %v", requestId, submitDuration, err)
		return fmt.Errorf("index task submit failed, err:%w", err)
	}

	l.Logger.Infof("成功提交任务到任务池 - RequestId: %s, 提交耗时: %v", requestId, submitDuration)
	tracer.WithTrace(ctx).Infof("index task submit successfully.")
	return nil
}

func (l *TaskLogic) initCodebaseIfNotExists(clientId, clientPath, userUid, codebaseName string) (*model.Codebase, error) {
	var codebase *model.Codebase
	var err error
	// 判断数据库记录是否存在 ，状态为 active
	codebase, err = l.svcCtx.Querier.Codebase.FindByClientIdAndPath(l.ctx, clientId, clientPath)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if codebase == nil {
		codebase, err = l.saveCodebase(clientId, clientPath, userUid, codebaseName)
	}

	return codebase, nil
}

/**
 * @Description: 初始化 codebase
 * @receiver l
 * @param clientId
 * @param clientPath
 * @param r
 * @param codebaseName
 * @param metadata
 * @return error
 * @return bool
 */
func (l *TaskLogic) saveCodebase(clientId, clientPath, userUId, codebaseName string) (*model.Codebase, error) {
	// 不存在则插入
	// clientId + codebasepath 为联合唯一索引
	// 保存到数据库
	codebaseModel := &model.Codebase{
		ClientID:   clientId,
		UserID:     userUId,
		Name:       codebaseName,
		ClientPath: clientPath,
		Status:     string(model.CodebaseStatusActive),
		Path:       clientPath,
	}
	err := l.svcCtx.Querier.Codebase.WithContext(l.ctx).Save(codebaseModel)
	if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
		// 不是 唯一索引冲突
		return nil, err
	}
	return codebaseModel, nil
}

// deleteFilesFromVectorDB 从向量数据库中删除指定的文件
func (l *TaskLogic) deleteFilesFromVectorDB(ctx context.Context, codebase *model.Codebase, filePaths []string) error {
	if len(filePaths) == 0 {
		return nil // 没有文件需要删除
	}

	l.Logger.Infof("准备从向量数据库中删除 %d 个文件，代码库ID: %d", len(filePaths), codebase.ID)

	for _, filePath := range filePaths {
		// 将文件路径转换为Linux格式（正斜杠）
		linuxPath := strings.ReplaceAll(filePath, "\\", "/")
		l.Logger.Debugf("添加文件到删除列表: %s", linuxPath)
		l.svcCtx.VectorStore.DeleteDictionary(ctx, filePath, vector.Options{CodebaseId: codebase.ID, CodebasePath: codebase.Path})
	}

	l.Logger.Infof("成功从向量数据库中删除了 %d 个文件", len(filePaths))
	return nil
}

// getSyncMetadata 获取同步元数据
func (l *TaskLogic) getSyncMetadata() *types.SyncMetadata {
	// 返回从ZIP文件中提取的元数据
	return l.syncMetadata
}

// executeRenameTasks 执行重命名任务
func (l *TaskLogic) executeRenameTasks(ctx context.Context, clientId string, codebasePath string, renameTasks []types.FileListItem) error {
	for _, task := range renameTasks {
		l.Logger.Infof("开始执行重命名任务 - 源路径: %s, 目标路径: %s", task.Path, task.TargetPath)

		// 更新向量数据库中的文件路径
		if err := l.renameFileInVectorDB(ctx, clientId, codebasePath, task.Path, task.TargetPath); err != nil {
			l.Logger.Errorf("重命名文件失败 - 源路径: %s, 目标路径: %s, 错误: %v",
				task.Path, task.TargetPath, err)
			return err
		}

		// 更新任务状态
		l.svcCtx.StatusManager.UpdateFileStatus(ctx, l.getTaskRequestId(ctx),
			func(status *types.FileStatusResponseData) {
				for i, item := range status.FileList {
					if item.Path == task.Path && item.Operate == "rename" {
						status.FileList[i].Status = "completed"
					}
				}
			})

		l.Logger.Infof("重命名任务执行成功 - 源路径: %s, 目标路径: %s", task.Path, task.TargetPath)
	}
	return nil
}

// renameFileInVectorDB 在向量数据库中重命名文件
func (l *TaskLogic) renameFileInVectorDB(ctx context.Context, clientId string, codebasePath string, sourcePath, targetPath string) error {
	l.Logger.Infof("开始执行向量数据库中的文件重命名 - 源路径: %s, 目标路径: %s", sourcePath, targetPath)

	// 将文件路径转换为Linux格式（正斜杠）
	sourceLinuxPath := strings.ReplaceAll(sourcePath, "\\", "/")
	targetLinuxPath := strings.ReplaceAll(targetPath, "\\", "/")

	l.svcCtx.VectorStore.UpdateCodeChunksDictionary(ctx, clientId, codebasePath, sourceLinuxPath, targetLinuxPath)

	return nil
}

// getTaskRequestId 获取当前任务的请求ID
func (l *TaskLogic) getTaskRequestId(ctx context.Context) string {
	if requestId := ctx.Value("requestId"); requestId != nil {
		if id, ok := requestId.(string); ok {
			return id
		}
	}
	return ""
}
