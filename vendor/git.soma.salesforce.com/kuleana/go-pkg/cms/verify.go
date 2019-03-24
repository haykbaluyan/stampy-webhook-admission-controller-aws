package cms

import (
	"bytes"
	"crypto/x509"

	"git.soma.salesforce.com/kuleana/go-pkg/cms/protocol"
	"github.com/juju/errors"
)

// Verify verifies the SingerInfos' signatures. Each signature's associated
// certificate is verified using the provided roots. UnsafeNoVerify may be
// specified to skip this verification. Nil may be provided to use system roots.
// The full chains for the certificates whose keys made the signatures are
// returned.
//
// WARNING: this function doesn't do any revocation checking.
func (sd *SignedData) Verify(opts x509.VerifyOptions, signer *x509.Certificate) ([][][]*x509.Certificate, error) {
	econtent, err := sd.psd.EncapContentInfo.EContentValue()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if econtent == nil {
		return nil, errors.New("detached signature")
	}

	return sd.VerifyContent(econtent, opts, signer)
}

// VerifyDetached verifies the SingerInfos' detached signatures over the
// provided data message. Each signature's associated certificate is verified
// using the provided roots. UnsafeNoVerify may be specified to skip this
// verification. Nil may be provided to use system roots. The full chains for
// the certificates whose keys made the signatures are returned.
//
// WARNING: this function doesn't do any revocation checking.
func (sd *SignedData) VerifyDetached(message []byte, opts x509.VerifyOptions, signer *x509.Certificate) ([][][]*x509.Certificate, error) {
	if sd.psd.EncapContentInfo.EContent.Bytes != nil {
		return nil, errors.New("signature not detached")
	}
	return sd.VerifyContent(message, opts, signer)
}

// VerifyContent verifies the SingerInfos' signatures over the
// provided data message.
func (sd *SignedData) VerifyContent(econtent []byte, opts x509.VerifyOptions, signer *x509.Certificate) ([][][]*x509.Certificate, error) {
	if len(sd.psd.SignerInfos) == 0 {
		return nil, protocol.ASN1Error{Message: "no signatures found"}
	}

	certs, err := sd.psd.X509Certificates()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if opts.Intermediates == nil {
		opts.Intermediates = x509.NewCertPool()
	}

	for _, cert := range certs {
		opts.Intermediates.AddCert(cert)
	}

	chains := make([][][]*x509.Certificate, 0, len(sd.psd.SignerInfos))

	for _, si := range sd.psd.SignerInfos {
		var signedMessage []byte

		// SignedAttrs is optional if EncapContentInfo eContentType isn't id-data.
		if si.SignedAttrs == nil {
			// SignedAttrs may only be absent if EncapContentInfo eContentType is
			// id-data.
			if !sd.psd.EncapContentInfo.IsTypeData() {
				return nil, protocol.ASN1Error{Message: "missing SignedAttrs"}
			}

			// If SignedAttrs is absent, the signature is over the original
			// encapsulated content itself.
			signedMessage = econtent
		} else {
			// If SignedAttrs is present, we validate the mandatory ContentType and
			// MessageDigest attributes.
			siContentType, err := si.GetContentTypeAttribute()
			if err != nil {
				return nil, errors.Trace(err)
			}
			if !siContentType.Equal(sd.psd.EncapContentInfo.EContentType) {
				return nil, protocol.ASN1Error{Message: "invalid SignerInfo ContentType attribute"}
			}

			// Calculate the digest over the actual message.
			hash, err := si.Hash()
			if err != nil {
				return nil, errors.Trace(err)
			}
			actualMessageDigest := hash.New()
			if _, err = actualMessageDigest.Write(econtent); err != nil {
				return nil, errors.Trace(err)
			}

			// Get the digest from the SignerInfo.
			messageDigestAttr, err := si.GetMessageDigestAttribute()
			if err != nil {
				return nil, errors.Trace(err)
			}

			// Make sure message digests match.
			if !bytes.Equal(messageDigestAttr, actualMessageDigest.Sum(nil)) {
				return nil, errors.New("invalid message digest")
			}

			// The signature is over the DER encoded signed attributes, minus the
			// leading class/tag/length bytes. This includes the digest of the
			// original message, so it is implicitly signed too.
			if signedMessage, err = si.SignedAttrs.MarshaledForSigning(); err != nil {
				return nil, errors.Trace(err)
			}
		}

		if signer != nil {
			certs = append(certs, signer)
		}

		cert, _ := si.FindCertificate(certs)
		if cert == nil {
			return nil, errors.Trace(protocol.ErrNoCertificate)
		}
		if signer != nil && !bytes.Equal(cert.Raw, signer.Raw) {
			return nil, errors.Errorf("the signer certificate does not match one in the signature: CN=%s",
				cert.Subject.CommonName)
		}

		algo := si.X509SignatureAlgorithm()
		if algo == x509.UnknownSignatureAlgorithm {
			return nil, errors.Trace(protocol.ErrUnsupported)
		}

		if err := cert.CheckSignature(algo, signedMessage, si.Signature); err != nil {
			return nil, errors.Trace(err)
		}

		// If the caller didn't specify the signature time, we'll use the verified
		// timestamp. If there's no timestamp we use the current time when checking
		// the cert validity window. This isn't perfect because the signature may
		// have been created before the cert's not-before date, but this is the best
		// we can do.
		optsCopy := opts

		hasTS, err := hasTimestamp(si)
		if err != nil {
			return nil, errors.Trace(err)
		}
		if hasTS {
			tsOpts := x509.VerifyOptions{
				Roots:         opts.Roots,
				Intermediates: opts.Intermediates,
				KeyUsages: []x509.ExtKeyUsage{
					x509.ExtKeyUsageTimeStamping,
				},
			}

			tsti, err := getTimestamp(si, tsOpts)
			if err != nil {
				return nil, errors.Annotatef(err, "failed to verify timestamp")
			}

			// This check is slightly redundant, given that the cert validity times
			// are checked by cert.Verify. We take the timestamp accuracy into account
			// here though, whereas cert.Verify will not.
			if !tsti.Before(cert.NotAfter) || !tsti.After(cert.NotBefore) {
				return nil, x509.CertificateInvalidError{Cert: cert, Reason: x509.Expired, Detail: ""}
			}

			if optsCopy.CurrentTime.IsZero() {
				optsCopy.CurrentTime = tsti.GenTime
			}
		}

		chain, err := cert.Verify(optsCopy)
		if err != nil {
			return nil, errors.Annotatef(err, "failed to verify: %s", cert.Subject.CommonName)
		}

		chains = append(chains, chain)
	}

	// OK
	return chains, nil
}
