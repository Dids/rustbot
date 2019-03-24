package webrcon

// Status represents the current status of the server
var Status StatusPacket

// Players represents the current players on the server
var Players PlayerPacket

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
	Message    string           `json:"Message,omitempty"`
	Identifier PacketIdentifier `json:"Identifier,omitempty"`
	Type       PacketType       `json:"Type,omitempty"`
	Stacktrace string           `json:"Stacktrace,omitempty"`
}

// ChatPacket represents a single webrcon chat packet
type ChatPacket struct {
	Message  string `json:"Message,omitempty"`
	UserID   uint64 `json:"UserId,omitempty"`
	Username string `json:"Username,omitempty"`
	Color    string `json:"Color,omitempty"`
	Time     uint64 `json:"Time,omitempty"`
}

// JoinPacket represents a single webrcon join packet
type JoinPacket struct {
	IP       string `json:"IP,omitempty"`
	Port     string `json:"Port,omitempty"`
	UserID   uint64 `json:"UserId,omitempty"`
	Username string `json:"Username,omitempty"`
	OS       string `json:"OS,omitempty"`
}

// DisconnectPacket represents a single webrcon disconnect packet
type DisconnectPacket struct {
	IP       string `json:"IP,omitempty"`
	Port     string `json:"Port,omitempty"`
	UserID   uint64 `json:"UserId,omitempty"`
	Username string `json:"Username,omitempty"`
}

// StatusPacket represents a single webrcon status packet
type StatusPacket struct {
	Hostname       string          `json:"hostname,omitempty"`
	Version        int             `json:"version,omitempty"`
	Secure         string          `json:"secure,omitempty"`
	Map            string          `json:"map,omitempty"`
	CurrentPlayers int             `json:"players_current,omitempty"`
	MaxPlayers     int             `json:"players_max,omitempty"`
	QueuedPlayers  int             `json:"players_queued,omitempty"`
	JoiningPlayers int             `json:"players_joining,omitempty"`
	Players        []*PlayerPacket `json:"players,omitempty"`
}

// PlayerPacket represents a single user in a StatusPacket
type PlayerPacket struct {
	SteamID    string  `json:"steamid,omitempty"`
	Username   string  `json:"username,omitempty"`
	Ping       int     `json:"ping,omitempty"`
	Connected  string  `json:"connected,omitempty"`
	IP         string  `json:"ip,omitempty"`
	Port       int     `json:"port,omitempty"`
	Violations float32 `json:"violations,omitempty"`
	Kicks      int     `json:"kicks,omitempty"`
}
