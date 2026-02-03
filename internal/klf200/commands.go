package klf200

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// SLIP protocol constants
const (
	SlipEnd     byte = 0xC0
	SlipEsc     byte = 0xDB
	SlipEscEnd  byte = 0xDC
	SlipEscEsc  byte = 0xDD
)

var (
	ErrInvalidFrame    = errors.New("invalid frame")
	ErrChecksumMismatch = errors.New("checksum mismatch")
	ErrFrameTooShort   = errors.New("frame too short")
)

// EncodeFrame creates a SLIP-encoded frame from command and data
func EncodeFrame(cmd CommandID, data []byte) []byte {
	buf := new(bytes.Buffer)

	// Protocol length (excluding length byte itself and checksum)
	buf.WriteByte(byte(len(data) + 3))

	// Command ID (big endian)
	binary.Write(buf, binary.BigEndian, uint16(cmd))

	// Data
	buf.Write(data)

	// Calculate checksum (XOR of all bytes)
	frame := buf.Bytes()
	checksum := byte(0)
	for _, b := range frame {
		checksum ^= b
	}
	buf.WriteByte(checksum)

	// SLIP encode
	return slipEncode(buf.Bytes())
}

// slipEncode applies SLIP encoding to data
func slipEncode(data []byte) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(SlipEnd)

	for _, b := range data {
		switch b {
		case SlipEnd:
			buf.WriteByte(SlipEsc)
			buf.WriteByte(SlipEscEnd)
		case SlipEsc:
			buf.WriteByte(SlipEsc)
			buf.WriteByte(SlipEscEsc)
		default:
			buf.WriteByte(b)
		}
	}

	buf.WriteByte(SlipEnd)
	return buf.Bytes()
}

// slipDecode removes SLIP encoding from data
func slipDecode(data []byte) ([]byte, error) {
	if len(data) < 2 {
		return nil, ErrFrameTooShort
	}

	// Remove leading and trailing END markers
	start := 0
	end := len(data)

	for start < end && data[start] == SlipEnd {
		start++
	}
	for end > start && data[end-1] == SlipEnd {
		end--
	}

	if start >= end {
		return nil, ErrInvalidFrame
	}

	data = data[start:end]
	buf := new(bytes.Buffer)

	for i := 0; i < len(data); i++ {
		if data[i] == SlipEsc {
			if i+1 >= len(data) {
				return nil, ErrInvalidFrame
			}
			switch data[i+1] {
			case SlipEscEnd:
				buf.WriteByte(SlipEnd)
			case SlipEscEsc:
				buf.WriteByte(SlipEsc)
			default:
				return nil, ErrInvalidFrame
			}
			i++
		} else {
			buf.WriteByte(data[i])
		}
	}

	return buf.Bytes(), nil
}

// DecodeFrame parses a SLIP-encoded frame
func DecodeFrame(raw []byte) (*Frame, error) {
	data, err := slipDecode(raw)
	if err != nil {
		return nil, err
	}

	if len(data) < 4 {
		return nil, ErrFrameTooShort
	}

	// Verify checksum
	checksum := byte(0)
	for _, b := range data[:len(data)-1] {
		checksum ^= b
	}
	if checksum != data[len(data)-1] {
		return nil, ErrChecksumMismatch
	}

	// Parse frame
	frameLen := data[0]
	if int(frameLen) > len(data)-1 {
		return nil, ErrInvalidFrame
	}

	cmd := CommandID(binary.BigEndian.Uint16(data[1:3]))

	frame := &Frame{
		Command: cmd,
		Data:    data[3 : len(data)-1],
	}

	return frame, nil
}

// BuildPasswordEnterRequest creates a password authentication request
func BuildPasswordEnterRequest(password string) []byte {
	// Password is max 32 bytes, padded with zeros
	data := make([]byte, 32)
	copy(data, []byte(password))
	return EncodeFrame(GW_PASSWORD_ENTER_REQ, data)
}

// BuildGetAllNodesRequest creates a request to get all node information
func BuildGetAllNodesRequest() []byte {
	return EncodeFrame(GW_GET_ALL_NODES_INFORMATION_REQ, nil)
}

// BuildCommandSendRequest creates a command to control a node
func BuildCommandSendRequest(sessionID uint16, commandOriginator uint8, priorityLevel Priority,
	nodeIDs []uint8, mainParameter uint16, functionalParameters []uint16) []byte {

	buf := new(bytes.Buffer)

	// Session ID (2 bytes)
	binary.Write(buf, binary.BigEndian, sessionID)

	// Command originator (1 byte) - typically 1 for user
	buf.WriteByte(commandOriginator)

	// Priority level (1 byte)
	buf.WriteByte(byte(priorityLevel))

	// Parameter active flags (1 byte) - bit 0 = main parameter
	paramFlags := byte(0x01)
	buf.WriteByte(paramFlags)

	// Functional parameter 1 active (2 bytes) - not used
	buf.Write([]byte{0x00, 0x00})

	// Functional parameter 2 active (2 bytes) - not used
	buf.Write([]byte{0x00, 0x00})

	// Main parameter (2 bytes)
	binary.Write(buf, binary.BigEndian, mainParameter)

	// Functional parameters (16 x 2 bytes) - set to ignore
	for i := 0; i < 16; i++ {
		if i < len(functionalParameters) {
			binary.Write(buf, binary.BigEndian, functionalParameters[i])
		} else {
			binary.Write(buf, binary.BigEndian, PositionIgnore)
		}
	}

	// Index array count (1 byte)
	buf.WriteByte(byte(len(nodeIDs)))

	// Node IDs (max 20, 1 byte each)
	for _, id := range nodeIDs {
		buf.WriteByte(id)
	}
	// Pad to 20 nodes
	for i := len(nodeIDs); i < 20; i++ {
		buf.WriteByte(0)
	}

	// Priority level lock (2 bytes)
	buf.Write([]byte{0x00, 0x00})

	// Lock priorities (8 bytes)
	buf.Write(make([]byte, 8))

	// Originator (1 byte)
	buf.WriteByte(0x00)

	return EncodeFrame(GW_COMMAND_SEND_REQ, buf.Bytes())
}

