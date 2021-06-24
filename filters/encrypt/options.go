package encrypt

// getOpts - iterate the inbound Options and return a struct.
func getOpts(opt ...Option) options {
	opts := getDefaultOptions()
	for _, o := range opt {
		if o != nil {
			o(&opts)
		}
	}
	return opts
}

// Option - how Options are passed as arguments.
type Option func(*options)

// options = how options are represented
type options struct {
	withFilterOperations map[DataClassification]FilterOperation
}

func getDefaultOptions() options {
	return options{}
}

func withFilterOperations(ops map[DataClassification]FilterOperation) Option {
	return func(o *options) {
		o.withFilterOperations = ops
	}
}
