package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sroopra/ghega/pkg/channel"
)

func TestGenerateMigrationReports_SingleChannel(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join(tmpDir, "export")
	outDir := filepath.Join(tmpDir, "out")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chXML := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-001</id>
  <name>Simple Channel</name>
  <description>A simple channel</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>MLLP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.tcp.TcpListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>6661</port>
      </listenerConnectorProperties>
    </properties>
    <transformer>
      <steps/>
    </transformer>
    <filter>
      <rules/>
    </filter>
  </sourceConnector>
  <destinationConnectors>
    <connector>
      <name>HTTP Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.http.HttpDispatcherProperties">
        <host>example.com</host>
        <port>80</port>
        <method>POST</method>
      </properties>
      <transformer>
        <steps/>
      </transformer>
      <filter>
        <rules/>
      </filter>
    </connector>
  </destinationConnectors>
  <properties/>
</channel>`

	if err := os.WriteFile(filepath.Join(exportDir, "simple.xml"), []byte(chXML), 0644); err != nil {
		t.Fatalf("write channel xml: %v", err)
	}

	summary, err := GenerateMigrationReports(exportDir, outDir, "", "")
	if err != nil {
		t.Fatalf("generate reports: %v", err)
	}

	if summary.TotalChannels != 1 {
		t.Errorf("expected 1 channel, got %d", summary.TotalChannels)
	}
	if summary.TotalAutoConverted != 1 {
		t.Errorf("expected 1 auto-converted, got %d", summary.TotalAutoConverted)
	}

	// Check summary report file.
	summaryPath := filepath.Join(outDir, "migration-report.yaml")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Fatalf("expected summary report")
	}

	// Check per-channel files.
	chDir := filepath.Join(outDir, "simple-channel")
	for _, f := range []string{"channel.yaml", "rewrite-tasks.yaml", "migration-report.yaml"} {
		p := filepath.Join(chDir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Fatalf("expected %s", p)
		}
	}
}

func TestGenerateMigrationReports_WithScripts(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join(tmpDir, "export")
	outDir := filepath.Join(tmpDir, "out")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chXML := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-002</id>
  <name>Script Channel</name>
  <description>Channel with scripts</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>MLLP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.tcp.TcpListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>6661</port>
      </listenerConnectorProperties>
    </properties>
    <transformer>
      <steps>
        <step>
          <sequenceNumber>0</sequenceNumber>
          <name>Set Name</name>
          <script>if (msg['PID']['PID.5']['PID.5.1'] == '') { msg['PID']['PID.5']['PID.5.1'] = 'UNKNOWN'; }</script>
          <type>JavaScript</type>
        </step>
      </steps>
    </transformer>
    <filter>
      <rules/>
    </filter>
  </sourceConnector>
  <destinationConnectors>
    <connector>
      <name>HTTP Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.http.HttpDispatcherProperties">
        <host>example.com</host>
        <port>80</port>
        <method>POST</method>
      </properties>
      <transformer>
        <steps/>
      </transformer>
      <filter>
        <rules/>
      </filter>
    </connector>
  </destinationConnectors>
  <properties/>
</channel>`

	if err := os.WriteFile(filepath.Join(exportDir, "script.xml"), []byte(chXML), 0644); err != nil {
		t.Fatalf("write channel xml: %v", err)
	}

	summary, err := GenerateMigrationReports(exportDir, outDir, "", "")
	if err != nil {
		t.Fatalf("generate reports: %v", err)
	}

	if summary.TotalChannels != 1 {
		t.Errorf("expected 1 channel, got %d", summary.TotalChannels)
	}
	if summary.TotalAutoConverted != 0 {
		t.Errorf("expected 0 auto-converted (script channel has JS), got %d", summary.TotalAutoConverted)
	}
	if summary.TotalNeedsRewrite != 1 {
		t.Errorf("expected 1 needs-rewrite, got %d", summary.TotalNeedsRewrite)
	}

	// Verify rewrite tasks exist.
	rtPath := filepath.Join(outDir, "script-channel", "rewrite-tasks.yaml")
	data, err := os.ReadFile(rtPath)
	if err != nil {
		t.Fatalf("read rewrite tasks: %v", err)
	}
	if !strings.Contains(string(data), "severity:") {
		t.Errorf("expected rewrite tasks with severity, got:\n%s", string(data))
	}
}