// BuildHouseStatusMonitorEnableRequest enables position change notifications
func BuildHouseStatusMonitorEnableRequest() []byte {
	return EncodeFrame(GW_HOUSE_STATUS_MONITOR_ENABLE_REQ, nil)
}

// BuildStatusRequest creates a status request for specific nodes
func BuildStatusRequest(sessionID uint16, nodeIDs []uint8) []byte {
	buf := new(bytes.Buffer)

	// Session ID
	binary.Write(buf, binary.BigEndian, sessionID)

	// Index array count
	buf.WriteByte(byte(len(nodeIDs)))

	// Node IDs (max 20)
	for _, id := range nodeIDs {
		buf.WriteByte(id)
	}
	for i := len(nodeIDs); i < 20; i++ {
		buf.WriteByte(0)
	}

	// Status type (0 = request target position)
	buf.WriteByte(0x00)

	// Functional parameter (0 = main info)
	buf.WriteByte(0x00)

	return EncodeFrame(GW_STATUS_REQUEST_REQ, buf.Bytes())
}

// ParsePasswordConfirm parses password confirmation response
func ParsePasswordConfirm(data []byte) (bool, error) {
	if len(data) < 1 {
		return false, ErrFrameTooShort
	}
	return data[0] == 0, nil
}

// ParseNodeInformation parses node information from notification
func ParseNodeInformation(data []byte) (*Node, error) {
	if len(data) < 127 {
		return nil, fmt.Errorf("node information too short: %d bytes", len(data))
	}

	node := &Node{
		ID: data[0],
	}

	// Order (1 byte) - skip
	// Placement (1 byte) - skip

	// Name (64 bytes, null-terminated UTF-8)
	nameEnd := 3
	for i := 3; i < 67 && data[i] != 0; i++ {
		nameEnd = i + 1
	}
	node.Name = string(data[3:nameEnd])

	// Velocity (1 byte at offset 67)
	node.Velocity = Velocity(data[67])

	// Node type (2 bytes at offset 68)
	node.NodeType = NodeType(binary.BigEndian.Uint16(data[68:70]))
	node.NodeTypeStr = node.NodeType.String()

	// Product group (1 byte) - skip
	// Product type (1 byte) - skip
	// Node variation (1 byte) - skip
	// Power mode (1 byte) - skip
	// Build number (1 byte) - skip
	// Serial number (8 bytes) - skip

	// State (1 byte at offset 82)
	node.State = NodeState(data[82])
	node.StateStr = node.State.String()

	// Current position (2 bytes at offset 83)
	node.CurrentPosition = binary.BigEndian.Uint16(data[83:85])
	node.PositionPercent = PositionToPercent(node.CurrentPosition)

	// Target position (2 bytes at offset 85)
	node.TargetPosition = binary.BigEndian.Uint16(data[85:87])
	node.TargetPercent = PositionToPercent(node.TargetPosition)

	return node, nil
}

// ParseNodeStatePositionChanged parses position change notification
func ParseNodeStatePositionChanged(data []byte) (nodeID uint8, position uint16, err error) {
	if len(data) < 6 {
		return 0, 0, ErrFrameTooShort
	}

	nodeID = data[0]
	// State at offset 1
	position = binary.BigEndian.Uint16(data[2:4])

	return nodeID, position, nil
}

// ParseCommandSendConfirm parses command confirmation
func ParseCommandSendConfirm(data []byte) (sessionID uint16, status ResponseStatus, err error) {
	if len(data) < 3 {
		return 0, 0, ErrFrameTooShort
	}

	sessionID = binary.BigEndian.Uint16(data[0:2])
	status = ResponseStatus(data[2])

	return sessionID, status, nil
}

// ParseRunStatusNotification parses run status notification
func ParseRunStatusNotification(data []byte) (sessionID uint16, nodeID uint8, runStatus RunStatus, statusReply StatusReply, err error) {
	if len(data) < 13 {
		return 0, 0, 0, 0, ErrFrameTooShort
	}

	sessionID = binary.BigEndian.Uint16(data[0:2])
	// Status ID at offset 2
	nodeID = data[3]
	// Parameter ID at offset 4
	runStatus = RunStatus(data[5])
	statusReply = StatusReply(data[6])

	return sessionID, nodeID, runStatus, statusReply, nil
}
