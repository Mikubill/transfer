package ece

// https://github.com/xakep666/ecego, MIT license, modified

import (
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"fmt"
)

var (
	errKeyIDTooLong       = fmt.Errorf("keyID too long")
	errTooSmallRecordSize = fmt.Errorf("record size too small")
	errPaddingToRecord    = fmt.Errorf("unable to pad to record size")
)

// Encrypt takes input plain text, encrypts it using provided parameters and appends result to target
func (e *Engine) Encrypt(content, target []byte, params operationalParams) ([]byte, error) {
	if len(params.Salt) != saltSize {
		return nil, ErrInvalidSaltSize
	}

	if params.RecordSize == 0 {
		params.RecordSize = DefaultRecordSize
	}

	if params.Version == "" {
		params.Version = AES128GCM
	}

	var err error

	if params.Version == AES128GCM {
		target, err = e.appendHeader(target, params)
		if err != nil {
			return nil, err
		}
	}

	publicKey := e.publicKey(params)
	var receiverPublicKey *ecdsa.PublicKey
	if params.StaticKey == nil {
		receiverPublicKey, err = unmarshalPublicKey(publicKey, params.DH)
		if err != nil {
			return nil, err
		}
	}

	key, nonce, err := e.deriveKey(
		params,
		receiverPublicKey,
		e.buildInfoContext(params.Version, publicKey, receiverPublicKey),
	)
	if err != nil {
		return nil, fmt.Errorf("derive key/nonce failed: %w", err)
	}

	aead, err := newGCM(key)
	if err != nil {
		return nil, err
	}

	return encrypt(content, target, nonce, aead, params)
}

func (e *Engine) appendHeader(target []byte, params operationalParams) ([]byte, error) {
	target = append(target, params.Salt...)

	target = append(target, make([]byte, recordSizeSize)...)
	binary.BigEndian.PutUint32(target[len(target)-recordSizeSize:], params.RecordSize)

	keyID := params.KeyID
	if keyID == nil && params.DH != nil {
		publicKey := e.publicKey(params)
		keyID = elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	}

	keyIDLen := len(keyID)
	if keyIDLen > 255 {
		return nil, errKeyIDTooLong
	}

	target = append(target, byte(keyIDLen))
	target = append(target, keyID...)

	return target, nil
}

func encrypt(content, target, nonce []byte, aead cipher.AEAD, params operationalParams) ([]byte, error) {
	padSize := params.Version.paddingSize()
	overhead := uint32(padSize)
	contentLen := uint32(len(content))
	if params.Version == AES128GCM {
		overhead += TagSize
	}

	if params.RecordSize < overhead {
		return nil, errTooSmallRecordSize
	}

	var (
		blockNonce = make([]byte, NonceSize)
		pad        = params.Pad
		lastBlock  = false
		counter    = uint64(0)
		blockStart = uint32(0)
	)

	for ; !lastBlock; counter++ {
		recordPad := calculateRecordPad(pad, overhead, params)
		pad -= recordPad
		blockSize := params.RecordSize - overhead - recordPad
		blockEnd := blockStart + blockSize
		lastBlock = isLastBlock(pad, blockEnd, contentLen, params)

		if blockEnd > contentLen {
			blockEnd = contentLen
		}

		padded, err := padPlainText(content[blockStart:blockEnd], params, recordPad, lastBlock)
		if err != nil {
			return nil, err
		}

		fillBlockNonce(counter, nonce, blockNonce)
		target = aead.Seal(target, blockNonce, padded, nil)

		blockStart += blockSize
	}

	return target, nil
}

func calculateRecordPad(pad, overhead uint32, params operationalParams) uint32 {
	padSize := params.Version.paddingSize()
	// Pad so that at least one data byte is in a block.
	recordPad := params.RecordSize - overhead - 1
	if recordPad > pad {
		recordPad = pad
	}

	if params.Version != AES128GCM && recordPad > (1<<(padSize*8))-1 {
		recordPad = (1 << (padSize * 8)) - 1
	}

	if pad > 0 && recordPad == 0 {
		recordPad++ // Deal with perverse case of rs=overhead+1 with padding.
	}

	return recordPad
}

func isLastBlock(pad, blockEnd, contentLen uint32, params operationalParams) bool {
	var lastBlock bool

	if params.Version != AES128GCM {
		// The > here ensures that we write out a padding-only block at the end
		// of a buffer.
		lastBlock = blockEnd > contentLen
	} else {
		lastBlock = blockEnd >= contentLen
	}

	lastBlock = lastBlock && pad == 0
	return lastBlock
}

func padPlainText(plainTextBlock []byte, params operationalParams, pad uint32, lastBlock bool) ([]byte, error) {
	padSize := params.Version.paddingSize()
	ret := make([]byte, uint32(len(plainTextBlock)+padSize)+pad)
	if params.Version != AES128GCM {
		if !lastBlock && uint32(len(ret)) < params.RecordSize {
			return nil, errPaddingToRecord
		}

		copy(ret[uint32(padSize)+pad:], plainTextBlock)
		switch padSize {
		case 1:
			ret[0] = byte(pad)
		case 2:
			binary.BigEndian.PutUint16(ret, uint16(pad))
		}

		return ret, nil
	}

	delimiter := byte(1)
	if lastBlock {
		delimiter = 2
	}

	copy(ret, plainTextBlock)
	ret[len(plainTextBlock)] = delimiter
	return ret, nil
}
