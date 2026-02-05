package klf200

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

// Client represents a connection to a KLF-200 gateway
type Client struct {
	host     string
	port     int
	password string
	logger   zerolog.Logger

	conn          *tls.Conn
	connMu        sync.Mutex
	connected     atomic.Bool
	authenticated atomic.Bool

	sessionID atomic.Uint32

	// Callbacks
	onNodeUpdate func(*Node)
	onDisconnect func(error)

	// Read buffer for SLIP framing
	readBuf bytes.Buffer
	readMu  sync.Mutex

	// Response channels
	responseChan chan *Frame
	stopChan     chan struct{}
	wg           sync.WaitGroup

	// Sensor status
	sensorStatus   SensorStatus
	sensorStatusMu sync.RWMutex
}

// ClientConfig holds configuration for the KLF-200 client
type ClientConfig struct {
	Host     string
	Port     int
	Password string
	Logger   zerolog.Logger
}

// NewClient creates a new KLF-200 client
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		host:         cfg.Host,
		port:         cfg.Port,
		password:     cfg.Password,
		logger:       cfg.Logger,
		responseChan: make(chan *Frame, 100),
		stopChan:     make(chan struct{}),
	}
}

// UpdateConfig updates the client configuration (should be disconnected first)
func (c *Client) UpdateConfig(cfg ClientConfig) {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	c.host = cfg.Host
	c.port = cfg.Port
	c.password = cfg.Password
	if cfg.Logger.GetLevel() != zerolog.Disabled {
		c.logger = cfg.Logger
	}
}

// SetNodeUpdateCallback sets the callback for node updates
func (c *Client) SetNodeUpdateCallback(cb func(*Node)) {
	c.onNodeUpdate = cb
}

// SetDisconnectCallback sets the callback for disconnection
func (c *Client) SetDisconnectCallback(cb func(error)) {
	c.onDisconnect = cb
}

// Connect establishes connection to the KLF-200
func (c *Client) Connect(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.connected.Load() {
		return nil
	}

	// Reset stopChan if it was closed
	select {
	case <-c.stopChan:
		c.stopChan = make(chan struct{})
	default:
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	c.logger.Info().Str("addr", addr).Msg("Connecting to KLF-200")

	// Configure TLS with Velux CA certificate
	// KLF-200 uses a self-signed certificate without proper CN/SAN
	// We verify using the CA certificate but skip hostname verification
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(VeluxCA))

	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: true, // Skip hostname verification (no CN/SAN in cert)
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS12,
	}

	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	// Connect with context
	netConn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to KLF-200: %w", err)
	}

	// Wrap with TLS
	conn := tls.Client(netConn, tlsConfig)

	// Perform TLS handshake with timeout
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	} else {
		conn.SetDeadline(time.Now().Add(10 * time.Second))
	}

	if err := conn.Handshake(); err != nil {
		netConn.Close()
		return fmt.Errorf("TLS handshake failed: %w", err)
	}

	// Clear deadline
	conn.SetDeadline(time.Time{})

	c.conn = conn
	c.connected.Store(true)

	// Start reader goroutine
	c.wg.Add(1)
	go c.readLoop()

	c.logger.Info().Msg("Connected to KLF-200")

	return nil
}

// Authenticate authenticates with the KLF-200
func (c *Client) Authenticate(ctx context.Context) error {
	if !c.connected.Load() {
		return fmt.Errorf("not connected")
	}

	c.logger.Debug().Msg("Authenticating with KLF-200")

	// Send password
	frame := BuildPasswordEnterRequest(c.password)
	c.logger.Debug().
		Hex("frame", frame).
		Int("len", len(frame)).
		Str("password", c.password).
		Msg("Sending password frame")

	if err := c.sendRaw(frame); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}
	c.logger.Debug().Msg("Password frame sent")

	// Wait for response
	resp, err := c.waitForResponse(ctx, GW_PASSWORD_ENTER_CFM, 10*time.Second)
	if err != nil {
		return fmt.Errorf("authentication timeout: %w", err)
	}

	ok, err := ParsePasswordConfirm(resp.Data)
	if err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	if !ok {
		return fmt.Errorf("authentication failed: invalid password")
	}

	c.authenticated.Store(true)
	c.logger.Info().Msg("Authenticated with KLF-200")

	// Enable house status monitor
	if err := c.enableHouseStatusMonitor(ctx); err != nil {
		c.logger.Warn().Err(err).Msg("Failed to enable house status monitor")
	}

	return nil
}

