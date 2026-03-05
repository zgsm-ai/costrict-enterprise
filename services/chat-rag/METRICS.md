# Prometheus Metrics Integration

This project integrates Prometheus metrics functionality for monitoring various aspects of chat completion requests.

## Features

### 1. Automatic Metrics Collection

- Automatically collects metrics data during log processing
- Reports metrics after `uploadToLoki`
- Based on data from `model.ChatLog`

### 2. Supported Metrics

#### Request Metrics

- `chat_rag_requests_total`: Total number of chat completion requests
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`, `category`

#### Token Metrics

- `chat_rag_original_tokens_total`: Total number of original tokens processed
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`, `token_scope` (system/user/all)
- `chat_rag_compressed_tokens_total`: Total number of compressed tokens processed
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`, `token_scope` (system/user/all)

#### Compression Metrics

- `chat_rag_compression_ratio`: Distribution of compression ratios (buckets: 0.1, 0.2, ..., 1.0)
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`
- `chat_rag_user_prompt_compressed_total`: Total number of requests where user prompt was compressed
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`

#### Latency Metrics

- `chat_rag_semantic_latency_ms`: Semantic processing latency in milliseconds (buckets: 10, 50, 100, 200, 500, 1000, 2000, 5000)
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`
- `chat_rag_summary_latency_ms`: Summary processing latency in milliseconds (buckets: 10, 50, 100, 200, 500, 1000, 2000, 5000)
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`
- `chat_rag_main_model_latency_ms`: Main model processing latency in milliseconds (buckets: 100, 500, 1000, 2000, 5000, 10000, 20000)
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`
- `chat_rag_total_latency_ms`: Total processing latency in milliseconds (buckets: 100, 500, 1000, 2000, 5000, 10000, 20000, 30000)
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`

#### Response Metrics

- `chat_rag_response_tokens_total`: Total number of response tokens generated
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`

#### Error Metrics

- `chat_rag_errors_total`: Total number of errors encountered
  - Labels: `client_id`, `client_ide`, `model`, `user`, `login_from`, `error_type` (from log.Error field)

## Usage

### 1. Accessing Metrics Endpoint

After starting the service, Prometheus metrics can be accessed via:

```
GET http://localhost:8080/metrics
```

### 2. Prometheus Configuration

Add the following job to your Prometheus configuration file:

```yaml
scrape_configs:
  - job_name: "chat-rag"
    static_configs:
      - targets: ["localhost:8080"]
    metrics_path: "/metrics"
    scrape_interval: 15s
```

### 3. Example Queries

#### Total Requests

```promql
chat_rag_requests_total
```

#### Compression Ratio Distribution

```promql
histogram_quantile(0.95, chat_rag_compression_ratio_bucket)
```

#### Average Latency

```promql
rate(chat_rag_total_latency_ms_sum[5m]) / rate(chat_rag_total_latency_ms_count[5m])
```

#### Requests by Client

```promql
sum(rate(chat_rag_requests_total[5m])) by (client_id)
```

## Architecture

### Components

1. **MetricsService**: Defines and records Prometheus metrics
2. **LoggerService**: Integrates with MetricsService, automatically reporting metrics during log processing
3. **MetricsHandler**: Provides the `/metrics` HTTP endpoint

### Integration Flow

1. Initialize `MetricsService` in `ServiceContext`
2. Inject `MetricsService` into `LoggerService`
3. Call `metricsService.RecordChatLog()` in `LoggerService.processLogs()` after successful `uploadToLoki`
4. Expose metrics to Prometheus via `/metrics` endpoint

## Considerations

1. **Performance Impact**: Metrics collection has minimal performance impact, but monitor memory usage in high-concurrency scenarios
2. **Label Cardinality**: Avoid high-cardinality labels (e.g., request_id) to prevent memory leaks
3. **Data Retention**: Prometheus defaults to 15-day retention (configurable)
4. **Security**: Implement access control for `/metrics` endpoint in production

## Troubleshooting

### Common Issues

1. **Metrics Not Updating**

   - Verify LoggerService is running
   - Check if log files are being processed correctly
   - Confirm successful Loki upload

2. **High Memory Usage**

   - Check for high label cardinality
   - Consider reducing number of histogram buckets

3. **Prometheus Scrape Failure**
   - Verify service port
   - Check firewall settings
   - Confirm `/metrics` endpoint is accessible
