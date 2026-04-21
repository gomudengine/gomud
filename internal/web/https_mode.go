package web

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
)

type httpsMode string

const (
	httpsModeDisabled httpsMode = "disabled"
	httpsModeHTTPOnly httpsMode = "http-only"
	httpsModeManual   httpsMode = "manual"
	httpsModeAuto     httpsMode = "auto"
)

type httpsPlan struct {
	mode              httpsMode
	host              string
	redirectPort      int
	cacheDir          string
	certFile          string
	keyFile           string
	email             string
	redirectEnabled   bool
	fallbackReason    string
	emailNoticeNeeded bool
}

type HTTPSStatus struct {
	Mode            string
	Summary         string
	Host            string
	HttpPort        int
	HttpsPort       int
	HttpEnabled     bool
	HttpsEnabled    bool
	RedirectEnabled bool
	CacheDir        string
	CertFile        string
	KeyFile         string
	EmailConfigured bool
	Checks          []string
	NextSteps       []string
	LastError       string
	CertificateHost string
	CertificateDNS  []string
	Issuer          string
	ExpiresAt       string
	DaysRemaining   int
}

var (
	httpsStatusLock sync.RWMutex
	httpsStatus     = HTTPSStatus{
		Mode:      "disabled",
		Summary:   "HTTPS is disabled.",
		Checks:    []string{},
		NextSteps: []string{},
	}
)

func resolveHTTPSPlan(network configs.Network, filePaths configs.FilePaths) httpsPlan {
	plan := httpsPlan{
		mode:            httpsModeDisabled,
		redirectPort:    int(network.HttpsPort),
		redirectEnabled: bool(network.HttpsRedirect),
		certFile:        filePaths.HttpsCertFile.String(),
		keyFile:         filePaths.HttpsKeyFile.String(),
		cacheDir:        filepath.Clean(filePaths.HttpsCacheDir.String()),
		email:           configs.GetSecret(filePaths.HttpsEmail),
		host:            normalizeHTTPSHost(filePaths.WebDomain.String()),
	}

	if network.HttpsPort <= 0 {
		if network.HttpPort > 0 {
			plan.mode = httpsModeHTTPOnly
		}
		return plan
	}

	hasManualCert := plan.certFile != ""
	hasManualKey := plan.keyFile != ""
	if hasManualCert && hasManualKey {
		plan.mode = httpsModeManual
		return plan
	}

	if hasManualCert != hasManualKey {
		plan.mode = fallbackHTTPMode(network)
		plan.fallbackReason = "manual HTTPS requires both HttpsCertFile and HttpsKeyFile"
		return plan
	}

	if plan.host == "" {
		plan.mode = fallbackHTTPMode(network)
		plan.fallbackReason = "automatic HTTPS requires FilePaths.WebDomain to be set to a public hostname"
		return plan
	}

	if !isPublicACMEHost(plan.host) {
		plan.mode = fallbackHTTPMode(network)
		plan.fallbackReason = fmt.Sprintf("automatic HTTPS requires a public hostname, got %q", plan.host)
		return plan
	}

	if network.HttpPort != 80 || network.HttpsPort != 443 {
		plan.mode = fallbackHTTPMode(network)
		plan.fallbackReason = fmt.Sprintf("automatic HTTPS requires Network.HttpPort=80 and Network.HttpsPort=443, got %d/%d", network.HttpPort, network.HttpsPort)
		return plan
	}

	plan.mode = httpsModeAuto
	plan.emailNoticeNeeded = plan.email == ""
	return plan
}

func fallbackHTTPMode(network configs.Network) httpsMode {
	if network.HttpPort > 0 {
		return httpsModeHTTPOnly
	}
	return httpsModeDisabled
}

func normalizeHTTPSHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimSuffix(host, ".")
	return host
}

func isPublicACMEHost(host string) bool {
	if host == "" {
		return false
	}

	switch host {
	case "localhost", "localhost.localdomain":
		return false
	}

	if ip := net.ParseIP(host); ip != nil {
		return false
	}

	if strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".localhost") {
		return false
	}

	return strings.Contains(host, ".")
}

func validateAutoHTTPSServerName(configuredHost string, requestedHost string) error {
	configuredHost = normalizeHTTPSHost(configuredHost)
	requestedHost = normalizeHTTPSHost(requestedHost)
	if requestedHost == configuredHost {
		return nil
	}

	if requestedHost == "" || !isPublicACMEHost(requestedHost) {
		return fmt.Errorf("automatic HTTPS only serves %q; use the public domain instead of localhost or an IP address", configuredHost)
	}

	return fmt.Errorf("automatic HTTPS only serves %q; got TLS request for %q", configuredHost, requestedHost)
}

