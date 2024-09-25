package filegunner

type RunModeEnum string

const (
	Email  RunModeEnum = "email"
	DryRun RunModeEnum = "dryrun"
)

type Configuration struct {
	Services           map[string]MailgunService `yaml:"services"`
	RunMode            RunModeEnum               `yaml:"runMode"`
	HistoryDir         *string                   `yaml:"historyDir"`
	InputDir           string                    `yaml:"inputDir"`
	VerboseFileWatcher bool                      `yaml:"verboseFileWatcher"`
}

type MailgunService struct {
	Url    string `yaml:"url"`
	ApiKey string `yaml:"apiKey"`
}
