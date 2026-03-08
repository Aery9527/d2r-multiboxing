package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"d2rhl/internal/common/d2r"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, d2r.DefaultGamePath, cfg.D2RPath)
	assert.Equal(t, LaunchDelayRange{MinSeconds: 30, MaxSeconds: 30}, cfg.LaunchDelay)
}
func TestSaveAndLoad(t *testing.T) {
	// 使用 temp 目錄模擬設定檔
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, configFileName)

	cfg := Config{
		D2RPath:     `D:\Games\D2R\D2R.exe`,
		LaunchDelay: LaunchDelayRange{MinSeconds: 30, MaxSeconds: 60},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	assert.NoError(t, err)

	err = os.WriteFile(cfgPath, data, 0o600)
	assert.NoError(t, err)

	// 驗證讀回
	readData, err := os.ReadFile(cfgPath)
	assert.NoError(t, err)

	var loaded Config
	err = json.Unmarshal(readData, &loaded)
	assert.NoError(t, err)
	assert.Equal(t, cfg.D2RPath, loaded.D2RPath)
	assert.Equal(t, cfg.LaunchDelay, loaded.LaunchDelay)
}

func TestConfigJSON(t *testing.T) {
	cfg := Config{
		D2RPath:     `C:\Program Files (x86)\Diablo II Resurrected\D2R.exe`,
		LaunchDelay: LaunchDelayRange{MinSeconds: 30, MaxSeconds: 60},
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	assert.NoError(t, err)

	var parsed Config
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)
	assert.Equal(t, cfg, parsed)
}

func TestParseLaunchDelayRange(t *testing.T) {
	delay, err := ParseLaunchDelayRange("30-60")
	assert.NoError(t, err)
	assert.Equal(t, LaunchDelayRange{MinSeconds: 30, MaxSeconds: 60}, delay)

	delay, err = ParseLaunchDelayRange("30")
	assert.NoError(t, err)
	assert.Equal(t, LaunchDelayRange{MinSeconds: 30, MaxSeconds: 30}, delay)
}

func TestParseLaunchDelayRangeRejectsTooSmallValues(t *testing.T) {
	_, err := ParseLaunchDelayRange("9-30")
	assert.EqualError(t, err, "啟動間隔下限不可小於 10 秒")

	_, err = ParseLaunchDelayRange("30-9")
	assert.EqualError(t, err, "啟動間隔上限不可小於 10 秒")
}

func TestParseLaunchDelayRangeRejectsReverseRange(t *testing.T) {
	_, err := ParseLaunchDelayRange("60-30")
	assert.EqualError(t, err, "啟動間隔下限不可大於上限")
}

func TestLaunchDelayRangeJSONCompatibility(t *testing.T) {
	var cfg Config

	assert.NoError(t, json.Unmarshal([]byte(`{"d2r_path":"C:\\D2R.exe","launch_delay":30}`), &cfg))
	assert.Equal(t, LaunchDelayRange{MinSeconds: 30, MaxSeconds: 30}, cfg.LaunchDelay)

	assert.NoError(t, json.Unmarshal([]byte(`{"d2r_path":"C:\\D2R.exe","launch_delay":"30-60"}`), &cfg))
	assert.Equal(t, LaunchDelayRange{MinSeconds: 30, MaxSeconds: 60}, cfg.LaunchDelay)
}
