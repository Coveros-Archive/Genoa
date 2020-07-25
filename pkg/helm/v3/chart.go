package v3

import (
	"coveros.com/pkg"
	"coveros.com/pkg/utils"
	"fmt"
	"helm.sh/helm/v3/pkg/repo"
	"os"
	"path/filepath"
)

func (h HelmV3) DownloadChart(repoUrl, repoAlias, chart, version, username, password, destDir string) (string, error) {

	_, err := os.Stat(destDir)
	if os.IsNotExist(err) {
		if errMakingDir := os.MkdirAll(destDir, os.ModeDir); errMakingDir != nil {
			helmInfoLogF("Failed to create dir to download chart into: %v", errMakingDir)
			return "", errMakingDir
		}
	}

	// first attempt to assume a download url, so we dont have to look up url in index file
	chartTarballName := fmt.Sprintf("%s-%s.tgz", chart, version)
	assumedChartPath := fmt.Sprintf("%s/%s", destDir, chartTarballName)
	assumedDownloadUrl := fmt.Sprintf("%s/%s", repoUrl, chartTarballName)
	logger.Info(fmt.Sprintf("attempting to download %s chart from %s", chart, repoUrl))

	if errDownloadingChart := utils.DownloadFile(assumedChartPath, assumedDownloadUrl, username, password); errDownloadingChart != nil {

		// assumed url did not work, lookup url specified in index file. Could be slow if a chart entry has too many versions
		assumedIndexFileName := repoAlias + "-index.yaml"
		indexFile := filepath.Join(h.settings.RepositoryCache, assumedIndexFileName)
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			return "", pkg.ErrorHelmRepoNeedsRefresh{Message: fmt.Sprintf("%s repo index file not found, a refresh can help", repoAlias)}
		}
		repoIndexFile, errLoadingIndex := repo.LoadIndexFile(indexFile)
		if errLoadingIndex != nil {
			logger.Error(errLoadingIndex, "Could not load index file")
			return assumedChartPath, errLoadingIndex
		}

		downloadUrl, errGettingDownloadUrl := h.FindDownloadUrlFromCacheFile(repoIndexFile, chart, version)
		if errGettingDownloadUrl != nil {
			logger.Error(errGettingDownloadUrl, "Could not find a download url")
			return assumedChartPath, errGettingDownloadUrl
		}
		logger.Info(fmt.Sprintf("attempting to download chart from index file %s", downloadUrl))

		return assumedChartPath, utils.DownloadFile(assumedChartPath, downloadUrl, username, password)
	}
	return assumedChartPath, nil
}
