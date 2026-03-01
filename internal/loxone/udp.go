package loxone

import (
	"fmt"
	"net"
	"sync"

	"github.com/rs/zerolog"
	"github.com/stefanbeyeler/loxone2velux/internal/config"
)

// UDPSender sends status updates to a Loxone Miniserver via UDP
type UDPSender struct {
	conn    *net.UDPConn
	mu      sync.Mutex
	enabled bool
	logger  zerolog.Logger
}

// NewUDPSender creates a new UDP sender (initially disabled)
func NewUDPSender(logger zerolog.Logger) *UDPSender {
	return &UDPSender{
		logger: logger.With().Str("component", "udp-sender").Logger(),
	}
}

// Configure sets up or reconfigures the UDP sender based on current config.
func (s *UDPSender) Configure(cfg config.UDPFeedbackConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close existing connection if any
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	s.enabled = cfg.Enabled
	if !cfg.Enabled || cfg.IP == "" || cfg.Port == 0 {
		s.enabled = false
		s.logger.Info().Msg("UDP feedback disabled")
		return nil
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", cfg.IP, cfg.Port))
	if err != nil {
		return fmt.Errorf("invalid UDP address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %w", err)
	}

	s.conn = conn
	s.logger.Info().Str("addr", addr.String()).Msg("UDP feedback configured")
	return nil
}

// IsEnabled returns whether UDP feedback is currently enabled
func (s *UDPSender) IsEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enabled && s.conn != nil
}

// Send sends a single property update to the Loxone Miniserver.
// Format: "<loxone_id>/<property>:<value>"
// Fire-and-forget: errors are logged but not returned.
func (s *UDPSender) Send(loxoneID, property string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled || s.conn == nil {
		return
	}

	msg := fmt.Sprintf("%s/%s:%v", loxoneID, property, value)

	_, err := s.conn.Write([]byte(msg))
	if err != nil {
		s.logger.Warn().Err(err).Str("msg", msg).Msg("Failed to send UDP feedback")
	} else {
		s.logger.Debug().Str("msg", msg).Msg("UDP feedback sent")
	}
}

// Close closes the UDP connection
func (s *UDPSender) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.enabled = false
}
