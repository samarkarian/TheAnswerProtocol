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
