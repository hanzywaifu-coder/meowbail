package buttons

import (
	"encoding/json"

	"go.mau.fi/whatsmeow/types/events"
)

// HandleButtonResponse mendeteksi semua respon tombol (ButtonsResponseMessage, ListResponseMessage,
// dan InteractiveResponseMessage / NativeFlow Response)
func HandleButtonResponse(evt interface{}) (string, bool) {
	e, ok := evt.(*events.Message)
	if !ok || e.Message == nil {
		return "", false
	}

	// 1. Tombol ButtonsResponseMessage lama (Quick Reply / Buttons)
	if e.Message.ButtonsResponseMessage != nil {
		selectedID := e.Message.ButtonsResponseMessage.GetSelectedButtonID()
		if selectedID != "" {
			return selectedID, true
		}
	}

	// 2. Tombol ListResponseMessage (Dropdown Klasik)
	if e.Message.ListResponseMessage != nil {
		singleSelect := e.Message.ListResponseMessage.GetSingleSelectReply()
		if singleSelect != nil {
			return singleSelect.GetSelectedRowID(), true
		}
	}

	// 3. Tombol InteractiveResponseMessage (NativeFlow single_select, quick_reply, dll)
	if e.Message.InteractiveResponseMessage != nil {
		nativeFlowResp := e.Message.InteractiveResponseMessage.GetNativeFlowResponseMessage()
		if nativeFlowResp != nil {
			paramsJSON := nativeFlowResp.GetParamsJSON()
			if paramsJSON != "" {
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(paramsJSON), &parsed); err == nil {
					// Dropdown single_select mengembalikan "id" di dalam params_json
					if id, ok := parsed["id"].(string); ok && id != "" {
						return id, true
					}
					// Atau di beberapa tipe flow: "selected_row_id" / "row_id"
					if id, ok := parsed["selected_row_id"].(string); ok && id != "" {
						return id, true
					}
				}
			}
		}
	}

	return "", false
}
