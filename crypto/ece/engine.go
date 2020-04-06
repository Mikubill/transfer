package ece

// https://github.com/xakep666/ecego, MIT license, modified

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const (
	KeySize = aes.BlockSize
	// NonceSize is standard nonce size for GCM (unexported in cipher package)
	NonceSize = 12
	// TagSize is standard auth tag size for GCM (unexported in cipher package)
	TagSize = 16
)

// Available content-encoding versions
const (
	// AES128GCM is a newest and recommended for usage, also used as default if not specified explicitly
	AES128GCM Version = "aes128gcm"

	// AESGCM is a widely used version (i.e. in Firebase Cloud Messaging)
	AESGCM Version = "aesgcm"

	// AESGCM128 is a legacy version but sometimes used
	AESGCM128 Version = "aesgcm128"
)

const (
	DefaultRecordSize uint32 = 4096
	DefaultKeyLabel   string = "P-256"
)

// aes128gcm content header format from RFC8188:
// +-----------+--------+-----------+---------------+
// | salt (16) | rs (4) | idlen (1) | keyid (idlen) |
// +-----------+--------+-----------+---------------+
const (
	saltSize       = 16
	recordSizeSize = 4
	idLenSize      = 1
)

var (
	ErrInvalidKeySize  = fmt.Errorf("invalid static key length")
	ErrInvalidSaltSize = fmt.Errorf("invaild salt size")
)

type (
	// Version determines encoding mode
	Version string

	// KeyLookupFunc is a function that fetches a private key by provided keyID
	// It must always return non-nil value even if input is nil
	KeyLookupFunc func(keyID []byte) *ecdsa.PrivateKey
)

// Engine represents encryption and decryption engine
type Engine struct {
	keyLookupFunc KeyLookupFunc
	keyLabel      string
	authSecret    []byte
}

// newEngine constructs an ecego engine. keyLookupFunc may be nil if only static key encryption will used.
func newEngine(keyLookupFunc KeyLookupFunc, options ...engineOption) *Engine {
	ret := &Engine{keyLookupFunc: keyLookupFunc}

	for _, o := range options {
		o.apply(ret)
	}

	if ret.keyLabel == "" {
		ret.keyLabel = DefaultKeyLabel
	}

	return ret
}

type operationalParams struct {
	Version    Version
	Salt       []byte
	DH         []byte
	StaticKey  []byte // If provided will be used instead of key derivation
	KeyID      []byte
	RecordSize uint32 // DefaultRecordSize used if not provided
	Pad        uint32
}

func (e *Engine) privateKey(params operationalParams) *ecdsa.PrivateKey {
	if params.StaticKey != nil {
		return nil
	}
	return e.keyLookupFunc(params.KeyID)
}

func (e *Engine) publicKey(params operationalParams) *ecdsa.PublicKey {
	if privateKey := e.privateKey(params); privateKey != nil {
		return &privateKey.PublicKey
	}
	return nil
}

func (e *Engine) buildInfoContext(version Version, senderPublicKey, receiverPublicKey *ecdsa.PublicKey) []byte {
	if senderPublicKey == nil || receiverPublicKey == nil {
		return nil
	}

	var builder bytes.Buffer

	var (
		receiverPKBytes = elliptic.Marshal(receiverPublicKey.Curve, receiverPublicKey.X, receiverPublicKey.Y)
		senderPKBytes   = elliptic.Marshal(senderPublicKey.Curve, senderPublicKey.X, senderPublicKey.Y)
	)

	if version == AES128GCM {
		builder.WriteString("WebPush: info\x00")
		builder.Write(receiverPKBytes)
		builder.Write(senderPKBytes)

		return builder.Bytes()
	}

	// 1st part - keyLabel + \x00
	builder.WriteString(e.keyLabel)
	builder.WriteByte(0)

	// 2nd part - receiver public key length (2 bytes in BigEndian) + receiver public key
	_ = binary.Write(&builder, binary.BigEndian, uint16(len(receiverPKBytes)))
	builder.Write(receiverPKBytes)

	// 3rd part - sender public key length (same as before) + sender public key
	_ = binary.Write(&builder, binary.BigEndian, uint16(len(senderPKBytes)))
	builder.Write(senderPKBytes)

	return builder.Bytes()
}

