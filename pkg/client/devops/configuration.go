package devops

// ConfigurationOperator provides APIs for operating devops configuration, like reloading.
type ConfigurationOperator interface {

	// ReloadConfiguration reload devops configuration
	ReloadConfiguration() error
}
