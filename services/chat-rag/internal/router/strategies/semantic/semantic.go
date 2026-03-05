package semantic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	ruleengine "github.com/monkeyDluffy6017/ai-llm-rule-engine/pkg/ruleengine"
	"github.com/tidwall/gjson"
	"github.com/zgsm-ai/chat-rag/internal/bootstrap"
	"github.com/zgsm-ai/chat-rag/internal/client"
	"github.com/zgsm-ai/chat-rag/internal/config"
	"github.com/zgsm-ai/chat-rag/internal/logger"
	"github.com/zgsm-ai/chat-rag/internal/types"
	"github.com/zgsm-ai/chat-rag/internal/utils"
	"go.uber.org/zap"
)

type Strategy struct {
	cfg config.SemanticConfig
}

func New(cfg config.SemanticConfig) *Strategy {
	return &Strategy{cfg: cfg}
}

func (s *Strategy) Name() string { return "semantic" }

func (s *Strategy) Run(
	ctx context.Context,
	svcCtx *bootstrap.ServiceContext,
	headers *http.Header,
	req *types.ChatCompletionRequest,
) (string, string, []string, error) {
	if req == nil || len(req.Messages) == 0 {
		return "", "", nil, nil
	}

	// Only trigger when request model is Auto
	if !strings.EqualFold(req.Model, "auto") {
		return "", "", nil, nil
	}

	start := time.Now()

	// 1) Extract inputs first (will also apply head-only truncation inside)
	current, history := s.extractInputs(req)

	// 2) Rule engine prefilter (align with plugin: rule first, then analyzer)
	cands := filterEnabled(s.cfg.Routing.Candidates)
	if s.cfg.RuleEngine.Enabled && len(s.cfg.RuleEngine.InlineRules) > 0 {
		filtered, forcedFallback := s.applyRuleEngine(ctx, svcCtx, headers, req, cands)
		if forcedFallback {
			return s.selectFallback(req), current, s.orderCandidatesByLabel("", req.Model, cands), nil
		}
		if len(filtered) > 0 {
			cands = filtered
		}
	}

	// 3) Build prompt for analyzer
	prompt := s.buildPrompt(current, history)
	logger.InfoC(ctx, "semantic router: analyzer prompt",
		zap.String("prompt", prompt),
	)

	// Analyzer timeout and total timeout with retry
	perTimeout := time.Duration(s.cfg.Analyzer.TimeoutMs) * time.Millisecond
	if perTimeout <= 0 {
		perTimeout = 3 * time.Second
	}
	totalWindow := time.Duration(s.cfg.Analyzer.TotalTimeoutMs) * time.Millisecond
	if totalWindow <= 0 {
		totalWindow = 10 * time.Second
	}
	deadline := time.Now().Add(totalWindow)

	// Build analyzer-specific LLM client (non-streaming) with optional endpoint/token overrides
	llmCfg := svcCtx.Config.LLM
	if s.cfg.Analyzer.Endpoint != "" {
		llmCfg.Endpoint = s.cfg.Analyzer.Endpoint
	}
	analyzerHeaders := headers
	if s.cfg.Analyzer.ApiToken != "" {
		cloned := headers.Clone()
		token := s.cfg.Analyzer.ApiToken
		// Ensure Bearer prefix
		if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = "Bearer " + token
		}
		cloned.Set("Authorization", token)
		cloned.Set(types.HeaderAuthorization, token)
		analyzerHeaders = &cloned
	}

	// Use default timeout config for analyzer
	timeoutCfg := config.LLMTimeoutConfig{
		IdleTimeoutMs:      30000,
		TotalIdleTimeoutMs: 30000,
	}
	llmClient, err := client.NewLLMClient(llmCfg, timeoutCfg, s.cfg.Analyzer.Model, analyzerHeaders)
	if err != nil {
		logger.WarnC(ctx, "semantic router: fallback used",
			zap.String("reason", "analyzer_client_error"),
			zap.Error(err),
			zap.String("selected_model", s.selectFallback(req)),
		)
		return s.selectFallback(req), current, s.orderCandidatesByLabel("", req.Model, cands), nil
	}

	retries := 0
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			logger.WarnC(ctx, "semantic router: fallback used",
				zap.String("reason", "total_timeout"),
				zap.String("selected_model", s.selectFallback(req)),
			)
			return s.selectFallback(req), current, s.orderCandidatesByLabel("", req.Model, cands), nil
		}

		// Use context.WithTimeout for hard timeout limit instead of idle timeout
		requestTimeout := minDuration(perTimeout, remaining)
		actx, cancel := context.WithTimeout(ctx, requestTimeout)

		logger.InfoC(ctx, "semantic router: analyzer request start",
			zap.String("model", s.cfg.Analyzer.Model),
			zap.Duration("timeout", requestTimeout),
			zap.Int("retry", retries),
		)

		attemptStart := time.Now()
		// Call without idleTimer, use timeout context instead
		r, err := llmClient.ChatLLMWithMessagesRaw(actx, types.LLMRequestParams{
			Messages: []types.Message{{Role: types.RoleUser, Content: prompt}},
		}, nil)

		cancel() // Cancel context immediately after request completes
		if err != nil {
			// retry on timeout/network up to 3 times total
			if isRetryableAnalyzerErr(err) && retries < 3 {
				retries++
				actualDuration := time.Since(attemptStart)
				logger.WarnC(ctx, "semantic router: analyzer retry due to error",
					zap.Error(err),
					zap.Int("retry", retries),
					zap.Duration("attempt_duration", actualDuration),
				)
				continue
			}
			logger.WarnC(ctx, "semantic router: fallback used",
				zap.String("reason", "analyzer_error"),
				zap.Error(err),
				zap.String("selected_model", s.selectFallback(req)),
			)
			return s.selectFallback(req), current, s.orderCandidatesByLabel("", req.Model, cands), nil
		}

		// success response, parse label
		var text string
		if len(r.Choices) > 0 {
			text = utils.GetContentAsString(r.Choices[0].Message.Content)
		}
		label := s.parseLabel(text)
		logger.InfoC(ctx, "semantic router: analyzer response",
			zap.String("label", label),
			zap.Int64("analyzer_latency_ms", time.Since(attemptStart).Milliseconds()),
		)
		if label == "" {
			// retry on empty/unknown label up to 3 times total
			if retries < 3 {
				retries++
				logger.WarnC(ctx, "semantic router: empty label from analyzer, retrying", zap.Int("retry", retries))
				continue
			}
			logger.WarnC(ctx, "semantic router: fallback used",
				zap.String("reason", "empty_label"),
				zap.String("selected_model", s.selectFallback(req)),
			)
			return s.selectFallback(req), current, s.orderCandidatesByLabel("", req.Model, cands), nil
		}

		selected := s.selectByLabelFromCandidates(label, req.Model, cands)
		ordered := s.orderCandidatesByLabel(label, req.Model, cands)
		logger.InfoC(ctx, "semantic router: selected model",
			zap.String("label", label),
			zap.String("selected_model", selected),
			zap.Int("candidates", len(cands)),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
		return selected, current, ordered, nil
	}
}

