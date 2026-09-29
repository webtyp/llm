package llm_test

import (
	"fmt"
	"testing"

	"webtyp.com/context"
	"webtyp.com/llm"
)

type scriptedClient struct {
	replies []llm.Response
	seen    []llm.Request
}

func (s *scriptedClient) Generate(ctx *context.Context, req llm.Request) (llm.Response, error) {
	s.seen = append(s.seen, req)
	if len(s.replies) == 0 {
		return llm.Response{}, nil
	}
	resp := s.replies[0]
	s.replies = s.replies[1:]
	return resp, nil
}

func Example_toolRoundTrip() {
	client := &scriptedClient{
		replies: []llm.Response{
			{
				StopReason: llm.StopToolUse,
				ToolCalls: []llm.ToolCall{
					{ID: "c1", Name: "clinic_hours", Input: "{}"},
				},
			},
			{
				StopReason: llm.StopEndTurn,
				Text:       "Abrimos de 8 a 20",
			},
		},
	}

	ctx := context.Background()
	req := llm.Request{
		System: "You are the receptionist of a clinic.",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "¿A qué hora abren?"},
		},
		Tools: []llm.ToolDef{
			{Name: "clinic_hours", Description: "Returns clinic operating hours", InputSchema: "{}"},
		},
		MaxOutputTokens: 256,
	}

	for {
		resp, err := client.Generate(ctx, req)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		if resp.StopReason == llm.StopToolUse {
			req.Messages = append(req.Messages, llm.Message{
				Role:      llm.RoleAssistant,
				ToolCalls: resp.ToolCalls,
			})

			for _, call := range resp.ToolCalls {
				if call.Name == "clinic_hours" {
					req.Messages = append(req.Messages, llm.Message{
						Role:       llm.RoleTool,
						ToolCallID: call.ID,
						ToolName:   call.Name,
						Content:    "8-20",
					})
				}
			}
			continue
		}

		fmt.Println(resp.Text)
		break
	}

	// Output:
	// Abrimos de 8 a 20
}

func TestRoundTrip_SecondRequestCarriesToolResult(t *testing.T) {
	client := &scriptedClient{
		replies: []llm.Response{
			{
				StopReason: llm.StopToolUse,
				ToolCalls: []llm.ToolCall{
					{ID: "c1", Name: "clinic_hours", Input: "{}"},
				},
			},
			{
				StopReason: llm.StopEndTurn,
				Text:       "Abrimos de 8 a 20",
			},
		},
	}

	ctx := context.Background()
	req := llm.Request{
		System: "You are the receptionist of a clinic.",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "¿A qué hora abren?"},
		},
		Tools: []llm.ToolDef{
			{Name: "clinic_hours", Description: "Returns clinic operating hours", InputSchema: "{}"},
		},
		MaxOutputTokens: 256,
	}

	for {
		resp, err := client.Generate(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.StopReason == llm.StopToolUse {
			req.Messages = append(req.Messages, llm.Message{
				Role:      llm.RoleAssistant,
				ToolCalls: resp.ToolCalls,
			})

			for _, call := range resp.ToolCalls {
				if call.Name == "clinic_hours" {
					req.Messages = append(req.Messages, llm.Message{
						Role:       llm.RoleTool,
						ToolCallID: call.ID,
						ToolName:   call.Name,
						Content:    "8-20",
					})
				}
			}
			continue
		}
		break
	}

	if len(client.seen) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(client.seen))
	}

	secondReqMsgs := client.seen[1].Messages
	if len(secondReqMsgs) != 3 {
		t.Fatalf("expected 3 messages in second request, got %d", len(secondReqMsgs))
	}

	lastMsg := secondReqMsgs[2]
	if lastMsg.Role != llm.RoleTool {
		t.Errorf("expected last message role %q, got %q", llm.RoleTool, lastMsg.Role)
	}
	if lastMsg.ToolCallID != "c1" {
		t.Errorf("expected ToolCallID %q, got %q", "c1", lastMsg.ToolCallID)
	}
	if lastMsg.ToolName != "clinic_hours" {
		t.Errorf("expected ToolName %q, got %q", "clinic_hours", lastMsg.ToolName)
	}
}

type chunkedStreamer struct {
	chunks []string
	resp   llm.Response
}

func (c chunkedStreamer) GenerateStream(ctx *context.Context, req llm.Request, onText func(text string)) (llm.Response, error) {
	for _, chunk := range c.chunks {
		onText(chunk)
	}
	return c.resp, nil
}

func TestStreamer_PiecesConcatenateToText(t *testing.T) {
	var _ llm.Streamer = chunkedStreamer{}

	s := chunkedStreamer{
		chunks: []string{"Abri", "mos a ", "las 8"},
		resp: llm.Response{
			Text:       "Abrimos a las 8",
			StopReason: llm.StopEndTurn,
		},
	}

	ctx := context.Background()
	var accumulated string
	resp, err := s.GenerateStream(ctx, llm.Request{}, func(text string) {
		accumulated += text
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if accumulated != resp.Text {
		t.Errorf("expected accumulated text %q to equal response text %q", accumulated, resp.Text)
	}
}
