package writer

// Config holds configuration for the AST writer
type Config struct {
	// Maximum position to scan when collecting files from FileSet
	MaxFileSetScan int // default: 1000000

	// Early termination: stop scanning after this many positions with no new files
	FileSearchWindow int // default: 10000
}

// DefaultConfig returns the default writer configuration
func DefaultConfig() *Config {
	return &Config{
		MaxFileSetScan:   1000000, //TODO: make these consts
		FileSearchWindow: 10000,   //TODO: make these consts
	}
}
