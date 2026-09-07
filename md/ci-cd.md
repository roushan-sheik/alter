# CI/CD — Go Backend

This document covers everything you need to understand, set up, and maintain the deployment pipeline for the Go backend.

---

## Architecture

```text
Developer (local)
      │
      │  git push / open PR
      ▼
GitHub Repository (feat/* or main branch)
      │
      │  GitHub Actions: .github/workflows/ci.yml
      ├─► [CI Job: build]
      │     go mod download → go vet → go build
      │     Runs on: PR to main + push to any non-main branch
      │     MUST PASS before merge is allowed (via Branch Protection)
      │
      │  GitHub Actions: .github/workflows/deploy.yml
      └─► [CD Job: deploy]  (only on push to main)
            │
            │  appleboy/ssh-action
            ▼
          VPS (Ubuntu)
            │
            ├─ Backup .env
            ├─ git pull origin main
            ├─ docker compose up -d --build --no-deps app
            └─ curl /health  ← must return 200 or deploy is flagged as failed (and rolled back)
            
          Nginx (reverse proxy)  →  Docker: gotickets_app → Go API
```

---

## How It Works

Every time code is pushed, two things can happen:

1. **CI runs** — The code is vetted (`go vet`) and built (`go build`). This runs on every branch and every PR to ensure the code is error-free and compiles correctly.
2. **CD runs** — If the push is to `main`, the server automatically pulls and redeploys the app. This only happens after CI passes (assuming Branch Protection rules are enabled).

```text
Your branch → PR → CI passes → Merge to main → CD auto-deploys → VPS live
```

The VPS runs Docker Compose. The deploy process is:
- SSH into the server
- Back up the `.env` file to `/tmp/.env.gotickets.backup`
- Pull the latest code via `git fetch` and `git reset --hard`
- Restore the `.env` file
- Rebuild the `app` Docker container using a zero-downtime strategy
- Hit the `/health` endpoint to confirm the app is up. If it fails to respond successfully within a minute, the server rolls back to the previous commit automatically.

---

## GitHub Secrets You Need to Add

Go to **GitHub → Repo → Settings → Secrets and variables → Actions**, then click **"New repository secret"** (in the Repository secrets section).

Add each of the 6 secrets below one by one.

| Secret | What it is |
|---|---|
| `VPS_HOST` | The IP or domain of your production server |
| `VPS_USER` | The SSH user on the server (e.g., `ubuntu`, `root`, or your username) |
| `VPS_SSH_KEY` | Your private SSH key (the full PEM content) |
| `VPS_PORT` | SSH port — usually `22` |
| `VPS_DEPLOY_PATH` | The exact folder where the project lives on the VPS (e.g., `/home/ubuntu/Zic/backend`) |
| `APP_PORT` | Port the app listens on (default in docker-compose is `5000`) |

---

### `VPS_HOST` — Your server's IP address

The public IP address of your live VPS. Find it in your cloud provider's dashboard.

```text
Example: 198.51.100.25
```

---

### `VPS_USER` — SSH username

The username you use to log into the server.

```text
Example: ubuntu
```

---

### `VPS_SSH_KEY` — Private SSH key

This allows GitHub Actions to log into your server automatically without a password.
Run this on your local machine to get your private key:

```bash
cat ~/.ssh/id_rsa
```

Copy the **entire output** — starting from `-----BEGIN OPENSSH PRIVATE KEY-----` all the way to the end — and paste it as the secret value.

---

### `VPS_PORT` — SSH port

The port used to connect to your server via SSH. Unless you changed it, this is:

```text
22
```

---

### `VPS_DEPLOY_PATH` — Project directory on the server

The exact folder where the project lives on the VPS:

```text
/home/ubuntu/Zic/backend
```

---

### `APP_PORT` — Application port

The port your Go app listens on. Check your `docker-compose.yml` or `.env` file — by default it is:

```text
5000
```

---

### 5. Protect the main branch

Go to **GitHub → Repo → Settings → Branches → Add rule** and:
- Target: `main`
- Enable: **Require status checks to pass before merging**
- Add check: `Build and Vet` (this is the CI job name)
- Enable: **Require a pull request before merging**

### 6. Set up the server (first time only)

SSH into the VPS and clone the repo:

```bash
git clone git@github.com:YOUR_ORG/YOUR_REPO.git /home/ubuntu/Zic/backend
cd /home/ubuntu/Zic/backend
```

Then create the `.env` file manually on the server. Do not copy it from your machine — type the values in directly:

```bash
nano .env
```

This file stays on the server permanently. `git pull` will never touch it.

---

## What Triggers a Deploy?

| Action | CI | Deploy |
|---|---|---|
| Push to any feature branch | Yes | No |
| Open a PR to `main` | Yes | No |
| Merge a PR into `main` | Yes | No |

---

## Managing Environment Variables

Your `.env` lives only on the VPS. It is never committed to git.

To update a value in production:

```bash
ssh ubuntu@YOUR_VPS_IP
cd /home/ubuntu/Zic/backend
nano .env
docker compose up -d --build app
```

That's all. The new value takes effect on the next container start.

---

## Common Problems

### `ssh: no key found` — Deploy fails with SSH authentication error

The `VPS_SSH_KEY` secret was either not copied correctly or the server does not recognise the key. 
1. **Incomplete or incorrect SSH key**: Make sure you copied the entire key including `-----BEGIN OPENSSH PRIVATE KEY-----` and `-----END OPENSSH PRIVATE KEY-----`.
2. **Public key is not on the server**: Make sure the corresponding public key is in `~/.ssh/authorized_keys` on your VPS.

### Port already in use

If another project is using the port, you'll get a conflict. Adjust the `PORT` in your `.env` to map to a different external port in your docker-compose.

### Container name conflict on deploy

If you see `The container name "/gotickets_app" is already in use`, a stopped container is still holding the name. Clean it up:

```bash
docker rm -f gotickets_app
docker compose up -d --build --remove-orphans app
```

---

## Security Reminders

- `.env` must be in `.gitignore` — check this before the very first commit on any machine
- Never paste a real secret into a PR comment, Slack message, or AI tool prompt
- The `VPS_SSH_KEY` in GitHub should be a deploy-only key, not your personal key
- If any credential was shared anywhere it shouldn't have been, rotate it immediately — don't wait
