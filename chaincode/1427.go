package main

import (
	"crypto/x509"
	"fmt"
	"github.com/davidkhala/fabric-common/golang/format"
	. "github.com/davidkhala/goutils"
	"github.com/davidkhala/goutils/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/pkg/attrmgr"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// ClientIdentity alternative of creator starting from 1.1
type ClientIdentity struct {
	MspID          format.MSPID
	CertificatePem string
	Attrs          map[string]string `json:"Attrs,omitempty"`
}

func NewClientIdentity(stub shim.ChaincodeStubInterface) (c ClientIdentity) {
	signingID := &msp.SerializedIdentity{}
	creator, err := stub.GetCreator()
	PanicError(err)
	if creator == nil {
		PanicString("failed to get transaction invoker's identity from the chaincode stub")
	}
	err = proto.Unmarshal(creator, signingID)
	PanicError(err)

	c.MspID = signingID.GetMspid()
	c.CertificatePem = string(signingID.GetIdBytes())
	attrs, err := attrmgr.New().GetAttributesFromCert(c.GetCertificate())
	PanicError(err)
	c.Attrs = attrs.Attrs
	return c
}

func (c ClientIdentity) GetAttributeValue(attrName string) string {
	return c.Attrs[attrName]
}

// GetID returns a unique ID associated with the invoking identity.
func (c ClientIdentity) GetID() format.CertificateID {
	// The leading "x509::" distinguishes this as an X509 certificate, and
	// the subject and issuer DNs uniquely identify the X509 certificate.
	// The resulting ID will remain the same if the certificate is renewed.
	var certificate = c.GetCertificate()
	id := fmt.Sprintf("x509::%s::%s", crypto.GetDN(certificate.Subject), crypto.GetDN(certificate.Issuer))
	return id
}

func (c ClientIdentity) GetCertificate() *x509.Certificate {
	return crypto.ParseCertPemOrPanic([]byte(c.CertificatePem))
}
