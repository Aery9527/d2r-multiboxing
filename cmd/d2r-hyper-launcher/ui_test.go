package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLIUIPrefixesAreConfiguredByMessageKind(t *testing.T) {
	testUI := newCLIUI()

	kinds := []uiMessageKind{
		uiMessageInfo,
		uiMessageCommand,
		uiMessagePrompt,
		uiMessageSuccess,
		uiMessageError,
		uiMessageWarning,
	}

	seen := make(map[string]bool, len(kinds))
	for _, kind := range kinds {
		prefix := testUI.prefix(kind)
		assert.NotEmpty(t, prefix)
		assert.False(t, seen[prefix])
		seen[prefix] = true
	}
}

func TestCLIUIOptionRendersBracketedChoice(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.option("a", "launch-all", "")
	})

	assert.Equal(t, []string{"[a] launch-all"}, nonEmptyOutputLines(output))
}

func TestCLIUIHeadRendersTitleBetweenDividers(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.headf("Main Menu")
	})

	lines := nonEmptyOutputLines(output)
	assert.Len(t, lines, 3)
	assert.Equal(t, testUI.style.headerDivider, lines[0])
	assert.Equal(t, testUI.style.headerDivider, lines[2])
	assert.Equal(t, "Main Menu", strings.TrimSpace(lines[1]))
	assert.Equal(t, displayWidth(testUI.style.headerDivider), displayWidth(lines[1]))

	leftPadding := displayWidth(lines[1]) - displayWidth(strings.TrimLeft(lines[1], " "))
	rightPadding := displayWidth(lines[1]) - displayWidth(strings.TrimRight(lines[1], " "))
	assert.LessOrEqual(t, absInt(leftPadding-rightPadding), 1)
}

func TestCLIUIMenuDividerUsesMenuStyle(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.menuDividerLine()
	})

	assert.Equal(t, []string{testUI.style.menuDivider}, nonEmptyOutputLines(output))
}

func TestCLIUIMenuBlockWrapsCallbackContentInMenuDividers(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.menuBlock(func() {
			testUI.option("1", "option", "")
		})
	})

	lines := nonEmptyOutputLines(output)
	assert.Len(t, lines, 3)
	assert.Equal(t, testUI.style.menuDivider, lines[0])
	assert.Equal(t, testUI.style.menuDivider, lines[2])

	option, ok := parseMenuOptionLine(lines[1])
	assert.True(t, ok)
	assert.Equal(t, "1", option.key)
	assert.Contains(t, option.line, "option")
}

func TestCLIMenuOptionsRenderAlignsPrefixesToLongestKey(t *testing.T) {
	testUI := newCLIUI()
	options := testUI.newMenuOptions()
	options.option("digits", "Launch target", "")
	options.option("a", "Launch all", "")
	options.option("0", "Offline", "")

	output := captureStdout(t, func() {
		options.render()
	})

	lines := nonEmptyOutputLines(output)
	assert.Len(t, lines, 3)

	targetPrefixWidth := displayWidth(lines[0][:strings.Index(lines[0], "Launch target")])
	allPrefixWidth := displayWidth(lines[1][:strings.Index(lines[1], "Launch all")])
	offlinePrefixWidth := displayWidth(lines[2][:strings.Index(lines[2], "Offline")])
	assert.Equal(t, targetPrefixWidth, allPrefixWidth)
	assert.Equal(t, targetPrefixWidth, offlinePrefixWidth)
}

func TestDisplayWidthTreatsCJKAsDoubleWidth(t *testing.T) {
	assert.Equal(t, 6, displayWidth("[數字]"))
	assert.Equal(t, 3, displayWidth("[a]"))
}

func TestCLIUISubMenuOptionsAppendsCommonNavAfterBlankLine(t *testing.T) {
	testUI := newCLIUI()
	options := testUI.subMenuOptions(func(options *cliMenuOptions) {
		options.option("1", "option", "")
	})

	output := captureStdout(t, func() {
		options.render()
	})

	lines := normalizedOutputLines(output)
	keys := make([]string, 0, 4)
	for _, line := range lines {
		option, ok := parseMenuOptionLine(line)
		if !ok {
			continue
		}
		keys = append(keys, option.key)
	}

	assert.Equal(t, []string{"1", "b", "h", "q"}, keys)
	navIndex := firstLineIndex(lines, func(line string) bool {
		option, ok := parseMenuOptionLine(line)
		return ok && option.key == "b"
	})
	assert.Greater(t, navIndex, 0)
	assert.Equal(t, "", lines[navIndex-1])
}

