// responses.go — OpenAI Responses 协议（Codex CLI）↔ 统一信封。
package gateway

import (
	"encoding/json"
	"fmt"

	pb "github.com/ShadowSmallBaby/ClawProxyHub/sdk/proto/cphv1"
)

// parseResponsesRequest 把 /v1/responses 请求体转成统一信封。
func parseResponsesRequest(body []byte) (*pb.ChatRequest, error) {
	var raw struct {
		Model           string          `json:"model"`
		Instructions    string          `json:"instructions"`
		Input           json.RawMessage `json:"input"`
		Tools           []respTool      `json:"tools"`
		ToolChoice      json.RawMessage `json:"tool_choice"`
		MaxOutputTokens int32           `json:"max_output_tokens"`
		Temperature     *float64        `json:"temperature"`
		TopP            *float64        `json:"top_p"`
		Stream          bool            `json:"stream"`
		Reasoning       json.RawMessage `json:"reasoning"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	req := &pb.ChatRequest{
		Model:       raw.Model,
		Stream:      raw.Stream,
		MaxTokens:   raw.MaxOutputTokens,
		Temperature: deref(raw.Temperature),
		Extra:       map[string]string{},
	}
	if raw.TopP != nil {
		req.Extra["top_p"] = fmt.Sprintf("%g", *raw.TopP)
	}
	// 不透传 reasoning / reasoning.effort：Responses 本无顶层 reasoning_effort，
	// 而 Codex 会发 "xhigh" 这类上游不认的私有值，透传过去直接 500。
	// （对齐上游 v1.0.2：该字段在归一化时丢弃。）
	if raw.Instructions != "" {
		req.Messages = append(req.Messages, &pb.EnvelopeMessage{Role: "system", Text: raw.Instructions})
	}

	// input 可能是纯字符串，也可能是消息数组
	var inputText string
	if err := json.Unmarshal(raw.Input, &inputText); err == nil && inputText != "" {
		req.Messages = append(req.Messages, &pb.EnvelopeMessage{Role: "user", Text: inputText})
	} else {
		var items []struct {
			Type    string          `json:"type"`
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
			// function_call（assistant 历史里的工具调用）
			CallID    string `json:"call_id"`
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
			// function_call_output（工具结果）
			Output string `json:"output"`
		}
		if err := json.Unmarshal(raw.Input, &items); err != nil {
			return nil, fmt.Errorf("input must be string or message array")
		}
		// 合并相邻 assistant message 与 function_call：并行调用必须归入同一条
		// assistant 的 tool_calls。否则会退化成「N 条各带 1 个 tool_call 的 assistant」，
		// 随后的 tool 消息与声明它的 assistant 错位，上游直接拒绝。
		var pendText string
		var pendAssistant bool
		var pendTools []*pb.ToolCall
		flushAssistant := func() {
			if !pendAssistant && len(pendTools) == 0 {
				return
			}
			req.Messages = append(req.Messages, &pb.EnvelopeMessage{
				Role: "assistant", Text: pendText, ToolCalls: pendTools,
			})
			pendText, pendAssistant, pendTools = "", false, nil
		}
		for _, it := range items {
			switch it.Type {
			case "message", "":
				role := normalizeRole(it.Role)
				if role == "assistant" {
					flushAssistant()
					pendText, pendAssistant = extractText(it.Content), true
					continue
				}
				flushAssistant()
				req.Messages = append(req.Messages, &pb.EnvelopeMessage{
					Role: role, Text: extractText(it.Content),
				})
			case "function_call":
				args := it.Arguments
				if args == "" {
					args = "{}" // 上游要求 arguments 是合法 JSON 文本
				}
				pendTools = append(pendTools, &pb.ToolCall{Id: it.CallID, Name: it.Name, Arguments: args})
			case "function_call_output":
				flushAssistant()
				req.Messages = append(req.Messages, &pb.EnvelopeMessage{
					Role: "tool", Text: it.Output, ToolCallId: it.CallID,
				})
			}
		}
		flushAssistant()
	}

	for _, t := range raw.Tools {
		if t.Type == "function" && t.Name != "" {
			req.Tools = append(req.Tools, &pb.ToolDefinition{
				Name: t.Name, Description: t.Description,
				ParametersSchema: string(t.Parameters),
			})
		}
	}
	if len(raw.ToolChoice) > 0 {
		var s string
		if err := json.Unmarshal(raw.ToolChoice, &s); err == nil {
			req.ToolChoice = &pb.ToolChoice{Type: s}
		}
	}
	return req, nil
}

type respTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ---------- 信封事件 → Responses SSE ----------

// respFnItem 一个 function_call output item 的累积状态。
type respFnItem struct {
	itemID string
	name   string
	idx    int
	args   string
}

type responsesSSEState struct {
	model    string
	respID   string
	nextItem int // 递增的 output_index；文本 item 与 function_call item 共用同一序列
	textItem string
	textIdx  int
	text     string
	fnItems  map[string]*respFnItem
	fnOrder  []string
	tools    toolCallTracker
}

func newResponsesSSEState(model string) *responsesSSEState {
	return &responsesSSEState{
		model: model, respID: "resp_" + randHex(16),
		textItem: "", textIdx: -1, fnItems: map[string]*respFnItem{},
	}
}

// addFn 首次出现的工具调用 → 建 item 并返回（needStart=true 表示要发 output_item.added）。
func (s *responsesSSEState) addFn(id, name string) (*respFnItem, bool) {
	if it, ok := s.fnItems[id]; ok {
		if it.name == "" && name != "" {
			it.name = name
		}
		return it, false
	}
	it := &respFnItem{itemID: fmt.Sprintf("item_%d", s.nextItem), name: name, idx: s.nextItem}
	s.nextItem++
	s.fnItems[id] = it
	s.fnOrder = append(s.fnOrder, id)
	return it, true
}

// outputItem function_call item 的 JSON 表示（added / done 共用，status 不同）。
func (it *respFnItem) outputItem(id, status string) map[string]interface{} {
	return map[string]interface{}{
		"type": "function_call", "id": it.itemID, "call_id": id,
		"name": it.name, "arguments": it.args, "status": status,
	}
}

func (s *responsesSSEState) convertEvent(ev *pb.StreamEvent) string {
	switch e := ev.Event.(type) {
	case *pb.StreamEvent_MessageStart:
		return respEvent("response.created", map[string]interface{}{
			"response": map[string]interface{}{
				"id": s.respID, "object": "response", "model": e.MessageStart.Model,
				"status": "in_progress", "output": []interface{}{},
			},
		})

	case *pb.StreamEvent_ContentDelta:
		var out string
		if s.textItem == "" {
			s.textItem = fmt.Sprintf("item_%d", s.nextItem)
			s.textIdx = s.nextItem
			s.nextItem++
			out += respEvent("response.output_item.added", map[string]interface{}{
				"output_index": s.textIdx, "item": map[string]interface{}{
					"type": "message", "id": s.textItem, "role": "assistant", "status": "in_progress",
					"content": []interface{}{map[string]interface{}{"type": "output_text", "text": ""}},
				},
			})
		}
		s.text += e.ContentDelta.Text
		out += respEvent("response.output_text.delta", map[string]interface{}{
			"item_id": s.textItem, "output_index": s.textIdx, "content_index": 0,
			"delta": e.ContentDelta.Text,
		})
		return out

	case *pb.StreamEvent_ToolCallDelta:
		id, name := s.tools.resolve(e.ToolCallDelta)
		it, needStart := s.addFn(id, name)
		var out string
		if needStart {
			out += respEvent("response.output_item.added", map[string]interface{}{
				"output_index": it.idx, "item": it.outputItem(id, "in_progress"),
			})
		}
		if e.ToolCallDelta.ArgumentsDelta != "" {
			it.args += e.ToolCallDelta.ArgumentsDelta
			out += respEvent("response.function_call_arguments.delta", map[string]interface{}{
				"item_id": it.itemID, "output_index": it.idx,
				"delta": e.ToolCallDelta.ArgumentsDelta,
			})
		}
		return out

	case *pb.StreamEvent_MessageFinish:
		var out string
		// 输出项必须逐个 output_item.done 收尾：Codex CLI 只认 done 事件里的
		// function_call（缺了它工具不会被调度执行，表现为「复杂操作无回复」）。
		var output []interface{}
		if s.textItem != "" {
			out += respEvent("response.output_text.done", map[string]interface{}{
				"item_id": s.textItem, "output_index": s.textIdx, "content_index": 0, "text": s.text,
			})
			item := map[string]interface{}{
				"type": "message", "id": s.textItem, "role": "assistant", "status": "completed",
				"content": []interface{}{map[string]interface{}{
					"type": "output_text", "text": s.text, "annotations": []interface{}{},
				}},
			}
			out += respEvent("response.output_item.done", map[string]interface{}{
				"output_index": s.textIdx, "item": item,
			})
			output = append(output, item)
		}
		for _, id := range s.fnOrder {
			it := s.fnItems[id]
			out += respEvent("response.function_call_arguments.done", map[string]interface{}{
				"item_id": it.itemID, "output_index": it.idx, "arguments": it.args,
			})
			item := it.outputItem(id, "completed")
			out += respEvent("response.output_item.done", map[string]interface{}{
				"output_index": it.idx, "item": item,
			})
			output = append(output, item)
		}
		// usage 为必填字段，缺失时补零值（Codex 严格反序列化，否则断流）。
		var inTok, outTok int64
		if e.MessageFinish.Usage != nil {
			inTok, outTok = e.MessageFinish.Usage.InputTokens, e.MessageFinish.Usage.OutputTokens
		}
		usage := map[string]interface{}{
			"input_tokens":  inTok,
			"output_tokens": outTok,
			"total_tokens":  inTok + outTok,
		}
		out += respEvent("response.completed", map[string]interface{}{
			"response": map[string]interface{}{
				"id": s.respID, "object": "response", "model": s.model,
				"status": "completed", "output": output, "usage": usage,
			},
		})
		return out
	}
	return ""
}

func (s *responsesSSEState) finish() string { return "" }

func respEvent(eventType string, payload map[string]interface{}) string {
	payload["type"] = eventType
	b, _ := json.Marshal(payload)
	return "event: " + eventType + "\ndata: " + string(b) + "\n\n"
}

// responsesAggregate Responses 非流式聚合。
type responsesAggregate struct {
	model  string
	text   string
	tools  map[string]*aggrTool
	track  toolCallTracker
	order  []string
	finish string
	input  int64
	output int64
}

func (a *responsesAggregate) feed(ev *pb.StreamEvent) {
	switch e := ev.Event.(type) {
	case *pb.StreamEvent_MessageStart:
		a.model = e.MessageStart.Model
	case *pb.StreamEvent_ContentDelta:
		a.text += e.ContentDelta.Text
	case *pb.StreamEvent_ToolCallDelta:
		id, name := a.track.resolve(e.ToolCallDelta)
		if a.tools == nil {
			a.tools = map[string]*aggrTool{}
		}
		t, ok := a.tools[id]
		if !ok {
			t = &aggrTool{id: id, name: name}
			a.tools[id] = t
			a.order = append(a.order, id)
		}
		t.input += e.ToolCallDelta.ArgumentsDelta
	case *pb.StreamEvent_MessageFinish:
		a.finish = e.MessageFinish.FinishReason
		if e.MessageFinish.Usage != nil {
			a.input, a.output = e.MessageFinish.Usage.InputTokens, e.MessageFinish.Usage.OutputTokens
		}
	}
}

func (a *responsesAggregate) result() map[string]interface{} {
	var output []interface{}
	if a.text != "" {
		output = append(output, map[string]interface{}{
			"type": "message", "id": "item_0", "role": "assistant", "status": "completed",
			"content": []interface{}{map[string]interface{}{
				"type": "output_text", "text": a.text, "annotations": []interface{}{},
			}},
		})
	}
	// 工具项按出现顺序输出（顺序与流式编码器一致）
	for _, id := range a.order {
		t := a.tools[id]
		output = append(output, map[string]interface{}{
			"type": "function_call", "id": "item_" + id, "call_id": t.id, "name": t.name,
			"arguments": t.input, "status": "completed",
		})
	}
	return map[string]interface{}{
		"id": "resp_" + randHex(16), "object": "response", "model": a.model,
		"status": "completed", "output": output,
		"usage": map[string]interface{}{
			"input_tokens": a.input, "output_tokens": a.output,
			"total_tokens": a.input + a.output,
		},
	}
}
