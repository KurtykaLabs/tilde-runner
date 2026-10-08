// Command approved-releases creates and verifies Tilde's signed approved-releases list.
//
//	approved-releases keygen -out KEY                 # writes the private key, prints the public key
//	approved-releases sign -key KEY -in DOC.json [-introduction PREVIOUS.json] [-previous CURRENT.json] -out LIST.json
//	approved-releases verify -pub BASE64 -in LIST.json  # prints the verified document
//
// The private key file holds the base64 Ed25519 seed. Keep it offline and out of Git.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"com.tilde/runner-client/internal/approval"
)

func main() {
	if len(os.Args) < 2 {
		fail(errors.New("usage: approved-releases keygen|sign|verify [flags]"))
	}
	var err error
	switch os.Args[1] {
	case "keygen":
		err = keygen(os.Args[2:])
	case "sign":
		err = sign(os.Args[2:])
	case "verify":
		err = verifyList(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func keygen(args []string) error {
	flags := flag.NewFlagSet("keygen", flag.ExitOnError)
	out := flags.String("out", "", "private key file to create (must not exist)")
	_ = flags.Parse(args)
	if *out == "" {
		return errors.New("-out is required")
	}
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(file, base64.StdEncoding.EncodeToString(private.Seed())); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	fmt.Println(base64.StdEncoding.EncodeToString(public))
	return nil
}

func readKey(path string) (ed25519.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	seed, err := base64.StdEncoding.Strict().DecodeString(strings.TrimSpace(string(raw)))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, errors.New("key file must hold a base64 32-byte Ed25519 seed")
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

func sign(args []string) error {
	flags := flag.NewFlagSet("sign", flag.ExitOnError)
	keyPath := flags.String("key", "", "private key file")
	in := flags.String("in", "", "unsigned document JSON")
	introduction := flags.String("introduction", "", "the signed list that named this key as successor_key (successor keys only)")
	previous := flags.String("previous", "", "the currently published signed list; the new version must exceed it")
	out := flags.String("out", "", "signed list to write")
	_ = flags.Parse(args)
	if *keyPath == "" || *in == "" || *out == "" {
		return errors.New("-key, -in and -out are required")
	}
	key, err := readKey(*keyPath)
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(*in)
	if err != nil {
		return err
	}
	document, _, err := approval.ParseDocument(raw)
	if err != nil {
		return fmt.Errorf("document: %w", err)
	}
	if *previous != "" {
		current, err := os.ReadFile(*previous)
		if err != nil {
			return err
		}
		var envelope approval.Envelope
		if err := json.Unmarshal(current, &envelope); err != nil {
			return fmt.Errorf("previous: %w", err)
		}
		payload, err := base64.StdEncoding.DecodeString(envelope.Payload)
		if err != nil {
			return fmt.Errorf("previous: %w", err)
		}
		published, _, err := approval.ParseDocument(payload)
		if err != nil {
			return fmt.Errorf("previous: %w", err)
		}
		if document.Version <= published.Version {
			return fmt.Errorf("version %d must exceed the published version %d", document.Version, published.Version)
		}
	}
	var intro []byte
	if *introduction != "" {
		if intro, err = os.ReadFile(*introduction); err != nil {
			return err
		}
	}
	signed, err := approval.Sign(document, key, intro)
	if err != nil {
		return err
	}
	return os.WriteFile(*out, append(signed, '\n'), 0o644)
}

func verifyList(args []string) error {
	flags := flag.NewFlagSet("verify", flag.ExitOnError)
	pub := flags.String("pub", "", "trusted approval public key (base64), normally the app's compiled key")
	in := flags.String("in", "", "signed list")
	_ = flags.Parse(args)
	anchor, err := approval.DecodeKey(*pub)
	if err != nil {
		return errors.New("-pub must be a base64 Ed25519 public key")
	}
	raw, err := os.ReadFile(*in)
	if err != nil {
		return err
	}
	verified, err := approval.Verify([]byte(strings.TrimSpace(string(raw))), anchor)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(verified.Document, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	fmt.Fprintf(os.Stderr, "verified: version %d, signer %s, chain length %d\n",
		verified.Document.Version, base64.StdEncoding.EncodeToString(verified.Signer), len(verified.Chain))
	return nil
}
