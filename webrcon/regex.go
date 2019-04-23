package webrcon

import "regexp"

var chatRegex = regexp.MustCompile(`\[CHAT\] (.+?)\[[0-9]+\/([0-9]+)\] : (.*)`)
var joinRegex = regexp.MustCompile(`(.*):([0-9]+)+\/([0-9]+)+\/(.+?) joined \[(.*)\/([0-9]+)]`)
var disconnectRegex = regexp.MustCompile(`(.*):([0-9]+)+\/([0-9]+)+\/(.+?) disconnecting: (.*)`)
var killRegex = regexp.MustCompile(`(?P<victim>.+?)(?:\[(?:[0-9]+?)\/(?P<victimid>[0-9]+?)\])(?: (?P<how>was killed by|died) )(?P<killer>(?:(?:[^\/\[\]]+)\[[0-9]+/(?P<killerid>[0-9]+)\]$)|(?P<reason>[^\/]*$))`)
var statusRegex = regexp.MustCompile(`(?:.*?hostname:\s*(?P<hostname>.*?)\\n)(?:.*?version\s*:\s*(?P<version>\d+) )(?:.*?secure\s*\((?P<secure>.*?)\)\\n)(?:.*?map\s*:\s*(?P<map>.*?)\\n)(?:.*?players\s*:\s*(?P<players_current>\d+) \((?P<players_max>\d+) max\) \((?P<players_queued>\d+) queued\) \((?P<players_joining>\d+) joining\)\\n)`)
var playerListRegex = regexp.MustCompile(`(?:\\n)(?P<SteamID>[0-9]+?)(?:[ ]+?)(?:\\\")(?P<Username>.+?)(?:\\\")(?:[ ]+?)(?P<Ping>[0-9]+?)(?:[ ]+?)(?P<Connected>[0-9.s]+?)(?:[ ]+?)(?P<IP>[0-9.]+?)(?::)(?P<Port>[0-9]+?)(?:[ ]+?)(?P<OwnerSteamID>[0-9]+?)(?:[ ]+?)(?P<Violations>[0-9.]+?)(?:[ ]+?)(?P<Kicks>[0-9]+?)(?:[ ]+?)`)
var removeIDsRegex = regexp.MustCompile(`\[.+?\/.+?\]`)
var removeBracesRegex = regexp.MustCompile(`(?:.+)( \(.+\))`)
