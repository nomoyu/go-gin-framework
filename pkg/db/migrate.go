package db

import "github.com/nomoyu/go-gin-framework/pkg/logger"

func AutoMigrate() error {
	d := DB()
	if d == nil || !HasRegisteredModels() {
		return nil
	}
	err := d.AutoMigrate(RegisteredModels()...)
	if err != nil {
		logger.Errorf("AutoMigrate 失败: %v", err)
		return err
	}
	logger.Info("✅ AutoMigrate 完成")
	return nil
}
