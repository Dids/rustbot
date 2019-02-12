package discord

import (
	"regexp"

	"github.com/Dids/rustbot/eventhandler"
)

// TODO: Refactor/move these to EventHandler, and if possible escape automatically!
func escapeMessage(message eventhandler.Message) eventhandler.Message {
	message.User = escapeMarkdown(message.User)
	message.Message = escapeMarkdown(message.Message)
	return message
}

func escapeMarkdown(markdown string) string {
	unescaped := unescapeBackslashRegex.ReplaceAllString(markdown, `$1`) // unescape any "backslashed" character
	escaped := escapeMarkdownRegex.ReplaceAllString(unescaped, `\$1`)
	return escaped
}

type caseInsensitiveReplacer struct {
	toReplace   *regexp.Regexp
	replaceWith string
}

func newCaseInsensitiveReplacer(toReplace, replaceWith string) *caseInsensitiveReplacer {
	return &caseInsensitiveReplacer{
		toReplace:   regexp.MustCompile("(?i)" + toReplace),
		replaceWith: replaceWith,
	}
}

func (cir *caseInsensitiveReplacer) Replace(str string) string {
	return cir.toReplace.ReplaceAllString(str, cir.replaceWith)
}
