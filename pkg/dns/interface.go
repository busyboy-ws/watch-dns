package dns

type Dns interface {
	//	query dns
	QueryDns(rr string, domain string) (string, bool)
	UpdateDns(rr string,domain string, ip string) bool
	AddDns(rr string, ip string, domain string) bool
}
