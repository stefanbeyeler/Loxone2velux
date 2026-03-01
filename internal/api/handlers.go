package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/stefanbeyeler/loxone2velux/internal/config"
	"github.com/stefanbeyeler/loxone2velux/internal/gateway"
	"github.com/stefanbeyeler/loxone2velux/internal/klf200"
)

// ConfigManager interface for configuration operations
type ConfigManager interface {
	GetConfig() *config.Config
	UpdateConfig(cfg *config.Config) error
	GetConfigPath() string
}

// Handlers holds all HTTP handlers
type Handlers struct {
	gateway    *gateway.Service
	logger     zerolog.Logger
	configMgr  ConfigManager
	version    string
}

// NewHandlers creates new handlers
func NewHandlers(gw *gateway.Service, logger zerolog.Logger, configMgr ConfigManager, version string) *Handlers {
	return &Handlers{
		gateway:   gw,
		logger:    logger.With().Str("component", "handlers").Logger(),
		configMgr: configMgr,
		version:   version,
	}
}

// Response types
type HealthResponse struct {
	Status    string `json:"status"`
	Connected bool   `json:"connected"`
	NodeCount int    `json:"node_count"`
	Version   string `json:"version"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

type PositionRequest struct {
	Position float64 `json:"position"`
}

type CommandResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	NodeID  uint8  `json:"node_id"`
}

type NodesResponse struct {
	Nodes []*klf200.Node `json:"nodes"`
	Count int            `json:"count"`
}

// Health returns the health status
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "ok",
		Connected: h.gateway.IsConnected(),
		NodeCount: h.gateway.GetNodeCount(),
		Version:   h.version,
	}

	if !resp.Connected {
		resp.Status = "degraded"
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListNodes returns all nodes
func (h *Handlers) ListNodes(w http.ResponseWriter, r *http.Request) {
	nodes := h.gateway.GetNodes()

	resp := NodesResponse{
		Nodes: nodes,
		Count: len(nodes),
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetNode returns a single node
func (h *Handlers) GetNode(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid node ID", err.Error())
		return
	}

	node, ok := h.gateway.GetNode(nodeID)
	if !ok {
		writeError(w, http.StatusNotFound, "Node not found", "")
		return
	}

	writeJSON(w, http.StatusOK, node)
}

// SetPosition sets the position of a node
func (h *Handlers) SetPosition(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid node ID", err.Error())
		return
	}

	var req PositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Position < 0 || req.Position > 100 {
		writeError(w, http.StatusBadRequest, "Position must be between 0 and 100", "")
		return
	}

	if err := h.gateway.SetPosition(r.Context(), nodeID, req.Position); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to set position", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, CommandResponse{
		Success: true,
		Message: "Position command sent",
		NodeID:  nodeID,
	})
}

// OpenNode fully opens a node
func (h *Handlers) OpenNode(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid node ID", err.Error())
		return
	}

	if err := h.gateway.Open(r.Context(), nodeID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to open node", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, CommandResponse{
		Success: true,
		Message: "Open command sent",
		NodeID:  nodeID,
	})
}

// CloseNode fully closes a node
func (h *Handlers) CloseNode(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid node ID", err.Error())
		return
	}

	if err := h.gateway.Close(r.Context(), nodeID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to close node", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, CommandResponse{
		Success: true,
		Message: "Close command sent",
		NodeID:  nodeID,
	})
}

// StopNode stops a node's movement
func (h *Handlers) StopNode(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid node ID", err.Error())
		return
	}

	if err := h.gateway.StopNode(r.Context(), nodeID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to stop node", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, CommandResponse{
		Success: true,
		Message: "Stop command sent",
		NodeID:  nodeID,
	})
}

// Loxone-friendly endpoints (GET requests with URL parameters)

// LoxoneSetPosition handles Loxone position requests via URL
func (h *Handlers) LoxoneSetPosition(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ERROR"))
		return
	}

	positionStr := chi.URLParam(r, "position")
	position, err := strconv.ParseFloat(positionStr, 64)
	if err != nil || position < 0 || position > 100 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ERROR"))
		return
	}

	if err := h.gateway.SetPosition(r.Context(), nodeID, position); err != nil {
		h.logger.Error().Err(err).Uint8("node", nodeID).Float64("pos", position).Msg("Failed to set position")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// LoxoneOpen handles Loxone open requests
func (h *Handlers) LoxoneOpen(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ERROR"))
		return
	}

	if err := h.gateway.Open(r.Context(), nodeID); err != nil {
		h.logger.Error().Err(err).Uint8("node", nodeID).Msg("Failed to open")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// LoxoneClose handles Loxone close requests
func (h *Handlers) LoxoneClose(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ERROR"))
		return
	}

	if err := h.gateway.Close(r.Context(), nodeID); err != nil {
		h.logger.Error().Err(err).Uint8("node", nodeID).Msg("Failed to close")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// LoxoneGetPosition returns the current position of a node as a plain number (0-100)
func (h *Handlers) LoxoneGetPosition(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ERROR"))
		return
	}

	node, ok := h.gateway.GetNode(nodeID)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("ERROR"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.Itoa(int(node.PositionPercent))))
}

// LoxoneStop handles Loxone stop requests
func (h *Handlers) LoxoneStop(w http.ResponseWriter, r *http.Request) {
	nodeID, err := parseNodeID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("ERROR"))
		return
	}

	if err := h.gateway.StopNode(r.Context(), nodeID); err != nil {
		h.logger.Error().Err(err).Uint8("node", nodeID).Msg("Failed to stop")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Sensor endpoints

// GetSensorStatus returns the current sensor status (rain, wind, etc.)
func (h *Handlers) GetSensorStatus(w http.ResponseWriter, r *http.Request) {
	status := h.gateway.GetSensorStatus()
	writeJSON(w, http.StatusOK, status)
}

// RefreshSensorStatus triggers a refresh of sensor data from the KLF-200
func (h *Handlers) RefreshSensorStatus(w http.ResponseWriter, r *http.Request) {
	if err := h.gateway.RefreshSensorStatus(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to refresh sensor status", err.Error())
		return
	}

	status := h.gateway.GetSensorStatus()
	writeJSON(w, http.StatusOK, status)
}

// Loxone-friendly sensor endpoints (return simple 0/1 values)

// LoxoneSensorStatus returns all sensor values in Loxone-friendly format
func (h *Handlers) LoxoneSensorStatus(w http.ResponseWriter, r *http.Request) {
	status := h.gateway.GetSensorStatus()
	rain := 0
	wind := 0
	if status.RainDetected {
		rain = 1
	}
	if status.WindDetected {
		wind = 1
	}
	// Format: rain;wind (e.g., "0;0" or "1;0")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.Itoa(rain) + ";" + strconv.Itoa(wind)))
}

// LoxoneRainStatus returns just the rain sensor value (0 or 1)
func (h *Handlers) LoxoneRainStatus(w http.ResponseWriter, r *http.Request) {
	status := h.gateway.GetSensorStatus()
	if status.RainDetected {
		w.Write([]byte("1"))
	} else {
		w.Write([]byte("0"))
	}
}

// LoxoneWindStatus returns just the wind sensor value (0 or 1)
func (h *Handlers) LoxoneWindStatus(w http.ResponseWriter, r *http.Request) {
	status := h.gateway.GetSensorStatus()
	if status.WindDetected {
		w.Write([]byte("1"))
	} else {
		w.Write([]byte("0"))
	}
}

// Configuration endpoints

// ConfigResponse is the JSON structure for config API
type ConfigResponse struct {
	KLF200  ConfigKLF200        `json:"klf200"`
	Server  ConfigServer        `json:"server"`
	Loxone  ConfigLoxone        `json:"loxone"`
	Logging ConfigLogging       `json:"logging"`
}

type ConfigLoxone struct {
	UDPFeedback config.UDPFeedbackConfig `json:"udp_feedback"`
	Mappings    []config.NodeMapping     `json:"mappings"`
}

type ConfigKLF200 struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Password          string `json:"password"`
	ReconnectInterval string `json:"reconnect_interval"`
	RefreshInterval   string `json:"refresh_interval"`
}

type ConfigServer struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	APIToken string `json:"api_token"`
}

type ConfigLogging struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// GetConfig returns the current configuration
func (h *Handlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	if h.configMgr == nil {
		writeError(w, http.StatusServiceUnavailable, "Configuration manager not available", "")
		return
	}

	cfg := h.configMgr.GetConfig()
	resp := ConfigResponse{
		KLF200: ConfigKLF200{
			Host:              cfg.KLF200.Host,
			Port:              cfg.KLF200.Port,
			Password:          cfg.KLF200.Password,
			ReconnectInterval: cfg.KLF200.ReconnectInterval.String(),
			RefreshInterval:   cfg.KLF200.RefreshInterval.String(),
		},
		Server: ConfigServer{
			Host:     cfg.Server.Host,
			Port:     cfg.Server.Port,
			APIToken: cfg.Server.APIToken,
		},
		Loxone: ConfigLoxone{
			UDPFeedback: cfg.Loxone.UDPFeedback,
			Mappings:    cfg.Loxone.Mappings,
		},
		Logging: ConfigLogging{
			Level:  cfg.Logging.Level,
			Format: cfg.Logging.Format,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

// UpdateConfig updates the configuration
func (h *Handlers) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if h.configMgr == nil {
		writeError(w, http.StatusServiceUnavailable, "Configuration manager not available", "")
		return
	}

	var req ConfigResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get current config and update it
	cfg := h.configMgr.GetConfig()

	// Update KLF200 settings
	if req.KLF200.Host != "" {
		cfg.KLF200.Host = req.KLF200.Host
	}
	if req.KLF200.Port > 0 {
		cfg.KLF200.Port = req.KLF200.Port
	}
	if req.KLF200.Password != "" {
		cfg.KLF200.Password = req.KLF200.Password
	}

	// Update server settings (API token can be empty to disable auth)
	cfg.Server.APIToken = req.Server.APIToken

	// Update logging settings
	if req.Logging.Level != "" {
		cfg.Logging.Level = req.Logging.Level
	}

	// Save config
	if err := h.configMgr.UpdateConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save configuration", err.Error())
		return
	}

	h.logger.Info().Msg("Configuration updated")

	// Return updated config
	h.GetConfig(w, r)
}

// Reconnect triggers a reconnection to the KLF-200
func (h *Handlers) Reconnect(w http.ResponseWriter, r *http.Request) {
	if err := h.gateway.Reconnect(r.Context()); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Reconnection initiated",
	})
}

// Helper functions

func parseNodeID(r *http.Request) (uint8, error) {
	nodeIDStr := chi.URLParam(r, "nodeID")
	nodeID, err := strconv.ParseUint(nodeIDStr, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(nodeID), nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message, details string) {
	writeJSON(w, status, ErrorResponse{
		Error:   message,
		Code:    status,
		Details: details,
	})
}

// generateUUID generates a random UUID v4
func generateUUID() string {
	var uuid [16]byte
	rand.Read(uuid[:])
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// Mapping CRUD endpoints

// ListMappings returns all node-to-Loxone mappings
func (h *Handlers) ListMappings(w http.ResponseWriter, r *http.Request) {
	mappings := h.gateway.GetMappingManager().GetAll()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"mappings": mappings,
		"count":    len(mappings),
	})
}

// CreateMapping creates a new node mapping
func (h *Handlers) CreateMapping(w http.ResponseWriter, r *http.Request) {
	var mapping config.NodeMapping
	if err := json.NewDecoder(r.Body).Decode(&mapping); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if mapping.LoxoneID == "" {
		writeError(w, http.StatusBadRequest, "loxone_id is required", "")
		return
	}

	mapping.ID = generateUUID()
	mapping.Enabled = true

	cfg := h.configMgr.GetConfig()
	cfg.Loxone.Mappings = append(cfg.Loxone.Mappings, mapping)
	if err := h.configMgr.UpdateConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save mapping", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, mapping)
}

// UpdateMapping updates an existing mapping
func (h *Handlers) UpdateMapping(w http.ResponseWriter, r *http.Request) {
	mappingID := chi.URLParam(r, "mappingID")

	var update config.NodeMapping
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	cfg := h.configMgr.GetConfig()
	found := false
	for i, m := range cfg.Loxone.Mappings {
		if m.ID == mappingID {
			update.ID = mappingID
			cfg.Loxone.Mappings[i] = update
			found = true
			break
		}
	}

	if !found {
		writeError(w, http.StatusNotFound, "Mapping not found", "")
		return
	}

	if err := h.configMgr.UpdateConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save mapping", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, update)
}

// DeleteMapping deletes a mapping
func (h *Handlers) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	mappingID := chi.URLParam(r, "mappingID")

	cfg := h.configMgr.GetConfig()
	newMappings := make([]config.NodeMapping, 0, len(cfg.Loxone.Mappings))
	found := false

	for _, m := range cfg.Loxone.Mappings {
		if m.ID == mappingID {
			found = true
			continue
		}
		newMappings = append(newMappings, m)
	}

	if !found {
		writeError(w, http.StatusNotFound, "Mapping not found", "")
		return
	}

	cfg.Loxone.Mappings = newMappings
	if err := h.configMgr.UpdateConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete mapping", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Loxone config endpoints

// GetLoxoneConfig returns the Loxone-specific configuration
func (h *Handlers) GetLoxoneConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.configMgr.GetConfig()
	writeJSON(w, http.StatusOK, ConfigLoxone{
		UDPFeedback: cfg.Loxone.UDPFeedback,
		Mappings:    cfg.Loxone.Mappings,
	})
}

// UpdateLoxoneUDPConfig updates UDP feedback settings
func (h *Handlers) UpdateLoxoneUDPConfig(w http.ResponseWriter, r *http.Request) {
	var req config.UDPFeedbackConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	cfg := h.configMgr.GetConfig()
	cfg.Loxone.UDPFeedback = req

	if err := h.configMgr.UpdateConfig(cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save config", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"udp_feedback": req,
	})
}

// TestUDP sends a test UDP message
func (h *Handlers) TestUDP(w http.ResponseWriter, r *http.Request) {
	udpSender := h.gateway.GetUDPSender()
	if udpSender == nil || !udpSender.IsEnabled() {
		writeError(w, http.StatusBadRequest, "UDP feedback is not enabled", "")
		return
	}

	udpSender.Send("test", "ping", 1)
	writeJSON(w, http.StatusOK, map[string]string{"status": "test sent"})
}
