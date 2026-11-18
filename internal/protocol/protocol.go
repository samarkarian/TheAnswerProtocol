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