// orderCandidatesByLabel returns candidate model names sorted by score(desc), tieBreakOrder, then name asc.
// If label is empty, order by tieBreakOrder then name asc.
// Always appends fallback model (if configured) at the end when not already included.
func (s *Strategy) orderCandidatesByLabel(label string, orig string, candidates []config.RoutingCandidate) []string {
	type scored struct {
		name  string
		score int
	}
	// default order: by tieBreakOrder then name
	if label == "" {
		names := make([]string, 0, len(candidates))
		for _, c := range candidates {
			names = append(names, c.ModelName)
		}
		if len(s.cfg.Routing.TieBreakOrder) > 0 {
			order := indexMap(s.cfg.Routing.TieBreakOrder)
			sort.SliceStable(names, func(i, j int) bool { return order[names[i]] < order[names[j]] })
		} else {
			sort.Strings(names)
		}
		// ensure fallback appended if not exists
		out := dedup(names)
		if s.cfg.Routing.FallbackModelName != "" && !contains(out, s.cfg.Routing.FallbackModelName) {
			out = append(out, s.cfg.Routing.FallbackModelName)
		}
		return out
	}

	arr := make([]scored, 0, len(candidates))
	for _, c := range candidates {
		arr = append(arr, scored{name: c.ModelName, score: c.Scores[label]})
	}
	// sort by score desc
	sort.SliceStable(arr, func(i, j int) bool { return arr[i].score > arr[j].score })
	// group by score for tie-break
	// apply tieBreakOrder inside same score bucket
	if len(s.cfg.Routing.TieBreakOrder) > 0 {
		order := indexMap(s.cfg.Routing.TieBreakOrder)
		i := 0
		for i < len(arr) {
			j := i + 1
			for j < len(arr) && arr[j].score == arr[i].score {
				j++
			}
			sort.SliceStable(arr[i:j], func(a, b int) bool {
				ai := i + a
				bi := i + b
				return order[arr[ai].name] < order[arr[bi].name]
			})
			i = j
		}
	}
	// produce list
	out := make([]string, 0, len(arr))
	for _, s2 := range arr {
		out = append(out, s2.name)
	}
	out = dedup(out)
	// append fallback if not present
	if s.cfg.Routing.FallbackModelName != "" && !contains(out, s.cfg.Routing.FallbackModelName) {
		out = append(out, s.cfg.Routing.FallbackModelName)
	}
	return out
}

