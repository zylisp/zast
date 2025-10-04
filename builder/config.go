package builder

// BuilderConfig holds configuration for the AST builder
type Config struct {
	// Strict mode: fail on unknown node types vs. skip them
	StrictMode bool // default: false

	// Maximum nesting depth to prevent stack overflow on malformed input
	MaxNestingDepth int // default: 1000

	// Collect all errors vs. fail on first error
	CollectAllErrors bool // default: true
}

// DefaultBuilderConfig returns the default builder configuration
func DefaultConfig() *Config {
	return &Config{
		StrictMode:       false,
		MaxNestingDepth:  1000, //TODO: make this a const
		CollectAllErrors: true,
	}
}
