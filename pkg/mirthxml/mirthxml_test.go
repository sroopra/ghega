package mirthxml

import (
	"os"
	"path/filepath"
	"testing"
)

// syntheticChannelXML is a minimal, non-PHI Mirth channel export.
const syntheticChannelXML = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>test-channel-001</id>
  <name>TestChannel</name>
  <description>A synthetic test channel</description>
  <enabled>true</enabled>
  <revision>2</revision>
  <sourceConnector>
    <name>source</name>
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
          <name>Set Patient Name</name>
          <script>msg['PID']['PID.5']['PID.5.1'] = 'SyntheticPatient';</script>
          <type>JavaScript</type>
        </step>
      </steps>
    </transformer>
    <filter>
      <rules>
        <rule>
          <sequenceNumber>0</sequenceNumber>
          <name>Always true</name>
          <script>true;</script>
          <operator>AND</operator>
        </rule>
      </rules>
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
    <connector>
      <name>File Destination</name>
      <enabled>false</enabled>
      <properties class="com.mirth.connect.connectors.file.FileDispatcherProperties">
        <scheme>FILE</scheme>
        <host/>
        <directory>/tmp/synthetic/out</directory>
      </properties>
      <transformer>
        <steps/>
      </transformer>
      <filter>
        <rules/>
      </filter>
    </connector>
  </destinationConnectors>
  <properties>
    <property>
      <name>myProp</name>
      <value>myValue</value>
    </property>
  </properties>
</channel>`

const syntheticCodeTemplateXML = `<?xml version="1.0" encoding="UTF-8"?>
<codeTemplate>
  <id>ct-001</id>
  <name>Helper Function</name>
  <description>A synthetic helper</description>
  <code>function helper() { return 42; }</code>
  <context>CHANNEL</context>
  <type>FUNCTION</type>
</codeTemplate>`

const syntheticCodeTemplateLibraryXML = `<?xml version="1.0" encoding="UTF-8"?>
<codeTemplateLibrary>
  <id>lib-001</id>
  <name>Test Library</name>
  <description>A synthetic library</description>
  <codeTemplates>
    <codeTemplate>
      <id>ct-002</id>
      <name>Logger</name>
      <description>Log a message</description>
      <code>logger.info('synthetic log');</code>
      <context>GLOBAL</context>
      <type>FUNCTION</type>
    </codeTemplate>
  </codeTemplates>
