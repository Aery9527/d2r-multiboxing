package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/width"
)

const (
	cliInfoIcon    = "•"
	cliCommandIcon = ">"
	cliPromptIcon  = "?"
	cliSuccessIcon = "✔"
	cliErrorIcon   = "✘"
	cliWarningIcon = "⚠"
)

type uiMessageKind string

const (
	uiMessageCommand uiMessageKind = "command"
	uiMessageInfo    uiMessageKind = "info"
	uiMessagePrompt  uiMessageKind = "prompt"
	uiMessageSuccess uiMessageKind = "success"
	uiMessageError   uiMessageKind = "error"
	uiMessageWarning uiMessageKind = "warning"
)

type cliUIStyle struct {
	messageIndent   string
	headerDivider   string
	menuDivider     string
	messagePrefixes map[uiMessageKind]string
}

type cliUI struct {
	style                cliUIStyle
	waitForAnyKey        func() error
	readLine             func() (string, bool)
	canSingleKeyContinue func() bool
}

type cliMenuEntryKind string

const (
	cliMenuEntryOption cliMenuEntryKind = "option"
	cliMenuEntryBlank  cliMenuEntryKind = "blank"
)

type cliMenuEntry struct {
	kind    cliMenuEntryKind
	key     string
	label   string
	comment string
}

type cliMenuOptions struct {
	ui      *cliUI
	entries []cliMenuEntry
}

var ui = newCLIUI()

func newCLIUI() *cliUI {
	scanner := bufio.NewScanner(os.Stdin)
	return &cliUI{
		style: cliUIStyle{
			messageIndent: "",
			headerDivider: "========================================================",
			menuDivider:   "--------------------------------------------------------",
			messagePrefixes: map[uiMessageKind]string{
				uiMessageCommand: cliCommandIcon,
				uiMessageInfo:    cliInfoIcon,
				uiMessagePrompt:  cliPromptIcon,
				uiMessageSuccess: cliSuccessIcon,
				uiMessageError:   cliErrorIcon,
				uiMessageWarning: cliWarningIcon,
			},
		},
		waitForAnyKey: waitForAnyKey,
		readLine: func() (string, bool) {
			if !scanner.Scan() {
				return "", false
			}
			return strings.TrimSpace(scanner.Text()), true
		},
		canSingleKeyContinue: consoleSupportsSingleKeyContinue,
	}
}

func (u *cliUI) prefix(kind uiMessageKind) string {
	return u.style.messagePrefixes[kind]
}

func (u *cliUI) renderMessage(kind uiMessageKind, format string, args ...any) string {
	prefix := fmt.Sprintf("%s%s ", u.style.messageIndent, u.prefix(kind))
	body := fmt.Sprintf(format, args...)
	continuationIndent := strings.Repeat(" ", displayWidth(prefix))
	return prefix + strings.ReplaceAll(body, "\n", "\n"+continuationIndent)
}

func (u *cliUI) line(kind uiMessageKind, format string, args ...any) {
	fmt.Printf("%s\n", u.renderMessage(kind, format, args...))
}

func (u *cliUI) lines(kind uiMessageKind, messages ...string) {
	group := make([]string, 0, len(messages))
	for _, message := range messages {
		if message == "" {
			continue
		}
		group = append(group, message)
	}
	if len(group) == 0 {
		return
	}
	u.line(kind, "%s", strings.Join(group, "\n"))
}

func (u *cliUI) rawln(text string) {
	fmt.Println(text)
}

func (u *cliUI) rawlnf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (u *cliUI) headf(format string, args ...any) {
	title := fmt.Sprintf(format, args...)
	width := displayWidth(u.style.headerDivider)
	if titleWidth := displayWidth(title); titleWidth < width {
		padding := width - titleWidth
		leftPadding := padding / 2
		rightPadding := padding - leftPadding
		title = strings.Repeat(" ", leftPadding) + title + strings.Repeat(" ", rightPadding)
	}

	u.blankLine()
	u.headerDividerLine()
	u.rawln(title)
	u.headerDividerLine()
	u.blankLine()
}

func (u *cliUI) menuBlock(render func()) {
	u.menuDividerLine()
	render()
	u.menuDividerLine()
}

func (u *cliUI) newMenuOptions() *cliMenuOptions {
	return &cliMenuOptions{
		ui:      u,
		entries: make([]cliMenuEntry, 0, 8),
	}
}

func (u *cliUI) subMenuOptions(build func(*cliMenuOptions)) *cliMenuOptions {
	options := u.newMenuOptions()
	if build != nil {
		build(options)
	}
	options.appendSubMenuNav()
	return options
}

func (u *cliUI) mainMenuOptions(build func(*cliMenuOptions)) *cliMenuOptions {
	options := u.newMenuOptions()
	if build != nil {
		build(options)
	}
	options.appendQuitOption()
	return options
}

func (u *cliUI) infof(format string, args ...any) {
	u.line(uiMessageInfo, format, args...)
}

