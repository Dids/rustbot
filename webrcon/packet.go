package webrcon

// Status represents the current status of the server
var Status StatusPacket

// PacketType represents the type of a webrcon packet
type PacketType string

const (
	// GenericType is a packet type
	GenericType PacketType = "Generic"
	// ChatType is a packet type
	ChatType PacketType = "Chat"
	// IgnoreType is a packet type
	IgnoreType PacketType = "Ignore"
)

// PacketIdentifier represents the identifier of a webrcon packet
type PacketIdentifier int

const (
	// GenericIdentifier marks packets as generic data
	GenericIdentifier PacketIdentifier = 0
	// ChatIdentifier marks packets as chat messages
	ChatIdentifier PacketIdentifier = -1
	// IgnoreIdentifier marks packets as ignored
	IgnoreIdentifier PacketIdentifier = -2
)

// Packet represents a single webrcon packet
type Packet struct {
	Message    string           `json:"Message"`
	Identifier PacketIdentifier `json:"Identifier"`
	Type       PacketType       `json:"Type"`
	Stacktrace string           `json:"Stacktrace"`
}

// ChatPacket represents a single webrcon chat packet
type ChatPacket struct {
	Message  string `json:"Message"`
	UserID   uint64 `json:"UserId"`
	Username string `json:"Username"`
	Color    string `json:"Color"`
	Time     uint64 `json:"Time"`
}

// JoinPacket represents a single webrcon join packet
type JoinPacket struct {
	IP       string `json:"IP"`
	Port     string `json:"Port"`
	UserID   uint64 `json:"UserId"`
	Username string `json:"Username"`
	OS       string `json:"OS"`
}

// DisconnectPacket represents a single webrcon disconnect packet
type DisconnectPacket struct {
	IP       string `json:"IP"`
	Port     string `json:"Port"`
	UserID   uint64 `json:"UserId"`
	Username string `json:"Username"`
}

// StatusPacket represents a single webrcon status packet
type StatusPacket struct {
	Hostname       string `json:"hostname"`
	Version        int    `json:"version"`
	Secure         string `json:"secure"`
	Map            string `json:"map"`
	CurrentPlayers int    `json:"players_current"`
	MaxPlayers     int    `json:"players_max"`
	QueuedPlayers  int    `json:"players_queued"`
	JoiningPlayers int    `json:"players_joining"`
}
