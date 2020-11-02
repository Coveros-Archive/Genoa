package v3

import (
	"fmt"
	"github.com/coveros/genoa/pkg"
	"github.com/coveros/genoa/pkg/utils"
	"helm.sh/helm/v3/pkg/repo"
	"os"
	"path/filepath"
	"strings"
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
	chartTarballName := fmt.Sprintf("%s-%s.tgz", strings.ReplaceAll(chart, "/", "-"), version)
	assumedChartPath := fmt.Sprintf("%s/%s", destDir, chartTarballName)
	assumedDownloadUrl := fmt.Sprintf("%s/%s", repoUrl, chartTarballName)
	if errDownloadingChart := utils.DownloadFile(assumedChartPath, assumedDownloadUrl, username, password); errDownloadingChart != nil {

		assumedIndexFileName := repoAlias + "-index.yaml"
		indexFile := filepath.Join(h.settings.RepositoryCache, assumedIndexFileName)
		// if repo cache file not found, throw an error that indicates to download repo index.
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			return "", pkg.ErrorHelmRepoNeedsRefresh{Message: fmt.Sprintf("%s repo index file not found, a refresh can help", repoAlias)}
		}

		// load index file in memory
		repoIndexFile, errLoadingIndex := repo.LoadIndexFile(indexFile)
		if errLoadingIndex != nil {
			logger.Error(errLoadingIndex, "Could not load index file")
			return assumedChartPath, errLoadingIndex
		}

		// attempt to find the chart version from repo index file
		downloadUrl, errGettingDownloadUrl := h.FindDownloadUrlFromCacheFile(repoIndexFile, chart, version)
		if errGettingDownloadUrl != nil {
			logger.Error(errGettingDownloadUrl, "Could not find a download url")
			return assumedChartPath, errGettingDownloadUrl
		}

		// some download urls are not really urls, so fix that ( based on different registry implementation
		if !strings.HasPrefix(downloadUrl, "https://") {
			downloadUrl = fmt.Sprintf("%s/%s", utils.TrimSuffix(repoUrl, "/"), downloadUrl)
		}
		logger.Info(fmt.Sprintf("attempting to download chart from index url %s", downloadUrl))

		return assumedChartPath, utils.DownloadFile(assumedChartPath, downloadUrl, username, password)
	}
	return assumedChartPath, nil
}