func (e *Engine) deriveSharedSecret(params operationalParams, publicKey *ecdsa.PublicKey) ([]byte, error) {
	if params.StaticKey != nil {
		if len(params.StaticKey) != KeySize {
			return nil, ErrInvalidKeySize
		}

		return params.StaticKey, nil
	}
	privateKey := e.keyLookupFunc(params.KeyID)
	x, _ := privateKey.Curve.ScalarMult(publicKey.X, publicKey.Y, privateKey.D.Bytes())
	// RFC5903 Section 9 states we should only return x.
	return x.Bytes(), nil
}

func (e *Engine) buildInfo(base string, infoContext []byte) []byte {
	var b bytes.Buffer

	b.WriteString("Content-Encoding: ")
	b.WriteString(base)
	b.WriteByte(0)
	b.Write(infoContext)

	return b.Bytes()
}

func (e *Engine) deriveKey(params operationalParams, otherPublicKey *ecdsa.PublicKey, infoContext []byte) (key, nonce []byte, err error) {
	var (
		keyInfo, nonceInfo []byte
	)

	if otherPublicKey == nil {
		otherPublicKey = e.publicKey(params)
	}

	secret, err := e.deriveSharedSecret(params, otherPublicKey)
	if err != nil {
		return nil, nil, err
	}
	authSecret := e.authSecret

	switch params.Version {
	case AESGCM:
		keyInfo = e.buildInfo(AESGCM.String(), infoContext)
		nonceInfo = e.buildInfo("nonce", infoContext)
	case AESGCM128:
		keyInfo = []byte("Content-Encoding: aesgcm128")
		nonceInfo = []byte("Content-Encoding: nonce")
	case AES128GCM:
		keyInfo = []byte("Content-Encoding: aes128gcm\x00")
		nonceInfo = []byte("Content-Encoding: nonce\x00")
		// Only mix the authentication secret when using DH for aes128gcm
		if len(params.DH) == 0 {
			authSecret = authSecret[:0]
		}
	}

	if len(authSecret) != 0 {
		info := infoContext
		if params.Version != AES128GCM {
			info = e.buildInfo("auth", nil)
		}

		tmpSecret := make([]byte, sha256.Size)
		_, err = io.ReadFull(hkdf.New(sha256.New, secret, authSecret, info), tmpSecret)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to expand secret with auth secret: %w", err)
		}

		secret = tmpSecret
	}

	key = make([]byte, KeySize)
	nonce = make([]byte, NonceSize)

	_, err = io.ReadFull(hkdf.New(sha256.New, secret, params.Salt, keyInfo), key)
	if err != nil {
		return nil, nil, fmt.Errorf("hkdf for key failed: %w", err)
	}

	_, err = io.ReadFull(hkdf.New(sha256.New, secret, params.Salt, nonceInfo), nonce)
	if err != nil {
		return nil, nil, fmt.Errorf("hkdf for nonce failed: %w", err)
	}

	return key, nonce, nil
}

func (v Version) paddingSize() int {
	switch v {
	case AES128GCM:
		return 1
	case AESGCM:
		return 2
	case AESGCM128:
		return 1
	default:
		return -1
	}
}

func (v Version) String() string { return string(v) }

func unmarshalPublicKey(curve elliptic.Curve, dh []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(curve, dh)
	if x == nil {
		return &ecdsa.PublicKey{}, ErrInvalidDH
	}

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

func SingleKey(key *ecdsa.PrivateKey) KeyLookupFunc {
	return func([]byte) *ecdsa.PrivateKey { return key }
}

func newGCM(key []byte) (cipher.AEAD, error) {
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes init failed: %w", err)
	}

	aead, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, fmt.Errorf("gcm init failed: %w", err)
	}

	return aead, nil
}

func fillBlockNonce(counter uint64, baseNonce, blockNonce []byte) {
	_ = blockNonce[NonceSize-1]
	copy(blockNonce[:4], baseNonce)
	binary.BigEndian.PutUint64(
		blockNonce[4:],
		counter^binary.BigEndian.Uint64(baseNonce[4:]),
	)
}
