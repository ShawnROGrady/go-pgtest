package connparams

import (
	"net/url"
	"os"
	"strconv"
	"strings"
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
func New(dbName string, opts ...Option) ConnectionParams {
	p := ConnectionParams{dbName: dbName}
	for _, opt := range opts {
		opt.apply(&p)
	}

	return p
}

type Factory func(dbName string, opts ...Option) ConnectionParams

func DefaultFactory() Factory {
	return func(dbName string, opts ...Option) ConnectionParams {
		var p ConnectionParams
		p.setHost(defaultHost)
		p.setPort(defaultPort)
		p.setSSLMode(defaultSSLMode)

		if u := os.Getenv("USER"); u != "" {
			p.setUser(u)
		}

		p.dbName = dbName
		for _, opt := range opts {
			opt.apply(&p)
		}

		return p
	}
}

// NewWithDefaults returns a ConnectionParams for the specified dbName with
// default params set.
func NewWithDefaults(dbName string, opts ...Option) ConnectionParams {
	return DefaultFactory()(dbName, opts...)
}

func (p ConnectionParams) DBName() string {
	return p.dbName
}

func (p ConnectionParams) URI() *url.URL {
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

type KeyValue struct {
	Key   string
	Value string
}

func (kv KeyValue) String() string {
	return kv.Key + "=" + kv.Value
}

func normalizeValue(v string) string {
	if v == "" {
		return `''`
	}

	var (
		b        strings.Builder
		hasSpace bool
	)

	b.Grow(len(v))

	for _, r := range v {
		switch r {
		case '\\', '\'':
			b.WriteRune('\\')
		case ' ':
			hasSpace = true
		default:
		}
		b.WriteRune(r)
	}

	if hasSpace {
		return "'" + b.String() + "'"
	}

	return b.String()
}

func keyValue(k, v string) KeyValue {
	return KeyValue{Key: k, Value: v}
}

type KeyValues []KeyValue

func (kvs KeyValues) String() string {
	parts := make([]string, len(kvs))
	for i, kv := range kvs {
		parts[i] = kv.String()
	}

	return strings.Join(parts, " ")
}

func (p ConnectionParams) KeyValues() KeyValues {
	kvs := []KeyValue{{
		Key:   "dbname",
		Value: normalizeValue(p.dbName),
	}}

	if host, ok := p.getHost(); ok {
		kvs = append(kvs, keyValue("host", normalizeValue(host)))
	}

	if port, ok := p.getPort(); ok {
		kvs = append(kvs, keyValue("port", strconv.Itoa(port)))
	}

	if u, ok := p.getUser(); ok {
		kvs = append(kvs, keyValue("user", normalizeValue(u)))
	}

	if pass, ok := p.getPassword(); ok {
		kvs = append(kvs, keyValue("password", normalizeValue(pass)))
	}

	if mode, ok := p.getSSLMode(); ok {
		kvs = append(kvs, keyValue("sslmode", normalizeValue(mode)))
	}

	if n, ok := p.getFallbackApplicationName(); ok {
		kvs = append(kvs, keyValue("fallbackApplicationName", normalizeValue(n)))
	}

	if t, ok := p.getConnectionTimeout(); ok {
		kvs = append(kvs, keyValue("connection_timeout", strconv.Itoa(t)))
	}

	if c, ok := p.getSSLCert(); ok {
		kvs = append(kvs, keyValue("sslcert", normalizeValue(c)))
	}

	if k, ok := p.getSSLKey(); ok {
		kvs = append(kvs, keyValue("sslkey", normalizeValue(k)))
	}

	if c, ok := p.getSSLRootCert(); ok {
		kvs = append(kvs, keyValue("sslrootcert", normalizeValue(c)))
	}

	if s, ok := p.getService(); ok {
		kvs = append(kvs, keyValue("service", normalizeValue(s)))
	}

	return kvs
}

// String returns a string representation of the connection params. Although
// this can be used as a connection string, it's main purpose is debugging
// and testing. Either 'p.URI().String()' or 'p.KeyValues().String()' should
// be used to construct the connection string.
func (p ConnectionParams) String() string {
	return p.KeyValues().String()
}

func (p ConnectionParams) Equal(other ConnectionParams) bool {
	if p.set != other.set || p.dbName != other.dbName {
		return false
	}

	if u, ok := p.getUser(); ok && !other.hasUserEqual(u) {
		return false
	}

	if x, ok := p.getPassword(); ok && !other.hasPasswordEqual(x) {
		return false
	}

	if x, ok := p.getHost(); ok && !other.hasHostEqual(x) {
		return false
	}

	if x, ok := p.getPort(); ok && !other.hasPortEqual(x) {
		return false
	}

	if x, ok := p.getSSLMode(); ok && !other.hasSSLModeEqual(x) {
		return false
	}

	if x, ok := p.getFallbackApplicationName(); ok && !other.hasFallbackApplicationNameEqual(x) {
		return false
	}

	if x, ok := p.getConnectionTimeout(); ok && !other.hasConnectionTimeoutEqual(x) {
		return false
	}

	if x, ok := p.getSSLCert(); ok && !other.hasSSLCertEqual(x) {
		return false
	}

	if x, ok := p.getSSLKey(); ok && !other.hasSSLKeyEqual(x) {
		return false
	}

	if x, ok := p.getSSLRootCert(); ok && !other.hasSSLRootCertEqual(x) {
		return false
	}

	if x, ok := p.getService(); ok && !other.hasServiceEqual(x) {
		return false
	}

	return true
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

func (p ConnectionParams) getUser() (string, bool) {
	return p.user, (p.set & userSet) != 0
}

func (p ConnectionParams) getPassword() (string, bool) {
	return p.password, (p.set & passwordSet) != 0
}

func (p ConnectionParams) getHost() (string, bool) {
	return p.host, (p.set & hostSet) != 0
}

func (p ConnectionParams) getPort() (int, bool) {
	return p.port, (p.set & portSet) != 0
}

func (p ConnectionParams) getSSLMode() (string, bool) {
	return p.sslMode, (p.set & sslModeSet) != 0
}

func (p ConnectionParams) getFallbackApplicationName() (string, bool) {
	return p.fallbackApplicationName, (p.set & fallbackApplicationNameSet) != 0
}

func (p ConnectionParams) getConnectionTimeout() (int, bool) {
	return p.connectionTimeout, (p.set & connectionTimeoutSet) != 0
}

func (p ConnectionParams) getSSLCert() (string, bool) {
	return p.sslCert, (p.set & sslCertSet) != 0
}

func (p ConnectionParams) getSSLKey() (string, bool) {
	return p.sslKey, (p.set & sslKeySet) != 0
}

func (p ConnectionParams) getSSLRootCert() (string, bool) {
	return p.sslRootCert, (p.set & sslRootCertSet) != 0
}

func (p ConnectionParams) getService() (string, bool) {
	return p.service, (p.set & serviceSet) != 0
}

func (p ConnectionParams) hasUserEqual(user string) bool {
	return (p.set&userSet) != 0 && p.user == user
}

func (p ConnectionParams) hasPasswordEqual(password string) bool {
	return (p.set&passwordSet) != 0 && p.password == password
}

func (p ConnectionParams) hasHostEqual(host string) bool {
	return (p.set&hostSet) != 0 && p.host == host
}

func (p ConnectionParams) hasPortEqual(port int) bool {
	return (p.set&portSet) != 0 && p.port == port
}

func (p ConnectionParams) hasSSLModeEqual(sslMode string) bool {
	return (p.set&sslModeSet) != 0 && p.sslMode == sslMode
}

func (p ConnectionParams) hasFallbackApplicationNameEqual(fallbackApplicationName string) bool {
	return (p.set&fallbackApplicationNameSet) != 0 && p.fallbackApplicationName == fallbackApplicationName
}

func (p ConnectionParams) hasConnectionTimeoutEqual(connectionTimeout int) bool {
	return (p.set&connectionTimeoutSet) != 0 && p.connectionTimeout == connectionTimeout
}

func (p ConnectionParams) hasSSLCertEqual(sslCert string) bool {
	return (p.set&sslCertSet) != 0 && p.sslCert == sslCert
}

func (p ConnectionParams) hasSSLKeyEqual(sslKey string) bool {
	return (p.set&sslKeySet) != 0 && p.sslKey == sslKey
}

func (p ConnectionParams) hasSSLRootCertEqual(sslRootCert string) bool {
	return (p.set&sslRootCertSet) != 0 && p.sslRootCert == sslRootCert
}

func (p ConnectionParams) hasServiceEqual(service string) bool {
	return (p.set&serviceSet) != 0 && p.service == service
}
