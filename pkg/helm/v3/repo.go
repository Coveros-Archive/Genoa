package v3

import (
	"coveros.com/pkg"
	"coveros.com/pkg/utils"
	"fmt"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
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
		for _, entry := range chartEntries {
			if entry.Version == chartVersion {
				if len(entry.URLs) == 0 {
					return "", pkg.ErrorInvalidChartDownloadUrl{Message: fmt.Sprintf("%v-%v has no urls in cache file.. do not know what to do anymore", chartName, chartVersion)}
				}
				return entry.URLs[0], nil
			}
		}
	}
	return "", pkg.ErrorHelmRepoNeedsRefresh{Message: fmt.Sprintf("%v-%v chart not found in repo index, a refresh might help", chartName, chartVersion)}
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
