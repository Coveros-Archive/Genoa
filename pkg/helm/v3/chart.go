package v3

import (
	"coveros.com/pkg/utils"
	"fmt"
	"helm.sh/helm/v3/pkg/repo"
	"os"
)

func (h HelmV3) DownloadChart(repoUrl, repoAlias, chart, version, username, password, destDir string) (string, error) {

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

		logger.Info(fmt.Sprintf("Assumed download url '%s' for downloading chart did not work", assumedDownloadUrl))

		// assumed url did not work, lets try downloading from a url specified in index file ( the slow helm way... )
		assumedIndexFileName := repoAlias + "-index.yaml"
		repoCacheFile, errLoadingIndex := repo.LoadIndexFile(fmt.Sprintf("%s/%s", h.settings.RepositoryCache, assumedIndexFileName))
		if errLoadingIndex != nil {
			logger.Error(errLoadingIndex, "Could not load index file")
			return assumedChartPath, errLoadingIndex
		}

		downloadUrl, errGettingDownloadUrl := h.FindDownloadUrlFromCacheFile(repoCacheFile, chart, version)
		if errGettingDownloadUrl != nil {
			logger.Error(errGettingDownloadUrl, "Could not find a download url")
			return assumedChartPath, errGettingDownloadUrl
		}
		logger.Info(fmt.Sprintf("attempting to download from %s", downloadUrl))

		return assumedChartPath, utils.DownloadFile(assumedChartPath, downloadUrl, username, password)
	}
	return assumedChartPath, nil
}
