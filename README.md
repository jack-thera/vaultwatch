# vaultwatch

A CLI tool that monitors HashiCorp Vault secret expiration and sends alerts before leases expire.

---

## Installation

```bash
go install github.com/yourusername/vaultwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultwatch.git
cd vaultwatch && go build -o vaultwatch .
```

---

## Usage

Set your Vault address and token, then run vaultwatch with a warning threshold:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.xxxxxxxxxxxxxxxx"

# Alert on secrets expiring within 48 hours
vaultwatch watch --threshold 48h --alert slack
```

**Common flags:**

| Flag | Description | Default |
|------|-------------|---------|
| `--threshold` | Warn when lease expires within this duration | `24h` |
| `--alert` | Alert method (`slack`, `email`, `stdout`) | `stdout` |
| `--interval` | How often to poll Vault | `5m` |
| `--path` | Vault secret path to monitor | `secret/` |

**Example — watch a specific path and post to Slack:**

```bash
vaultwatch watch \
  --path secret/production \
  --threshold 72h \
  --alert slack \
  --slack-webhook https://hooks.slack.com/services/xxx
```

---

## Configuration

vaultwatch can also be configured via a `vaultwatch.yaml` file in the working directory. See [docs/configuration.md](docs/configuration.md) for full options.

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

---

## License

[MIT](LICENSE)