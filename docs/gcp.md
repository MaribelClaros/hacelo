# Self-hosting on a Google Cloud e2-micro VM

Google Cloud's Always Free tier includes one e2-micro VM (1 GB RAM, 30 GB
standard disk) in `us-west1`, `us-central1` or `us-east1`. Limits change, so
check https://cloud.google.com/free before starting.

## Cost warnings

- Use a **Standard persistent disk** (not Balanced or SSD): only standard is free.
- Google may charge for the VM's **external IPv4 address** (a few USD per
  month) even on the free tier. Check the price estimate when creating the VM
  and set a **budget alert** (Billing -> Budgets) of 1 USD so you notice.
- Outbound traffic beyond 1 GB per month is billed. This app uses very little.

## 1. Create the VM

1. Create a project and enable billing (a card is required).
2. Compute Engine -> Create instance:
   - Region `us-central1`, machine type **e2-micro**.
   - Boot disk: Debian 12, **Standard persistent disk**, 30 GB.
   - Firewall: allow **HTTP** and **HTTPS** traffic.
3. Note the external IP.

## 2. Get a free domain

Caddy needs a domain name to issue the HTTPS certificate. Create a free
subdomain at https://www.duckdns.org and point it to the VM's external IP.

## 3. Prepare the VM (SSH from the console)

```bash
# 2 GB swap: building the image needs more than the 1 GB of RAM
sudo fallocate -l 2G /swapfile && sudo chmod 600 /swapfile
sudo mkswap /swapfile && sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# Docker and git
curl -fsSL https://get.docker.com | sudo sh
sudo apt-get install -y git
sudo usermod -aG docker $USER && newgrp docker
```

## 4. Deploy

```bash
git clone https://github.com/<your-user>/hacelo.git
cd hacelo
cp .env.example .env
nano .env        # DOMAIN, API_TOKEN, VAPID keys, VAPID_SUBSCRIBER
docker compose up -d --build
```

The first build takes several minutes. Then open `https://<DOMAIN>`.

## 5. Install on the phone

Open the URL in Chrome -> **Add to Home screen** -> open **Ajustes**, paste the
`API_TOKEN` and tap **Activar notificaciones**.

## Updating

```bash
cd hacelo && git pull && docker compose up -d --build
```

## Backups

The database is the `hacelo_data` Docker volume. To copy it out:

```bash
docker compose cp hacelo:/data/hacelo.db ./hacelo-backup.db
```
