package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLIUIPrefixesAreConfiguredByMessageKind(t *testing.T) {
	testUI := newCLIUI()

	assert.Equal(t, "•", testUI.prefix(uiMessageInfo))
	assert.Equal(t, ">", testUI.prefix(uiMessageCommand))
	assert.Equal(t, "?", testUI.prefix(uiMessagePrompt))
	assert.Equal(t, "✔", testUI.prefix(uiMessageSuccess))
	assert.Equal(t, "✘", testUI.prefix(uiMessageError))
	assert.Equal(t, "⚠", testUI.prefix(uiMessageWarning))
}

func TestCLIUIOptionRendersBracketedChoice(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.option("a", "啟動所有帳號", "")
	})

	assert.Equal(t, "[a] 啟動所有帳號\n", output)
}

func TestCLIUIHeadRendersTitleBetweenDividers(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.headf("主選單")
	})

	assert.Equal(t, "\n========================================================\n"+strings.Repeat(" ", 25)+"主選單"+strings.Repeat(" ", 25)+"\n========================================================\n\n", output)
}

func TestCLIUIMenuDividerUsesMenuStyle(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.menuDividerLine()
	})

	assert.Equal(t, "--------------------------------------------------------\n", output)
}

func TestCLIUIMenuBlockWrapsCallbackContentInMenuDividers(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.menuBlock(func() {
			testUI.option("1", "測試選項", "")
		})
	})

	assert.Equal(t, "--------------------------------------------------------\n[1] 測試選項\n--------------------------------------------------------\n", output)
}

func TestCLIMenuOptionsRenderAlignsPrefixesToLongestKey(t *testing.T) {
	testUI := newCLIUI()
	options := testUI.newMenuOptions()
	options.option("數字", "啟動指定帳號", "")
	options.option("a", "啟動所有帳號", "")
	options.option("0", "離線遊玩", "")

	output := captureStdout(t, func() {
		options.render()
	})

	assert.Equal(t, "[數字] 啟動指定帳號\n[a]    啟動所有帳號\n[0]    離線遊玩\n", output)
}

func TestDisplayWidthTreatsCJKAsDoubleWidth(t *testing.T) {
	assert.Equal(t, 6, displayWidth("[數字]"))
	assert.Equal(t, 3, displayWidth("[a]"))
}

func TestCLIUISubMenuOptionsAppendsCommonNavAfterBlankLine(t *testing.T) {
	testUI := newCLIUI()
	options := testUI.subMenuOptions(func(options *cliMenuOptions) {
		options.option("1", "測試選項", "")
	})

	output := captureStdout(t, func() {
		options.render()
	})

	assert.Equal(t, "[1] 測試選項\n\n[b] 回上一層\n[h] 回主選單\n[q] 離開程式\n", output)
}

func TestCLIUIMainMenuOptionsAppendsQuitAfterBlankLine(t *testing.T) {
	testUI := newCLIUI()
	options := testUI.mainMenuOptions(func(options *cliMenuOptions) {
		options.option("1", "測試選項", "")
	})

	output := captureStdout(t, func() {
		options.render()
	})

	assert.Equal(t, "[1] 測試選項\n\n[q] 退出\n", output)
}

func TestCLIMenuOptionsRenderAlignsCommentColumn(t *testing.T) {
	testUI := newCLIUI()
	options := testUI.newMenuOptions()
	options.option("0", "離線遊玩", "可選 mod，不需帳密")
	options.option("d", "設定啟動間隔", "目前：30-60 秒（隨機）")

	output := captureStdout(t, func() {
		options.render()
	})

	assert.Equal(t, "[0] 離線遊玩      可選 mod，不需帳密\n[d] 設定啟動間隔  目前：30-60 秒（隨機）\n", output)
}

func TestCLIUIInputPromptUsesPromptRenderer(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.inputf("請選擇：")
	})

	assert.Equal(t, "? 請選擇：", output)
}

func TestCLIUICommandUsesCommandRenderer(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.commandf("%s %s", `C:\Games\D2R\D2R.exe`, "-uid osi")
	})

	assert.Equal(t, "> C:\\Games\\D2R\\D2R.exe -uid osi\n", output)
}

func TestCLIUIWarningLinesRendersGroupedMessageWithSinglePrefix(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.warningLines("第一行", "第二行", "", "第三行")
	})

	assert.Equal(t, "⚠ 第一行\n  第二行\n  第三行\n", output)
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
