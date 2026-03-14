package locale_test

import (
	"reflect"
	"testing"

	"d2rhl/internal/common/locale"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCatalogsAllFieldsNonEmpty verifies that every string field in every
// sub-catalog is populated for all known locales.  This prevents silent
// regressions when a new string is added to the Catalog struct but forgotten
// in one of the locale files.
func TestCatalogsAllFieldsNonEmpty(t *testing.T) {
	for _, l := range locale.KnownLocales() {
		l := l
		t.Run(string(l), func(t *testing.T) {
			cat := locale.Get(l)
			require.NotNil(t, cat)
			assertAllStringFieldsNonEmpty(t, reflect.ValueOf(*cat), string(l))
		})
	}
}

// assertAllStringFieldsNonEmpty recurses into structs and asserts every string
// field is non-empty.  The Locale field itself (type locale.Locale) is skipped
// since it is checked separately.
func assertAllStringFieldsNonEmpty(t *testing.T, v reflect.Value, path string) {
	t.Helper()
	switch v.Kind() {
	case reflect.Struct:
		vt := v.Type()
		for i := 0; i < v.NumField(); i++ {
			fieldName := vt.Field(i).Name
			// Skip the top-level Locale field (it is a named type, not string).
			if vt.Field(i).Type == reflect.TypeOf(locale.Locale("")) {
				continue
			}
			assertAllStringFieldsNonEmpty(t, v.Field(i), path+"."+fieldName)
		}
	case reflect.String:
		assert.NotEmpty(t, v.String(), "empty catalog string at %s", path)
	}
}

func TestGetKnownLocalesReturnsCatalog(t *testing.T) {
	for _, l := range locale.KnownLocales() {
		cat := locale.Get(l)
		assert.NotNil(t, cat, "Get(%s) returned nil", l)
		assert.Equal(t, l, cat.Locale, "catalog Locale field mismatch for %s", l)
	}
}

func TestGetUnknownLocaleFallsBackToZhTW(t *testing.T) {
	cat := locale.Get("xx-XX")
	require.NotNil(t, cat)
	assert.Equal(t, locale.LocaleZhTW, cat.Locale)
}

func TestParseLocaleAcceptsKnownLocales(t *testing.T) {
	cases := []struct {
		input    string
		expected locale.Locale
		ok       bool
	}{
		{"zh-TW", locale.LocaleZhTW, true},
		{"ZH-TW", locale.LocaleZhTW, true},
		{"zh-tw", locale.LocaleZhTW, true},
		{"en", locale.LocaleEn, true},
		{"EN", locale.LocaleEn, true},
		{"fr", locale.LocaleZhTW, false},
		{"", locale.LocaleZhTW, false},
	}
	for _, tc := range cases {
		got, ok := locale.ParseLocale(tc.input)
		assert.Equal(t, tc.ok, ok, "ParseLocale(%q) ok mismatch", tc.input)
		assert.Equal(t, tc.expected, got, "ParseLocale(%q) locale mismatch", tc.input)
	}
}

func TestKnownLocalesContainsExpectedLocales(t *testing.T) {
	known := locale.KnownLocales()
	assert.Contains(t, known, locale.LocaleZhTW)
	assert.Contains(t, known, locale.LocaleEn)
}
