package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zgsm-ai/codebase-indexer/internal/config"
	"github.com/zgsm-ai/codebase-indexer/internal/types"
)

const (
	rerankQuery     = "query"
	rerankDocuments = "documents"
	rerankScores    = "scores"
)

type Reranker interface {
	Rerank(ctx context.Context, query string, docs []*types.SemanticFileItem) ([]*types.SemanticFileItem, error)
}

type customReranker struct {
	config config.RerankerConf
}

type rerankerRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n"`
}

type rerankerResponse struct {
	Model string `json:"model"`

	Results []struct {
		Index    int `json:"index"`
		Document struct {
			Text string `json:"text"`
		} `json:"document"`
		RelevanceScore float32 `json:"relevance_score"`
	} `json:"results"`
	Usage struct {
		TotalTokens  int `json:"total_tokens"`
		PromptTokens int `json:"prompt_tokens"`
	} `json:"usage"`
}

// request
//
//	{
//	   "model": "gte-reranker-modernbert-base",
//	   "query": "What is the capital of the United States?",
//	   "documents": [
//	       "The Commonwealth of the .",
//	       "Carson City is the capital city ."
//	   ],
//	   "top_n": 10
//	}

// response
// {
//    "model": "gte-reranker-modernbert-base",
//    "usage": {
//        "total_tokens": 38,
//        "prompt_tokens": 38
//    },
//    "results": [
//        {
//            "index": 1,
//            "document": {
//                "text": "Carson City is the capital city ."
//            },
//            "relevance_score": 0.822102963924408
//        },
//        {
//            "index": 0,
//            "document": {
//                "text": "The Commonwealth of the ."
//            },
//            "relevance_score": 0.8087424039840698
//        }
//    ]
//}

func (r *customReranker) Rerank(ctx context.Context, query string, docs []*types.SemanticFileItem) ([]*types.SemanticFileItem, error) {
	if len(docs) == 0 {
		return docs, nil
	}

	contents := make([]string, len(docs))
	for i, doc := range docs {
		contents[i] = doc.Content
	}

	requestBody := &rerankerRequest{
		Model:     r.config.Model,
		Query:     query,
		Documents: contents,
		TopN:      len(docs), // Request all documents to be ranked
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal customReranker request body: %w", err)
	}

	rerankEndpoint := r.config.APIBase

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rerankEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create customReranker request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.config.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send customReranker request to %s: %w", rerankEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody := new(bytes.Buffer)
		_, _ = errorBody.ReadFrom(resp.Body)
		return nil, fmt.Errorf("customReranker API returned non-OK status %d: %s, body: %s", resp.StatusCode, resp.Status, errorBody.String())
	}

	var responseBody rerankerResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode customReranker response body: %w", err)
	}

	if len(responseBody.Results) == 0 {
		return nil, fmt.Errorf("reranker returned empty results")
	}

	// Create a mapping from original index to reranked position
	rerankedDocs := make([]*types.SemanticFileItem, len(docs))
	for i, result := range responseBody.Results {
		if result.Index >= len(docs) {
			return nil, fmt.Errorf("invalid index %d in reranker response", result.Index)
		}
		rerankedDocs[i] = docs[result.Index]
		// Store the relevance score in the document if needed
		rerankedDocs[i].Score = result.RelevanceScore
	}

	return rerankedDocs, nil
}

func NewReranker(c config.RerankerConf) Reranker {
	return &customReranker{
		config: c,
	}
}
