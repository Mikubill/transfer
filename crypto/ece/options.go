package ece

// https://github.com/xakep666/ecego, MIT license, modified

type engineOption interface {
	apply(*Engine)
}

type keyLabelOption string

func (k keyLabelOption) apply(e *Engine) { e.keyLabel = string(k) }

// WithKeyLabel sets a key label to use
func WithKeyLabel(keyLabel string) engineOption { return keyLabelOption(keyLabel) }

type authSecretOption []byte

func (a authSecretOption) apply(e *Engine) { e.authSecret = a }

// WithAuthSecret specifies auth secret for shared key derivation
func WithAuthSecret(authSecret []byte) engineOption { return authSecretOption(authSecret) }
