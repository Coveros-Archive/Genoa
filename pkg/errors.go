package pkg

type ErrorHelmRepoNeedsRefresh struct {
	Message string
}

func (e ErrorHelmRepoNeedsRefresh) Error() string {
	return e.Message
}

type ErrorHelmRepoNotFoundInRepoConfig struct {
	Message string
}

func (e ErrorHelmRepoNotFoundInRepoConfig) Error() string {
	return e.Message
}

type ErrorInvalidChartDownloadUrl struct {
	Message string
}

func (e ErrorInvalidChartDownloadUrl) Error() string {
	return e.Message
}
