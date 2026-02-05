package klf200

import (
	"fmt"
	"time"
)

// Command IDs for KLF-200 protocol
type CommandID uint16

const (
	// Error notification
	GW_ERROR_NTF CommandID = 0x0000

	// Authentication
	GW_PASSWORD_ENTER_REQ CommandID = 0x3000
	GW_PASSWORD_ENTER_CFM CommandID = 0x3001

	// Single node information
	GW_GET_NODE_INFORMATION_REQ CommandID = 0x0200
	GW_GET_NODE_INFORMATION_CFM CommandID = 0x0201

	// All nodes discovery
	GW_GET_ALL_NODES_INFORMATION_REQ         CommandID = 0x0202
	GW_GET_ALL_NODES_INFORMATION_CFM         CommandID = 0x0203
	GW_GET_ALL_NODES_INFORMATION_NTF         CommandID = 0x0204
	GW_GET_ALL_NODES_INFORMATION_FINISHED_NTF CommandID = 0x0205

	// Node information notification
	GW_GET_NODE_INFORMATION_NTF CommandID = 0x0210

	// Commands
	GW_COMMAND_SEND_REQ CommandID = 0x0300
	GW_COMMAND_SEND_CFM CommandID = 0x0301
	GW_COMMAND_RUN_STATUS_NTF CommandID = 0x0302
	GW_COMMAND_REMAINING_TIME_NTF CommandID = 0x0303

	// Session
	GW_SESSION_FINISHED_NTF CommandID = 0x0304

	// Status
	GW_STATUS_REQUEST_REQ CommandID = 0x0305
	GW_STATUS_REQUEST_CFM CommandID = 0x0306
	GW_STATUS_REQUEST_NTF CommandID = 0x0307

	// State changes
	GW_NODE_STATE_POSITION_CHANGED_NTF CommandID = 0x0211

	// Reboot
	GW_REBOOT_REQ CommandID = 0x0001
	GW_REBOOT_CFM CommandID = 0x0002

	// House status monitor
	GW_HOUSE_STATUS_MONITOR_ENABLE_REQ  CommandID = 0x0240
	GW_HOUSE_STATUS_MONITOR_ENABLE_CFM  CommandID = 0x0241
	GW_HOUSE_STATUS_MONITOR_DISABLE_REQ CommandID = 0x0242
	GW_HOUSE_STATUS_MONITOR_DISABLE_CFM CommandID = 0x0243
)

// NodeType represents the type of Velux device
type NodeType uint16

const (
	NodeTypeInteriorVenetianBlind NodeType = 0x0040
	NodeTypeRollerShutter         NodeType = 0x0080
	NodeTypeAwningBlind           NodeType = 0x0081
	NodeTypeWindowOpener          NodeType = 0x0101
	NodeTypeGarageOpener          NodeType = 0x0102
	NodeTypeLight                 NodeType = 0x0103
	NodeTypeGateLock              NodeType = 0x0104
	NodeTypeWindowLock            NodeType = 0x0105
	NodeTypeVerticalExteriorAwning NodeType = 0x0106
	NodeTypeDualShutter           NodeType = 0x0180
	NodeTypeHeatingControl        NodeType = 0x0200
	NodeTypeOnOffSwitch           NodeType = 0x0300
	NodeTypeHorizontalAwning      NodeType = 0x0340
	NodeTypeExteriorVenetianBlind NodeType = 0x0380
	NodeTypeLouverBlind           NodeType = 0x03C0
	NodeTypeCurtainTrack          NodeType = 0x0400
	NodeTypeVentilationPoint      NodeType = 0x0440
	NodeTypeExteriorHeating       NodeType = 0x0480
	NodeTypeSwingingShutter       NodeType = 0x0500
)

