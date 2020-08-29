package git

type ErrorGitAccessTokenNotFound struct {
	Message string
}

func (e ErrorGitAccessTokenNotFound) Error() string {
	return e.Message
}

type ErrorGitDeployDirNotSpecified struct {
	Message string
}

func (e ErrorGitDeployDirNotSpecified) Error() string {
	return e.Message
}

type ErrorEnvVarNotFound struct {
	Message string
}

func (e ErrorEnvVarNotFound) Error() string {
	return e.Message
}
