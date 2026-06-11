/*
Package zabbix provides bindings to interoperate between programs written in Go
language and the Zabbix monitoring API.

This project aims to provide a stable, fast Go client for the Zabbix JSON-RPC API
with support for both loose typing (using interface{} or map[string]interface{})
and strong types (such as zabbix.Host or zabbix.Event).

The package supports Zabbix API versions from 4.0 through to 7.0 with automatic
version detection and version-specific behavior:

  - Zabbix 4.x, 5.x: auth token in JSON-RPC body
  - Zabbix 6.x: object-based hosts/groups in maintenance, auth token in JSON-RPC body
  - Zabbix 7.x: Bearer token in Authorization HTTP header (preferred over auth in body),
    Content-Type: application/json, new proxy management fields

# Basic usage

	package main

	import (
		"context"
		"fmt"

		zabbix "path/to/go-zabbix"
	)

	func main() {
		ctx := context.Background()

		// Basic session creation
		session, err := zabbix.NewSession(ctx, "http://zabbix/api_jsonrpc.php", "Admin", "zabbix", "")
		if err != nil {
			panic(err)
		}
		fmt.Printf("Connected to Zabbix API v%s\n", session.GetVersion(ctx))

		// Or use API token (recommended for Zabbix 6.4+ and 7.0+)
		session, err = zabbix.NewSessionToken(ctx, "http://zabbix/api_jsonrpc.php", "your-api-token")
		if err != nil {
			panic(err)
		}

		// Or use session builder with caching
		cache := zabbix.NewSessionFileCache().SetFilePath("./zabbix_session")
		session, err = zabbix.CreateClient("http://zabbix/api_jsonrpc.php").
			WithCache(cache).
			WithCredentials("Admin", "zabbix").
			Connect(ctx)

		fmt.Printf("Connected to Zabbix API v%s\n", session.GetVersion(ctx))
	}

For more information see: https://github.com/cavaliercoder/go-zabbix

*/
package zabbix
