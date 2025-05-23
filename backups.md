---

# 🛡️ Automating SQLite Backups to Google Drive on Linux (Orange Pi)

This guide walks you through how to:

- Automatically dump a SQLite database (e.g., from a Docker volume)
- Upload the dump to Google Drive using `rclone`
- Schedule the process with `cron`

---

## 📦 Prerequisites

- A Linux-based system (e.g., Orange Pi running Ubuntu)
- Docker volume containing your SQLite DB
- [rclone](https://rclone.org/) configured for Google Drive  
  See [rclone Google Drive setup](https://rclone.org/drive/) if not done yet.

---

## 🔐 Allow `sqlite3` to Run with `sudo` (No Password)

You’ll need `sudo` privileges to access the SQLite file inside a Docker volume. To allow passwordless `sudo` for `sqlite3`:

```bash
sudo visudo
```

Add the following line at the end:

```text
orangepi ALL=(ALL) NOPASSWD: /usr/bin/sqlite3
```

---

## 📝 Create the Backup Script

Create a script at `/home/orangepi/sqlite_backup.sh`:

```bash
#!/bin/bash

# Set paths
DB_PATH="/var/lib/docker/volumes/calendar-app_calendar-data/_data/app.db"
BACKUP_DIR="/home/orangepi/backups"
BACKUP_NAME="backup__$(date +%F_%H-%M-%S).db.txt"

# Dump the SQLite DB
sudo sqlite3 "$DB_PATH" .dump > "$BACKUP_DIR/$BACKUP_NAME"

# Upload to Google Drive
rclone copy "$BACKUP_DIR/$BACKUP_NAME" gdrive:sqlite-backups
```

Make it executable:

```bash
chmod +x /home/orangepi/sqlite_backup.sh
```

---

## 🕑 Schedule Daily Cron Backup

Open your crontab:

```bash
crontab -e
```

Add the following line to run the backup every day at 2:00 AM:

```cron
0 2 * * * /home/orangepi/sqlite_backup.sh >> /home/orangepi/backup.log 2>&1
```

---

## 🔁 Restoring a Dumped Database

To reconstruct the SQLite database from the dump:

```bash
cat backup__2025-05-01_02-00-00.db.txt | sqlite3 my_reconstructed_database.db
```

---

## 🗜️ Optional: Compress the Dump

SQLite dumps compress very well. You can gzip them like this:

```bash
sqlite3 explorer.db .dump | gzip -c > explorer.db.txt.gz
```

---