package migration

import (
	"strings"
	"testing"

	"github.com/sroopra/ghega/pkg/mirthxml"
)

// syntheticMLLPChannel is a Mirth channel with an MLLP source and HTTP destination.
const syntheticMLLPChannel = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-001</id>
  <name>ADT A01 Feed</name>
  <description>ADT A01 from Mirth</description>
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
        <host> downstream.example.com</host>
        <port>8080</port>
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

// syntheticHTTPChannel is a Mirth channel with an HTTP source and file destination.
const syntheticHTTPChannel = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-002</id>
  <name>HTTP_Ingest</name>
  <description>HTTP ingestion channel</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>HTTP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.http.HttpListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>8080</port>
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
      <name>File Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.file.FileDispatcherProperties">
        <scheme>FILE</scheme>
        <host/>
        <directory>/tmp/ghega/out</directory>
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

// syntheticFileChannel is a Mirth channel with a File Reader source and HTTP destination.
const syntheticFileChannel = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-003</id>
  <name>File_Poller</name>
  <description>File polling channel</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>File Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.file.FileReceiverProperties">
      <scheme>FILE</scheme>
      <host/>
      <directory>/tmp/ghega/in</directory>
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
        <host>api.example.com</host>
        <port>443</port>
        <method>PUT</method>
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

// syntheticUnsupportedChannel uses an unsupported source connector.
const syntheticUnsupportedChannel = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-004</id>
  <name>Unsupported_Source</name>
  <description>Channel with unsupported connector</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>Web Service Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.ws.WebServiceListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>9000</port>
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
        <method>GET</method>
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

// syntheticMultiDestChannel has multiple destinations.
const syntheticMultiDestChannel = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-005</id>
  <name>Multi_Dest</name>
  <description>Channel with multiple destinations</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>MLLP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.tcp.TcpListenerProperties">
      <listenerConnectorProperties>
        <host>0.0.0.0</host>
        <port>2575</port>
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
      <name>HTTP Primary</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.http.HttpDispatcherProperties">
        <host>primary.example.com</host>
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
      <name>File Secondary</name>
      <enabled>false</enabled>
      <properties class="com.mirth.connect.connectors.file.FileDispatcherProperties">
        <scheme>FILE</scheme>
        <host/>
        <directory>/tmp/secondary</directory>
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

// syntheticScriptChannel has transformer/filter scripts.
const syntheticScriptChannel = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-006</id>
  <name>Script_Channel</name>
  <description>Channel with JS scripts</description>
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
          <script>msg['PID']['PID.5']['PID.5.1'] = 'Test';</script>
          <type>JavaScript</type>
        </step>
      </steps>
    </transformer>
    <filter>
      <rules>
        <rule>
          <sequenceNumber>0</sequenceNumber>
          <name>Always True</name>
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
  </destinationConnectors>
  <preprocessingScript>logger.info('pre');</preprocessingScript>
  <postprocessingScript>logger.info('post');</postprocessingScript>
  <properties/>
