package vector

import (
	"github.com/weaviate/weaviate/entities/models"
	"github.com/weaviate/weaviate/entities/schema"
	"github.com/zgsm-ai/codebase-indexer/pkg/utils"
)

const (
	MetadataCodebaseId   = "codebase_id"
	MetadataCodebaseName = "codebase_name"
	MetadataSyncId       = "sync_id"
	MetadataCodebasePath = "codebase_path"
	MetadataFilePath     = "file_path"
	MetadataLanguage     = "language"
	MetadataRange        = "range"
	MetadataTokenCount   = "token_count"
	Content              = "content"
)

const (
	schemeHttp  = "http"
	schemeHttps = "https"
	Verbose     = "verbose"
	Normal      = "normal"
)

var classProperties = []*models.Property{
	{
		Name:            MetadataFilePath,
		DataType:        schema.DataTypeText.PropString(),
		IndexFilterable: utils.BoolPtr(true),
	},
	{
		Name:     MetadataLanguage,
		DataType: schema.DataTypeText.PropString(),
	},
	{
		Name:            MetadataCodebaseId,
		DataType:        schema.DataTypeInt.PropString(),
		IndexFilterable: utils.BoolPtr(true),
	},
	{
		Name:            MetadataCodebasePath,
		DataType:        schema.DataTypeText.PropString(),
		IndexFilterable: utils.BoolPtr(true),
	},
	{
		Name:     MetadataCodebaseName,
		DataType: schema.DataTypeText.PropString(),
	},
	{
		Name:     MetadataSyncId,
		DataType: schema.DataTypeInt.PropString(),
	},
	{
		Name:     MetadataTokenCount,
		DataType: schema.DataTypeInt.PropString(),
	},
	{
		Name:     MetadataRange,
		DataType: schema.DataTypeIntArray.PropString(),
	},
	{
		Name:            Content,
		DataType:        schema.DataTypeText.PropString(),
		IndexSearchable: utils.BoolPtr(true),
	},
}
