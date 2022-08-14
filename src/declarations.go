package main

type BackupDestinationConfig struct {
	Local BackupDestinationLocal
	AWS   BackupDestinationAws
}

var DestinationList = []string{"aws", "local"}

type BackupTargetConfig struct {
	Type                      string   `mapstructure:"type"`
	Path                      string   `mapstructure:"path"`
	Format                    string   `mapstructure:"format"`
	Frequency                 string   `mapstructure:"frequency"`
	Cron                      string   `mapstructure:"cron"`
	Replace                   bool     `mapstructure:"replace"`
	DateSuffix                bool     `mapstructure:"date_suffix"`
	ExcludeVcs                bool     `mapstructure:"exclude_vcs"`
	Destinations              []string `mapstructure:"destinations"`
	ExcludeDirs               []string `mapstructure:"exclude_dirs"`
	PreserveAbsoluteHierarchy bool     `mapstructure:"preserve_absolute_hierarchy"`
}

type BackupDestination interface {
	runBackup(*BackupTarget)
	getName() string
}

type BackupTarget struct {
	Name              string
	TmpWorkdir        string
	Archive           string
	Files             []string
	Config            BackupTargetConfig
	DestinationConfig []BackupDestination
}
