package core

import (
	"fmt"

	backupAPI "github.com/mycontroller-org/backend/v2/pkg/backup"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/config"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

func RunSystemStartJobs() {
	dir := model.GetDirectoryInternal()
	filename := fmt.Sprintf("%s/%s", dir, config.SystemStartJobsFilename)
	if !utils.IsFileExists(filename) {
		return
	}

	bytes, err := utils.ReadFile(dir, config.SystemStartJobsFilename)
	if err != nil {
		zap.L().Fatal("error on loading system startup file", zap.String("filename", filename), zap.Error(err))
		return
	}

	jobs := &config.SystemStartupJobs{}
	err = yaml.Unmarshal(bytes, jobs)
	if err != nil {
		zap.L().Fatal("error on loading system startup file", zap.String("filename", filename), zap.Error(err))
		return
	}

	// execute restore operation
	systemRestoreOperation(&jobs.Restore)

	err = utils.RemoveFileOrEmptyDir(filename)
	if err != nil {
		zap.L().Fatal("error on removing file", zap.Any("filename", filename), zap.Error(err))
		return
	}

	// remove internal directory
	err = utils.RemoveDir(model.GetDirectoryInternal())
	if err != nil {
		zap.L().Fatal("error on removing internal direcotry", zap.String("path", model.GetDirectoryInternal()), zap.Error(err))
	}
}

func systemRestoreOperation(cfg *config.StartupRestore) {
	if !cfg.Enabled {
		return
	}

	zap.L().Info("found a restore setup on startup. Performaing restore operation", zap.Any("config", cfg))
	err := backupAPI.ExecuteRestore(cfg.ExtractedDirectory)
	if err != nil {
		zap.L().Fatal("error on restore", zap.Error(err))
		return
	}

	// clean extracted files
	err = utils.RemoveDir(cfg.ExtractedDirectory)
	if err != nil {
		zap.L().Fatal("error on deleting extracted backup files", zap.Any("restoreConfig", cfg), zap.Error(err))
		return
	}
}