</codeTemplateLibrary>`

func TestParseChannel(t *testing.T) {
	ch, err := ParseChannel([]byte(syntheticChannelXML))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	if ch.Version != "3.12.0" {
		t.Errorf("version: got %q, want %q", ch.Version, "3.12.0")
	}
	if ch.ID != "test-channel-001" {
		t.Errorf("id: got %q", ch.ID)
	}
	if ch.Name != "TestChannel" {
		t.Errorf("name: got %q", ch.Name)
	}
	if ch.Description != "A synthetic test channel" {
		t.Errorf("description: got %q", ch.Description)
	}
	if !ch.Enabled {
		t.Error("expected enabled=true")
	}
	if ch.Revision != 2 {
		t.Errorf("revision: got %d", ch.Revision)
	}
}

func TestParseChannelSourceConnector(t *testing.T) {
	ch, err := ParseChannel([]byte(syntheticChannelXML))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	src := ch.SourceConnector
	if src.Name != "source" {
		t.Errorf("source name: got %q", src.Name)
	}
	if !src.Enabled {
		t.Error("expected source enabled=true")
	}
	if src.Properties.Class != "com.mirth.connect.connectors.tcp.TcpListenerProperties" {
		t.Errorf("source properties class: got %q", src.Properties.Class)
	}

	// Verify typed property unmarshaling.
	var tcpProps TcpListenerProperties
	if err := src.Properties.UnmarshalInto(&tcpProps); err != nil {
		t.Fatalf("unmarshal tcp listener properties: %v", err)
	}
	if tcpProps.ListenerConnectorProperties.Host != "0.0.0.0" {
		t.Errorf("tcp host: got %q", tcpProps.ListenerConnectorProperties.Host)
	}
	if tcpProps.ListenerConnectorProperties.Port != 6661 {
		t.Errorf("tcp port: got %d", tcpProps.ListenerConnectorProperties.Port)
	}
}

func TestParseChannelTransformerAndFilter(t *testing.T) {
	ch, err := ParseChannel([]byte(syntheticChannelXML))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	src := ch.SourceConnector
	if len(src.Transformer.Steps) != 1 {
		t.Fatalf("expected 1 transformer step, got %d", len(src.Transformer.Steps))
	}
	step := src.Transformer.Steps[0]
	if step.SequenceNumber != 0 {
		t.Errorf("step sequence: got %d", step.SequenceNumber)
	}
	if step.Name != "Set Patient Name" {
		t.Errorf("step name: got %q", step.Name)
	}
	if step.Type != "JavaScript" {
		t.Errorf("step type: got %q", step.Type)
	}
	wantScript := "msg['PID']['PID.5']['PID.5.1'] = 'SyntheticPatient';"
	if step.Script != wantScript {
		t.Errorf("step script: got %q, want %q", step.Script, wantScript)
	}

	if len(src.Filter.Rules) != 1 {
		t.Fatalf("expected 1 filter rule, got %d", len(src.Filter.Rules))
	}
	rule := src.Filter.Rules[0]
	if rule.SequenceNumber != 0 {
		t.Errorf("rule sequence: got %d", rule.SequenceNumber)
	}
	if rule.Name != "Always true" {
		t.Errorf("rule name: got %q", rule.Name)
	}
	if rule.Operator != "AND" {
		t.Errorf("rule operator: got %q", rule.Operator)
	}
}

func TestParseChannelDestinationConnectors(t *testing.T) {
	ch, err := ParseChannel([]byte(syntheticChannelXML))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	if len(ch.DestinationConnectors) != 2 {
		t.Fatalf("expected 2 destination connectors, got %d", len(ch.DestinationConnectors))
	}

	// First destination: HTTP
	httpDest := ch.DestinationConnectors[0]
	if httpDest.Name != "HTTP Destination" {
		t.Errorf("dest[0] name: got %q", httpDest.Name)
	}
	if !httpDest.Enabled {
		t.Error("expected dest[0] enabled=true")
	}
	if httpDest.Properties.Class != "com.mirth.connect.connectors.http.HttpDispatcherProperties" {
		t.Errorf("dest[0] class: got %q", httpDest.Properties.Class)
	}
	var httpProps HttpDispatcherProperties
	if err := httpDest.Properties.UnmarshalInto(&httpProps); err != nil {
		t.Fatalf("unmarshal http dispatcher properties: %v", err)
	}
	if httpProps.Host != "example.com" {
		t.Errorf("http host: got %q", httpProps.Host)
	}
	if httpProps.Port != 80 {
		t.Errorf("http port: got %d", httpProps.Port)
	}
	if httpProps.Method != "POST" {
		t.Errorf("http method: got %q", httpProps.Method)
	}

	// Second destination: File
	fileDest := ch.DestinationConnectors[1]
	if fileDest.Name != "File Destination" {
		t.Errorf("dest[1] name: got %q", fileDest.Name)
	}
	if fileDest.Enabled {
		t.Error("expected dest[1] enabled=false")
	}
	var fileProps FileDispatcherProperties
	if err := fileDest.Properties.UnmarshalInto(&fileProps); err != nil {
		t.Fatalf("unmarshal file dispatcher properties: %v", err)
	}
	if fileProps.Scheme != "FILE" {
		t.Errorf("file scheme: got %q", fileProps.Scheme)
	}
	if fileProps.Directory != "/tmp/synthetic/out" {
		t.Errorf("file directory: got %q", fileProps.Directory)
	}
}

func TestParseChannelProperties(t *testing.T) {
	ch, err := ParseChannel([]byte(syntheticChannelXML))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	if len(ch.Properties) != 1 {
		t.Fatalf("expected 1 channel property, got %d", len(ch.Properties))
	}
	prop := ch.Properties[0]
	if prop.Name != "myProp" {
		t.Errorf("prop name: got %q", prop.Name)
	}
	if prop.Value != "myValue" {
		t.Errorf("prop value: got %q", prop.Value)
	}
}

func TestParseCodeTemplate(t *testing.T) {
	ct, err := ParseCodeTemplate([]byte(syntheticCodeTemplateXML))
	if err != nil {
		t.Fatalf("parse code template: %v", err)
	}
	if ct.ID != "ct-001" {
		t.Errorf("id: got %q", ct.ID)
	}
	if ct.Name != "Helper Function" {
		t.Errorf("name: got %q", ct.Name)
	}
	if ct.Context != "CHANNEL" {
		t.Errorf("context: got %q", ct.Context)
	}
	if ct.Type != "FUNCTION" {
		t.Errorf("type: got %q", ct.Type)
	}
}

func TestParseCodeTemplateLibrary(t *testing.T) {
	lib, err := ParseCodeTemplateLibrary([]byte(syntheticCodeTemplateLibraryXML))
	if err != nil {
		t.Fatalf("parse code template library: %v", err)
	}
	if lib.ID != "lib-001" {
		t.Errorf("library id: got %q", lib.ID)
	}
	if len(lib.CodeTemplates) != 1 {
		t.Fatalf("expected 1 code template, got %d", len(lib.CodeTemplates))
	}
	ct := lib.CodeTemplates[0]
	if ct.Name != "Logger" {
		t.Errorf("template name: got %q", ct.Name)
	}
}

func TestParseChannelsFromDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Write two valid channel files.
	if err := os.WriteFile(filepath.Join(tmpDir, "channel1.xml"), []byte(syntheticChannelXML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "channel2.xml"), []byte(syntheticChannelXML), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write a non-XML file — should be ignored.
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write an invalid XML file — should be skipped.
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.xml"), []byte("<not-a-channel/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	channels, err := ParseChannelsFromDir(tmpDir)
	if err != nil {
		t.Fatalf("parse channels from dir: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(channels))
	}
}

func TestParseCodeTemplatesFromDir(t *testing.T) {
	tmpDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpDir, "template1.xml"), []byte(syntheticCodeTemplateXML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "template2.xml"), []byte(syntheticCodeTemplateXML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	templates, err := ParseCodeTemplatesFromDir(tmpDir)
	if err != nil {
		t.Fatalf("parse code templates from dir: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}
}

func TestParseChannelFileNotFound(t *testing.T) {
	_, err := ParseChannelFile("/nonexistent/path/channel.xml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseCodeTemplateFileNotFound(t *testing.T) {
	_, err := ParseCodeTemplateFile("/nonexistent/path/template.xml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPropertiesUnmarshalIntoEmpty(t *testing.T) {
	p := Properties{Class: "test"}
	var out struct {
		Host string `xml:"host"`
	}
	if err := p.UnmarshalInto(&out); err != nil {
		t.Fatalf("unmarshal empty properties: %v", err)
	}
	if out.Host != "" {
		t.Errorf("expected empty host, got %q", out.Host)
	}
}

func TestDatabaseReaderPropertiesUnmarshal(t *testing.T) {
	raw := []byte(`<driver>org.postgresql.Driver</driver>
		<url>jdbc:postgresql://localhost:5432/testdb</url>
		<username>admin</username>
		<password>secret</password>
		<query>SELECT * FROM patients</query>
		<pollingInterval>5000</pollingInterval>`)
	p := Properties{Class: "com.mirth.connect.connectors.jdbc.DatabaseReaderProperties", Raw: raw}

	var props DatabaseReaderProperties
	if err := p.UnmarshalInto(&props); err != nil {
		t.Fatalf("unmarshal database reader properties: %v", err)
	}
	if props.Driver != "org.postgresql.Driver" {
		t.Errorf("driver: got %q, want %q", props.Driver, "org.postgresql.Driver")
	}
	if props.URL != "jdbc:postgresql://localhost:5432/testdb" {
		t.Errorf("url: got %q", props.URL)
	}
	if props.Username != "admin" {
		t.Errorf("username: got %q", props.Username)
	}
	if props.Password != "secret" {
		t.Errorf("password: got %q", props.Password)
	}
	if props.Query != "SELECT * FROM patients" {
		t.Errorf("query: got %q", props.Query)
	}
	if props.PollingInterval != 5000 {
		t.Errorf("pollingInterval: got %d, want %d", props.PollingInterval, 5000)
	}
}

func TestDatabaseWriterPropertiesUnmarshal(t *testing.T) {
	raw := []byte(`<driver>com.mysql.jdbc.Driver</driver>
		<url>jdbc:mysql://db.example.com:3306/prod</url>
		<username>writer</username>
		<password>hunter2</password>
		<query>INSERT INTO logs (msg) VALUES (?)</query>`)
	p := Properties{Class: "com.mirth.connect.connectors.jdbc.DatabaseWriterProperties", Raw: raw}

	var props DatabaseWriterProperties
	if err := p.UnmarshalInto(&props); err != nil {
		t.Fatalf("unmarshal database writer properties: %v", err)
	}
	if props.Driver != "com.mysql.jdbc.Driver" {
		t.Errorf("driver: got %q", props.Driver)
	}
	if props.URL != "jdbc:mysql://db.example.com:3306/prod" {
		t.Errorf("url: got %q", props.URL)
	}
	if props.Username != "writer" {
		t.Errorf("username: got %q", props.Username)
	}
	if props.Password != "hunter2" {
		t.Errorf("password: got %q", props.Password)
	}
	if props.Query != "INSERT INTO logs (msg) VALUES (?)" {
		t.Errorf("query: got %q", props.Query)
	}
}

func TestSftpReceiverPropertiesUnmarshal(t *testing.T) {
	raw := []byte(`<host>sftp.example.com</host>
		<port>22</port>
		<username>sftpuser</username>
		<password>sftppass</password>
		<remotePath>/inbox/pending</remotePath>
		<pollingInterval>10000</pollingInterval>`)
	p := Properties{Class: "com.mirth.connect.connectors.sftp.SftpReceiverProperties", Raw: raw}

	var props SftpReceiverProperties
	if err := p.UnmarshalInto(&props); err != nil {
		t.Fatalf("unmarshal sftp receiver properties: %v", err)
	}
	if props.Host != "sftp.example.com" {
		t.Errorf("host: got %q", props.Host)
	}
	if props.Port != 22 {
		t.Errorf("port: got %d, want %d", props.Port, 22)
	}
	if props.Username != "sftpuser" {
		t.Errorf("username: got %q", props.Username)
	}
	if props.Password != "sftppass" {
		t.Errorf("password: got %q", props.Password)
	}
	if props.RemotePath != "/inbox/pending" {
		t.Errorf("remotePath: got %q", props.RemotePath)
	}
	if props.PollingInterval != 10000 {
		t.Errorf("pollingInterval: got %d, want %d", props.PollingInterval, 10000)
	}
}

func TestSftpDispatcherPropertiesUnmarshal(t *testing.T) {
	raw := []byte(`<host>sftp.example.com</host>
		<port>22</port>
		<username>sftpuser</username>
		<password>sftppass</password>
		<remotePath>/outbox/completed</remotePath>`)
	p := Properties{Class: "com.mirth.connect.connectors.sftp.SftpDispatcherProperties", Raw: raw}

	var props SftpDispatcherProperties
	if err := p.UnmarshalInto(&props); err != nil {
		t.Fatalf("unmarshal sftp dispatcher properties: %v", err)
	}
	if props.Host != "sftp.example.com" {
		t.Errorf("host: got %q", props.Host)
	}
	if props.Port != 22 {
		t.Errorf("port: got %d, want %d", props.Port, 22)
	}
	if props.Username != "sftpuser" {
		t.Errorf("username: got %q", props.Username)
	}
	if props.Password != "sftppass" {
		t.Errorf("password: got %q", props.Password)
	}
	if props.RemotePath != "/outbox/completed" {
		t.Errorf("remotePath: got %q", props.RemotePath)
	}
}
