package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLIUIPrefixesAreConfiguredByMessageKind(t *testing.T) {
	testUI := newCLIUI()

	assert.Equal(t, "•", testUI.prefix(uiMessageInfo))
	assert.Equal(t, "?", testUI.prefix(uiMessagePrompt))
	assert.Equal(t, "✔", testUI.prefix(uiMessageSuccess))
	assert.Equal(t, "✘", testUI.prefix(uiMessageError))
	assert.Equal(t, "⚠", testUI.prefix(uiMessageWarning))
}

func TestCLIUIOptionRendersBracketedChoice(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.option("a", "啟動所有帳號")
	})

	assert.Equal(t, "[a] 啟動所有帳號\n", output)
}

func TestCLIUIHeadRendersTitleBetweenDividers(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.headf("主選單")
	})

	assert.Equal(t, "============================================\n主選單\n============================================\n\n", output)
}

func TestCLIUIMenuDividerUsesMenuStyle(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.menuDividerLine()
	})

	assert.Equal(t, "--------------------------------------------\n", output)
}

func TestCLIUISubMenuNavKeepsBackHomeQuitLast(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.subMenuNav()
	})

	assert.Contains(t, output, "\n[b] 回上一層\n[h] 回主選單\n[q] 離開程式\n")
}

func TestCLIUIInputPromptUsesPromptRenderer(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.inputf("請選擇：")
	})

	assert.Equal(t, "? 請選擇：", output)
}

func TestCLIUIReadInputUsesDefaultPrompt(t *testing.T) {
	testUI := newCLIUI()
	testUI.readLine = func() (string, bool) {
		return "a", true
	}

	var (
		input string
		ok    bool
	)
	output := captureStdout(t, func() {
		input, ok = testUI.readInput()
	})

	assert.True(t, ok)
	assert.Equal(t, "a", input)
	assert.Equal(t, "? 請選擇：", output)
}

func TestCLIUILineIndentsMultilineMessagesAfterIcon(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.warningf("第一行\n第二行")
	})

	assert.Equal(t, "⚠ 第一行\n  第二行\n", output)
}

func TestCLIUIInputIndentsMultilinePromptAfterIcon(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.inputf("第一行\n第二行：")
	})

	assert.Equal(t, "? 第一行\n  第二行：", output)
}
