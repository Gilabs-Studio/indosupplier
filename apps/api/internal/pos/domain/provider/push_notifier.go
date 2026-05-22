package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gilabs/gims/api/internal/pos/data/models"
)

type PushPayload struct {
	Title string
	Body  string
	Data  map[string]string
}

type PushNotifier interface {
	Send(ctx context.Context, tokens []models.POSDeviceToken, payload PushPayload) error
}

type MultiPlatformPushNotifier struct{}

func NewMultiPlatformPushNotifier() PushNotifier {
	return &MultiPlatformPushNotifier{}
}

func (n *MultiPlatformPushNotifier) Send(ctx context.Context, tokens []models.POSDeviceToken, payload PushPayload) error {
	if len(tokens) == 0 {
		return nil
	}
	fcmServerKey := os.Getenv("FCM_SERVER_KEY")
	fcmConfigured := fcmServerKey != "" || os.Getenv("FCM_SERVICE_ACCOUNT_JSON") != "" || os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != ""
	apnsConfigured := os.Getenv("APNS_KEY_ID") != "" && os.Getenv("APNS_TEAM_ID") != ""
	wnsConfigured := os.Getenv("WNS_PACKAGE_SID") != "" && os.Getenv("WNS_CLIENT_SECRET") != ""

	for _, token := range tokens {
		switch token.Platform {
		case "android":
			if !fcmConfigured {
				log.Printf("[pos_push] FCM not configured; skipped token_id=%s", token.ID)
				continue
			}
			if fcmServerKey != "" {
				if err := sendFCMLegacy(ctx, fcmServerKey, token.Token, payload); err != nil {
					log.Printf("[pos_push] FCM send failed token_id=%s err=%v", token.ID, err)
				}
				continue
			}
			log.Printf("[pos_push] FCM ready token_id=%s title=%q", token.ID, payload.Title)
		case "ios":
			if !apnsConfigured {
				log.Printf("[pos_push] APNs not configured; skipped token_id=%s", token.ID)
				continue
			}
			log.Printf("[pos_push] APNs ready token_id=%s title=%q", token.ID, payload.Title)
		case "windows":
			if !wnsConfigured {
				log.Printf("[pos_push] WNS not configured; skipped token_id=%s", token.ID)
				continue
			}
			log.Printf("[pos_push] WNS ready token_id=%s title=%q", token.ID, payload.Title)
		case "linux":
			log.Printf("[pos_push] Linux background push is not provider-backed; skipped token_id=%s", token.ID)
		}
	}
	return nil
}

func sendFCMLegacy(ctx context.Context, serverKey, token string, payload PushPayload) error {
	body, err := json.Marshal(map[string]interface{}{
		"to": token,
		"notification": map[string]string{
			"title": payload.Title,
			"body":  payload.Body,
			"sound": "notification",
		},
		"data": payload.Data,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "key="+serverKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return &pushHTTPError{status: resp.Status}
	}
	return nil
}

type pushHTTPError struct {
	status string
}

func (e *pushHTTPError) Error() string {
	return "push provider returned " + e.status
}
