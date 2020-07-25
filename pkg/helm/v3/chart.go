package v3

import (
	"coveros.com/pkg/utils"
	"fmt"
	"os"
)

func (h HelmV3) DownloadChart(repoUrl, chart, version, username, password, destDir string) (string, error) {

	_, err := os.Stat(destDir)
	if os.IsNotExist(err) {
		if errMakingDir := os.MkdirAll(destDir, os.ModeDir); errMakingDir != nil {
			helmInfoLogF("Failed to create dir to download chart into: %v", errMakingDir)
			return "", errMakingDir
		}
	}

	chartTarballName := fmt.Sprintf("%s-%s.tgz", chart, version)
	assumedChartPath := fmt.Sprintf("%s/%s", destDir, chartTarballName)
	assumedDownloadUrl := fmt.Sprintf("%s/%s", repoUrl, chartTarballName)
	if errDownloadingChart := utils.DownloadFile(assumedChartPath, assumedDownloadUrl, username, password); errDownloadingChart != nil {
		return "", errDownloadingChart
	}
	return assumedChartPath, nil
}
