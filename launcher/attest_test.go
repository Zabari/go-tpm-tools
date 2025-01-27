package main

import (
	"bytes"
	"context"
	"log"
	"net"
	"testing"

	"github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm-tools/internal/test"
	"github.com/google/go-tpm-tools/launcher/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	servgrpc "github.com/google/go-tpm-tools/launcher/proto/attestation_verifier/v0"
)

func TestAttest(t *testing.T) {
	tpm := test.GetTPM(t)
	defer client.CheckedClose(t, tpm)

	server := grpc.NewServer()

	fakeServer := service.New()
	servgrpc.RegisterAttestationVerifierServer(server, &fakeServer)

	lis := bufconn.Listen(1024 * 1024)
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(context.Background(), "bufconn", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect to attestation service: %v", err)
	}
	// Cannot test a GCE key on the simulator.
	agent := CreateAttestationAgent(tpm, client.AttestationKeyECC, conn, placeholderFetcher)

	token, err := agent.Attest(context.Background())
	if err != nil {
		t.Errorf("failed to attest to Attestation Service: %v", err)
	}

	if !bytes.Equal(token, service.FakeToken) {
		t.Errorf("received unexpected token: %v, expected: %v", token, service.FakeToken)
	}
}

func placeholderFetcher(audience string) ([][]byte, error) {
	return [][]byte{}, nil
}
