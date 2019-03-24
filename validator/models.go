package validator

import "time"

// GitCommitInfo provides information about the commit
type GitCommitInfo struct {
	// Repo specifies the repo name
	Repo string `json:"repo"`

	// Team specifies the team that owns the repo
	Team string `json:"team"`

	// Commit specifies GIT's hash of the commit
	Commit string `json:"commit"`

	// Author specifies email of the author
	Author string `json:"author"`

	// Committer specifies email of the commiter
	Committer string `json:"committer"`

	// Approver specifies email of the approver
	Approver string `json:"approver"`
}

// RoleInfo provides information about the Stampy client or server
type RoleInfo struct {
	// Role specifies the role of the requestor. It must match the role in auth certificate.
	Role string `json:"role"`

	// Host specifies the host of the client originating the request
	Host string `json:"host"`

	// IP specifies the IP of the client originating the request
	IP string `json:"ip"`
}

// SignatureInfo provides information about the file's signature
type SignatureInfo struct {
	// CorrelationID specifies the unique identifier of the file,
	// to bind a signature to the file for Audit purposes.
	// If it's not set, then SHA1 of the file will be used.
	CorrelationID string `json:"correlation_id"`

	// Name specifies the file name
	Name string `json:"name"`

	// SignatureFormat specifies the format of the signature [gpg|uefi|authenticode|cms|pkcs7|cms-detached|pkcs7-detached]
	SignatureFormat string `json:"signature_format"`

	// Size specifies the size of the file in bytes
	Size uint64 `json:"size"`

	// CreatedAt specifies time when the file was created
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// ModifiedAt specifies time when the file was last modified
	ModifiedAt *time.Time `json:"modified_at,omitempty"`

	// HashAlg specifies the hash algorithm
	HashAlg string `json:"hash_alg"`

	// Hash specifies the hash of the file to be signed
	Hash string `json:"hash"`

	// SigID specifies the unique signature identifier
	SigID string `json:"sig_id"`

	// SigAlg specifies the signature algorythm `RSA2048_SHA256`, `ECDSA_P256`
	SigAlg string `json:"sig_alg"`

	// SignedAt specifies time when the signature was produced
	SignedAt time.Time `json:"signed_at"`

	// Signature specifies the base64 encoded value for detached signatures
	//  gpg : PEM encoded OpenPGP
	//  cms: Base64 encoded CMS Signed Data
	Signature string `json:"signature,omitempty"`

	// SignedArtifactURL specifies a URL of short-lived location of signed files
	// with embedded signature: rpm, cms
	SignedArtifactURL string `json:"signed_url,omitempty"`

	// Certificate specifies the signing certificate in PEM format
	Certificate string `json:"certificate"`

	// CA specifies the issuing CA bundle in PEM format
	CA string `json:"ca,omitempty"`

	// GPGPubKey provides PEM encoded GPG public key, for GPG keys
	GPGPubKey string `json:"gpg_pubkey,omitempty"`
}

// SignatureResponse specifies a sign response with signed files
type SignatureResponse struct {
	// Requestor provides information about the requestor
	Requestor *RoleInfo `json:"requestor,omitempty"`

	// Signer provides information about the signer
	Signer *RoleInfo `json:"signer,omitempty"`

	// Commit provides information about the commit
	Commit *GitCommitInfo `json:"commit,omitempty"`

	// Signatures specifies a list of files to sign
	Signatures []*SignatureInfo `json:"signatures,omitempty"`

	// Locations specifies a list of locations for signed files
	Locations map[string]string `json:"locations,omitempty"`
}