type automaticHTTPSLogFilter struct{}

func (automaticHTTPSLogFilter) Write(p []byte) (int, error) {
	if isExpectedAutomaticHTTPSTLSLog(string(p)) {
		return len(p), nil
	}

	if _, err := os.Stderr.Write(p); err != nil {
		return 0, err
	}

	return len(p), nil
}

func isExpectedAutomaticHTTPSTLSLog(message string) bool {
	return strings.Contains(message, "automatic HTTPS only serves")
}

func buildHTTPSRedirectTarget(host string, httpsPort int, requestURI string) string {
	if strings.Contains(host, ":") {
		if splitHost, _, err := net.SplitHostPort(host); err == nil {
			host = splitHost
		}
	}

	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = "[" + host + "]"
	}

	if httpsPort == 443 {
		return fmt.Sprintf("https://%s%s", host, requestURI)
	}

	return fmt.Sprintf("https://%s:%d%s", host, httpsPort, requestURI)
}

func newHTTPSStatus(plan httpsPlan, network configs.Network) HTTPSStatus {
	status := HTTPSStatus{
		Mode:            string(plan.mode),
		Host:            plan.host,
		HttpPort:        int(network.HttpPort),
		HttpsPort:       int(network.HttpsPort),
		HttpEnabled:     network.HttpPort > 0,
		HttpsEnabled:    plan.mode == httpsModeManual || plan.mode == httpsModeAuto,
		RedirectEnabled: false,
		CacheDir:        plan.cacheDir,
		CertFile:        plan.certFile,
		KeyFile:         plan.keyFile,
		EmailConfigured: plan.email != "",
		Checks:          []string{},
		NextSteps:       []string{},
	}

	switch plan.mode {
	case httpsModeManual:
		status.Summary = "Manual HTTPS is enabled using the configured certificate files."
		status.Checks = append(status.Checks,
			fmt.Sprintf("HTTPS will listen on port %d.", network.HttpsPort),
			fmt.Sprintf("Manual certificate file: %s", plan.certFile),
			fmt.Sprintf("Manual private key file: %s", plan.keyFile),
		)
	case httpsModeAuto:
		status.Summary = "Automatic HTTPS is enabled. GoMud will request and renew a Let's Encrypt certificate for this hostname."
		status.Checks = append(status.Checks,
			fmt.Sprintf("Automatic HTTPS hostname: %s", plan.host),
			fmt.Sprintf("HTTP challenge port: %d", network.HttpPort),
			fmt.Sprintf("HTTPS serving port: %d", network.HttpsPort),
			fmt.Sprintf("Certificate cache directory: %s", plan.cacheDir),
		)
		status.NextSteps = append(status.NextSteps,
			fmt.Sprintf("Point DNS for %s at this server.", plan.host),
			"Make sure inbound ports 80 and 443 are open to the internet.",
		)
		if !status.EmailConfigured {
			status.NextSteps = append(status.NextSteps, "Optionally set FilePaths.HttpsEmail to receive certificate expiry notices.")
		}
	case httpsModeHTTPOnly:
		status.Summary = "GoMud is serving HTTP only."
		if plan.fallbackReason != "" {
			status.Checks = append(status.Checks, plan.fallbackReason)
		}
		if plan.host == "localhost" || plan.host == "localhost.localdomain" {
			status.Checks = append(status.Checks, "Localhost is treated as development mode, so automatic HTTPS is skipped on purpose.")
		}
		status.NextSteps = append(status.NextSteps, defaultHTTPSGuidance(plan, network)...)
	default:
		status.Summary = "HTTPS is disabled."
		if plan.fallbackReason != "" {
			status.Checks = append(status.Checks, plan.fallbackReason)
		}
		if plan.host == "localhost" || plan.host == "localhost.localdomain" {
			status.Checks = append(status.Checks, "Localhost is treated as development mode, so automatic HTTPS is skipped on purpose.")
		}
		status.NextSteps = append(status.NextSteps, defaultHTTPSGuidance(plan, network)...)
	}

	return status
}