// NodeTypeString returns human-readable name for node type
func (t NodeType) String() string {
	switch t {
	case NodeTypeInteriorVenetianBlind:
		return "Interior Venetian Blind"
	case NodeTypeRollerShutter:
		return "Roller Shutter"
	case NodeTypeAwningBlind:
		return "Awning Blind"
	case NodeTypeWindowOpener:
		return "Window Opener"
	case NodeTypeGarageOpener:
		return "Garage Opener"
	case NodeTypeLight:
		return "Light"
	case NodeTypeGateLock:
		return "Gate Lock"
	case NodeTypeWindowLock:
		return "Window Lock"
	case NodeTypeVerticalExteriorAwning:
		return "Vertical Exterior Awning"
	case NodeTypeDualShutter:
		return "Dual Shutter"
	case NodeTypeHeatingControl:
		return "Heating Control"
	case NodeTypeOnOffSwitch:
		return "On/Off Switch"
	case NodeTypeHorizontalAwning:
		return "Horizontal Awning"
	case NodeTypeExteriorVenetianBlind:
		return "Exterior Venetian Blind"
	case NodeTypeLouverBlind:
		return "Louver Blind"
	case NodeTypeCurtainTrack:
		return "Curtain Track"
	case NodeTypeVentilationPoint:
		return "Ventilation Point"
	case NodeTypeExteriorHeating:
		return "Exterior Heating"
	case NodeTypeSwingingShutter:
		return "Swinging Shutter"
	default:
		return fmt.Sprintf("Unknown (0x%04X)", uint16(t))
	}
}

// NodeState represents the operational state of a node
type NodeState uint8

const (
	NodeStateNonExecuting          NodeState = 0
	NodeStateErrorWhileExecution   NodeState = 1
	NodeStateNotUsed               NodeState = 2
	NodeStateWaitingForPower       NodeState = 3
	NodeStateExecuting             NodeState = 4
	NodeStateDone                  NodeState = 5
	NodeStateUnknown               NodeState = 255
)

func (s NodeState) String() string {
	switch s {
	case NodeStateNonExecuting:
		return "Non-Executing"
	case NodeStateErrorWhileExecution:
		return "Error"
	case NodeStateNotUsed:
		return "Not Used"
	case NodeStateWaitingForPower:
		return "Waiting for Power"
	case NodeStateExecuting:
		return "Executing"
	case NodeStateDone:
		return "Done"
	default:
		return "Unknown"
	}
}

// RunStatus represents the run status of a command
type RunStatus uint8

const (
	RunStatusExecutionCompleted RunStatus = 0
	RunStatusExecutionFailed    RunStatus = 1
	RunStatusExecutionActive    RunStatus = 2
)

// StatusReply represents the status reply type
type StatusReply uint8

const (
	StatusReplyUnknownStatusReply           StatusReply = 0x00
	StatusReplyCommandCompletedOk           StatusReply = 0x01
	StatusReplyNoContact                    StatusReply = 0x02
	StatusReplyManuallyOperated             StatusReply = 0x03
	StatusReplyBlocked                      StatusReply = 0x04
	StatusReplyWrongSystemKey               StatusReply = 0x05
	StatusReplyPriorityLevelLocked          StatusReply = 0x06
	StatusReplyReachedWrongPosition         StatusReply = 0x07
	StatusReplyErrorDuringExecution         StatusReply = 0x08
	StatusReplyNoExecution                  StatusReply = 0x09
	StatusReplyCalibrating                  StatusReply = 0x0A
	StatusReplyPowerConsumptionTooHigh      StatusReply = 0x0B
	StatusReplyPowerConsumptionTooLow       StatusReply = 0x0C
	StatusReplyLockPositionOpen             StatusReply = 0x0D
	StatusReplyMotionTimeTooLongCommunError StatusReply = 0x0E
	StatusReplyThermalProtection            StatusReply = 0x0F
	StatusReplyProductNotOperational        StatusReply = 0x10
	StatusReplyFilterMaintenanceNeeded      StatusReply = 0x11
	StatusReplyBatteryLevel                 StatusReply = 0x12
	StatusReplyTargetModified               StatusReply = 0x13
	StatusReplyModeNotImplemented           StatusReply = 0x14
	StatusReplyCommandIncompatibleToMovement StatusReply = 0x15
	StatusReplyUserAction                   StatusReply = 0x16
	StatusReplyDeadBoltError                StatusReply = 0x17
	StatusReplyAutomaticCycleEngaged        StatusReply = 0x18
	StatusReplyWrongLoadConnected           StatusReply = 0x19
	StatusReplyColourNotReachable           StatusReply = 0x1A
	StatusReplyTargetNotReachable           StatusReply = 0x1B
	StatusReplyBadIndexReceived             StatusReply = 0x1C
	StatusReplyCommandOverruled             StatusReply = 0x1D
	StatusReplyNodeWaitingForPower          StatusReply = 0x1E
	StatusReplyInformationCode              StatusReply = 0xDF
	StatusReplyParameterLimited             StatusReply = 0xE0
	StatusReplyLimitationByLocalUser        StatusReply = 0xE1
	StatusReplyLimitationByUser             StatusReply = 0xE2
	StatusReplyLimitationByRain             StatusReply = 0xE3
	StatusReplyLimitationByTimer            StatusReply = 0xE4
	StatusReplyLimitationByUPS              StatusReply = 0xE6
	StatusReplyLimitationByUnknown          StatusReply = 0xE7
	StatusReplyLimitationBySAAC             StatusReply = 0xEA
	StatusReplyLimitationByWind             StatusReply = 0xEB
	StatusReplyLimitationByMyself           StatusReply = 0xEC
	StatusReplyLimitationByAutomaticCycle   StatusReply = 0xED
	StatusReplyLimitationByEmergency        StatusReply = 0xEE
)

