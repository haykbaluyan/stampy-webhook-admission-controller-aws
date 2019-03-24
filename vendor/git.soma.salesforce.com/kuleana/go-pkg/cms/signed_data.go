package cms

import (
	"crypto/x509"
	"encoding/asn1"

	"git.soma.salesforce.com/kuleana/go-pkg/cms/protocol"
	"github.com/juju/errors"
)

// SignedData represents a signed message or detached signature.
type SignedData struct {
	psd *protocol.SignedData
}

// NewSignedData creates a new SignedData from the given data.
func NewSignedData(data []byte) (*SignedData, error) {
	eci, err := protocol.NewDataEncapsulatedContentInfo(data)
	if err != nil {
		return nil, errors.Trace(err)
	}

	psd, err := protocol.NewSignedData(eci)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &SignedData{psd}, nil
}

// ParseSignedData parses a SignedData from BER encoded data.
func ParseSignedData(ber []byte) (*SignedData, error) {
	ci, err := protocol.ParseContentInfo(ber)
	if err != nil {
		return nil, errors.Trace(err)
	}

	psd, err := ci.SignedDataContent()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &SignedData{psd}, nil
}

// FromEncapsulatedContentInfo set a SignedData from EncapsulatedContentInfo.
func FromEncapsulatedContentInfo(eci protocol.EncapsulatedContentInfo) (*SignedData, error) {
	psd, err := protocol.NewSignedData(eci)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &SignedData{psd}, nil
}

// FromContentInfo set a SignedData from ContentInfo.
func FromContentInfo(ci *protocol.ContentInfo) (*SignedData, error) {
	psd, err := ci.SignedDataContent()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &SignedData{psd}, nil
}

// GetData gets the encapsulated data from the SignedData. Nil will be returned
// if this is a detached signature. A protocol.ErrWrongType will be returned if
// the SignedData encapsulates something other than data (1.2.840.113549.1.7.1).
func (sd *SignedData) GetData() ([]byte, error) {
	return sd.psd.EncapContentInfo.DataEContent()
}

// GetEncapsulatedContent returns encapsulated data
func (sd *SignedData) GetEncapsulatedContent() *protocol.EncapsulatedContentInfo {
	return &sd.psd.EncapContentInfo
}

// PrintSignerInfos returns list of friendly SignerInfo
func (sd *SignedData) PrintSignerInfos() []string {
	list := []string{}
	for _, si := range sd.psd.SignerInfos {
		list = append(list, si.String())
	}
	return list
}

// GetCertificates gets all the certificates stored in the SignedData.
func (sd *SignedData) GetCertificates() ([]*x509.Certificate, error) {
	return sd.psd.X509Certificates()
}

// SetCertificates replaces the certificates stored in the SignedData with new
// ones.
func (sd *SignedData) SetCertificates(certs []*x509.Certificate) error {
	sd.psd.ClearCertificates()
	for _, cert := range certs {
		if err := sd.psd.AddCertificate(cert); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

// Detached removes the data content from this SignedData. No more signatures
// can be added after this method has been called.
func (sd *SignedData) Detached() {
	sd.psd.EncapContentInfo.EContent = asn1.RawValue{}
}

// IsDetached checks if this SignedData has data content.
func (sd *SignedData) IsDetached() bool {
	return sd.psd.EncapContentInfo.EContent.Bytes == nil
}

// ToDER encodes this SignedData message using DER.
func (sd *SignedData) ToDER() ([]byte, error) {
	return sd.psd.ContentInfoDER()
}

// Version of protocol
func (sd *SignedData) Version() int {
	return sd.psd.Version
}
