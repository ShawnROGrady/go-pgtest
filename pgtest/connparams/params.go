package connparams

import (
	"net/url"
	"os"
	"strconv"
)

// The possible SSL modes.
const (
	// No SSL.
	SSLModeDisable = "disable"

	// Always SSL (skip verification).
	SSLModeRequire = "require"

	// Always SSL (verify that the certificate presented by the server was
	// signed by a trusted CA)
	SSLModeVerifyCA = "verify-ca"

	// Always SSL (verify that the certification presented by the server
	// was signed by a trusted CA and the server host name matches the one
	// in the certificate)
	SSLModeVerifyFull = "verify-full"
)

const (
	defaultHost    = "localhost"
	defaultPort    = 5432
	defaultSSLMode = SSLModeDisable
)

const (
	userSet uint = 1 << iota
	passwordSet
	hostSet
	portSet
	sslModeSet
	fallbackApplicationNameSet
	connectionTimeoutSet
	sslCertSet
	sslKeySet
	sslRootCertSet
	serviceSet
)

type ConnectionParams struct {
	dbName                  string
	user                    string
	password                string
	host                    string
	port                    int
	sslMode                 string
	fallbackApplicationName string
	connectionTimeout       int
	sslCert                 string
	sslKey                  string
	sslRootCert             string
	service                 string

	set uint
}

// New constructs a new ConnectionParams for the specified dbName.
func New(dbName string, opts ...Option) *ConnectionParams {
	p := &ConnectionParams{dbName: dbName}
	for _, opt := range opts {
		opt.apply(p)
	}

	return p
}

type Factory func(dbName string, opts ...Option) *ConnectionParams

func DefaultFactory() Factory {
	return func(dbName string, opts ...Option) *ConnectionParams {
		p := new(ConnectionParams)
		p.setHost(defaultHost)
		p.setPort(defaultPort)
		p.setSSLMode(defaultSSLMode)

		if u := os.Getenv("USER"); u != "" {
			p.setUser(u)
		}

		p.dbName = dbName
		for _, opt := range opts {
			opt.apply(p)
		}

		return p
	}
}

// NewWithDefaults returns a ConnectionParams for the specified dbName with
// default params set.
func NewWithDefaults(dbName string, opts ...Option) *ConnectionParams {
	return DefaultFactory()(dbName, opts...)
}

func (p *ConnectionParams) DBName() string {
	return p.dbName
}

func (p *ConnectionParams) URI() *url.URL {
	uri := &url.URL{
		Scheme: "postgres",
		Path:   p.DBName(),
	}

	if (p.set & hostSet) != 0 {
		uri.Host = p.host
		if (p.set & portSet) != 0 {
			uri.Host = uri.Host + ":" + strconv.Itoa(p.port)
		}
	}

	if (p.set & userSet) != 0 {
		if (p.set & passwordSet) != 0 {
			uri.User = url.UserPassword(p.user, p.password)
		} else {
			uri.User = url.User(p.user)
		}
	}

	q := url.Values{}

	// NOTE: I'm assuming there will be an error actually connecting here,
	// but since postgres allows us to specify heirarchical uri params in
	// the query we should at least attempt.
	if (p.set&portSet) != 0 && (p.set&hostSet) == 0 {
		q.Set("port", strconv.Itoa(p.port))
	}

	if (p.set & sslModeSet) != 0 {
		q.Set("sslmode", p.sslMode)
	}

	if (p.set & fallbackApplicationNameSet) != 0 {
		q.Set("fallback_application_name", p.fallbackApplicationName)
	}

	if (p.set & connectionTimeoutSet) != 0 {
		q.Set("connection_timeout", strconv.Itoa(p.connectionTimeout))
	}

	if (p.set & sslCertSet) != 0 {
		q.Set("sslcert", p.sslCert)
	}

	if (p.set & sslKeySet) != 0 {
		q.Set("sslkey", p.sslKey)
	}

	if (p.set & sslRootCertSet) != 0 {
		q.Set("sslrootcert", p.sslRootCert)
	}

	if (p.set & serviceSet) != 0 {
		q.Set("service", p.service)
	}

	uri.RawQuery = q.Encode()

	return uri
}

func (p *ConnectionParams) setUser(user string) {
	p.user = user
	p.set |= userSet
}

func (p *ConnectionParams) setPassword(password string) {
	p.password = password
	p.set |= passwordSet
}

func (p *ConnectionParams) setHost(host string) {
	p.host = host
	p.set |= hostSet
}

func (p *ConnectionParams) setPort(port int) {
	p.port = port
	p.set |= portSet
}

func (p *ConnectionParams) setSSLMode(sslMode string) {
	p.sslMode = sslMode
	p.set |= sslModeSet
}

func (p *ConnectionParams) setFallbackApplicationName(fallbackApplicationName string) {
	p.fallbackApplicationName = fallbackApplicationName
	p.set |= fallbackApplicationNameSet
}

func (p *ConnectionParams) setConnectionTimeout(connectionTimeout int) {
	p.connectionTimeout = connectionTimeout
	p.set |= connectionTimeoutSet
}

func (p *ConnectionParams) setSSLCert(sslCert string) {
	p.sslCert = sslCert
	p.set |= sslCertSet
}

func (p *ConnectionParams) setSSLKey(sslKey string) {
	p.sslKey = sslKey
	p.set |= sslKeySet
}

func (p *ConnectionParams) setSSLRootCert(sslRootCert string) {
	p.sslRootCert = sslRootCert
	p.set |= sslRootCertSet
}

func (p *ConnectionParams) setService(service string) {
	p.service = service
	p.set |= serviceSet
}