func TestCLIUIMainMenuOptionsAppendsQuitAfterBlankLine(t *testing.T) {
	testUI := newCLIUI()
	options := testUI.mainMenuOptions(func(options *cliMenuOptions) {
		options.option("1", "option", "")
	})

	output := captureStdout(t, func() {
		options.render()
	})

	lines := normalizedOutputLines(output)
	keys := make([]string, 0, 2)
	for _, line := range lines {
		option, ok := parseMenuOptionLine(line)
		if !ok {
			continue
		}
		keys = append(keys, option.key)
	}

	assert.Equal(t, []string{"1", "q"}, keys)
	quitIndex := firstLineIndex(lines, func(line string) bool {
		option, ok := parseMenuOptionLine(line)
		return ok && option.key == "q"
	})
	assert.Greater(t, quitIndex, 0)
	assert.Equal(t, "", lines[quitIndex-1])
}

func TestCLIMenuOptionsRenderAlignsCommentColumn(t *testing.T) {
	testUI := newCLIUI()
	options := testUI.newMenuOptions()
	options.option("0", "Offline", "No login")
	options.option("d", "Delay", "30-60 sec")

	output := captureStdout(t, func() {
		options.render()
	})

	lines := nonEmptyOutputLines(output)
	assert.Len(t, lines, 2)

	firstCommentColumn := displayWidth(lines[0][:strings.Index(lines[0], "No login")])
	secondCommentColumn := displayWidth(lines[1][:strings.Index(lines[1], "30-60 sec")])
	assert.Equal(t, firstCommentColumn, secondCommentColumn)
}

func TestCLIUIInputPromptUsesPromptRenderer(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.inputf("Select:")
	})

	assert.True(t, strings.HasPrefix(output, testUI.prefix(uiMessagePrompt)+" "))
	assert.True(t, strings.HasSuffix(output, "Select:"))
	assert.NotContains(t, output, "\n")
}

func TestCLIUICommandUsesCommandRenderer(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.commandf("%s %s", `C:\Games\D2R\D2R.exe`, "-uid osi")
	})

	assert.True(t, strings.HasPrefix(output, testUI.prefix(uiMessageCommand)+" "))
	assert.True(t, strings.Contains(output, `C:\Games\D2R\D2R.exe`))
	assert.True(t, strings.HasSuffix(output, "-uid osi\n"))
}

func TestCLIUIWarningLinesRendersGroupedMessageWithSinglePrefix(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.warningLines("line 1", "line 2", "", "line 3")
	})

	lines := nonEmptyOutputLines(output)
	assert.Len(t, lines, 3)
	assert.True(t, strings.HasPrefix(lines[0], testUI.prefix(uiMessageWarning)+" "))

	continuationIndent := strings.Repeat(" ", displayWidth(testUI.prefix(uiMessageWarning)+" "))
	assert.True(t, strings.HasPrefix(lines[1], continuationIndent))
	assert.True(t, strings.HasPrefix(lines[2], continuationIndent))
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
	assert.True(t, strings.HasPrefix(output, testUI.prefix(uiMessagePrompt)+" "))
	assert.NotContains(t, output, "\n")
}

func TestCLIUILineIndentsMultilineMessagesAfterIcon(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.warningf("line 1\nline 2")
	})

	lines := nonEmptyOutputLines(output)
	assert.Len(t, lines, 2)
	assert.True(t, strings.HasPrefix(lines[0], testUI.prefix(uiMessageWarning)+" "))

	continuationIndent := strings.Repeat(" ", displayWidth(testUI.prefix(uiMessageWarning)+" "))
	assert.True(t, strings.HasPrefix(lines[1], continuationIndent))
}

func TestCLIUIInputIndentsMultilinePromptAfterIcon(t *testing.T) {
	testUI := newCLIUI()

	output := captureStdout(t, func() {
		testUI.inputf("line 1\nline 2:")
	})

	lines := nonEmptyOutputLines(output)
	assert.Len(t, lines, 2)
	assert.True(t, strings.HasPrefix(lines[0], testUI.prefix(uiMessagePrompt)+" "))

	continuationIndent := strings.Repeat(" ", displayWidth(testUI.prefix(uiMessagePrompt)+" "))
	assert.True(t, strings.HasPrefix(lines[1], continuationIndent))
}
