// Package anthropicup — Anthropic 兼容上游的通用适配：统一信封 ↔ messages 协议。
package anthropicup

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	pb "github.com/ShadowSmallBaby/ClawProxyHub/sdk/proto/cphv1"
)

// ChatBody 信封请求 → Anthropic Messages 请求体（stream=true）。
func ChatBody(req *pb.ChatRequest) map[string]interface{} {
	var system string
	var messages []map[string]interface{}
	for _, m := range req.Messages {
		switch {
		case m.Role == "system":
			system += m.Text
		case m.Role == "tool":
			// 工具结果以 tool_result 块包进 user 消息；多模态工具输出保留 content blocks。
			resultContent := interface{}(m.Text)
			if len(m.ContentJson) > 0 {
				resultContent = anthropicContent(m.ContentJson, m.Text)
			}
			messages = append(messages, map[string]interface{}{
				"role": "user",
				"content": []interface{}{map[string]interface{}{
					"type": "tool_result", "tool_use_id": m.ToolCallId, "content": resultContent,
				}},
			})
		case m.Role == "assistant" && len(m.ToolCalls) > 0:
			blocks := anthropicContent(m.ContentJson, m.Text)
			if len(m.ContentJson) == 0 && m.Text == "" {
				blocks = nil
			}
			for _, tc := range m.ToolCalls {
				blocks = append(blocks, map[string]interface{}{
					"type": "tool_use", "id": tc.Id, "name": tc.Name,
					"input": rawJSON(tc.Arguments),
				})
			}
			messages = append(messages, map[string]interface{}{"role": "assistant", "content": blocks})
		default:
			messages = append(messages, map[string]interface{}{
				"role":    m.Role,
				"content": anthropicContent(m.ContentJson, m.Text),
			})
		}
	}
	body := map[string]interface{}{
		"model":      req.Model,
		"messages":   messages,
		"max_tokens": orInt(req.MaxTokens, 8192),
		"stream":     true,
	}
	if system != "" {
		body["system"] = system
	}
	if len(req.Tools) > 0 {
		var tools []interface{}
		for _, t := range req.Tools {
			tools = append(tools, map[string]interface{}{
				"name": t.Name, "description": t.Description,
				"input_schema": rawJSON(orDefault(t.ParametersSchema, `{"type":"object"}`)),
			})
		}
		body["tools"] = tools
		if tc := req.ToolChoice; tc != nil {
			switch tc.Type {
			case "auto":
				body["tool_choice"] = map[string]interface{}{"type": "auto"}
			case "none":
				body["tool_choice"] = map[string]interface{}{"type": "none"}
			case "tool":
				body["tool_choice"] = map[string]interface{}{"type": "tool", "name": tc.ToolName}
			}
		}
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	return body
}

// anthropicContent 把 OpenAI 风格的 content_json 转成 Anthropic content blocks。
// 纯文本/老插件路径仍返回单个 text block；未知类型原样保留，避免静默丢数据。
func anthropicContent(raw []byte, fallback string) []interface{} {
	if len(raw) == 0 {
		return []interface{}{map[string]interface{}{"type": "text", "text": fallback}}
	}
	var parts []map[string]interface{}
	if json.Unmarshal(raw, &parts) != nil || len(parts) == 0 {
		return []interface{}{map[string]interface{}{"type": "text", "text": fallback}}
	}
	out := make([]interface{}, 0, len(parts))
	for _, p := range parts {
		switch p["type"] {
		case "text", "input_text", "output_text":
			out = append(out, map[string]interface{}{"type": "text", "text": stringValue(p["text"])})
		case "image_url":
			url := ""
			if v, ok := p["image_url"].(map[string]interface{}); ok {
				url = stringValue(v["url"])
			} else {
				url = stringValue(p["image_url"])
			}
			out = append(out, anthropicImageBlock(url))
		case "file":
			out = append(out, anthropicFileBlock(p["file"]))
		default:
			out = append(out, p)
		}
	}
	return out
}

func anthropicImageBlock(url string) map[string]interface{} {
	if strings.HasPrefix(url, "data:") {
		header, data := splitDataURL(url)
		return map[string]interface{}{"type": "image", "source": map[string]interface{}{
			"type": "base64", "media_type": header, "data": data,
		}}
	}
	return map[string]interface{}{"type": "image", "source": map[string]interface{}{
		"type": "url", "url": url,
	}}
}

func anthropicFileBlock(value interface{}) map[string]interface{} {
	file, _ := value.(map[string]interface{})
	data := stringValue(file["file_data"])
	if data == "" {
		data = stringValue(file["data"])
	}
	if strings.HasPrefix(data, "data:") {
		media, encoded := splitDataURL(data)
		return map[string]interface{}{"type": "document", "source": map[string]interface{}{
			"type": "base64", "media_type": media, "data": encoded,
		}}
	}
	if url := stringValue(file["file_url"]); url != "" {
		return map[string]interface{}{"type": "document", "source": map[string]interface{}{
			"type": "url", "url": url,
		}}
	}
	return map[string]interface{}{"type": "document", "source": map[string]interface{}{
		"type": "text", "media_type": "text/plain", "data": data,
	}}
}

func splitDataURL(raw string) (media, data string) {
	const prefix = "data:"
	if !strings.HasPrefix(raw, prefix) {
		return "application/octet-stream", raw
	}
	parts := strings.SplitN(strings.TrimPrefix(raw, prefix), ",", 2)
	if len(parts) != 2 {
		return "application/octet-stream", raw
	}
	media = strings.TrimSuffix(parts[0], ";base64")
	data = parts[1]
	if decoded, err := base64.StdEncoding.DecodeString(data); err == nil {
		data = base64.StdEncoding.EncodeToString(decoded)
	}
	return media, data
}

func stringValue(value interface{}) string {
	if s, ok := value.(string); ok {
		return s
	}
	if value == nil {
		return ""
	}
	b, _ := json.Marshal(value)
	return string(b)
}

// Parser 把上游 Anthropic SSE 行解析为信封事件。
type Parser struct {
	emit       func(*pb.StreamEvent)
	blocks     map[int]blockInfo // content block index → 身份
	nextToolID int
	pendingUse *pb.Usage
	sentFinish bool
}

type blockInfo struct {
	kind string // text / tool_use
	id   string
	name string
}

func NewParser(emit func(*pb.StreamEvent)) *Parser {
	return &Parser{emit: emit, blocks: map[int]blockInfo{}}
}

// Feed 处理一行（"event: xxx" 与 "data: {...}"）。
func (p *Parser) Feed(line string) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "data:") {
		return
	}
	payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	if payload == "" {
		return
	}
	var ev struct {
		Type    string `json:"type"`
		Index   int    `json:"index"`
		Message struct {
			Model string `json:"model"`
			Usage struct {
				InputTokens              int64 `json:"input_tokens"`
				CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
				CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
			} `json:"usage"`
		} `json:"message"`
		ContentBlock struct {
			Type  string          `json:"type"`
			ID    string          `json:"id"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
			Text  string          `json:"text"`
		} `json:"content_block"`
		Delta struct {
			Type        string `json:"type"`
			Text        string `json:"text"`
			PartialJSON string `json:"partial_json"`
			StopReason  string `json:"stop_reason"`
		} `json:"delta"`
		Usage struct {
			OutputTokens             int64 `json:"output_tokens"`
			InputTokens              int64 `json:"input_tokens"`
			CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
			CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal([]byte(payload), &ev); err != nil {
		return
	}
	switch ev.Type {
	case "message_start":
		// message_start 携带输入侧用量（含 prompt 缓存命中/写入），先在本地挂起，
		// 等 message_delta / message_stop 拿到 output_tokens 后一起上报。
		p.pendingUse = mergeUsage(p.pendingUse, &pb.Usage{
			InputTokens:  ev.Message.Usage.InputTokens,
			CachedTokens: ev.Message.Usage.CacheReadInputTokens,
		})
		p.emit(&pb.StreamEvent{Event: &pb.StreamEvent_MessageStart{
			MessageStart: &pb.MessageStart{Model: ev.Message.Model},
		}})
	case "content_block_start":
		p.blocks[ev.Index] = blockInfo{kind: ev.ContentBlock.Type, id: ev.ContentBlock.ID, name: ev.ContentBlock.Name}
	case "content_block_delta":
		switch ev.Delta.Type {
		case "text_delta":
			p.emit(&pb.StreamEvent{Event: &pb.StreamEvent_ContentDelta{
				ContentDelta: &pb.ContentDelta{Text: ev.Delta.Text},
			}})
		case "input_json_delta":
			info := p.blocks[ev.Index]
			p.emit(&pb.StreamEvent{Event: &pb.StreamEvent_ToolCallDelta{
				ToolCallDelta: &pb.ToolCallDelta{
					Id: info.id, Name: info.name, ArgumentsDelta: ev.Delta.PartialJSON,
				},
			}})
		}
	case "message_delta":
		// stop_reason + output_tokens 通常都在这里；部分上游会在 message_delta 里再带一次
		// 输入侧用量（Anthropic 官方在新版本里会补 input_tokens / cache_read_input_tokens）。
		p.pendingUse = mergeUsage(p.pendingUse, &pb.Usage{
			InputTokens:  ev.Usage.InputTokens,
			CachedTokens: ev.Usage.CacheReadInputTokens,
			OutputTokens: ev.Usage.OutputTokens,
		})
		if ev.Delta.StopReason != "" {
			p.finish(mapStop(ev.Delta.StopReason))
		}
	case "message_stop":
		p.finish("stop")
	}
}

// Finish 流结束兜底。
func (p *Parser) Finish() {
	if !p.sentFinish {
		p.emit(&pb.StreamEvent{Event: &pb.StreamEvent_MessageFinish{
			MessageFinish: &pb.MessageFinish{FinishReason: "stop"},
		}})
	}
}

// FinishWithError 流异常结束：发失败事件。
func (p *Parser) FinishWithError(code int32, message string) {
	p.emit(&pb.StreamEvent{Event: &pb.StreamEvent_TaskFailed{
		TaskFailed: &pb.TaskFailed{Error: &pb.Error{Code: code, Message: message}},
	}})
}

func (p *Parser) finish(reason string) {
	if p.sentFinish {
		return
	}
	p.sentFinish = true
	if reason == "tool_use" {
		// 信封语义用 tool_calls
		reason = "tool_calls"
	}
	p.emit(&pb.StreamEvent{Event: &pb.StreamEvent_MessageFinish{
		MessageFinish: &pb.MessageFinish{
			FinishReason: reason,
			Usage:        p.pendingUse,
		},
	}})
}

// mergeUsage 合并两次上游用量快照，取每个字段的较大值（上游分片上报，非累加语义）。
func mergeUsage(dst, src *pb.Usage) *pb.Usage {
	if dst == nil {
		return src
	}
	if src == nil {
		return dst
	}
	if src.InputTokens > dst.InputTokens {
		dst.InputTokens = src.InputTokens
	}
	if src.OutputTokens > dst.OutputTokens {
		dst.OutputTokens = src.OutputTokens
	}
	if src.CachedTokens > dst.CachedTokens {
		dst.CachedTokens = src.CachedTokens
	}
	return dst
}

// ---------- 工具 ----------

func rawJSON(s string) interface{} {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return map[string]interface{}{}
	}
	return v
}

func mapStop(reason string) string {
	switch reason {
	case "tool_use":
		return "tool_calls"
	case "max_tokens":
		return "length"
	default:
		return "stop"
	}
}

func orInt(v, def int32) int32 {
	if v > 0 {
		return v
	}
	return def
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