func (u *cliUI) commandf(format string, args ...any) {
	u.line(uiMessageCommand, format, args...)
}

func (u *cliUI) infoLines(messages ...string) {
	u.lines(uiMessageInfo, messages...)
}

func (u *cliUI) promptf(format string, args ...any) {
	u.line(uiMessagePrompt, format, args...)
}

func (u *cliUI) promptLines(messages ...string) {
	u.lines(uiMessagePrompt, messages...)
}

func (u *cliUI) successf(format string, args ...any) {
	u.line(uiMessageSuccess, format, args...)
}

func (u *cliUI) successLines(messages ...string) {
	u.lines(uiMessageSuccess, messages...)
}

func (u *cliUI) errorf(format string, args ...any) {
	u.line(uiMessageError, format, args...)
}

func (u *cliUI) errorLines(messages ...string) {
	u.lines(uiMessageError, messages...)
}

func (u *cliUI) warningf(format string, args ...any) {
	u.line(uiMessageWarning, format, args...)
}

func (u *cliUI) warningLines(messages ...string) {
	u.lines(uiMessageWarning, messages...)
}

func (u *cliUI) inputf(format string, args ...any) {
	fmt.Print(u.renderMessage(uiMessagePrompt, format, args...))
}

func (u *cliUI) readInput() (string, bool) {
	return u.readInputf("請選擇：")
}

func (u *cliUI) readInputf(format string, args ...any) (string, bool) {
	u.inputf(format, args...)
	return u.readLine()
}

func (u *cliUI) option(key, label, comment string) {
	if comment == "" {
		fmt.Printf("[%s] %s\n", key, label)
		return
	}
	fmt.Printf("[%s] %s  %s\n", key, label, comment)
}

func (o *cliMenuOptions) option(key, label, comment string) {
	o.entries = append(o.entries, cliMenuEntry{
		kind:    cliMenuEntryOption,
		key:     key,
		label:   label,
		comment: comment,
	})
}

func (o *cliMenuOptions) blankLine() {
	o.entries = append(o.entries, cliMenuEntry{kind: cliMenuEntryBlank})
}

func (o *cliMenuOptions) appendSubMenuNav() {
	o.blankLine()
	o.option(menuBack, "回上一層", "")
	o.option(menuHome, "回主選單", "")
	o.option(menuQuit, "離開程式", "")
}

func (o *cliMenuOptions) appendQuitOption() {
	o.blankLine()
	o.option(menuQuit, "退出", "")
}

func (o *cliMenuOptions) render() {
	maxPrefixWidth := 0
	maxLabelWidth := 0
	for _, entry := range o.entries {
		if entry.kind != cliMenuEntryOption {
			continue
		}
		prefixWidth := displayWidth(fmt.Sprintf("[%s]", entry.key))
		if prefixWidth > maxPrefixWidth {
			maxPrefixWidth = prefixWidth
		}
		labelWidth := displayWidth(entry.label)
		if labelWidth > maxLabelWidth {
			maxLabelWidth = labelWidth
		}
	}

	for _, entry := range o.entries {
		switch entry.kind {
		case cliMenuEntryBlank:
			o.ui.blankLine()
		case cliMenuEntryOption:
			prefix := fmt.Sprintf("[%s]", entry.key)
			prefixPadding := strings.Repeat(" ", maxPrefixWidth-displayWidth(prefix))
			line := fmt.Sprintf("%s%s %s", prefix, prefixPadding, entry.label)
			if entry.comment != "" {
				labelPadding := strings.Repeat(" ", maxLabelWidth-displayWidth(entry.label))
				line += fmt.Sprintf("%s  %s", labelPadding, entry.comment)
			}
			o.ui.rawln(line)
		}
	}
}

func displayWidth(s string) int {
	widthSum := 0
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		switch width.LookupRune(r).Kind() {
		case width.EastAsianWide, width.EastAsianFullwidth:
			widthSum += 2
		default:
			widthSum++
		}
	}
	return widthSum
}

func (u *cliUI) headerDividerLine() {
	fmt.Println(u.style.headerDivider)
}

func (u *cliUI) menuDividerLine() {
	fmt.Println(u.style.menuDivider)
}

func (u *cliUI) blankLine() {
	fmt.Println()
}

func (u *cliUI) anyKeyContinue() error {
	u.blankLine()
	var err error
	if u.canSingleKeyContinue() {
		u.inputf("請按任意鍵繼續...")
		err = u.waitForAnyKey()
		u.blankLine()
	} else {
		_, _ = u.readInputf("請按 Enter 繼續...")
	}
	return err
}

func showInvalidInputAndPause() {
	showInputErrorAndPause("無效輸入，請重試。")
}

func showInputErrorAndPause(message string) {
	ui.errorf("%s", message)
	if err := ui.anyKeyContinue(); err != nil {
		ui.warningf("等待按鍵失敗：%v", err)
	}
	ui.blankLine()
}
