package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"d2r-multiboxing/internal/d2r"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, d2r.DefaultGamePath, cfg.D2RPath)
}

func TestSaveAndLoad(t *testing.T) {
	// 使用 temp 目錄模擬設定檔
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, configFileName)

	cfg := Config{
		D2RPath: `D:\Games\D2R\D2R.exe`,
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
}

func TestConfigJSON(t *testing.T) {
	cfg := Config{
		D2RPath: `C:\Program Files (x86)\Diablo II Resurrected\D2R.exe`,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	assert.NoError(t, err)

	var parsed Config
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)
	assert.Equal(t, cfg, parsed)
}
