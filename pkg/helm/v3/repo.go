package v3

import (
	"fmt"
	"github.com/coveros/genoa/pkg"
	"github.com/coveros/genoa/pkg/utils"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"os"
	"sort"
	"strings"
)

func (h *HelmV3) GetRepoUrlFromRepoConfig(repoAliasName string) (string, string, string, error) {
	repoFile, errLoadingRepoFile := repo.LoadFile(h.settings.RepositoryConfig)
	if errLoadingRepoFile != nil {
		return "", "", "", errLoadingRepoFile
	}

	for _, eachRepo := range repoFile.Repositories {
		if strings.ToLower(eachRepo.Name) == strings.ToLower(repoAliasName) {
			return utils.TrimSuffix(eachRepo.URL, "/"), eachRepo.Username, eachRepo.Password, nil
		}
	}

	return "", "", "", pkg.ErrorHelmRepoNotFoundInRepoConfig{Message: fmt.Sprintf("%v repo not found repo config, please add it first", repoAliasName)}
}

func (h *HelmV3) FindDownloadUrlFromCacheFile(repoCacheFile *repo.IndexFile, chartName, chartVersion string) (string, error) {
	if chartEntries, chartFound := repoCacheFile.Entries[chartName]; chartFound {
		sort.Slice(chartEntries, func(i, j int) bool {
			return chartEntries[i].Version < chartEntries[j].Version
		})
		idx, err := getIdxOfChartVersionFromChartEntries(chartEntries, chartVersion, 0, len(chartEntries)-1)
		if err != nil {
			return "", err
		}

		return chartEntries[idx].URLs[0], nil
	}
	return "", pkg.ErrorHelmRepoNeedsRefresh{Message: fmt.Sprintf("%v-%v chart not found in repo index, a refresh might help", chartName, chartVersion)}
}

func AddReposFromFile(customRepoFile string) error {
	repoFile, errGettingRepoFile := repo.LoadFile(DefaultEnvSettings().RepositoryConfig)
	if errGettingRepoFile != nil {
		return errGettingRepoFile
	}

	customRepo, errGettingCustomRepoFile := repo.LoadFile(customRepoFile)
	if errGettingCustomRepoFile != nil {
		return errGettingCustomRepoFile
	}

	for _, customRepoEntry := range customRepo.Repositories {
		repoFile.Update(customRepoEntry)
	}
	return repoFile.WriteFile(DefaultEnvSettings().RepositoryConfig, os.ModePerm)
}

func (h *HelmV3) RefreshRepoIndex(repoAlias string) error {
	repoFile, errGettingRepoFile := h.getRepoFile()
	if errGettingRepoFile != nil {
		return errGettingRepoFile
	}

	for _, repoEntry := range repoFile.Repositories {
		if repoEntry.Name == repoAlias {
			newChartRepo, err := repo.NewChartRepository(repoEntry, getter.All(h.settings))
			if err != nil {
				return err
			}
			logger.Info(fmt.Sprintf("refreshing repo index for %s", repoAlias))
			if _, errDownloadingIndexFile := newChartRepo.DownloadIndexFile(); errDownloadingIndexFile != nil {
				return errDownloadingIndexFile
			}
			return nil
		}
	}

	return pkg.ErrorHelmRepoNotFoundInRepoConfig{Message: fmt.Sprintf("%s repo does not exist in repo config file, please add it first", repoAlias)}
}

func (h *HelmV3) getRepoFile() (*repo.File, error) {
	return repo.LoadFile(h.settings.RepositoryConfig)
}

func getIdxOfChartVersionFromChartEntries(slice repo.ChartVersions, lookupVersion string, left, right int) (int, error) {
	if left > right {
		return -1, pkg.ErrorHelmRepoNeedsRefresh{Message: fmt.Sprintf("chart not found in repo index, a refresh might help")}
	}
	mid := left + (right-left)/2
	if slice[mid].Version == lookupVersion {
		return mid, nil
	} else if lookupVersion < slice[mid].Version {
		return getIdxOfChartVersionFromChartEntries(slice, lookupVersion, left, mid-1)
	}

	return getIdxOfChartVersionFromChartEntries(slice, lookupVersion, mid+1, right)
}
