# Whitelist/Blacklist Go Service Middleware

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go/listauth)

Middleware supporting Whitelist/Blacklist authentication using Distinguished Names.

## Distinguished Name (DN)

DNs are comprised of zero or more comma-separated components called relative distinguished names, or RDNs. For example, the DN `uid=john.doe,ou=People,dc=example,dc=com` has four RDNs:

* uid=john.doe
* ou=People
* dc=example
* dc=com

## Whitelist/Blacklist

The lists are constructed from JSON content:

```json
{
  "userBlacklist": [
    "cn=alec.holmes,dc=deciphernow,dc=com",
    "cn=doug.fort,dc=deciphernow,dc=com"
  ],
  "userWhitelist": [
    "dc=deciphernow,dc=com"
  ]
}
```

Where the above would allow all users in `dc=deciphernow,dc=com` except `alec.holmes` and `doug.fort`.