// Velocity represents movement speed
type Velocity uint8

const (
	VelocityDefault   Velocity = 0
	VelocitySilent    Velocity = 1
	VelocityFast      Velocity = 2
	VelocityNotUsed   Velocity = 255
)

// Priority level for commands
type Priority uint8

const (
	PriorityHumanProtection       Priority = 0
	PriorityEnvironmentProtection Priority = 1
	PriorityUserLevel1            Priority = 2
	PriorityUserLevel2            Priority = 3
	PriorityComfortLevel1         Priority = 4
	PriorityComfortLevel2         Priority = 5
	PriorityComfortLevel3         Priority = 6
	PriorityComfortLevel4         Priority = 7
)

// Special position values
const (
	PositionMin     uint16 = 0x0000 // Fully open
	PositionMax     uint16 = 0xC800 // Fully closed (51200)
	PositionCurrent uint16 = 0xD100 // Keep current position
	PositionDefault uint16 = 0xD200 // Use default position
	PositionIgnore  uint16 = 0xD400 // Ignore this parameter
)

// Node represents a Velux device
type Node struct {
	ID            uint8      `json:"id"`
	Name          string     `json:"name"`
	NodeType      NodeType   `json:"node_type"`
	NodeTypeStr   string     `json:"node_type_str"`
	State         NodeState  `json:"state"`
	StateStr      string     `json:"state_str"`
	CurrentPosition uint16   `json:"current_position_raw"`
	PositionPercent float64  `json:"position_percent"`
	TargetPosition  uint16   `json:"target_position_raw"`
	TargetPercent   float64  `json:"target_percent"`
	Velocity      Velocity   `json:"velocity"`
	LastUpdate    time.Time  `json:"last_update"`
}

// PositionToPercent converts raw position (0-51200) to percentage (0-100)
// 0 = fully open, 100 = fully closed
func PositionToPercent(raw uint16) float64 {
	if raw > PositionMax {
		return 100.0
	}
	return float64(raw) / float64(PositionMax) * 100.0
}

// PercentToPosition converts percentage (0-100) to raw position (0-51200)
func PercentToPosition(percent float64) uint16 {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return uint16(percent / 100.0 * float64(PositionMax))
}

// Frame represents a KLF-200 protocol frame
type Frame struct {
	Command CommandID
	Data    []byte
}

// Response status codes
type ResponseStatus uint8

const (
	StatusOK                ResponseStatus = 0
	StatusErrorSystem       ResponseStatus = 1
	StatusErrorInvalidIndex ResponseStatus = 2
	StatusErrorOutOfRange   ResponseStatus = 3
)
