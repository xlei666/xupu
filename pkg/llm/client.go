package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/xlei/xupu/pkg/config"
)

// 初始化时加载环境变量
func init() {
	godotenv.Load()
}

// Client LLM客户端
type Client struct {
	APIKey  string
	BaseURL string
	Model   string
	httpCli *http.Client
}

// Message 聊天消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// SetModel 设置模型
func (c *Client) SetModel(model string) {
	c.Model = model
}

// SetTemperature 设置温度参数
func (c *Client) SetTemperature(temp float64) {
	// 存储在客户端中，用于后续请求
	// 注意：这需要在Generate时传递，暂时先不实现
}

// SetMaxTokens 设置最大token数
func (c *Client) SetMaxTokens(maxTokens int) {
	// 存储在客户端中，用于后续请求
	// 注意：这需要在Generate时传递，暂时先不实现
}

// NewClientWithConfig 使用配置创建LLM客户端
func NewClientWithConfig(providerName string) (*Client, error) {
	cfg := config.Get()
	provider, ok := cfg.LLM.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("未找到提供商 %s 的配置", providerName)
	}

	apiKey, err := provider.GetAPIKey()
	if err != nil {
		return nil, err
	}

	return &Client{
		APIKey:  apiKey,
		BaseURL: provider.BaseURL,
		Model:   provider.Models.Default,
		httpCli: &http.Client{Timeout: getTimeout()},
	}, nil
}

// NewClientForModule 为特定模块创建LLM客户端
// 自动从配置中获取该模块对应的模型设置
func NewClientForModule(moduleName string) (*Client, *config.ModuleMapping, error) {
	cfg := config.Get()

	mapping, provider, err := cfg.LLM.GetModuleConfig(moduleName)
	if err != nil {
		return nil, nil, err
	}

	apiKey, err := provider.GetAPIKey()
	if err != nil {
		return nil, nil, err
	}

	client := &Client{
		APIKey:  apiKey,
		BaseURL: provider.BaseURL,
		Model:   mapping.Model,
		httpCli: &http.Client{Timeout: getTimeout()},
	}

	return client, mapping, nil
}

// getTimeout 从配置获取超时时间，默认120秒
func getTimeout() time.Duration {
	cfg := config.Get()
	if cfg.System.Timeout.LLMRequest > 0 {
		return time.Duration(cfg.System.Timeout.LLMRequest) * time.Second
	}
	return 120 * time.Second // 默认120秒
}

// Generate 生成文本
func (c *Client) Generate(prompt string, systemPrompt string) (string, error) {
	return c.GenerateWithParams(prompt, systemPrompt, 0.7, 2000)
}

// GenerateWithParams 使用指定参数生成文本
func (c *Client) GenerateWithParams(prompt string, systemPrompt string, temperature float64, maxTokens int) (string, error) {
	messages := []Message{}
	if systemPrompt != "" {
		messages = append(messages, Message{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, Message{Role: "user", Content: prompt})

	reqBody := ChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	return c.SendRequest(reqBody)
}

// GenerateJSON 生成JSON格式输出
func (c *Client) GenerateJSON(prompt string, systemPrompt string) (map[string]interface{}, error) {
	return c.GenerateJSONWithParams(prompt, systemPrompt, 0.5, 2000)
}

// GenerateJSONWithParams 使用指定参数生成JSON格式输出
func (c *Client) GenerateJSONWithParams(prompt string, systemPrompt string, temperature float64, maxTokens int) (map[string]interface{}, error) {
	// 添加JSON格式要求
	jsonPrompt := prompt + "\n\n请直接以JSON格式返回结果，不要包含任何其他内容。"

	messages := []Message{}
	if systemPrompt != "" {
		messages = append(messages, Message{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, Message{Role: "user", Content: jsonPrompt})

	reqBody := ChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	content, err := c.SendRequest(reqBody)
	if err != nil {
		return nil, err
	}

	// 解析JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(content), &result)
	if err != nil {
		// 尝试提取 ```json``` 中的内容
		content = extractJSON(content)
		err = json.Unmarshal([]byte(content), &result)
		if err != nil {
			return nil, fmt.Errorf("无法解析JSON: %w, 原始内容: %s", err, content[:min(200, len(content))])
		}
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SendRequest 发送请求
func (c *Client) SendRequest(req ChatRequest) (string, error) {
	resp, err := c.sendRequestInternal(req)
	if err != nil {
		return "", err
	}

	var chatResp ChatResponse
	err = json.Unmarshal([]byte(resp), &chatResp)
	if err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("API返回无内容")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// sendRequestInternal 内部请求方法
func (c *Client) sendRequestInternal(req ChatRequest) (string, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.httpCli.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误: %d, %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

// StreamCallback 定义流式回调函数
// 返回 false 停止流式传输
type StreamCallback func(content string) bool

// GenerateStream 流式生成文本
func (c *Client) GenerateStream(prompt string, systemPrompt string, callback StreamCallback) error {
	return c.GenerateStreamWithParams(prompt, systemPrompt, 0.7, 2000, callback)
}

// GenerateStreamWithParams 使用指定参数流式生成文本
func (c *Client) GenerateStreamWithParams(prompt string, systemPrompt string, temperature float64, maxTokens int, callback StreamCallback) error {
	messages := []Message{}
	if systemPrompt != "" {
		messages = append(messages, Message{Role: "system", Content: systemPrompt})
	}
	messages = append(messages, Message{Role: "user", Content: prompt})

	// 为了最小化修改，我们临时构建 map
	reqMap := map[string]interface{}{
		"model":       c.Model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
		"stream":      true,
	}

	return c.sendStreamRequest(reqMap, callback)
}

// sendStreamRequest 发送流式请求
func (c *Client) sendStreamRequest(reqBody interface{}, callback StreamCallback) error {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpCli.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API返回错误: %d, %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		data := bytes.TrimPrefix(line, []byte("data: "))
		if string(data) == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(data, &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				if !callback(content) {
					break
				}
			}
		}
	}

	return nil
}

// extractJSON 从响应中提取JSON
func extractJSON(s string) string {
	// 查找 ```json```
	start := bytes.Index([]byte(s), []byte("```json"))
	if start >= 0 {
		start += 7
		end := bytes.Index([]byte(s[start:]), []byte("```"))
		if end >= 0 {
			return s[start : start+end]
		}
	}

	// 查找 ````
	start = bytes.Index([]byte(s), []byte("```"))
	if start >= 0 {
		start += 3
		end := bytes.Index([]byte(s[start:]), []byte("```"))
		if end >= 0 {
			return s[start : start+end]
		}
	}

	// 查找 { }
	start = bytes.Index([]byte(s), []byte("{"))
	if start >= 0 {
		end := bytes.LastIndex([]byte(s), []byte("}"))
		if end >= 0 {
			return s[start : end+1]
		}
	}

	return s
}
