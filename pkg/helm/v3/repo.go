package v3

import (
	"coveros.com/pkg"
	"coveros.com/pkg/utils"
	"fmt"
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
				return entry.URLs[0], nil
			}
		}
	}
	return "", pkg.ErrorChartEntryNotFoundInRepoIndex{Message: fmt.Sprintf("%v-%v chart not found in repo index, a refresh might help", chartName, chartVersion)}
}