// enableHouseStatusMonitor enables notifications for position changes
func (c *Client) enableHouseStatusMonitor(ctx context.Context) error {
	frame := BuildHouseStatusMonitorEnableRequest()
	if err := c.sendRaw(frame); err != nil {
		return err
	}

	_, err := c.waitForResponse(ctx, GW_HOUSE_STATUS_MONITOR_ENABLE_CFM, 5*time.Second)
	return err
}

// GetAllNodes retrieves all nodes from the KLF-200
func (c *Client) GetAllNodes(ctx context.Context) ([]*Node, error) {
	if !c.authenticated.Load() {
		return nil, fmt.Errorf("not authenticated")
	}

	c.logger.Debug().Msg("Getting all nodes")

	frame := BuildGetAllNodesRequest()
	if err := c.sendRaw(frame); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for confirmation
	_, err := c.waitForResponse(ctx, GW_GET_ALL_NODES_INFORMATION_CFM, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get confirmation: %w", err)
	}

	// Collect node notifications
	var nodes []*Node
	for {
		resp, err := c.waitForResponse(ctx, 0, 5*time.Second) // Accept any response
		if err != nil {
			return nodes, nil // Timeout means no more nodes
		}

		switch resp.Command {
		case GW_GET_ALL_NODES_INFORMATION_NTF:
			node, err := ParseNodeInformation(resp.Data)
			if err != nil {
				c.logger.Warn().Err(err).Msg("Failed to parse node info")
				continue
			}
			node.LastUpdate = time.Now()
			nodes = append(nodes, node)
			c.logger.Debug().Uint8("id", node.ID).Str("name", node.Name).Msg("Found node")

		case GW_GET_ALL_NODES_INFORMATION_FINISHED_NTF:
			c.logger.Debug().Int("count", len(nodes)).Msg("Finished getting nodes")
			return nodes, nil
		}
	}
}

// SetPosition sets the position of a node (0-100%)
func (c *Client) SetPosition(ctx context.Context, nodeID uint8, percent float64) error {
	if !c.authenticated.Load() {
		return fmt.Errorf("not authenticated")
	}

	position := PercentToPosition(percent)
	sessionID := uint16(c.sessionID.Add(1))

	c.logger.Debug().
		Uint8("node", nodeID).
		Float64("percent", percent).
		Uint16("position", position).
		Msg("Setting position")

	frame := BuildCommandSendRequest(
		sessionID,
		1, // User originated
		PriorityUserLevel2,
		[]uint8{nodeID},
		position,
		nil,
	)

	c.logger.Debug().
		Hex("frame", frame).
		Int("len", len(frame)).
		Msg("Sending command frame")

	if err := c.sendRaw(frame); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	// Wait for confirmation (GW_COMMAND_SEND_CFM) or error (GW_ERROR_NTF)
	// Skip async notifications like GW_NODE_STATE_POSITION_CHANGED_NTF
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("command timeout")
		case resp := <-c.responseChan:
			switch resp.Command {
			case GW_ERROR_NTF:
				errorCode := uint8(0)
				if len(resp.Data) > 0 {
					errorCode = resp.Data[0]
				}
				c.logger.Error().
					Uint8("errorCode", errorCode).
					Msg("KLF-200 returned error")
				return fmt.Errorf("KLF-200 error: code %d", errorCode)

			case GW_COMMAND_SEND_CFM:
				sessionID, status, err := ParseCommandSendConfirm(resp.Data)
				c.logger.Debug().
					Uint16("sessionID", sessionID).
					Uint8("status", uint8(status)).
					Hex("data", resp.Data).
					Msg("Received command confirmation")
				if err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				// Status 0 = accepted, Status 1 = accepted but busy (command still executes)
				if status > 1 {
					return fmt.Errorf("command failed with status: %d", status)
				}
				if status == 1 {
					c.logger.Debug().Msg("Command accepted (node busy)")
				} else {
					c.logger.Debug().Msg("Command confirmed")
				}
				return nil

			default:
				// Handle async notifications (position changes, etc.)
				c.handleAsyncFrame(resp)
			}
		}
	}
}