func defaultHTTPSGuidance(plan httpsPlan, network configs.Network) []string {
	steps := []string{}

	if network.HttpsPort <= 0 {
		steps = append(steps, "Set Network.HttpsPort to 443 to enable HTTPS.")
	}

	if plan.host == "" {
		steps = append(steps, "Set FilePaths.WebDomain to the public hostname players will use.")
	}

	if plan.host != "" && !isPublicACMEHost(plan.host) {
		steps = append(steps, "Use a public DNS hostname instead of localhost, a private-only name, or a raw IP address.")
	}

	if network.HttpPort <= 0 && network.HttpsPort > 0 {
		steps = append(steps, "Set Network.HttpPort to 80 so GoMud can serve HTTP or complete automatic HTTPS challenges.")
	}

	if network.HttpPort != 80 || network.HttpsPort != 443 {
		steps = append(steps, "Automatic HTTPS requires Network.HttpPort=80 and Network.HttpsPort=443.")
	}

	if plan.certFile != "" || plan.keyFile != "" {
		steps = append(steps, "Set both HttpsCertFile and HttpsKeyFile for manual HTTPS, or clear both to let GoMud use Let's Encrypt.")
	}

	if len(steps) == 0 {
		steps = append(steps, "Review the HTTPS checks below for the exact reason automatic HTTPS is not active.")
	}

	return steps
}

func markHTTPSStartupFailure(status *HTTPSStatus, err error) {
	if err == nil {
		return
	}

	status.LastError = err.Error()
	status.HttpsEnabled = false
	status.RedirectEnabled = false

	switch status.Mode {
	case string(httpsModeManual):
		status.Summary = "Manual HTTPS is configured, but the HTTPS listener is unavailable because startup failed."
	case string(httpsModeAuto):
		status.Summary = "Automatic HTTPS is configured, but the HTTPS listener is unavailable because startup failed."
	default:
		status.Summary = "HTTPS is unavailable because startup failed."
	}

	startupCheck := "HTTPS startup failed, so GoMud is not currently serving HTTPS."
	if !containsString(status.Checks, startupCheck) {
		status.Checks = append(status.Checks, startupCheck)
	}
}

func markHTTPSListenerReady(status *HTTPSStatus, redirectConfigured bool) {
	if status == nil {
		return
	}

	if redirectConfigured && (status.Mode == string(httpsModeManual) || status.Mode == string(httpsModeAuto)) {
		status.RedirectEnabled = true
	}
}

func markAutoHTTPSHTTPFailure(status *HTTPSStatus, err error) {
	if err == nil {
		return
	}

	status.LastError = err.Error()
	status.RedirectEnabled = false
	status.Summary = "Automatic HTTPS is serving cached certificates, but the required HTTP listener is unavailable for ACME challenges and redirects."

	startupCheck := "Automatic HTTPS requires a working HTTP listener for ACME challenges and redirects."
	if !containsString(status.Checks, startupCheck) {
		status.Checks = append(status.Checks, startupCheck)
	}
}

func SetHTTPSStatus(status HTTPSStatus) {
	httpsStatusLock.Lock()
	defer httpsStatusLock.Unlock()
	httpsStatus = status
}

func UpdateHTTPSStatus(mutator func(*HTTPSStatus)) {
	httpsStatusLock.Lock()
	defer httpsStatusLock.Unlock()
	mutator(&httpsStatus)
}