</channel>`

// syntheticDBChannel has database connectors.
const syntheticDBChannel = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-007</id>
  <name>DB_Channel</name>
  <description>Database channel</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>DB Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.jdbc.DatabaseReaderProperties">
      <driver>org.postgresql.Driver</driver>
      <url>jdbc:postgresql://localhost/db</url>
      <query>SELECT * FROM patients</query>
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
      <name>DB Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.jdbc.DatabaseWriterProperties">
        <driver>org.postgresql.Driver</driver>
        <url>jdbc:postgresql://localhost/db</url>
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

func TestConvertChannel_Nil(t *testing.T) {
	_, err := ConvertChannel(nil)
	if err == nil {
		t.Fatal("expected error for nil channel")
	}
}

func TestConvertChannel_MLLPToMLLP(t *testing.T) {
	mch, err := mirthxml.ParseChannel([]byte(syntheticMLLPChannel))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Name != "adt-a01-feed" {
		t.Errorf("name: got %q, want %q", res.Channel.Name, "adt-a01-feed")
	}
	if res.Channel.Description != "ADT A01 from Mirth" {
		t.Errorf("description: got %q", res.Channel.Description)
	}
	if res.Channel.Source.Type != "mllp" {
		t.Errorf("source type: got %q, want %q", res.Channel.Source.Type, "mllp")
	}
	if res.Channel.Source.Config["host"] != "0.0.0.0" {
		t.Errorf("source host: got %v", res.Channel.Source.Config["host"])
	}
	if res.Channel.Source.Config["port"] != 6661 {
		t.Errorf("source port: got %v", res.Channel.Source.Config["port"])
	}
	if res.Channel.Destination.Type != "http" {
		t.Errorf("dest type: got %q, want %q", res.Channel.Destination.Type, "http")
	}
	if res.Channel.Destination.Config["method"] != "POST" {
		t.Errorf("dest method: got %v", res.Channel.Destination.Config["method"])
	}
	if res.OriginalSourceType != "com.mirth.connect.connectors.tcp.TcpListenerProperties" {
		t.Errorf("original source type: got %q", res.OriginalSourceType)
	}
	if len(res.OriginalDestTypes) != 1 {
		t.Fatalf("expected 1 original dest type, got %d", len(res.OriginalDestTypes))
	}
}

func TestConvertChannel_HTTPToHTTP(t *testing.T) {
	mch, err := mirthxml.ParseChannel([]byte(syntheticHTTPChannel))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Source.Type != "http" {
		t.Errorf("source type: got %q, want %q", res.Channel.Source.Type, "http")
	}
	if res.Channel.Source.Config["host"] != "0.0.0.0" {
		t.Errorf("source host: got %v", res.Channel.Source.Config["host"])
	}
	if res.Channel.Source.Config["port"] != 8080 {
		t.Errorf("source port: got %v", res.Channel.Source.Config["port"])
	}
	if res.Channel.Destination.Type != "file" {
		t.Errorf("dest type: got %q, want %q", res.Channel.Destination.Type, "file")
	}
	if res.Channel.Destination.Config["path"] != "/tmp/ghega/out" {
		t.Errorf("dest path: got %v", res.Channel.Destination.Config["path"])
	}
}

func TestConvertChannel_FileToFile(t *testing.T) {
	mch, err := mirthxml.ParseChannel([]byte(syntheticFileChannel))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Source.Type != "file" {
		t.Errorf("source type: got %q, want %q", res.Channel.Source.Type, "file")
	}
	if res.Channel.Source.Config["path"] != "/tmp/ghega/in" {
		t.Errorf("source path: got %v", res.Channel.Source.Config["path"])
	}
	if res.Channel.Destination.Type != "http" {
		t.Errorf("dest type: got %q, want %q", res.Channel.Destination.Type, "http")
	}
	wantURL := "https://api.example.com/"
	if res.Channel.Destination.Config["url"] != wantURL {
		t.Errorf("dest url: got %v, want %v", res.Channel.Destination.Config["url"], wantURL)
	}
	if res.Channel.Destination.Config["method"] != "PUT" {
		t.Errorf("dest method: got %v", res.Channel.Destination.Config["method"])
	}
}

func TestConvertChannel_UnsupportedSource(t *testing.T) {
	mch, err := mirthxml.ParseChannel([]byte(syntheticUnsupportedChannel))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Source.Type != "" {
		t.Errorf("expected empty source type for unsupported connector, got %q", res.Channel.Source.Type)
	}
	if res.Channel.Destination.Type != "http" {
		t.Errorf("dest type: got %q, want %q", res.Channel.Destination.Type, "http")
	}

	found := false
	for _, w := range res.Warnings {
		if w == `unsupported source connector type "com.mirth.connect.connectors.ws.WebServiceListenerProperties"` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about unsupported source type, got: %v", res.Warnings)
	}
}

func TestConvertChannel_MultipleDestinations(t *testing.T) {
	mch, err := mirthxml.ParseChannel([]byte(syntheticMultiDestChannel))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Destination.Type != "http" {
		t.Errorf("dest type: got %q, want %q", res.Channel.Destination.Type, "http")
	}

	found := false
	for _, w := range res.Warnings {
		if w == `Mirth channel has 2 destinations; only the primary ("HTTP Primary") was converted` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about multiple destinations, got: %v", res.Warnings)
	}
}

func TestConvertChannel_ScriptsPresent(t *testing.T) {
	mch, err := mirthxml.ParseChannel([]byte(syntheticScriptChannel))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	found := false
	for _, w := range res.Warnings {
		if w == "JavaScript transformers/filters are present but not converted in this step" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected warning about scripts, got: %v", res.Warnings)
	}
}

func TestConvertChannel_EmptyScriptsNoWarning(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-empty</id>
  <name>Empty_Scripts</name>
  <description>Steps and rules with empty scripts</description>
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
          <name>Empty Step</name>
          <script>   </script>
          <type>JavaScript</type>
        </step>
      </steps>
    </transformer>
    <filter>
      <rules>
        <rule>
          <sequenceNumber>0</sequenceNumber>
          <name>Empty Rule</name>
          <script></script>
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
        <steps>
          <step>
            <sequenceNumber>0</sequenceNumber>
            <name>Empty Dest Step</name>
            <script>   </script>
            <type>JavaScript</type>
          </step>
        </steps>
      </transformer>
      <filter>
        <rules>
          <rule>
            <sequenceNumber>0</sequenceNumber>
            <name>Empty Dest Rule</name>
            <script>   </script>
            <operator>AND</operator>
          </rule>
        </rules>
      </filter>
    </connector>
  </destinationConnectors>
  <preprocessingScript>   </preprocessingScript>
  <postprocessingScript></postprocessingScript>
  <deployScript></deployScript>
  <undeployScript>   </undeployScript>
  <properties/>
</channel>`

	mch, err := mirthxml.ParseChannel([]byte(xml))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	for _, w := range res.Warnings {
		if w == "JavaScript transformers/filters are present but not converted in this step" {
			t.Errorf("unexpected script warning for empty/whitespace-only scripts")
		}
	}
}

