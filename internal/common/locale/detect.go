package locale

import "syscall"

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultLCID = kernel32.NewProc("GetUserDefaultLCID")
)

// DetectSystemLocale probes the Windows user-default UI language and maps it
// to a supported Locale.  Returns LocaleZhTW for any Traditional-Chinese LCID,
// LocaleEn for everything else.
func DetectSystemLocale() Locale {
	r, _, _ := procGetUserDefaultLCID.Call()
	return lcidToLocale(uint32(r))
}

// lcidToLocale maps a Windows LCID to a supported Locale.
// Primary language IDs are in the low 10 bits of the LCID.
// 0x04 = Chinese; sub-language 0x01/0x03/0x05 = Traditional Chinese variants.
func lcidToLocale(lcid uint32) Locale {
	primaryLang := lcid & 0x3FF
	if primaryLang == 0x04 {
		subLang := (lcid >> 10) & 0x3F
		const (
			subLangChineseTW = 0x01 // zh-TW  (LCID 0x0404)
			subLangChineseHK = 0x03 // zh-HK  (LCID 0x0C04)
			subLangChineseMO = 0x05 // zh-MO  (LCID 0x1404)
		)
		switch subLang {
		case subLangChineseTW, subLangChineseHK, subLangChineseMO:
			return LocaleZhTW
		}
	}
	return LocaleEn
}
