package main

import "time"

const (
	_authHeader             = "Authorization"
	_authUrl                = "https://api.fastmail.com/.well-known/jmap"
	_configDirPerm          = 0o700
	_configFilePerm         = 0o600
	_configFilePermReadOnly = 0o400
	_contentTypeHeader      = "Content-Type"
	_fastmaskRequestId      = "fastmask"
	_httpTimeout            = 5 * time.Second
	_jmapCallId             = "0"
	_jmapStateEnabled       = "enabled"
	_jsonContentType        = "application/json"
	_maskedEmailCapability  = "https://www.fastmail.com/dev/maskedemail"
	_maskedEmailSetMethod   = "MaskedEmail/set"
	_prefixLen              = 5
	_primaryAccountKey      = "urn:ietf:params:jmap:core"
	_programName            = "fastmask"
)
