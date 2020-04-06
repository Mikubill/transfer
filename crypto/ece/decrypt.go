package ece

// https://github.com/xakep666/ecego, MIT license, modified

import (
	"bytes"
	"crypto/cipher"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
)

var (
	ErrInvalidDH      = fmt.Errorf("dh sequence is invalid")
	ErrTruncated      = fmt.Errorf("content truncated")
	ErrInvalidPadding = fmt.Errorf("invalid padding")
)

// Decrypt takes input cipher text, decrypts it using provided parameters and appends result to target
func (e *Engine) Decrypt(content, target []byte, params operationalParams) ([]byte, error) {
	if len(params.Salt) != saltSize {
		return nil, ErrInvalidSaltSize
	}

	if params.RecordSize == 0 {
		params.RecordSize = DefaultRecordSize
	}

	if params.Version == "" {
		params.Version = AES128GCM
	}

	// in aes128gcm mode content contains special header before cipher text
	if params.Version == AES128GCM {
		var err error

		content, err = readHeader(content, &params)
		if err != nil {
			return nil, fmt.Errorf("read aes128gcm header failed: %w", err)
		}

		if params.DH == nil {
			// Public key may be in KeyID
			params.DH = params.KeyID
		}
	}

	publicKey := e.publicKey(params)
	var senderPublicKey *ecdsa.PublicKey
	if params.StaticKey == nil {
		var err error
		senderPublicKey, err = unmarshalPublicKey(publicKey, params.DH)
		if err != nil {
			return nil, err
		}
	}

	key, nonce, err := e.deriveKey(
		params,
		senderPublicKey,
		e.buildInfoContext(params.Version, senderPublicKey, publicKey),
	)
	if err != nil {
		return nil, fmt.Errorf("derive key/nonce failed: %w", err)
	}

	aead, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	return decrypt(content, target, nonce, aead, params)
}

// readHeader reads special header for aes128gcm and modifies params with parsed values
func readHeader(in []byte, params *operationalParams) ([]byte, error) {
	if len(in) < saltSize+recordSizeSize+idLenSize {
		return nil, ErrTruncated
	}

	params.Salt = make([]byte, saltSize)
	in = in[copy(params.Salt, in):]

	params.RecordSize = binary.BigEndian.Uint32(in[:recordSizeSize])
	in = in[recordSizeSize:]

	idLen := in[0]
	in = in[idLenSize:]

	if len(in) < int(idLen) {
		return nil, ErrTruncated
	}

	params.KeyID = make([]byte, idLen)
	in = in[copy(params.KeyID, in):]

	return in, nil
}

func decrypt(content, target, nonce []byte, aead cipher.AEAD, params operationalParams) ([]byte, error) {
	blockNonce := make([]byte, NonceSize)
	plainTextBlock := make([]byte, 0)
	contentLen := uint32(len(content))

	for start, counter := uint32(0), uint64(0); start < contentLen; counter++ {
		end, err := calculateCipherTextBlockEnd(start, contentLen, params)
		if err != nil {
			return nil, err
		}

		cipherTextBlock := content[start:end]
		fillBlockNonce(counter, nonce, blockNonce)

		plainTextBlock, err = aead.Open(plainTextBlock[:0], blockNonce, cipherTextBlock, nil)
		if err != nil {
			return nil, fmt.Errorf("block %d decrypt failed: %w", counter, err)
		}

		if params.Version != AES128GCM {
			target, err = unpadLegacy(plainTextBlock, target, params.Version.paddingSize())
		} else {
			target, err = unpad(plainTextBlock, target, end >= contentLen)
		}

		if err != nil {
			return nil, err
		}

		start = end
	}

	return target, nil
}

func calculateCipherTextBlockEnd(start, contentLen uint32, params operationalParams) (uint32, error) {
	blockSize := params.RecordSize
	if params.Version != AES128GCM {
		blockSize += TagSize
	}

	end := start + blockSize
	if params.Version != AES128GCM && end == contentLen {
		return 0, ErrTruncated
	}

	if contentLen < end {
		end = contentLen
	}

	if end-start <= TagSize {
		return 0, ErrTruncated
	}

	return end, nil
}

func unpadLegacy(plainText, target []byte, padSize int) ([]byte, error) {
	var pad uint32
	switch padSize {
	case 1:
		pad = uint32(plainText[0])
	case 2:
		pad = uint32(binary.BigEndian.Uint16(plainText[:2]))
	default:
		return nil, fmt.Errorf("unknown padding size %d", padSize)
	}

	if int(pad)+padSize > len(plainText) {
		return nil, fmt.Errorf("padding exceeds block size: %w", ErrInvalidPadding)
	}

	if !bytes.Equal(
		plainText[padSize:int(pad)+padSize],
		make([]byte, pad),
	) {
		return nil, ErrInvalidPadding
	}

	return append(target, plainText[int(pad)+padSize:]...), nil
}

func unpad(plainText, target []byte, lastBlock bool) ([]byte, error) {
	for i := len(plainText) - 1; i >= 0; i-- {
		switch {
		case plainText[i] == 0:
			continue
		case lastBlock && plainText[i] != 2:
			return nil, fmt.Errorf("last block must start padding with 2: %w", ErrInvalidPadding)
		case !lastBlock && plainText[i] != 1:
			return nil, fmt.Errorf("non-last block must start padding with 1: %w", ErrInvalidPadding)
		default:
			return append(target, plainText[:i]...), nil
		}
	}

	return nil, fmt.Errorf("all zero plaintext")
}