func GetHTTPSStatus() HTTPSStatus {
	httpsStatusLock.RLock()
	defer httpsStatusLock.RUnlock()

	status := httpsStatus
	status.Checks = append([]string{}, status.Checks...)
	status.NextSteps = append([]string{}, status.NextSteps...)
	return status
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func httpsDiagnosticHost(host string) string {
	host = normalizeHTTPSHost(host)
	if isPublicACMEHost(host) {
		return host
	}
	return "play.example.com"
}

func httpsUsesExampleHost(host string) bool {
	return httpsDiagnosticHost(host) == "play.example.com" && !isPublicACMEHost(normalizeHTTPSHost(host))
}

func runAutoHTTPSPreflight(status *HTTPSStatus) {
	if status.Host == "" {
		return
	}

	localIPs := localInterfaceIPs()
	if len(localIPs) > 0 {
		status.Checks = append(status.Checks, fmt.Sprintf("Local interface IPs: %s", strings.Join(localIPs, ", ")))
	}

	addrs, err := net.LookupHost(status.Host)
	if err != nil {
		status.Checks = append(status.Checks, fmt.Sprintf("DNS lookup failed for %s: %v", status.Host, err))
		status.NextSteps = append(status.NextSteps, fmt.Sprintf("Create an A or AAAA record for %s before expecting Let's Encrypt to succeed.", status.Host))
		return
	}

	appendAutoHTTPSDNSChecks(status, addrs, localIPs)
}

func appendAutoHTTPSDNSChecks(status *HTTPSStatus, addrs []string, localIPs []string) {
	if status == nil {
		return
	}

	status.Checks = append(status.Checks, fmt.Sprintf("DNS lookup for %s returned: %s", status.Host, strings.Join(addrs, ", ")))
	if len(localIPs) > 0 && sharesAnyAddress(addrs, localIPs) {
		status.Checks = append(status.Checks, "DNS includes a local interface address on this machine.")
	}
}

func logHTTPSStatus(status HTTPSStatus) {
	fields := []any{
		"mode", status.Mode,
		"summary", status.Summary,
		"httpPort", status.HttpPort,
		"httpsPort", status.HttpsPort,
	}

	if status.Host != "" {
		fields = append(fields, "host", status.Host)
	}

	mudlog.Info("HTTPS", fields...)

	for _, check := range status.Checks {
		mudlog.Info("HTTPS", "check", check)
	}

	for _, step := range status.NextSteps {
		mudlog.Info("HTTPS", "next", step)
	}
}

func localInterfaceIPs() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	seen := map[string]struct{}{}
	ips := []string{}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil || ip == nil {
				continue
			}

			ip = normalizeComparableIP(ip)
			if ip == nil || ip.IsLoopback() {
				continue
			}

			ipStr := ip.String()
			if _, ok := seen[ipStr]; ok {
				continue
			}
			seen[ipStr] = struct{}{}
			ips = append(ips, ipStr)
		}
	}

	return ips
}

func sharesAnyAddress(resolved []string, local []string) bool {
	localSet := map[string]struct{}{}
	for _, ipStr := range local {
		if ip := normalizeComparableIP(net.ParseIP(ipStr)); ip != nil {
			localSet[ip.String()] = struct{}{}
		}
	}

	for _, ipStr := range resolved {
		if ip := normalizeComparableIP(net.ParseIP(ipStr)); ip != nil {
			if _, ok := localSet[ip.String()]; ok {
				return true
			}
		}
	}

	return false
}

func normalizeComparableIP(ip net.IP) net.IP {
	if ip == nil {
		return nil
	}
	if v4 := ip.To4(); v4 != nil {
		return v4
	}
	return ip.To16()
}

func setCertificateInfo(status *HTTPSStatus, leaf *x509.Certificate) {
	if leaf == nil {
		return
	}

	status.CertificateHost = leaf.Subject.CommonName
	status.CertificateDNS = append([]string{}, leaf.DNSNames...)
	status.Issuer = leaf.Issuer.CommonName
	if status.Issuer == "" {
		status.Issuer = leaf.Issuer.String()
	}
	status.ExpiresAt = leaf.NotAfter.Format(time.RFC3339)
	status.DaysRemaining = int(time.Until(leaf.NotAfter).Hours() / 24)
}

func firstLeafCertificate(cert *tls.Certificate) (*x509.Certificate, error) {
	if cert == nil {
		return nil, nil
	}
	if cert.Leaf != nil {
		return cert.Leaf, nil
	}
	if len(cert.Certificate) == 0 {
		return nil, nil
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}
	return leaf, nil
}

func describeListenError(port int, err error) []string {
	steps := []string{}
	if err == nil {
		return steps
	}

	switch {
	case isErrno(err, syscall.EADDRINUSE):
		steps = append(steps,
			fmt.Sprintf("Port %d is already in use by another process.", port),
			fmt.Sprintf("Run `ss -ltnp | rg ':%d'` to find the conflicting process.", port),
		)
	case isErrno(err, syscall.EACCES):
		steps = append(steps,
			fmt.Sprintf("GoMud does not have permission to bind port %d.", port),
			"Use a privileged launch method, grant the binary bind capability, or choose a higher port.",
		)
	}

	return steps
}

func isErrno(err error, target syscall.Errno) bool {
	if err == nil {
		return false
	}
	if opErr, ok := err.(*net.OpError); ok {
		err = opErr.Err
	}
	if sysErr, ok := err.(*os.SyscallError); ok {
		err = sysErr.Err
	}
	return err == target
}
