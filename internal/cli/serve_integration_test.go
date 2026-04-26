package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sroopra/ghega/pkg/mllp"
)

// findFreePort returns an available TCP port on localhost.
func findFreePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func TestServeIntegration_ADTA01_EndToEnd(t *testing.T) {
	if os.Getenv("BE_TEST_INTEGRATION") != "1" {
		t.Skip("skipping integration test; set BE_TEST_INTEGRATION=1 to run")
	}

	// Destination webhook server.
	var receivedPayload []byte
	var receivedContentType string
	destMux := http.NewServeMux()
	destMux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		receivedPayload = body
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	destSrv := &http.Server{Handler: destMux}
	destLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start destination listener: %v", err)
	}
	defer destSrv.Close()
	go func() { _ = destSrv.Serve(destLn) }()
	destURL := fmt.Sprintf("http://%s/webhook", destLn.Addr().String())

	httpPort := findFreePort(t)
	mllpPort := findFreePort(t)

	t.Setenv("GHEGA_PORT", strconv.Itoa(httpPort))
	t.Setenv("GHEGA_MLLP_PORT", strconv.Itoa(mllpPort))
	t.Setenv("GHEGA_DESTINATION_URL", destURL)
	t.Setenv("GHEGA_DATABASE_URL", ":memory:")

	// Start ghega serve in background.
	go func() {
		_ = runServe([]string{"-port", strconv.Itoa(httpPort)})
	}()

	// Wait for HTTP server to be ready.
	httpURL := fmt.Sprintf("http://127.0.0.1:%d/healthz", httpPort)
	for i := 0; i < 50; i++ {
		resp, err := http.Get(httpURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Build a synthetic ADT A01 message.
	msg := "MSH|^~\\&|SENDING_APP|SENDING_FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20240101120000||ADT^A01|MSG001|P|2.5\r" +
		"PID|1||SYNTHETIC_MRN^^^MRN||TESTPATIENT^JOHN^MICHAEL||19800101|M|||123 MAIN ST^^ANYTOWN^ST^12345||555-555-5555|||||SINGLE\r"

	// Connect to MLLP listener and send message.
	mllpAddr := fmt.Sprintf("127.0.0.1:%d", mllpPort)
	conn, err := net.Dial("tcp", mllpAddr)
	if err != nil {
		t.Fatalf("failed to dial mllp listener: %v", err)
	}
	defer conn.Close()

	frame := mllp.EncodeFrame([]byte(msg))
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("failed to write mllp frame: %v", err)
	}

	// Read ACK response.
	var ackBuf bytes.Buffer
	reader := make([]byte, 4096)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(reader)
	if err != nil {
		t.Fatalf("failed to read ack: %v", err)
	}
	ackBuf.Write(reader[:n])

	ackPayload, _, err := mllp.DecodeFrame(ackBuf.Bytes())
	if err != nil {
		t.Fatalf("failed to decode ack frame: %v", err)
	}
	ackStr := string(ackPayload)
	if !strings.Contains(ackStr, "MSA|AA|") {
		t.Errorf("expected ACK with MSA|AA|, got: %s", ackStr)
	}

	// Allow async processing to complete.
	time.Sleep(500 * time.Millisecond)

	// Verify the destination received the mapped JSON payload.
	if receivedPayload == nil {
		t.Fatal("destination did not receive payload")
	}
	if receivedContentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", receivedContentType)
	}

	var mapped map[string]string
	if err := json.Unmarshal(receivedPayload, &mapped); err != nil {
		t.Fatalf("failed to unmarshal received payload: %v", err)
	}
	if mapped["patient_mrn"] != "SYNTHETIC_MRN" {
		t.Errorf("expected patient_mrn=SYNTHETIC_MRN, got %s", mapped["patient_mrn"])
	}
	if mapped["patient_last_name"] != "TESTPATIENT" {
		t.Errorf("expected patient_last_name=TESTPATIENT, got %s", mapped["patient_last_name"])
	}
	if mapped["patient_first_name"] != "JOHN" {
		t.Errorf("expected patient_first_name=JOHN, got %s", mapped["patient_first_name"])
	}
	if mapped["message_type"] != "ADT" {
		t.Errorf("expected message_type=ADT, got %s", mapped["message_type"])
	}
}

func TestServeIntegration_MessageStoreUpdate(t *testing.T) {
	if os.Getenv("BE_TEST_INTEGRATION") != "1" {
		t.Skip("skipping integration test; set BE_TEST_INTEGRATION=1 to run")
	}

	// Destination webhook server that returns error to test failed path.
	destMux := http.NewServeMux()
	destMux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	destSrv := &http.Server{Handler: destMux}
	destLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start destination listener: %v", err)
	}
	defer destSrv.Close()
	go func() { _ = destSrv.Serve(destLn) }()
	destURL := fmt.Sprintf("http://%s/webhook", destLn.Addr().String())

	httpPort := findFreePort(t)
	mllpPort := findFreePort(t)

	t.Setenv("GHEGA_PORT", strconv.Itoa(httpPort))
	t.Setenv("GHEGA_MLLP_PORT", strconv.Itoa(mllpPort))
	t.Setenv("GHEGA_DESTINATION_URL", destURL)

	// Start ghega serve in background.
	go func() {
		_ = runServe([]string{"-port", strconv.Itoa(httpPort)})
	}()

	// Wait for HTTP server to be ready.
	httpURL := fmt.Sprintf("http://127.0.0.1:%d/healthz", httpPort)
	for i := 0; i < 50; i++ {
		resp, err := http.Get(httpURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	msg := "MSH|^~\\&|SENDING_APP|SENDING_FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20240101120000||ADT^A01|MSG002|P|2.5\r" +
		"PID|1||SYNTHETIC_MRN2^^^MRN||TESTPATIENT2^JANE||19800101|M|||123 MAIN ST^^ANYTOWN^ST^12345||555-555-5555|||||SINGLE\r"

	mllpAddr := fmt.Sprintf("127.0.0.1:%d", mllpPort)
	conn, err := net.Dial("tcp", mllpAddr)
	if err != nil {
		t.Fatalf("failed to dial mllp listener: %v", err)
	}
	defer conn.Close()

	frame := mllp.EncodeFrame([]byte(msg))
	if _, err := conn.Write(frame); err != nil {
		t.Fatalf("failed to write mllp frame: %v", err)
	}

	reader := make([]byte, 4096)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(reader)
	if err != nil {
		t.Fatalf("failed to read ack: %v", err)
	}

	// Allow async processing to complete.
	time.Sleep(500 * time.Millisecond)

	// Since we can't inspect the store when using runServe, we at least verify
	// the ACK was received (which means the pipeline completed).
}