// Open fully opens a node (position 0%)
func (c *Client) Open(ctx context.Context, nodeID uint8) error {
	return c.SetPosition(ctx, nodeID, 0)
}

// Close fully closes a node (position 100%)
func (c *Client) Close(ctx context.Context, nodeID uint8) error {
	return c.SetPosition(ctx, nodeID, 100)
}

// Stop stops a node's movement
func (c *Client) Stop(ctx context.Context, nodeID uint8) error {
	if !c.authenticated.Load() {
		return fmt.Errorf("not authenticated")
	}

	sessionID := uint16(c.sessionID.Add(1))

	c.logger.Debug().Uint8("node", nodeID).Msg("Stopping node")

	// Use current position to stop
	frame := BuildCommandSendRequest(
		sessionID,
		1,
		PriorityUserLevel2,
		[]uint8{nodeID},
		PositionCurrent, // Keep current = stop
		nil,
	)

	if err := c.sendRaw(frame); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	return nil
}

// GetLimitationStatus queries the limitation status for nodes (sensor data)
func (c *Client) GetLimitationStatus(ctx context.Context, nodeIDs []uint8) ([]*LimitationStatus, error) {
	if !c.authenticated.Load() {
		return nil, fmt.Errorf("not authenticated")
	}

	sessionID := uint16(c.sessionID.Add(1))

	c.logger.Debug().Interface("nodes", nodeIDs).Msg("Getting limitation status")

	// Request both min and max limitations
	frame := BuildGetLimitationStatusRequest(sessionID, nodeIDs, 0) // 0 = min limitation
	if err := c.sendRaw(frame); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for confirmation
	_, err := c.waitForResponse(ctx, GW_GET_LIMITATION_STATUS_CFM, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get confirmation: %w", err)
	}

	// Collect limitation notifications
	var limitations []*LimitationStatus
	for {
		resp, err := c.waitForResponse(ctx, 0, 2*time.Second)
		if err != nil {
			// Timeout means no more notifications
			break
		}

		if resp.Command == GW_LIMITATION_STATUS_NTF {
			status, err := ParseLimitationStatusNotification(resp.Data)
			if err != nil {
				c.logger.Warn().Err(err).Msg("Failed to parse limitation status")
				continue
			}

			c.logger.Debug().
				Uint8("nodeID", status.NodeID).
				Uint8("origin", uint8(status.LimitationOrigin)).
				Str("originStr", status.LimitationOrigin.String()).
				Uint16("minValue", status.MinValue).
				Uint16("maxValue", status.MaxValue).
				Msg("Limitation status received")

			limitations = append(limitations, status)

			// Update sensor status based on limitation origin
			c.updateSensorStatus(status)
		}
	}

	return limitations, nil
}

// updateSensorStatus updates the internal sensor status based on limitation data
func (c *Client) updateSensorStatus(status *LimitationStatus) {
	c.sensorStatusMu.Lock()
	defer c.sensorStatusMu.Unlock()

	c.sensorStatus.LastUpdate = time.Now()

	switch status.LimitationOrigin {
	case LimitationTypeRain:
		c.sensorStatus.RainDetected = true
	case LimitationTypeWind:
		c.sensorStatus.WindDetected = true
	}

	// If no limitation, sensors are clear
	if status.LimitationOrigin == LimitationTypeNone {
		c.sensorStatus.RainDetected = false
		c.sensorStatus.WindDetected = false
	}
}

