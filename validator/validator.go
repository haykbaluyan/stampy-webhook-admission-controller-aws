package validator

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	"git.soma.salesforce.com/kuleana/go-pkg/cms"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
)

// ValidateManifestSignature validates manifest signature
func ValidateManifestSignature(manifest, manifestSig string) (bool, string, error) {
	manifestSigBytes := []byte(manifestSig)
	sig, err := loadSignatureResponse(manifestSigBytes)
	if err != nil {
		return false, "", errors.Errorf("api=ValidateManifestSignature, reason=loadSignatureResponse, err=%v", err)
	}

	manifestBytes := []byte(manifest)
	manifestDigest := hex.EncodeToString(certutil.SHA256(manifestBytes))
	artifact, err := findArtifactInSignatureResponse(sig, manifestDigest)
	if err != nil {
		return false, "", errors.Errorf("api=ValidateManifestSignature, reason=findArtifactInSignatureResponse, err=%v", err)
	}

	bundle, bundleStatus, err := certutil.VerifyBundleFromPEM([]byte(artifact.Certificate), []byte(artifact.CA), nil)
	if err != nil {
		return false, "", errors.Errorf("api=ValidateManifestSignature, reason=VerifyBundleFromPEM, err=%v", err)
	}

	if bundleStatus.IsUntrusted() {
		return false, "", errors.Errorf("api=ValidateManifestSignature, reason='signing certificate is not trusted', certificate=%q, ca=%q", artifact.Certificate, artifact.CA)
	}

	switch artifact.SignatureFormat {
	case "cms-detached", "pkcs7-detached":
		err = verifyDetached(manifestBytes, artifact, bundle.RootCert)
		if err != nil {
			return false, "", errors.Errorf("api=ValidateManifestSignature, reason=verifyDetached, artifactName=%q, err=%v", artifact.Name, err)
		}
	default:
		return false, "", errors.Errorf("api=ValidateManifestSignature, reason='not supported signature format', signatureFormat=%q, err=%v", artifact.SignatureFormat, err)
	}

	return true, manifestDigest, nil
}

// loadSignatureResponse loads and decodes a SignatureResponse
func loadSignatureResponse(b []byte) (*SignatureResponse, error) {
	r := bytes.NewReader(b)
	res := new(SignatureResponse)
	return res, json.NewDecoder(r).Decode(res)
}

// findArtifactInSignatureResponse finds corresponding artifact in the signature response
func findArtifactInSignatureResponse(sig *SignatureResponse, hash string) (*SignatureInfo, error) {
	artifact, err := sig.getSignatureInfoWithHash(hash)
	if err != nil {
		return nil, errors.Errorf("api=findArtifactInSignatureResponse, reason=getSignatureInfoWithHash, hash=%q", hash)
	}
	return artifact, nil
}

func verifyDetached(input []byte, artifact *SignatureInfo, rootCert *x509.Certificate) error {
	opts := x509.VerifyOptions{
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageCodeSigning,
		},
	}

	if rootCert != nil {
		roots := x509.NewCertPool()
		roots.AddCert(rootCert)
		opts.Roots = roots
	}

	der, err := base64.StdEncoding.DecodeString(artifact.Signature)
	if err != nil {
		return errors.Annotatef(err, "unable to decode signature")
	}

	sd, err := cms.ParseSignedData(der)
	if err != nil {
		return errors.Annotatef(err, "unable to parse signed data")
	}

	_, err = sd.VerifyDetached(input, opts, nil)
	if err != nil {
		return errors.Annotatef(err, "reason=verifyDetached, artifact=%q, sig_id=%s", artifact.Name, artifact.SigID)
	}
	return nil
}

// getSignatureInfoWithHash finds a signature info in the response by hash
func (s *SignatureResponse) getSignatureInfoWithHash(hash string) (*SignatureInfo, error) {
	if s != nil && len(s.Signatures) > 0 {
		for _, a := range s.Signatures {
			if a.Hash == hash {
				return a, nil
			}
		}
	}
	return nil, errors.Errorf("api=getSignatureInfoWithHash, hash=%q", hash)
}
