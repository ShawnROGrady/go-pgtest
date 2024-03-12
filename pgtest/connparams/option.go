package connparams

// Option is a parameter to configure the connection.
type Option interface {
	apply(*ConnectionParams)
}

type optionFunc func(*ConnectionParams)

func (f optionFunc) apply(p *ConnectionParams) { f(p) }

// WithUser returns a Option specifying the name of the database to
// connect to.
func WithUser(user string) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setUser(user)
	})
}

// WithPassword returns a Option specifying the user's password.
func WithPassword(password string) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setPassword(password)
	})
}

// WithHost returns a Option specifying Are for unix domain sockets.
func WithHost(host string) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setHost(host)
	})
}

// WithPort returns a Option specifying the port to bind to.
func WithPort(port int) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setPort(port)
	})
}

// WithSSLMode returns a Option specifying whether or not to use SSL.
func WithSSLMode(sslMode string) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setSSLMode(sslMode)
	})
}

// WithFallbackApplicationName returns a Option specifying an
// application_name to fall back to if one isn't provided.
func WithFallbackApplicationName(fallbackApplicationName string) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setFallbackApplicationName(fallbackApplicationName)
	})
}

// WithConnectionTimeout returns a Option specifying maximum wait for
// connection, in seconds. Zero or not specified means wait indefinitely.
func WithConnectionTimeout(connectionTimeout int) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setConnectionTimeout(connectionTimeout)
	})
}

// WithSSLCert returns a Option specifying cert file location. The
// file must contain PEM encoded data.
func WithSSLCert(sslCert string) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setSSLCert(sslCert)
	})
}

// WithSSLKey returns a Option specifying key file location. The file
// must contain PEM encoded data.
func WithSSLKey(sslKey string) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setSSLKey(sslKey)
	})
}

// WithSSLRootCert returns a Option specifying the location of the
// root certificate file. The file must contain PEM encoded data
func WithSSLRootCert(sslRootCert string) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setSSLRootCert(sslRootCert)
	})
}

// WithService returns a Option specifying gSS (Kerberos) service name
// to use when constructing the SPN.
func WithService(service string) Option {
	return optionFunc(func(p *ConnectionParams) {
		p.setService(service)
	})
}
