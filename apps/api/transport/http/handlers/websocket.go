package handlers

import "net/http"

// WebSocketAccountUpdatesDoc godoc
// @Summary Subscribe to account realtime updates
// @Description Opens an account-scoped websocket that emits data-changed notifications.
// @Tags realtime
// @Produce json
// @Param account_id path string true "Account ID"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} map[string]string
// @Router /ws/accounts/{account_id} [get]
func WebSocketAccountUpdatesDoc(_ http.ResponseWriter, _ *http.Request) {}
