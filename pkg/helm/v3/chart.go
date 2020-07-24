package v3

import (
	"fmt"
	"io"
	"net/http"
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
	if errDownloadingChart := downloadFile(assumedChartPath, assumedDownloadUrl, username, password); errDownloadingChart != nil {
		return "", errDownloadingChart
	}
	return assumedChartPath, nil
}

func downloadFile(filepath, url, username, password string) (err error) {

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err

}
