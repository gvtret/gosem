package dlms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeAARE(t *testing.T) {
	src := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")
	aare, err := DecodeAARE(nil, &src)
	assert.NoError(t, err)
	assert.Equal(t, ApplicationContextLNNoCiphering, aare.ApplicationContext)
	assert.Equal(t, AssociationResultAccepted, aare.AssociationResult)
	assert.Equal(t, SourceDiagnosticNone, aare.SourceDiagnostic)

	assert.NotNil(t, aare.InitiateResponse)
	assert.Equal(t, uint16(128), aare.InitiateResponse.ServerMaxReceivePduSize)
	assert.Nil(t, aare.ConfirmedServiceError)
}

func TestDecodeRejectedAARE(t *testing.T) {
	src := decodeHexString("611FA109060760857405080101A203020101A305A10302010DBE0604040E010600")

	aare, err := DecodeAARE(nil, &src)
	assert.NoError(t, err)
	assert.Equal(t, ApplicationContextLNNoCiphering, aare.ApplicationContext)
	assert.Equal(t, AssociationResultPermanentRejected, aare.AssociationResult)
	assert.Equal(t, SourceDiagnosticAuthenticationFailure, aare.SourceDiagnostic)
	assert.Nil(t, aare.InitiateResponse)
	assert.Nil(t, aare.ConfirmedServiceError)

	// Sagemcom reply
	src = decodeHexString("6129A109060760857405080101A203020101A305A10302010DBE10040E0800065F1F040000101400800080")
	aare, err = DecodeAARE(nil, &src)
	assert.NoError(t, err)
	assert.Equal(t, ApplicationContextLNNoCiphering, aare.ApplicationContext)
	assert.Equal(t, AssociationResultPermanentRejected, aare.AssociationResult)
	assert.Equal(t, SourceDiagnosticAuthenticationFailure, aare.SourceDiagnostic)
	assert.Nil(t, aare.InitiateResponse)
	assert.Nil(t, aare.ConfirmedServiceError)
}

func TestDecodeAAREWithSecurity(t *testing.T) {
	ciphering, _ := NewCiphering(
		SecurityLevelDedicatedKey,
		SecurityEncryption|SecurityAuthentication,
		decodeHexString("4349520000000001"),
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
		1,
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
	)

	settings := &Settings{
		Ciphering: ciphering,
	}

	src := decodeHexString("6148A109060760857405080103A203020100A305A103020100A40A04084C475A2022604828BE230421281F300000003149963E23D6DA824A369644B66A9A17C60C3CA3F63E58608FA192")
	aare, err := DecodeAARE(settings, &src)
	assert.NoError(t, err)
	assert.Equal(t, ApplicationContextLNCiphering, aare.ApplicationContext)
	assert.Equal(t, AssociationResultAccepted, aare.AssociationResult)
	assert.Equal(t, SourceDiagnosticNone, aare.SourceDiagnostic)

	sourceSystemTitle := decodeHexString("4C475A2022604828")
	assert.Equal(t, sourceSystemTitle, aare.SourceSystemTitle)
}
