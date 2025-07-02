package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

func NotifyWebhook(userID int64, ip string) {
	webhookURL := "http://webhook:9000/hooks/token-alert"

	payload := map[string]interface{}{
		"user_id":  userID,
		"new_ip":   ip,
		"time":     time.Now().Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(payload)
	http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
}
