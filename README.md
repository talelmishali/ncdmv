## ncdmv

Monitor NCDMV for new appointment slots and get notified through Discord.

<img src="https://i.imgur.com/pW9Vxio.png" alt="Discord message example" width="75%"/>

### Usage

```
$ go run ./cmd/server -h
Usage of ncdmv:
  -appt_type string
        appointment type (options: permit,driver-license,driver-license-duplicate,driver-license-renewal,id-card,knowledge-test,motorcycle-skills-test) (default "permit")
  -database_path string
        path to database file (default "./ncdmv.db")
  -debug
        enable debug mode
  -disable_gpu
        disable GPU
  -discord_webhook string
        Discord webhook URL for notifications (optional)
  -headless
        enable headless browser (default true)
  -interval int
        interval between checks, in minutes (default 30)
  -locations string
        comma-seperated list of locations to check (options: durham-south,hillsborough,raleigh-east,raleigh-north,ahoskie,durham-east,fuquay-varina,garner,raleigh-west,cary) (default "cary,durham-east,durham-south")
  -migrations_path string
        path to migrations directory
  -notify_unavailable
        if true, notify when a previously available appointment becomes unavailable (default true)
  -stop_on_failure
        if true, stop completely on a failure instead of just logging
  -timeout int
        timeout, in seconds (default 120)
```

### Examples

Run in headless mode:

```
go run ./cmd/server -locations=cary,durham-east,durham-south -discord_webhook=[WEBHOOK]
```

Show the browser with a timeout of 2 minutes each check (across all locations) and an interval of 10 minutes:

```
go run ./cmd/server -headless=false -locations=cary,durham-east,durham-south -discord_webhook=[WEBHOOK] -timeout=120 -interval=10
```

Run on Docker:

```
docker run --rm ghcr.io/aksiksi/ncdmv:latest ncdmv -h
```

### Setup (Debian)

1. Install Google Chrome: https://www.if-not-true-then-false.com/2021/install-google-chrome-on-debian/. Quick summary:

```
sudo su -

cat  /etc/apt/sources.list.d/google-chrome.list
deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main
EOF

wget -O- https://dl.google.com/linux/linux_signing_key.pub |gpg --dearmor > /etc/apt/trusted.gpg.d/google.gpg

apt update

apt install google-chrome-stable
```

2. Run!