func dedup(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func contains(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

func (s *Strategy) extractInputs(req *types.ChatCompletionRequest) (current string, history string) {
	// Follow plugin logic: extractHistoryAndCurrent for protocol==openai
	if !strings.EqualFold(s.cfg.InputExtraction.Protocol, "openai") {
		// Follow plugin fallback behavior for non-openai: take latest user message,
		// strip code fences if configured, then apply head-only MaxInputBytes if set.
		lastUserIdx := -1
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == types.RoleUser {
				lastUserIdx = i
				break
			}
		}
		if lastUserIdx >= 0 {
			current = utils.GetContentAsString(req.Messages[lastUserIdx].Content)
		}
		if s.cfg.InputExtraction.StripCodeFences {
			current = stripCodeFences(current, s.cfg.InputExtraction.CodeFenceRegex)
		}
		if s.cfg.Analyzer.MaxInputBytes > 0 && len([]byte(current)) > s.cfg.Analyzer.MaxInputBytes {
			bs := []byte(current)
			if s.cfg.Analyzer.MaxInputBytes < len(bs) {
				bs = bs[:s.cfg.Analyzer.MaxInputBytes]
			}
			current = string(bs)
		}
		return current, ""
	}

	// explicit tags per plugin (support <task>, <user_message>, <answer>, <feedback> with @-prefix trim)
	extractExplicit := func(s string) (string, bool) {
		if re := regexp.MustCompile(`(?s)<task>\n?\s*(.*?)\s*\n?</task>`); re != nil {
			if m := re.FindStringSubmatch(s); len(m) >= 2 {
				return strings.TrimSpace(m[1]), true
			}
		}
		if re := regexp.MustCompile(`(?s)<user_message>\n?\s*(.*?)\s*\n?</user_message>`); re != nil {
			if m := re.FindStringSubmatch(s); len(m) >= 2 {
				return strings.TrimSpace(m[1]), true
			}
		}
		if re := regexp.MustCompile(`(?s)<answer>\n?\s*(.*?)\s*\n?</answer>`); re != nil {
			if m := re.FindStringSubmatch(s); len(m) >= 2 {
				return strings.TrimSpace(m[1]), true
			}
		}
		if re := regexp.MustCompile(`(?s)<feedback>\n?\s*(.*?)\s*\n?</feedback>`); re != nil {
			if m := re.FindStringSubmatch(s); len(m) >= 2 {
				fs := strings.TrimSpace(m[1])
				if re2 := regexp.MustCompile(`^@/?\s*`); re2 != nil {
					fs = re2.ReplaceAllString(fs, "")
				}
				return strings.TrimSpace(fs), true
			}
		}
		return "", false
	}

	// helper to get raw content
	getRaw := func(c any) string { return utils.GetContentAsString(c) }

	msgs := req.Messages
	if len(msgs) == 0 {
		return "", ""
	}

	// unified scan from newest to oldest
	// Optimization: disable default assistant inclusion; record at most ONE candidate assistant for disambiguation use
	found := false
	histParts := make([]string, 0)
	firstAssistantRaw := ""
	for i := len(msgs) - 1; i >= 0; i-- {
		role := msgs[i].Role
		raw := strings.TrimSpace(getRaw(msgs[i].Content))
		if raw == "" {
			continue
		}
		if !found {
			if role == types.RoleAssistant {
				current = raw
				found = true
				continue
			}
			if role == types.RoleUser {
				if v, ok := extractExplicit(raw); ok && strings.TrimSpace(v) != "" {
					current = strings.TrimSpace(v)
					found = true
					continue
				}
			}
			continue
		}
		// after current found, collect older tagged user into history
		// assistant messages are NOT included by default; only record the first candidate for possible disambiguation later
		if role == types.RoleAssistant {
			if firstAssistantRaw == "" {
				firstAssistantRaw = raw
			}
			continue
		}
		if role == types.RoleUser {
			if v, ok := extractExplicit(raw); ok && strings.TrimSpace(v) != "" {
				histParts = append(histParts, v)
			}
		}
	}
	if !found {
		return "", ""
	}

	// reverse to older -> newer
	for i, j := 0, len(histParts)-1; i < j; i, j = i+1, j-1 {
		histParts[i], histParts[j] = histParts[j], histParts[i]
	}
	sep := s.cfg.InputExtraction.UserJoinSep
	if sep == "" {
		sep = "\n\n"
	}

	// Decide whether to include at most one recent assistant for disambiguation
	isDisambig := func(s2 string) bool {
		st := strings.TrimSpace(s2)
		if st == "" {
			return false
		}
		// short current input
		if len([]rune(st)) <= 8 {
			return true
		}
		// keyword triggers
		re := regexp.MustCompile(`(?i)^(继续|同上|重试|再来一次|go on|continue|same as above|retry)\b`)
		return re.MatchString(st)
	}

	if isDisambig(current) && strings.TrimSpace(firstAssistantRaw) != "" {
		// append as the newest historical item
		histParts = append(histParts, firstAssistantRaw)
	}

	// Enforce count limit: keep only the most recent N history items (older->newer order means keep tail)
	if s.cfg.InputExtraction.MaxHistoryMessages > 0 && len(histParts) > s.cfg.InputExtraction.MaxHistoryMessages {
		histParts = histParts[len(histParts)-s.cfg.InputExtraction.MaxHistoryMessages:]
	}

	history = strings.Join(histParts, sep)

	// Strictly clean tool/noise blocks from assistant/user history before code fence stripping
	cleanHistoryNoise := func(s2 string) string {
		patterns := []string{
			`(?s)<think>.*?</think>`,
			`(?s)<attempt_completion>.*?</attempt_completion>`,
			`(?s)<environment_details>.*?</environment_details>`,
			`(?m)^\[attempt_completion\].*$\n?`,
		}
		out := s2
		for _, p := range patterns {
			re := regexp.MustCompile(p)
			out = re.ReplaceAllString(out, "")
		}
		return out
	}

	history = cleanHistoryNoise(history)

	// strip code fences if configured, using plugin default pattern when empty
	if s.cfg.InputExtraction.StripCodeFences {
		pattern := s.cfg.InputExtraction.CodeFenceRegex
		if pattern == "" {
			pattern = "(?s)```{3,4}[\\s\\S]*?```{3,4}"
		}
		re := regexp.MustCompile(pattern)
		history = re.ReplaceAllString(history, "")
		current = re.ReplaceAllString(current, "")
	}

	// apply byte limits: current by Analyzer.MaxInputBytes; history by InputExtraction.MaxHistoryBytes
	if s.cfg.Analyzer.MaxInputBytes > 0 && len([]byte(current)) > s.cfg.Analyzer.MaxInputBytes {
		bs := []byte(current)
		if s.cfg.Analyzer.MaxInputBytes < len(bs) {
			bs = bs[:s.cfg.Analyzer.MaxInputBytes]
		}
		current = string(bs)
	}
	if s.cfg.InputExtraction.MaxHistoryBytes > 0 && len([]byte(history)) > s.cfg.InputExtraction.MaxHistoryBytes {
		bs := []byte(history)
		if s.cfg.InputExtraction.MaxHistoryBytes < len(bs) {
			bs = bs[:s.cfg.InputExtraction.MaxHistoryBytes]
		}
		history = string(bs)
	}

	return current, history
}

func (s *Strategy) buildPrompt(current, history string) string {
	if s.cfg.Analyzer.PromptTemplate != "" {
		return strings.ReplaceAll(strings.ReplaceAll(s.cfg.Analyzer.PromptTemplate, "{HISTORY}", history), "{CURRENT}", current)
	}
	// default prompt from design doc
	return fmt.Sprintf(
		"You are a classification specialist. Classify ONLY based on the CURRENT turn. Use history strictly for disambiguation of short messages (e.g., \"retry\", \"continue\", \"same as above\").\n\nLabels and definitions:\n1) simple_request: Non-technical, conversational queries. Includes greetings (e.g., \"hello\"), identity questions (e.g., \"who are you?\"), or general chat not involving programming/code/dev tasks.\n2) planning_request: Requests for analysis/planning/explanation about code or a task without directly writing or editing code. Examples: review code and give feedback, create a technical plan/outline, discuss architecture, explain an algorithm.\n3) code_modification: Requests that require generating, editing, or modifying code. Examples: implement a function, fix a bug, add a new feature, refactor, translate comments, or convert code between languages.\n\nSpecial rules:\n- Classify ONLY using the Current section below. Do NOT summarize or rewrite anything.\n- History may be referenced only to interpret very short Current inputs.\n- If Current contains file paths, line ranges (e.g., foo.go:12-20), diffs, or code blocks indicating an edit intent, prefer code_modification unless clearly chit-chat.\n- If the Current contains imperative phrases indicating immediate implementation (e.g., \"实施\", \"实现\", \"开始实现\", \"开始编码\", \"按计划实施\", \"落地\", \"修改\", \"修复\", \"apply the plan\", \"go ahead and implement\", \"implement now\"), classify as code_modification.\n- If the Current requests or instructs USING TOOLS (e.g., calling/using tools, running step-by-step tool executions, specifying tool tags/steps), classify as code_modification.\n\nHistory:\n%s\n\nCurrent:\n%s\n\nInstructions:\n- Output exactly one of: simple_request, planning_request, code_modification\n- Output the label only. No extra words.", history, current)
}

// buildRequestContextFromRules builds request_context by scanning rule facts with prefixes
func (s *Strategy) buildRequestContextFromRules(
	ctx context.Context,
	headers *http.Header,
	req *types.ChatCompletionRequest,
	rules []ruleengine.Rule,
) ruleengine.RequestContext {
	rc := ruleengine.RequestContext{}
	// collect request.* facts
	facts := collectRequestFactsFromRules(rules)
	// marshal request body for gjson extraction
	bodyBytes, _ := json.Marshal(req)

	// Align prefix defaults with plugin logic
	bodyPrefix := s.cfg.RuleEngine.BodyPrefix
	if strings.TrimSpace(bodyPrefix) == "" {
		bodyPrefix = "body."
	}
	headerPrefix := s.cfg.RuleEngine.HeaderPrefix
	if strings.TrimSpace(headerPrefix) == "" {
		headerPrefix = "header."
	}

	for _, f := range facts {
		if !strings.HasPrefix(f, "request.") {
			continue
		}
		sub := strings.TrimPrefix(f, "request.")
		switch {
		case strings.HasPrefix(sub, bodyPrefix):
			path := strings.TrimPrefix(sub, bodyPrefix)
			if path == "" {
				continue
			}
			v := gjson.GetBytes(bodyBytes, path)
			if v.Exists() {
				val := gjsonToInterface(v)
				keys := append([]string{"body"}, strings.Split(path, ".")...)
				setNested(rc, keys, val)
			}
		case strings.HasPrefix(sub, headerPrefix):
			h := strings.TrimPrefix(sub, headerPrefix)
			if h == "" || headers == nil {
				continue
			}
			if hv := headers.Get(h); hv != "" {
				setNested(rc, []string{"header", h}, hv)
			}
		default:
			// leave empty
		}
	}
	return rc
}

func collectRequestFactsFromRules(rules []ruleengine.Rule) []string {
	set := map[string]struct{}{}
	add := func(s string) {
		if strings.HasPrefix(s, "request.") {
			set[s] = struct{}{}
		}
	}
	for _, r := range rules {
		for _, c := range r.Conditions.All {
			add(c.Fact)
		}
		for _, c := range r.Conditions.Any {
			add(c.Fact)
		}
		if r.Conditions.Not != nil {
			add(r.Conditions.Not.Fact)
		}
		for _, sk := range r.Action.SortBy {
			if strings.HasPrefix(sk.Fact, "request.") {
				add(sk.Fact)
			}
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

// buildAvailableModelsFromBodyOrCandidates parses routing.available_models from request body; falls back to candidates
func (s *Strategy) buildAvailableModelsFromBodyOrCandidates(_ *types.ChatCompletionRequest, cands []config.RoutingCandidate) []ruleengine.Model {
	// Per requirement, available_models are defined in config; ignore request body
	models := make([]ruleengine.Model, 0, len(cands))
	for _, c := range cands {
		models = append(models, ruleengine.Model{
			"model_name": c.ModelName,
		})
	}
	return models
}

// collectDynamicMetricsFromRules scans rules for model.dy_* facts
func (s *Strategy) collectDynamicMetricsFromRules(rules []ruleengine.Rule) []string {
	set := map[string]struct{}{}
	add := func(f string) {
		if strings.HasPrefix(f, "model.dy_") {
			set[f[len("model."):]] = struct{}{}
		}
	}
	for _, r := range rules {
		for _, c := range r.Conditions.All {
			add(c.Fact)
		}
		for _, c := range r.Conditions.Any {
			add(c.Fact)
		}
		if r.Conditions.Not != nil {
			add(r.Conditions.Not.Fact)
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

// gjsonToInterface converts gjson.Result to native Go types
func gjsonToInterface(v gjson.Result) any {
	switch v.Type {
	case gjson.False, gjson.True:
		return v.Bool()
	case gjson.Number:
		// prefer float64
		if f, err := strconv.ParseFloat(v.Raw, 64); err == nil {
			return f
		}
		return v.Value()
	case gjson.String:
		return v.String()
	case gjson.JSON:
		if v.IsArray() {
			arr := make([]any, 0, len(v.Array()))
			for _, it := range v.Array() {
				arr = append(arr, gjsonToInterface(it))
			}
			return arr
		}
		m := map[string]any{}
		v.ForEach(func(k, val gjson.Result) bool {
			m[k.String()] = gjsonToInterface(val)
			return true
		})
		return m
	default:
		return v.Value()
	}
}

func setNested(rc ruleengine.RequestContext, keys []string, value any) {
	if len(keys) == 0 {
		return
	}
	cur := map[string]any(rc)
	for i := 0; i < len(keys)-1; i++ {
		k := keys[i]
		next, ok := cur[k]
		if !ok {
			child := map[string]any{}
			cur[k] = child
			cur = child
			continue
		}
		if m, ok := next.(map[string]any); ok {
			cur = m
		} else {
			child := map[string]any{}
			cur[k] = child
			cur = child
		}
	}
	cur[keys[len(keys)-1]] = value
}

func pickMetricValue(h map[string]string, metric string) any {
	// preference: value, v, current, metric name itself, first entry
	if v, ok := h["value"]; ok {
		if f, ok2 := tryParseFloat(v); ok2 {
			return f
		}
		return v
	}
	if v, ok := h["v"]; ok {
		if f, ok2 := tryParseFloat(v); ok2 {
			return f
		}
		return v
	}
	if v, ok := h["current"]; ok {
		if f, ok2 := tryParseFloat(v); ok2 {
			return f
		}
		return v
	}
	if v, ok := h[metric]; ok {
		if f, ok2 := tryParseFloat(v); ok2 {
			return f
		}
		return v
	}
	for _, v := range h {
		if f, ok2 := tryParseFloat(v); ok2 {
			return f
		}
		return v
	}
	return nil
}

func tryParseFloat(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, true
	}
	return 0, false
}

func (s *Strategy) parseLabel(text string) string {
	labels := []string{"simple_request", "planning_request", "code_modification"}
	if len(s.cfg.Analyzer.AnalysisLabels) > 0 {
		labels = s.cfg.Analyzer.AnalysisLabels
	}
	// Build word-boundary regex like plugin
	var b strings.Builder
	b.WriteString(`\b(`)
	for i, sv := range labels {
		if i > 0 {
			b.WriteString(`|`)
		}
		b.WriteString(regexp.QuoteMeta(sv))
	}
	b.WriteString(`)\b`)
	re := regexp.MustCompile(b.String())
	if m := re.FindString(text); m != "" {
		return m
	}
	// fallback: substring scan
	for _, s := range labels {
		if s != "" && strings.Contains(text, s) {
			return s
		}
	}
	return ""
}

func (s *Strategy) selectByLabelFromCandidates(label string, orig string, candidates []config.RoutingCandidate) string {
	if len(candidates) == 0 {
		return s.selectFallbackWithOrig(orig)
	}

	// already filtered above

	// scoring
	type scored struct {
		name  string
		score int
	}
	var arr []scored
	for _, c := range candidates {
		score := c.Scores[label]
		arr = append(arr, scored{name: c.ModelName, score: score})
	}
	// find max score
	max := -1
	for _, s := range arr {
		if s.score > max {
			max = s.score
		}
	}
	if max < 0 || max < s.cfg.Routing.MinScore {
		return s.selectFallbackWithOrig(orig)
	}
	// tie-break
	var winners []string
	for _, s := range arr {
		if s.score == max {
			winners = append(winners, s.name)
		}
	}
	if len(winners) == 1 {
		return winners[0]
	}
	// order by tieBreakOrder
	if len(s.cfg.Routing.TieBreakOrder) > 0 {
		order := indexMap(s.cfg.Routing.TieBreakOrder)
		sort.SliceStable(winners, func(i, j int) bool { return order[winners[i]] < order[winners[j]] })
		return winners[0]
	}
	sort.Strings(winners)
	return winners[0]
}

func (s *Strategy) selectFallback(req *types.ChatCompletionRequest) string {
	if s.cfg.Routing.FallbackModelName != "" {
		return s.cfg.Routing.FallbackModelName
	}
	return req.Model
}

func (s *Strategy) selectFallbackWithOrig(orig string) string {
	if s.cfg.Routing.FallbackModelName != "" {
		return s.cfg.Routing.FallbackModelName
	}
	return orig
}

func filterEnabled(cands []config.RoutingCandidate) []config.RoutingCandidate {
	var out []config.RoutingCandidate
	for _, c := range cands {
		if c.Enabled {
			out = append(out, c)
		}
	}
	return out
}

func indexMap(arr []string) map[string]int {
	m := make(map[string]int, len(arr))
	for i, v := range arr {
		m[v] = i
	}
	return m
}

func stripCodeFences(s string, custom string) string {
	if s == "" {
		return s
	}
	if custom != "" {
		re := regexp.MustCompile(custom)
		return re.ReplaceAllString(s, "")
	}
	// default: remove ``` blocks (align with plugin)
	re := regexp.MustCompile("(?s)```{3,4}[\\s\\S]*?```{3,4}")
	return re.ReplaceAllString(s, "")
}

// minDuration returns the smaller of two durations
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// isRetryableAnalyzerErr determines if analyzer error is retryable
func isRetryableAnalyzerErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline") || strings.Contains(msg, "connection") || strings.Contains(msg, "network")
}

// applyRuleEngine integrates the original rule engine style prefilter (simplified adapter)
func (s *Strategy) applyRuleEngine(
	ctx context.Context,
	svcCtx *bootstrap.ServiceContext,
	headers *http.Header,
	req *types.ChatCompletionRequest,
	cands []config.RoutingCandidate,
) ([]config.RoutingCandidate, bool) {
	if !s.cfg.RuleEngine.Enabled || len(s.cfg.RuleEngine.InlineRules) == 0 {
		return cands, false
	}

	logger.InfoC(ctx, "semantic router: rule engine enabled",
		zap.Int("rules", len(s.cfg.RuleEngine.InlineRules)),
	)

	// 1) parse inline rules (JSON strings) into ruleengine.Rule
	rules := make([]ruleengine.Rule, 0, len(s.cfg.RuleEngine.InlineRules))
	for _, raw := range s.cfg.RuleEngine.InlineRules {
		var r ruleengine.Rule
		if err := json.Unmarshal([]byte(raw), &r); err == nil {
			rules = append(rules, r)
		} else {
			logger.WarnC(ctx, "invalid inline rule, skipped", zap.Error(err))
		}
	}
	if len(rules) == 0 {
		return cands, false
	}

	// 2) build request_context from rules facts (dynamic based on facts)
	reqCtx := s.buildRequestContextFromRules(ctx, headers, req, rules)

	// 3) build available_models: prefer from request body `routing.available_models`, else from candidates
	models := s.buildAvailableModelsFromBodyOrCandidates(req, cands)

	// 4) load dynamic metrics into models if configured
	dm := s.cfg.Analyzer.DynamicMetrics
	if dm.Enabled && dm.RedisPrefix != "" && svcCtx.RedisClient != nil {
		// Scan rules to discover dynamic metrics facts used: model.dy_*
		dynMetrics := s.collectDynamicMetricsFromRules(rules)
		for i := range models {
			modelName, _ := models[i]["model_name"].(string)
			if modelName == "" {
				continue
			}
			for _, metric := range dynMetrics {
				key := fmt.Sprintf("%s:%s:%s", dm.RedisPrefix, metric, modelName)
				if valStr, err := svcCtx.RedisClient.GetString(ctx, key); err == nil && valStr != "" {
					name := metric
					// store as float when possible, else string
					if f, ok := tryParseFloat(valStr); ok {
						models[i]["dy_"+strings.TrimPrefix(name, "dy_")] = f
					} else {
						models[i]["dy_"+strings.TrimPrefix(name, "dy_")] = valStr
					}
				}
			}
		}
	}

	// 5) evaluate
	res, err := ruleengine.New().Evaluate(reqCtx, models, rules)
	if err != nil || len(res.QualifiedModels) == 0 {
		if err != nil {
			logger.WarnC(ctx, "rule engine evaluate error", zap.Error(err))
		}
		// Align with original plugin: if fallback configured, short-circuit analyzer
		if s.cfg.Routing.FallbackModelName != "" {
			return nil, true
		}
		// 无 fallback 时继续沿用原始候选，交由后续流程决定
		return cands, false
	}

	// 6) filter candidates by qualified models order
	order := map[string]int{}
	qnames := make([]string, 0, len(res.QualifiedModels))
	for i, m := range res.QualifiedModels {
		if name, ok := m["model_name"].(string); ok {
			order[name] = i
			qnames = append(qnames, name)
		}
	}
	var filtered []config.RoutingCandidate
	for _, c := range cands {
		if _, ok := order[c.ModelName]; ok {
			filtered = append(filtered, c)
		}
	}
	if len(filtered) == 0 {
		return cands, false
	}
	sort.SliceStable(filtered, func(i, j int) bool { return order[filtered[i].ModelName] < order[filtered[j].ModelName] })
	logger.InfoC(ctx, "rule engine filtered candidates", zap.Int("count", len(filtered)))
	return filtered, false
}

// applyDynamicMetrics filters candidates with Redis-backed dynamic metrics
// Note: Dynamic metrics behavior aligns with plugin logic via rule engine evaluation path only.