func TestConvertChannel_DBConnectors(t *testing.T) {
	mch, err := mirthxml.ParseChannel([]byte(syntheticDBChannel))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Source.Type != "db" {
		t.Errorf("source type: got %q, want %q", res.Channel.Source.Type, "db")
	}
	if res.Channel.Destination.Type != "db" {
		t.Errorf("dest type: got %q, want %q", res.Channel.Destination.Type, "db")
	}
	if res.Channel.Source.Config["driver"] != "org.postgresql.Driver" {
		t.Errorf("source driver: got %v", res.Channel.Source.Config["driver"])
	}
	if res.Channel.Source.Config["url"] != "jdbc:postgresql://localhost/db" {
		t.Errorf("source url: got %v", res.Channel.Source.Config["url"])
	}
	if res.Channel.Source.Config["query"] != "SELECT * FROM patients" {
		t.Errorf("source query: got %v", res.Channel.Source.Config["query"])
	}
	if res.Channel.Destination.Config["driver"] != "org.postgresql.Driver" {
		t.Errorf("dest driver: got %v", res.Channel.Destination.Config["driver"])
	}
	if res.Channel.Destination.Config["url"] != "jdbc:postgresql://localhost/db" {
		t.Errorf("dest url: got %v", res.Channel.Destination.Config["url"])
	}

	for _, w := range res.Warnings {
		if strings.Contains(w, "configuration is not fully extracted") {
			t.Errorf("unexpected generic config warning: %s", w)
		}
	}
}

// syntheticDBChannelWithPassword has database connectors with passwords.
const syntheticDBChannelWithPassword = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-007p</id>
  <name>DB_Password_Channel</name>
  <description>Database channel with passwords</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>DB Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.jdbc.DatabaseReaderProperties">
      <driver>org.postgresql.Driver</driver>
      <url>jdbc:postgresql://localhost/db</url>
      <username>admin</username>
      <password>secret123</password>
      <query>SELECT * FROM patients</query>
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
      <name>DB Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.jdbc.DatabaseWriterProperties">
        <driver>org.postgresql.Driver</driver>
        <url>jdbc:postgresql://localhost/db</url>
        <username>admin</username>
        <password>secret456</password>
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

