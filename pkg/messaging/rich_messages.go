package messaging

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	protocol "github.com/hanzywaifu-coder/dongtube-meowbail/internal/protocol"

	"go.mau.fi/whatsmeow/proto/waAICommon"
	"go.mau.fi/whatsmeow/proto/waAICommonDeprecated"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func generateRandomUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// SendAITableResponse mengirim pesan tabel AI bergaya Meta AI / rich response message
func SendAITableResponse(c *protocol.Client, ctx context.Context, chat protocol.JID, title string, headers []string, rows [][]string) error {
	var tableRows []*waAICommonDeprecated.AIRichResponseTableMetadata_AIRichResponseTableRow

	// Header row
	if len(headers) > 0 {
		tableRows = append(tableRows, &waAICommonDeprecated.AIRichResponseTableMetadata_AIRichResponseTableRow{
			Items:     headers,
			IsHeading: proto.Bool(true),
		})
	}

	// Data rows
	for _, r := range rows {
		tableRows = append(tableRows, &waAICommonDeprecated.AIRichResponseTableMetadata_AIRichResponseTableRow{
			Items:     r,
			IsHeading: proto.Bool(false),
		})
	}

	tableSubMsgType := waAICommonDeprecated.AIRichResponseSubMessageType_AI_RICH_RESPONSE_TABLE
	submessages := []*waAICommonDeprecated.AIRichResponseSubMessage{
		{
			MessageType: &tableSubMsgType,
			TableMetadata: &waAICommonDeprecated.AIRichResponseTableMetadata{
				Title: proto.String(title),
				Rows:  tableRows,
			},
		},
	}

	// Bangun format unified response JSON untuk engine render UI WhatsApp
	var unifiedRows []map[string]interface{}
	for _, tr := range tableRows {
		var cells []map[string]string
		for _, it := range tr.Items {
			cells = append(cells, map[string]string{"text": it})
		}
		unifiedRows = append(unifiedRows, map[string]interface{}{
			"is_header":      tr.GetIsHeading(),
			"cells":          tr.Items,
			"markdown_cells": cells,
		})
	}

	unifiedObj := map[string]interface{}{
		"response_id": generateRandomUUID(),
		"sections": []map[string]interface{}{
			{
				"view_model": map[string]interface{}{
					"primitive": map[string]interface{}{
						"title":      title,
						"rows":       unifiedRows,
						"__typename": "GenATableUXPrimitive",
					},
					"__typename": "GenAISingleLayoutViewModel",
				},
			},
		},
	}
	unifiedJSON, _ := json.Marshal(unifiedObj)

	richMsgType := waAICommonDeprecated.AIRichResponseMessageType_AI_RICH_RESPONSE_TYPE_STANDARD
	botMsg := &waE2E.Message{
		BotForwardedMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				RichResponseMessage: &waE2E.AIRichResponseMessage{
					MessageType: &richMsgType,
					Submessages: submessages,
					UnifiedResponse: &waAICommon.AIRichResponseUnifiedResponse{
						Data: unifiedJSON,
					},
					ContextInfo: &waE2E.ContextInfo{
						IsForwarded:   proto.Bool(false),
						ForwardOrigin: waE2E.ContextInfo_META_AI.Enum(),
						ForwardedAiBotMessageInfo: &waAICommon.ForwardedAIBotMessageInfo{
							BotJID: proto.String("867051314767696@bot"),
						},
					},
				},
			},
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, botMsg)
	return err
}

// SendAICodeResponse mengirim pesan format code block dengan syntax highlight khas Meta AI
func SendAICodeResponse(c *protocol.Client, ctx context.Context, chat protocol.JID, title, code, language string) error {
	if language == "" {
		language = "javascript"
	}

	codeSubMsgType := waAICommonDeprecated.AIRichResponseSubMessageType_AI_RICH_RESPONSE_CODE
	defaultHighlight := waAICommonDeprecated.AIRichResponseCodeMetadata_AI_RICH_RESPONSE_CODE_HIGHLIGHT_DEFAULT

	submessages := []*waAICommonDeprecated.AIRichResponseSubMessage{
		{
			MessageType: &codeSubMsgType,
			CodeMetadata: &waAICommonDeprecated.AIRichResponseCodeMetadata{
				CodeLanguage: proto.String(language),
				CodeBlocks: []*waAICommonDeprecated.AIRichResponseCodeMetadata_AIRichResponseCodeBlock{
					{
						HighlightType: &defaultHighlight,
						CodeContent:   proto.String(code),
					},
				},
			},
		},
	}

	unifiedObj := map[string]interface{}{
		"response_id": generateRandomUUID(),
		"sections": []map[string]interface{}{
			{
				"view_model": map[string]interface{}{
					"primitive": map[string]interface{}{
						"language": language,
						"code_blocks": []map[string]string{
							{
								"content": code,
								"type":    "DEFAULT",
							},
						},
						"__typename": "GenAICodeUXPrimitive",
					},
					"__typename": "GenAISingleLayoutViewModel",
				},
			},
		},
	}
	unifiedJSON, _ := json.Marshal(unifiedObj)

	richMsgType := waAICommonDeprecated.AIRichResponseMessageType_AI_RICH_RESPONSE_TYPE_STANDARD
	botMsg := &waE2E.Message{
		BotForwardedMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				RichResponseMessage: &waE2E.AIRichResponseMessage{
					MessageType: &richMsgType,
					Submessages: submessages,
					UnifiedResponse: &waAICommon.AIRichResponseUnifiedResponse{
						Data: unifiedJSON,
					},
					ContextInfo: &waE2E.ContextInfo{
						IsForwarded:   proto.Bool(false),
						ForwardOrigin: waE2E.ContextInfo_META_AI.Enum(),
						ForwardedAiBotMessageInfo: &waAICommon.ForwardedAIBotMessageInfo{
							BotJID: proto.String("867051314767696@bot"),
						},
					},
				},
			},
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, botMsg)
	return err
}

// SendAILatexResponse mengirim pesan formula matematika / LaTeX khas Meta AI
func SendAILatexResponse(c *protocol.Client, ctx context.Context, chat protocol.JID, text, expression string) error {
	latexSubMsgType := waAICommonDeprecated.AIRichResponseSubMessageType_AI_RICH_RESPONSE_LATEX

	submessages := []*waAICommonDeprecated.AIRichResponseSubMessage{
		{
			MessageType: &latexSubMsgType,
			LatexMetadata: &waAICommonDeprecated.AIRichResponseLatexMetadata{
				Text: proto.String(text),
				Expressions: []*waAICommonDeprecated.AIRichResponseLatexMetadata_AIRichResponseLatexExpression{
					{
						LatexExpression: proto.String(expression),
					},
				},
			},
		},
	}

	richMsgType := waAICommonDeprecated.AIRichResponseMessageType_AI_RICH_RESPONSE_TYPE_STANDARD
	botMsg := &waE2E.Message{
		BotForwardedMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				RichResponseMessage: &waE2E.AIRichResponseMessage{
					MessageType: &richMsgType,
					Submessages: submessages,
					ContextInfo: &waE2E.ContextInfo{
						IsForwarded:   proto.Bool(false),
						ForwardedAiBotMessageInfo: &waAICommon.ForwardedAIBotMessageInfo{
							BotJID: proto.String("867051314767696@bot"),
						},
					},
				},
			},
		},
	}

	_, err := c.Client.SendMessage(ctx, chat, botMsg)
	return err
}
