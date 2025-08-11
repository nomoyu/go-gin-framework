package db

import "github.com/nomoyu/go-gin-framework/pkg/logger"

func AutoMigrate() error {
	d := DB()
	if d == nil || !HasRegisteredModels() {
		return nil
	}
	err := d.AutoMigrate(RegisteredModels()...)
	if err != nil {
		logger.Errorf("AutoMigrate fail: %v", err)
		return err
	}
	logger.Info("AutoMigrate success")
	return nil
}