// syntheticSFTPChannel has SFTP connectors.
const syntheticSFTPChannel = `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-011</id>
  <name>SFTP_Channel</name>
  <description>SFTP channel</description>
  <enabled>true</enabled>
  <revision>1</revision>
  <sourceConnector>
    <name>SFTP Source</name>
    <enabled>true</enabled>
    <properties class="com.mirth.connect.connectors.sftp.SftpReceiverProperties">
      <host>sftp.example.com</host>
      <port>22</port>
      <username>sftpuser</username>
      <path>/inbound</path>
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
      <name>SFTP Destination</name>
      <enabled>true</enabled>
      <properties class="com.mirth.connect.connectors.sftp.SftpDispatcherProperties">
        <host>sftp.example.com</host>
        <port>22</port>
        <username>sftpuser</username>
        <path>/outbound</path>
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

func TestConvertChannel_DBConnectorsWithPassword(t *testing.T) {
	mch, err := mirthxml.ParseChannel([]byte(syntheticDBChannelWithPassword))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Source.Type != "db" {
		t.Errorf("source type: got %q, want %q", res.Channel.Source.Type, "db")
	}
	if res.Channel.Destination.Type != "db" {
		t.Errorf("dest type: got %q, want %q", res.Channel.Destination.Type, "db")
	}
	if res.Channel.Source.Config["username"] != "admin" {
		t.Errorf("source username: got %v", res.Channel.Source.Config["username"])
	}
	if res.Channel.Destination.Config["username"] != "admin" {
		t.Errorf("dest username: got %v", res.Channel.Destination.Config["username"])
	}

	srcPassWarn := false
	dstPassWarn := false
	for _, w := range res.Warnings {
		if strings.Contains(w, "Password field detected in connector config") {
			if strings.Contains(w, "secrets management") {
				if !srcPassWarn {
					srcPassWarn = true
				} else {
					dstPassWarn = true
				}
			}
		}
	}
	if !srcPassWarn {
		t.Errorf("expected source password warning, got: %v", res.Warnings)
	}
	if !dstPassWarn {
		t.Errorf("expected destination password warning, got: %v", res.Warnings)
	}
}

func TestConvertChannel_SFTPConnectors(t *testing.T) {
	mch, err := mirthxml.ParseChannel([]byte(syntheticSFTPChannel))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Source.Type != "sftp" {
		t.Errorf("source type: got %q, want %q", res.Channel.Source.Type, "sftp")
	}
	if res.Channel.Destination.Type != "sftp" {
		t.Errorf("dest type: got %q, want %q", res.Channel.Destination.Type, "sftp")
	}
	if res.Channel.Source.Config["host"] != "sftp.example.com" {
		t.Errorf("source host: got %v", res.Channel.Source.Config["host"])
	}
	if res.Channel.Source.Config["port"] != 22 {
		t.Errorf("source port: got %v", res.Channel.Source.Config["port"])
	}
	if res.Channel.Source.Config["path"] != "/inbound" {
		t.Errorf("source path: got %v", res.Channel.Source.Config["path"])
	}
	if res.Channel.Source.Config["username"] != "sftpuser" {
		t.Errorf("source username: got %v", res.Channel.Source.Config["username"])
	}
	if res.Channel.Destination.Config["host"] != "sftp.example.com" {
		t.Errorf("dest host: got %v", res.Channel.Destination.Config["host"])
	}
	if res.Channel.Destination.Config["port"] != 22 {
		t.Errorf("dest port: got %v", res.Channel.Destination.Config["port"])
	}
	if res.Channel.Destination.Config["path"] != "/outbound" {
		t.Errorf("dest path: got %v", res.Channel.Destination.Config["path"])
	}
	if res.Channel.Destination.Config["username"] != "sftpuser" {
		t.Errorf("dest username: got %v", res.Channel.Destination.Config["username"])
	}

	for _, w := range res.Warnings {
		if strings.Contains(w, "configuration is not fully extracted") {
			t.Errorf("unexpected generic config warning: %s", w)
		}
	}
}

func TestConvertChannel_NameSanitization(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-008</id>
  <name>ADT_A01  Feed!!!</name>
  <description>Bad name</description>
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

	mch, err := mirthxml.ParseChannel([]byte(xml))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	want := "adt-a01-feed"
	if res.Channel.Name != want {
		t.Errorf("name: got %q, want %q", res.Channel.Name, want)
	}
	found := false
	for _, w := range res.Warnings {
		if w == `channel name sanitized from "ADT_A01  Feed!!!" to "adt-a01-feed"` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected name sanitization warning, got: %v", res.Warnings)
	}
}

func TestConvertChannel_EmptyDestinations(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-009</id>
  <name>No_Dest</name>
  <description>No destinations</description>
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
  <destinationConnectors/>
  <properties/>
</channel>`

	mch, err := mirthxml.ParseChannel([]byte(xml))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Destination.Type != "" {
		t.Errorf("expected empty destination type, got %q", res.Channel.Destination.Type)
	}
	found := false
	for _, w := range res.Warnings {
		if w == "no destination connectors found" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected no-destination warning, got: %v", res.Warnings)
	}
}

func TestConvertChannel_DefaultMethod(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<channel version="3.12.0">
  <id>ch-010</id>
  <name>Default_Method</name>
  <description>Missing method</description>
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
        <method/>
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

	mch, err := mirthxml.ParseChannel([]byte(xml))
	if err != nil {
		t.Fatalf("parse channel: %v", err)
	}

	res, err := ConvertChannel(mch)
	if err != nil {
		t.Fatalf("convert channel: %v", err)
	}

	if res.Channel.Destination.Config["method"] != "POST" {
		t.Errorf("method: got %v, want POST", res.Channel.Destination.Config["method"])
	}
	found := false
	for _, w := range res.Warnings {
		if w == "HTTP destination method defaulted to POST" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected default method warning, got: %v", res.Warnings)
	}
}
