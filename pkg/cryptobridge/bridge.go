// Package cryptobridge calls the local Rust Olm helper. It does not persist keys.
package cryptobridge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"time"
)

const maxRequest = 1 << 20
const maxResponse = 2 << 20

type boundedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, errors.New("crypto response too large")
	}
	return b.Buffer.Write(p)
}

// Call exchanges one JSON request over pipes. Responses can contain private keys.
func Call(ctx context.Context, binary string, request any) (json.RawMessage, error) {
	data, err := json.Marshal(request)
	if err != nil || len(data) > maxRequest {
		return nil, errors.New("invalid crypto request")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary)
	cmd.Stdin = bytes.NewReader(data)
	output := &boundedBuffer{limit: maxResponse}
	cmd.Stdout = output
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil || !json.Valid(output.Bytes()) {
		return nil, errors.New("local crypto operation failed")
	}
	return json.RawMessage(bytes.Clone(output.Bytes())), nil
}
