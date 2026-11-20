// Package protocol implements parsing and formatting of RFC 42TAP lines.
package protocol

const MaxLineLen = 1024

const (
	CmdConnect   = "CONNECT"
	CmdLook      = "LOOK"
	CmdMove      = "MOVE"
	CmdQuit      = "QUIT"
	CmdChat      = "CHAT"
	CmdWho       = "WHO"
	CmdGroup     = "GROUP"
	CmdTake      = "TAKE"
	CmdDrop      = "DROP"
	CmdInventory = "INVENTORY"
	CmdTalk      = "TALK"
	CmdAttack    = "ATTACK"
	CmdStatus    = "STATUS"
	CmdQuest     = "QUEST"
	CmdQuests    = "QUESTS"
)

// CHAT scopes and events categories
const (
	ScopeGlobal = "GLOBAL"
	ScopeRoom   = "ROOM"
	ScopeGroup  = "GROUP"
	ScopeStats  = "STATS"
)

// PRESENCE actions
const (
	PresenceEnter = "ENTER"
	PresenceLeave = "LEAVE"
)

// GROUP sub-commands
const (
	GroupCreate = "CREATE"
	GroupInvite = "INVITE"
	GroupJoin   = "JOIN"
	GroupLeave  = "LEAVE"
)

// Kind is the type of a line sent by the server.
// Its value is the keyword that starts the line.
type Kind string

// Kinds of server lines. KindUnknown is the zero value
// of Kind; ParseLine never returns it.
const (
	KindUnknown Kind = ""
	KindOK      Kind = "OK"
	KindErr     Kind = "ERR"
	KindEvent   Kind = "EVT"
)

// ErrorCode struct contains the code and the associated error message
type ErrorCode struct {
	Code    int
	Message string
}

var (
	ErrNameInUse          = ErrorCode{201, "NAME_IN_USE"}
	ErrNoExit             = ErrorCode{301, "NO_EXIT"}
	ErrBadRequest         = ErrorCode{400, "BAD_REQUEST"}
	ErrNotInGroup         = ErrorCode{401, "NOT_IN_GROUP"}
	ErrAlreadyInGroup     = ErrorCode{402, "ALREADY_IN_GROUP"}
	ErrNotAuthenticated   = ErrorCode{403, "NOT_AUTHENTICATED"}
	ErrItemNotFound       = ErrorCode{404, "ITEM_NOT_FOUND"}
	ErrItemNotInInventory = ErrorCode{404, "ITEM_NOT_IN_INVENTORY"}
	ErrNPCNotFound        = ErrorCode{404, "NPC_NOT_FOUND"}
	ErrNPCNotHostile      = ErrorCode{405, "NPC_NOT_HOSTILE"}
	ErrNoQuestAvailable   = ErrorCode{406, "NO_QUEST_AVAILABLE"}
)
