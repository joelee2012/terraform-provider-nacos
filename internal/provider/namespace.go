package provider

import (
	"github.com/joelee2012/go-nacos"
)

// serverNamespaceID maps a user-facing namespace id (where "" means the public
// namespace) to the id the Nacos server expects when addressing it by id.
// Nacos v1/v2 (v1 console) use the empty string for public; Nacos v3 (v3
// console) uses "public".
//
// Only GetNamespace needs this: go-nacos fetches the namespace list and matches
// on the exact id, so a "" lookup would miss the v3 public namespace. The config
// methods pass the value straight through because v3 accepts an empty namespace
// id and treats it as public (verified against Nacos 3.2.4). GetAPIVersion is
// cached by the client and performs no I/O.
func serverNamespaceID(client *nacos.Client, ns string) string {
	if ns != "" {
		return ns
	}
	ver, err := client.GetAPIVersion()
	if err != nil {
		// Cannot determine version: fall back to the v1/v2 convention ("").
		return ""
	}
	if ver == "v3" {
		return "public"
	}
	return ""
}

// userNamespaceID maps a namespace id read back from the server to the value
// surfaced to users, preserving the form the user chose for the public
// namespace. The server always reports the public namespace as "public" on
// Nacos v3, but the provider's contract is that an empty namespace_id means
// public. Keeping the user's form avoids two problems:
//
//   - When the user omits the attribute (or sets ""), state must hold "" or
//     Terraform reports "provider produced inconsistent result after apply".
//   - When the user explicitly sets "public" (a valid v3 id), state must keep
//     "public": rewriting a config-known value is rejected by Terraform.
//
// For every other namespace the server value is returned unchanged.
func userNamespaceID(serverNS, userNS string) string {
	if serverNS != "public" {
		return serverNS
	}
	if userNS == "public" {
		return "public"
	}
	return ""
}