// GetSensorStatus returns the current sensor status
func (c *Client) GetSensorStatus() SensorStatus {
	c.sensorStatusMu.RLock()
	defer c.sensorStatusMu.RUnlock()
	return c.sensorStatus
}

// RefreshSensorStatus queries all nodes for limitation status to update sensor readings
// Returns nil even on timeout - the sensor status will keep its last known values
func (c *Client) RefreshSensorStatus(ctx context.Context, nodeIDs []uint8) error {
	_, err := c.GetLimitationStatus(ctx, nodeIDs)
	if err != nil {
		// Log the error but don't fail - limitation status may not be supported
		// by all KLF-200 firmware versions or may timeout
		c.logger.Warn().Err(err).Msg("Sensor status refresh failed, keeping previous values")
		// Update last update time to indicate we tried
		c.sensorStatusMu.Lock()
		c.sensorStatus.LastUpdate = time.Now()
		c.sensorStatusMu.Unlock()
	}
	return nil
}

// sendRaw sends raw bytes to the KLF-200
func (c *Client) sendRaw(data []byte) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	_, err := c.conn.Write(data)
	return err
}

// waitForResponse waits for a specific response or any response if cmd is 0
func (c *Client) waitForResponse(ctx context.Context, cmd CommandID, timeout time.Duration) (*Frame, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case frame := <-c.responseChan:
			if cmd == 0 || frame.Command == cmd {
				return frame, nil
			}
			// Put back other responses (or handle them)
			c.handleAsyncFrame(frame)
		}
	}
}

// handleAsyncFrame handles frames that were not expected
func (c *Client) handleAsyncFrame(frame *Frame) {
	switch frame.Command {
	case GW_NODE_STATE_POSITION_CHANGED_NTF:
		nodeID, state, position, target, err := ParseNodeStatePositionChangedFull(frame.Data)
		if err != nil {
			c.logger.Warn().Err(err).Msg("Failed to parse position change")
			return
		}
		c.logger.Debug().
			Uint8("nodeID", nodeID).
			Uint8("state", uint8(state)).
			Uint16("position", position).
			Uint16("target", target).
			Float64("percent", PositionToPercent(position)).
			Msg("Node position changed notification")
		if c.onNodeUpdate != nil {
			c.onNodeUpdate(&Node{
				ID:              nodeID,
				State:           state,
				StateStr:        state.String(),
				CurrentPosition: position,
				PositionPercent: PositionToPercent(position),
				TargetPosition:  target,
				TargetPercent:   PositionToPercent(target),
				LastUpdate:      time.Now(),
			})
		}

	case GW_COMMAND_RUN_STATUS_NTF:
		sessionID, nodeID, runStatus, statusReply, err := ParseRunStatusNotification(frame.Data)
		if err != nil {
			c.logger.Warn().Err(err).Msg("Failed to parse run status")
			return
		}
		c.logger.Debug().
			Uint16("sessionID", sessionID).
			Uint8("nodeID", nodeID).
			Uint8("runStatus", uint8(runStatus)).
			Uint8("statusReply", uint8(statusReply)).
			Msg("Command run status notification")

		// Check for sensor-related limitations
		c.sensorStatusMu.Lock()
		c.sensorStatus.LastUpdate = time.Now()
		switch statusReply {
		case StatusReplyLimitationByRain:
			c.sensorStatus.RainDetected = true
			c.logger.Info().Msg("Rain sensor triggered - rain detected")
		case StatusReplyLimitationByWind:
			c.sensorStatus.WindDetected = true
			c.logger.Info().Msg("Wind sensor triggered - wind detected")
		}
		c.sensorStatusMu.Unlock()

		// Update state based on run status
		var state NodeState
		switch runStatus {
		case RunStatusExecutionCompleted:
			state = NodeStateDone
		case RunStatusExecutionFailed:
			state = NodeStateErrorWhileExecution
		case RunStatusExecutionActive:
			state = NodeStateExecuting
		}
		if c.onNodeUpdate != nil {
			c.onNodeUpdate(&Node{
				ID:         nodeID,
				State:      state,
				StateStr:   state.String(),
				LastUpdate: time.Now(),
			})
		}

	case GW_LIMITATION_STATUS_NTF:
		status, err := ParseLimitationStatusNotification(frame.Data)
		if err != nil {
			c.logger.Warn().Err(err).Msg("Failed to parse limitation status")
			return
		}
		c.logger.Debug().
			Uint8("nodeID", status.NodeID).
			Str("origin", status.LimitationOrigin.String()).
			Msg("Limitation status notification")
		c.updateSensorStatus(status)
	}
}

