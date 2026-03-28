package sshkey

// SSH Key sweeper is intentionally omitted.
//
// The SSH key API does not expose a name or label field that could be used
// to identify test resources by the "tfacc-" prefix convention. The SSHKey
// struct only contains: ID, SSHKey (public key content), CreatedAt, MachineID,
// and PublicKey. Without a name field, there is no safe way to distinguish
// test-created keys from user-created keys.
//
// Acceptance tests that create SSH keys should clean up in their own
// CheckDestroy functions.
