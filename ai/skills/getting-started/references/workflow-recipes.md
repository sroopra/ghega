# Workflow Recipes

Step-by-step recipes for common Ghega tasks. Each recipe is self-contained.

---

## Recipe 1: Build and Run Ghega from Source

**Goal:** Get a running Ghega instance from a fresh clone.

```bash
# 1. Install dependencies
go mod download
cd ui && npm install && cd ..

# 2. Build UI and binary
make ui-build
make build

# 3. Start the server
./ghega serve

# 4. Verify
curl http://localhost:8080/healthz
# → {"status":"ok"}
```

**Result:** HTTP API on `:8080`, MLLP on `:2575`, UI at `http://localhost:8080`.

---

## Recipe 2: Create and Deploy an MLLP-to-HTTP Channel

**Goal:** Scaffold a channel, validate, test, and deploy it.

```bash
# 1. Generate the channel scaffold
./ghega generate channel mllp-to-http \
    --name adt-inbound \
    --message-type ADT_A01 \
    --out ./channels/adt-inbound

# 2. Validate the definition
./ghega channel validate ./channels/adt-inbound/channel.yaml

# 3. Run test fixtures
./ghega channel test ./channels/adt-inbound/channel.yaml

# 4. Deploy the revision
./ghega channel deploy ./channels/adt-inbound/channel.yaml
```

**Generated files:**
- `channel.yaml` — channel definition with source, destination, mappings
- `tests/fixture.yaml` — test fixtures with input/expected pairs
- `fixtures/sample.hl7` — sample HL7v2 message
- `fixtures/minimal.hl7` — minimal HL7v2 message

**Next steps:** Edit `channel.yaml` to add/modify mappings, then re-validate
and re-test. Use the `creating-mllp-channels` skill for mapping guidance.

---

## Recipe 3: Create and Deploy an HL7v2-to-FHIR Channel

**Goal:** Scaffold a channel that transforms HL7v2 to FHIR bundles.

```bash
# 1. Generate the FHIR channel scaffold
./ghega generate channel hl7v2-to-fhir \
    --name fhir-adt \
    --message-type ADT_A01 \
    --out ./channels/fhir-adt

# 2. Validate, test, deploy
./ghega channel validate ./channels/fhir-adt/channel.yaml
./ghega channel test ./channels/fhir-adt/channel.yaml
./ghega channel deploy ./channels/fhir-adt/channel.yaml
```

Use the `mapping-hl7v2-to-fhir` and `creating-fhir-channels` skills for
advanced FHIR mapping work.

---

## Recipe 4: Migrate a Mirth Connect Export

**Goal:** Convert Mirth XML exports into Ghega channel artifacts.

```bash
# 1. Run the migration
./ghega migrate mirth ./mirth-export/ --out ./migrated

# 2. Review the generated reports
ls ./migrated/
# → channel directories, migration reports, rewrite tasks

# 3. For each generated channel, validate and test
./ghega channel validate ./migrated/my-channel/channel.yaml
./ghega channel test ./migrated/my-channel/channel.yaml

# 4. Generate rewrite tasks for unconverted scripts
./ghega generate migration-task \
    --channel my-channel \
    --out ./migrated/my-channel \
    --category script-rewrite \
    --severity high \
    --description "Convert E4X date formatting to CEL"
```

Use the `migrating-from-mirth` and `writing-typed-rewrite-tasks` skills for
detailed migration guidance.

---

## Recipe 5: Continuous Validation with Watch Mode

**Goal:** Automatically validate and test channels as you edit them.

```bash
# 1. Start the watch loop (runs in foreground)
./ghega watch ./channels/

# 2. Edit channel files in another terminal
# Watch re-runs validation + tests on every change to channel.yaml files
# Press Ctrl+C to stop
```

**Note:** Watch is a validate/test loop only. It does not deploy channels or
affect the running server.

---

## Recipe 6: Channel Diff and Rollback

**Goal:** Compare changes and roll back if needed.

```bash
# 1. Make changes to a deployed channel
# (edit channel.yaml)

# 2. See what changed vs the deployed version
./ghega channel diff ./channels/adt-inbound/channel.yaml

# 3. If the changes are good, deploy the new revision
./ghega channel deploy ./channels/adt-inbound/channel.yaml

# 4. If something went wrong, rollback
./ghega channel rollback adt-inbound --to <previous-hash>
```

---

## Recipe 7: Send a Test HL7v2 Message via MLLP

**Goal:** Send an HL7v2 message to the running MLLP listener and verify.

```bash
# 1. Ensure ghega serve is running
# 2. Send a message (MLLP framing: 0x0b = start, 0x1c 0x0d = end)
printf '\x0bMSH|^~\\&|TestApp|TestFac|GhegaApp|GhegaFac|20240101120000||ADT^A01|MSG001|P|2.5\rPID|1||MRN12345^^^Hosp||TESTA^SYNTHETICA\r\x1c\r' \
    | nc localhost 2575

# 3. Check the message in the API
curl http://localhost:8080/api/v1/messages
```

**Note:** The ACK is always returned. For delivery to succeed, set
`GHEGA_DESTINATION_URL` to a running HTTP endpoint. Without a destination
receiver, the message metadata will show a `failed` delivery status.

---

## Recipe 8: Run the UI in Development Mode

**Goal:** Run the React UI with hot reload for frontend development.

```bash
# Terminal 1: Start the backend
./ghega serve

# Terminal 2: Start the UI dev server
cd ui
npm install
npm run dev
# → Vite server on http://localhost:3000
# → API requests proxied to http://localhost:8080
```

The UI has pages for: Home, Channels, Messages, Alerts, Operations,
Migrations, Settings, and Login. Note that Channels, Settings, and Operations
pages are currently placeholders or return empty data.

---

## Recipe 9: Run All Checks (Build + Test + Lint + Validate Skills)

**Goal:** Full validation before committing changes.

```bash
make build && make test && make lint && make validate-skills
```

All four should pass cleanly on a healthy codebase.
