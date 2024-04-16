/*
Package logger Логер всего приложения
*/
package logger

import (
	"go.uber.org/zap"
)

// Log Логгер приложения
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует логгер с необходимым уровнем логирования
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// устанавливаем синглтон
	Log = zl
	return nil
}
