package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	cliInfoIcon    = "•"
	cliPromptIcon  = "?"
	cliSuccessIcon = "✔"
	cliErrorIcon   = "✘"
	cliWarningIcon = "⚠"
)

type uiMessageKind string

const (
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

var ui = newCLIUI()

func newCLIUI() *cliUI {
	scanner := bufio.NewScanner(os.Stdin)
	return &cliUI{
		style: cliUIStyle{
			messageIndent: "",
			headerDivider: "============================================",
			menuDivider:   "--------------------------------------------",
			messagePrefixes: map[uiMessageKind]string{
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
	continuationIndent := strings.Repeat(" ", utf8.RuneCountInString(prefix))
	return prefix + strings.ReplaceAll(body, "\n", "\n"+continuationIndent)
}

func (u *cliUI) line(kind uiMessageKind, format string, args ...any) {
	fmt.Printf("%s\n", u.renderMessage(kind, format, args...))
}

func (u *cliUI) rawln(text string) {
	fmt.Println(text)
}

func (u *cliUI) rawlnf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func (u *cliUI) headf(format string, args ...any) {
	u.headerDividerLine()
	u.rawlnf(format, args...)
	u.headerDividerLine()
	ui.blankLine()
}

func (u *cliUI) infof(format string, args ...any) {
	u.line(uiMessageInfo, format, args...)
}

func (u *cliUI) promptf(format string, args ...any) {
	u.line(uiMessagePrompt, format, args...)
}

func (u *cliUI) successf(format string, args ...any) {
	u.line(uiMessageSuccess, format, args...)
}

func (u *cliUI) errorf(format string, args ...any) {
	u.line(uiMessageError, format, args...)
}

func (u *cliUI) warningf(format string, args ...any) {
	u.line(uiMessageWarning, format, args...)
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

func (u *cliUI) option(key, label string) {
	fmt.Printf("[%s] %s\n", key, label)
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

func (u *cliUI) subMenuNav() {
	u.blankLine()
	u.option(menuBack, "回上一層")
	u.option(menuHome, "回主選單")
	u.option(menuQuit, "離開程式")
}

func (u *cliUI) anyKeyContinue() error {
	var err error
	if u.canSingleKeyContinue() {
		u.inputf("請按任意鍵繼續...")
		err = u.waitForAnyKey()
	} else {
		_, _ = u.readInputf("請按 Enter 繼續...")
	}
	u.blankLine()
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
