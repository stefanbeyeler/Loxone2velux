package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/stefanbeyeler/loxone2velux/internal/config"
	"github.com/stefanbeyeler/loxone2velux/internal/klf200"
)

// Service is the main gateway service
type Service struct {
	cfg    *config.KLF200Config
	client *klf200.Client
	nodes  *klf200.NodeManager
	logger zerolog.Logger

	mu       sync.RWMutex
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewService creates a new gateway service
func NewService(cfg *config.KLF200Config, logger zerolog.Logger) *Service {
	clientCfg := klf200.ClientConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Password: cfg.Password,
		Logger:   logger.With().Str("component", "klf200-client").Logger(),
	}

	return &Service{
		cfg:      cfg,
		client:   klf200.NewClient(clientCfg),
		nodes:    klf200.NewNodeManager(),
		logger:   logger.With().Str("component", "gateway").Logger(),
		stopChan: make(chan struct{}),
	}
}

// Start starts the gateway service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info().
		Str("host", s.cfg.Host).
		Int("port", s.cfg.Port).
		Msg("Starting gateway service")

	// Set callbacks
	s.client.SetNodeUpdateCallback(s.handleNodeUpdate)
	s.client.SetDisconnectCallback(s.handleDisconnect)

	// Try initial connection (non-blocking on failure)
	var connectErr error
	if err := s.connect(ctx); err != nil {
		connectErr = err
		s.logger.Warn().Err(err).Msg("Initial connection failed, will retry in background")
	}

	// Start refresh loop
	s.wg.Add(1)
	go s.refreshLoop()

	// Start reconnect loop
	s.wg.Add(1)
	go s.reconnectLoop()

	return connectErr
}

// connect establishes connection to KLF-200
func (s *Service) connect(ctx context.Context) error {
	if err := s.client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if err := s.client.Authenticate(ctx); err != nil {
		s.client.Disconnect()
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Get initial nodes
	if err := s.refreshNodes(ctx); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to get initial nodes")
	}

	return nil
}

// refreshNodes retrieves all nodes from KLF-200
func (s *Service) refreshNodes(ctx context.Context) error {
	nodes, err := s.client.GetAllNodes(ctx)
	if err != nil {
		return err
	}

	s.nodes.SetNodes(nodes)
	s.logger.Info().Int("count", len(nodes)).Msg("Refreshed nodes")

	return nil
}

// refreshLoop periodically refreshes node information
func (s *Service) refreshLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cfg.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if s.client.IsAuthenticated() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := s.refreshNodes(ctx); err != nil {
					s.logger.Warn().Err(err).Msg("Failed to refresh nodes")
				}
				cancel()
			}
		}
	}
}

// reconnectLoop handles reconnection
func (s *Service) reconnectLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.cfg.ReconnectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if !s.client.IsConnected() {
				s.logger.Info().Msg("Attempting to reconnect")
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := s.connect(ctx); err != nil {
					s.logger.Warn().Err(err).Msg("Reconnect failed")
				} else {
					s.logger.Info().Msg("Reconnected successfully")
				}
				cancel()
			}
		}
	}
}

// handleNodeUpdate handles node position updates
func (s *Service) handleNodeUpdate(node *klf200.Node) {
	s.nodes.UpdateNode(node)
	s.logger.Debug().
		Uint8("id", node.ID).
		Float64("position", node.PositionPercent).
		Msg("Node position updated")
}

// handleDisconnect handles disconnection
func (s *Service) handleDisconnect(err error) {
	if err != nil {
		s.logger.Warn().Err(err).Msg("Disconnected from KLF-200")
	} else {
		s.logger.Info().Msg("Disconnected from KLF-200")
	}
}

// Stop stops the gateway service
func (s *Service) Stop() error {
	s.logger.Info().Msg("Stopping gateway service")

	close(s.stopChan)
	s.wg.Wait()

	return s.client.Disconnect()
}

// IsConnected returns true if connected to KLF-200
func (s *Service) IsConnected() bool {
	return s.client.IsAuthenticated()
}

// GetNodes returns all nodes
func (s *Service) GetNodes() []*klf200.Node {
	return s.nodes.GetAllNodes()
}

// GetNode returns a node by ID
func (s *Service) GetNode(id uint8) (*klf200.Node, bool) {
	return s.nodes.GetNode(id)
}

// GetNodeCount returns the number of nodes
func (s *Service) GetNodeCount() int {
	return s.nodes.NodeCount()
}

// SetPosition sets the position of a node
func (s *Service) SetPosition(ctx context.Context, nodeID uint8, percent float64) error {
	if !s.client.IsAuthenticated() {
		return fmt.Errorf("not connected to KLF-200")
	}

	return s.client.SetPosition(ctx, nodeID, percent)
}

// Open fully opens a node
func (s *Service) Open(ctx context.Context, nodeID uint8) error {
	if !s.client.IsAuthenticated() {
		return fmt.Errorf("not connected to KLF-200")
	}

	return s.client.Open(ctx, nodeID)
}

// Close fully closes a node
func (s *Service) Close(ctx context.Context, nodeID uint8) error {
	if !s.client.IsAuthenticated() {
		return fmt.Errorf("not connected to KLF-200")
	}

	return s.client.Close(ctx, nodeID)
}

// StopNode stops a node's movement
func (s *Service) StopNode(ctx context.Context, nodeID uint8) error {
	if !s.client.IsAuthenticated() {
		return fmt.Errorf("not connected to KLF-200")
	}

	return s.client.Stop(ctx, nodeID)
}

// GetSensorStatus returns the current sensor status
func (s *Service) GetSensorStatus() klf200.SensorStatus {
	return s.client.GetSensorStatus()
}

// RefreshSensorStatus queries the KLF-200 for current sensor/limitation status
func (s *Service) RefreshSensorStatus(ctx context.Context) error {
	if !s.client.IsAuthenticated() {
		return fmt.Errorf("not connected to KLF-200")
	}

	// Get all node IDs
	nodes := s.nodes.GetAllNodes()
	nodeIDs := make([]uint8, len(nodes))
	for i, n := range nodes {
		nodeIDs[i] = n.ID
	}

	return s.client.RefreshSensorStatus(ctx, nodeIDs)
}

// Reconnect disconnects and reconnects to the KLF-200
func (s *Service) Reconnect(ctx context.Context) error {
	s.logger.Info().Msg("Manual reconnect requested")

	// Disconnect if connected
	if s.client.IsConnected() {
		s.client.Disconnect()
	}

	// Reconnect
	if err := s.connect(ctx); err != nil {
		return fmt.Errorf("reconnect failed: %w", err)
	}

	s.logger.Info().Msg("Reconnected successfully")
	return nil
}

// UpdateConfig updates the KLF-200 configuration (requires reconnect)
func (s *Service) UpdateConfig(cfg *config.KLF200Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg = cfg

	// Update client config
	clientCfg := klf200.ClientConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Password: cfg.Password,
		Logger:   s.logger.With().Str("component", "klf200-client").Logger(),
	}
	s.client.UpdateConfig(clientCfg)
}
