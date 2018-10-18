package network

import ()

const (
	tbRaw      = "raw"
	tbMangle   = "mangle"
	tbNat      = "nat"
	tbFilter   = "filter"
	tbSecurity = "security"
)

const (
	chPrerouting  = "PREROUTING"
	chInput       = "INPUT"
	chForward     = "FORWARD"
	chOutput      = "OUTPUT"
	chPostrouting = "POSTROUTING"
)

const (
	tgAccept      = "ACCEPT"
	tgClassify    = "CLASSIFY"
	tgClusterip   = "CLUSTERIP"
	tgConnmark    = "CONNMARK"
	tgConnsecmark = "CONNSECMARK"
	tgDnat        = "DNAT"
	tgDrop        = "DROP"
	tgDscp        = "DSCP"
	tgEcn         = "ECN"
	tgLog         = "LOG"
	tgMark        = "MARK"
	tgMasquerade  = "MASQUERADE"
	tgMirror      = "MIRROR"
	tgNetmap      = "NETMAP"
	tgNfqueue     = "NFQUEUE"
	tgNotrack     = "NOTRACK"
	tgQueue       = "QUEUE"
	tgRedirect    = "REDIRECT"
	tgReject      = "REJECT"
	tgReturn      = "RETURN"
	tgSame        = "SAME"
	tgSecmark     = "SECMARK"
	tgTcpmss      = "TCPMSS"
	tgTos         = "TOS"
	tgTtl         = "TTL"
	tgUlog        = "ULOG"
)
