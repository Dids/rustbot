package webrcon

import "regexp"

var chatRegex = regexp.MustCompile(`\[CHAT\] (.+?)\[[0-9]+\/([0-9]+)\] : (.*)`)
var joinRegex = regexp.MustCompile(`(.*):([0-9]+)+\/([0-9]+)+\/(.+?) joined \[(.*)\/([0-9]+)]`)
var disconnectRegex = regexp.MustCompile(`(.*):([0-9]+)+\/([0-9]+)+\/(.+?) disconnecting: (.*)`)
var killRegex = regexp.MustCompile(`(?P<victim>.+?)(?:\[(?:[0-9]+?)\/(?P<victimid>[0-9]+?)\])(?: (?P<how>was killed by|died) )(?P<killer>(?:(?:[^\/\[\]]+)\[[0-9]+/(?P<killerid>[0-9]+)\]$)|(?P<reason>[^\/]*$))`)
var statusRegex = regexp.MustCompile(`(?:.*?hostname:\s*(?P<hostname>.*?)\\n)(?:.*?version\s*:\s*(?P<version>\d+) )(?:.*?secure\s*\((?P<secure>.*?)\)\\n)(?:.*?map\s*:\s*(?P<map>.*?)\\n)(?:.*?players\s*:\s*(?P<players_current>\d+) \((?P<players_max>\d+) max\) \((?P<players_queued>\d+) queued\) \((?P<players_joining>\d+) joining\)\\n)`)
var removeIDsRegex = regexp.MustCompile(`\[.+?\/.+?\]`)
var removeBracesRegex = regexp.MustCompile(`(?:.+)( \(.+\))`)