func TestGenerateMigrationReports_NoPHI(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join(tmpDir, "export")
	outDir := filepath.Join(tmpDir, "out")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Use a channel name that could look like PHI but is synthetic.
	chXML := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-phi-test</id>
  <name>Test Patient Channel</name>
  <description>Synthetic test data only</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>MLLP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.tcp.TcpListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>6661</port>
      </listenerConnectorProperties>
    </properties>
    <transformer>
      <steps/>
    </transformer>
    <filter>
      <rules/>
    </filter>
  </sourceConnector>
  <destinationConnectors>
    <connector>
      <name>HTTP Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.http.HttpDispatcherProperties">
        <host>example.com</host>
        <port>80</port>
        <method>POST</method>
      </properties>
      <transformer>
        <steps/>
      </transformer>
      <filter>
        <rules/>
      </filter>
    </connector>
  </destinationConnectors>
  <properties/>
</channel>`

	if err := os.WriteFile(filepath.Join(exportDir, "phi.xml"), []byte(chXML), 0644); err != nil {
		t.Fatalf("write channel xml: %v", err)
	}

	_, err := GenerateMigrationReports(exportDir, outDir, "", "")
	if err != nil {
		t.Fatalf("generate reports: %v", err)
	}

	// Walk output and ensure no real PHI is present.
	// Since the input is synthetic, this is a sanity check.
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("read out dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(outDir, entry.Name()))
		if err != nil {
			t.Fatalf("read file: %v", err)
		}
		content := string(data)
		if strings.Contains(content, "SSN") || strings.Contains(content, "123-45-6789") {
			t.Errorf("potential PHI found in %s", entry.Name())
		}
	}
}

func TestWriteChannelYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "channel.yaml")

	// Actually test with a real Ghega channel.
	gch := channel.Channel{
		Name: "test-channel",
		Source: channel.Source{
			Type: "mllp",
			Config: map[string]any{
				"host": "0.0.0.0",
				"port": 2575,
			},
		},
		Destination: channel.Destination{
			Type: "http",
			Config: map[string]any{
				"url":    "http://example.com/",
				"method": "POST",
			},
		},
	}

	if err := WriteChannelYAML(gch, path); err != nil {
		t.Fatalf("write channel yaml: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read channel yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "name: test-channel") {
		t.Errorf("expected name in yaml, got:\n%s", content)
	}
	if !strings.Contains(content, "type: mllp") {
		t.Errorf("expected source type in yaml, got:\n%s", content)
	}
}

func TestGenerateMigrationReports_NameCollision(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join(tmpDir, "export")
	outDir := filepath.Join(tmpDir, "out")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	baseXML := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>%s</id>
  <name>%s</name>
  <description>Collision test</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>MLLP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.tcp.TcpListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>6661</port>
      </listenerConnectorProperties>
    </properties>
    <transformer>
      <steps/>
    </transformer>
    <filter>
      <rules/>
    </filter>
  </sourceConnector>
  <destinationConnectors>
    <connector>
      <name>HTTP Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.http.HttpDispatcherProperties">
        <host>example.com</host>
        <port>80</port>
        <method>POST</method>
      </properties>
      <transformer>
        <steps/>
      </transformer>
      <filter>
        <rules/>
      </filter>
    </connector>
  </destinationConnectors>
  <properties/>
</channel>`

	// "ADT Feed" and "ADT_Feed" both sanitize to "adt-feed"
	ch1 := fmt.Sprintf(baseXML, "ch-001", "ADT Feed")
	ch2 := fmt.Sprintf(baseXML, "ch-002", "ADT_Feed")

	if err := os.WriteFile(filepath.Join(exportDir, "adt1.xml"), []byte(ch1), 0644); err != nil {
		t.Fatalf("write channel xml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(exportDir, "adt2.xml"), []byte(ch2), 0644); err != nil {
		t.Fatalf("write channel xml: %v", err)
	}

	summary, err := GenerateMigrationReports(exportDir, outDir, "", "")
	if err != nil {
		t.Fatalf("generate reports: %v", err)
	}

	if summary.TotalChannels != 2 {
		t.Errorf("expected 2 channels, got %d", summary.TotalChannels)
	}

	// Both should be auto-converted.
	if summary.TotalAutoConverted != 2 {
		t.Errorf("expected 2 auto-converted, got %d", summary.TotalAutoConverted)
	}

	// Verify both directories exist.
	firstDir := filepath.Join(outDir, "adt-feed")
	secondDir := filepath.Join(outDir, "adt-feed-2")
	if _, err := os.Stat(firstDir); os.IsNotExist(err) {
		t.Fatalf("expected first directory %s to exist", firstDir)
	}
	if _, err := os.Stat(secondDir); os.IsNotExist(err) {
		t.Fatalf("expected second directory %s to exist", secondDir)
	}

	// Verify warning in second report.
	rptPath := filepath.Join(secondDir, "migration-report.yaml")
	data, err := os.ReadFile(rptPath)
	if err != nil {
		t.Fatalf("read migration report: %v", err)
	}
	if !strings.Contains(string(data), "collision") {
		t.Errorf("expected name collision warning in report, got:\n%s", string(data))
	}

	// Verify channel.yaml name matches the renamed directory.
	chPath := filepath.Join(secondDir, "channel.yaml")
	chData, err := os.ReadFile(chPath)
	if err != nil {
		t.Fatalf("read channel.yaml: %v", err)
	}
	if !strings.Contains(string(chData), "name: adt-feed-2") {
		t.Errorf("expected channel.yaml name to match renamed directory, got:\n%s", string(chData))
	}
}

func TestGenerateMigrationReports_WithSamplesNoExpected(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join(tmpDir, "export")
	outDir := filepath.Join(tmpDir, "out")
	samplesDir := filepath.Join(tmpDir, "samples")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(samplesDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chXML := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-003</id>
  <name>Sample Channel</name>
  <description>Channel with a static mapping</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>MLLP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.tcp.TcpListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>6661</port>
      </listenerConnectorProperties>
    </properties>
    <transformer>
      <steps>
        <step>
          <sequenceNumber>0</sequenceNumber>
          <name>Set MRN</name>
          <script>msg['PID']['PID.3']['PID.3.1'] = 'SYNTHETIC_MRN';</script>
          <type>JavaScript</type>
        </step>
      </steps>
    </transformer>
    <filter>
      <rules/>
    </filter>
  </sourceConnector>
  <destinationConnectors>
    <connector>
      <name>HTTP Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.http.HttpDispatcherProperties">
        <host>example.com</host>
        <port>80</port>
        <method>POST</method>
      </properties>
      <transformer>
        <steps/>
      </transformer>
      <filter>
        <rules/>
      </filter>
    </connector>
  </destinationConnectors>
  <properties/>
</channel>`

	if err := os.WriteFile(filepath.Join(exportDir, "sample.xml"), []byte(chXML), 0644); err != nil {
		t.Fatalf("write channel xml: %v", err)
	}

	hl7 := "MSH|^~\\&|SENDING|FACILITY|RECEIVING|FACILITY|20240101||ADT^A01|12345|P|2.3\rPID|1||REAL_MRN^^^MRN||DOE^JOHN"
	if err := os.WriteFile(filepath.Join(samplesDir, "sample1.hl7"), []byte(hl7), 0644); err != nil {
		t.Fatalf("write sample hl7: %v", err)
	}

	summary, err := GenerateMigrationReports(exportDir, outDir, samplesDir, "")
	if err != nil {
		t.Fatalf("generate reports: %v", err)
	}

	if summary.TotalSamples != 1 {
		t.Errorf("expected 1 total sample, got %d", summary.TotalSamples)
	}
	if summary.TotalMatched != 0 {
		t.Errorf("expected 0 matched without expected, got %d", summary.TotalMatched)
	}

	// Verify per-channel report contains sample results.
	rptPath := filepath.Join(outDir, "sample-channel", "migration-report.yaml")
	rptData, err := os.ReadFile(rptPath)
	if err != nil {
		t.Fatalf("read migration report: %v", err)
	}
	if !strings.Contains(string(rptData), "sampleResults:") {
		t.Errorf("expected sampleResults in report, got:\n%s", string(rptData))
	}
	if !strings.Contains(string(rptData), "status: no-expected") {
		t.Errorf("expected no-expected status, got:\n%s", string(rptData))
	}
	if !strings.Contains(string(rptData), "sample: sample1.hl7") {
		t.Errorf("expected sample name, got:\n%s", string(rptData))
	}
}

func TestGenerateMigrationReports_WithSamplesAndExpected(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join(tmpDir, "export")
	outDir := filepath.Join(tmpDir, "out")
	samplesDir := filepath.Join(tmpDir, "samples")
	expectedDir := filepath.Join(tmpDir, "expected")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(samplesDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(expectedDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chXML := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-004</id>
  <name>Expected Channel</name>
  <description>Channel with a static mapping</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>MLLP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.tcp.TcpListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>6661</port>
      </listenerConnectorProperties>
    </properties>
    <transformer>
      <steps>
        <step>
          <sequenceNumber>0</sequenceNumber>
          <name>Set MRN</name>
          <script>msg['PID']['PID.3']['PID.3.1'] = 'SYNTHETIC_MRN';</script>
          <type>JavaScript</type>
        </step>
      </steps>
    </transformer>
    <filter>
      <rules/>
    </filter>
  </sourceConnector>
  <destinationConnectors>
    <connector>
      <name>HTTP Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.http.HttpDispatcherProperties">
        <host>example.com</host>
        <port>80</port>
        <method>POST</method>
      </properties>
      <transformer>
        <steps/>
      </transformer>
      <filter>
        <rules/>
      </filter>
    </connector>
  </destinationConnectors>
  <properties/>
</channel>`

	if err := os.WriteFile(filepath.Join(exportDir, "expected.xml"), []byte(chXML), 0644); err != nil {
		t.Fatalf("write channel xml: %v", err)
	}

	hl7 := "MSH|^~\\&|SENDING|FACILITY|RECEIVING|FACILITY|20240101||ADT^A01|12345|P|2.3\rPID|1||REAL_MRN^^^MRN||DOE^JOHN"
	if err := os.WriteFile(filepath.Join(samplesDir, "sample1.hl7"), []byte(hl7), 0644); err != nil {
		t.Fatalf("write sample hl7: %v", err)
	}
	if err := os.WriteFile(filepath.Join(samplesDir, "sample2.hl7"), []byte(hl7), 0644); err != nil {
		t.Fatalf("write sample2 hl7: %v", err)
	}

	// Expected for sample1 matches the engine output.
	if err := os.WriteFile(filepath.Join(expectedDir, "sample1.expected"), []byte("PID-3.1: SYNTHETIC_MRN"), 0644); err != nil {
		t.Fatalf("write expected file: %v", err)
	}
	// Expected for sample2 is intentionally different to trigger mismatch.
	if err := os.WriteFile(filepath.Join(expectedDir, "sample2.txt"), []byte("PID-3.1: WRONG_MRN"), 0644); err != nil {
		t.Fatalf("write expected file: %v", err)
	}

	summary, err := GenerateMigrationReports(exportDir, outDir, samplesDir, expectedDir)
	if err != nil {
		t.Fatalf("generate reports: %v", err)
	}

	if summary.TotalSamples != 2 {
		t.Errorf("expected 2 total samples, got %d", summary.TotalSamples)
	}
	if summary.TotalMatched != 1 {
		t.Errorf("expected 1 matched, got %d", summary.TotalMatched)
	}

	rptPath := filepath.Join(outDir, "expected-channel", "migration-report.yaml")
	rptData, err := os.ReadFile(rptPath)
	if err != nil {
		t.Fatalf("read migration report: %v", err)
	}
	content := string(rptData)
	if !strings.Contains(content, "status: matched") {
		t.Errorf("expected matched status in report, got:\n%s", content)
	}
	if !strings.Contains(content, "status: mismatched") {
		t.Errorf("expected mismatched status in report, got:\n%s", content)
	}
}

func TestGenerateMigrationReports_SamplesMissingExpectedFile(t *testing.T) {
	tmpDir := t.TempDir()
	exportDir := filepath.Join(tmpDir, "export")
	outDir := filepath.Join(tmpDir, "out")
	samplesDir := filepath.Join(tmpDir, "samples")
	expectedDir := filepath.Join(tmpDir, "expected")

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(samplesDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(expectedDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	chXML := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-005</id>
  <name>Missing Expected Channel</name>
  <description>Channel with a static mapping</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>MLLP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.tcp.TcpListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>6661</port>
      </listenerConnectorProperties>
    </properties>
    <transformer>
      <steps>
        <step>
          <sequenceNumber>0</sequenceNumber>
          <name>Set MRN</name>
          <script>msg['PID']['PID.3']['PID.3.1'] = 'SYNTHETIC_MRN';</script>
          <type>JavaScript</type>
        </step>
      </steps>
    </transformer>
    <filter>
      <rules/>
    </filter>
  </sourceConnector>
  <destinationConnectors>
    <connector>
      <name>HTTP Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.http.HttpDispatcherProperties">
        <host>example.com</host>
        <port>80</port>
        <method>POST</method>
      </properties>
      <transformer>
        <steps/>
      </transformer>
      <filter>
        <rules/>
      </filter>
    </connector>
  </destinationConnectors>
  <properties/>
</channel>`

	if err := os.WriteFile(filepath.Join(exportDir, "missing.xml"), []byte(chXML), 0644); err != nil {
		t.Fatalf("write channel xml: %v", err)
	}

	hl7 := "MSH|^~\\&|SENDING|FACILITY|RECEIVING|FACILITY|20240101||ADT^A01|12345|P|2.3\rPID|1||REAL_MRN^^^MRN||DOE^JOHN"
	if err := os.WriteFile(filepath.Join(samplesDir, "orphan.hl7"), []byte(hl7), 0644); err != nil {
		t.Fatalf("write sample hl7: %v", err)
	}

	// No expected file for orphan.hl7

	summary, err := GenerateMigrationReports(exportDir, outDir, samplesDir, expectedDir)
	if err != nil {
		t.Fatalf("generate reports: %v", err)
	}

	if summary.TotalSamples != 1 {
		t.Errorf("expected 1 total sample, got %d", summary.TotalSamples)
	}
	if summary.TotalMatched != 0 {
		t.Errorf("expected 0 matched, got %d", summary.TotalMatched)
	}

	rptPath := filepath.Join(outDir, "missing-expected-channel", "migration-report.yaml")
	rptData, err := os.ReadFile(rptPath)
	if err != nil {
		t.Fatalf("read migration report: %v", err)
	}
	if !strings.Contains(string(rptData), "status: no-expected") {
		t.Errorf("expected no-expected status, got:\n%s", string(rptData))
	}
}