// isAsyncNotification returns true if the frame is an async notification that should be handled immediately
func (c *Client) isAsyncNotification(cmd CommandID) bool {
	switch cmd {
	case GW_NODE_STATE_POSITION_CHANGED_NTF, GW_COMMAND_RUN_STATUS_NTF, GW_LIMITATION_STATUS_NTF:
		return true
	default:
		return false
	}
}

// readLoop continuously reads from the TLS connection and extracts SLIP frames
func (c *Client) readLoop() {
	defer c.wg.Done()

	buf := make([]byte, 1024)
	var frameBuf bytes.Buffer
	inFrame := false

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		n, err := c.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				c.logger.Info().Msg("Connection closed by KLF-200")
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Read timeout is normal, continue
				continue
			} else {
				c.logger.Error().Err(err).Msg("Read error")
			}
			c.handleDisconnect(err)
			return
		}

		// Process received bytes and extract SLIP frames
		for i := 0; i < n; i++ {
			b := buf[i]

			if b == SlipEnd {
				if inFrame && frameBuf.Len() > 0 {
					// End of frame - process it
					frameData := make([]byte, frameBuf.Len()+2)
					frameData[0] = SlipEnd
					copy(frameData[1:], frameBuf.Bytes())
					frameData[len(frameData)-1] = SlipEnd

					frame, err := DecodeFrame(frameData)
					if err != nil {
						c.logger.Warn().Err(err).Msg("Failed to decode frame")
					} else {
						c.logger.Debug().
							Uint16("cmd", uint16(frame.Command)).
							Int("dataLen", len(frame.Data)).
							Msg("Received frame")

						// Handle async notifications directly, don't queue them
						if c.isAsyncNotification(frame.Command) {
							c.handleAsyncFrame(frame)
						} else {
							select {
							case c.responseChan <- frame:
							default:
								c.logger.Warn().Msg("Response channel full, dropping frame")
							}
						}
					}
					frameBuf.Reset()
				}
				inFrame = true
			} else if inFrame {
				frameBuf.WriteByte(b)
			}
		}
	}
}

// handleDisconnect handles disconnection
func (c *Client) handleDisconnect(err error) {
	c.connected.Store(false)
	c.authenticated.Store(false)

	if c.onDisconnect != nil {
		c.onDisconnect(err)
	}
}

// Disconnect closes the connection
func (c *Client) Disconnect() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	// Safe close of stopChan (only close if not already closed)
	select {
	case <-c.stopChan:
		// Already closed
	default:
		close(c.stopChan)
	}

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.connected.Store(false)
		c.authenticated.Store(false)
		return err
	}

	return nil
}

// IsConnected returns true if connected
func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

// IsAuthenticated returns true if authenticated
func (c *Client) IsAuthenticated() bool {
	return c.authenticated.Load()
}

// Wait waits for the client to finish
func (c *Client) Wait() {
	c.wg.Wait()
}
