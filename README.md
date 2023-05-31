## ncdmv

### Usage

```
$ go run . -h
Usage of ncdmv:
  -appt_type string
        appointment type (one of: license,license-duplicate,license-renewal,permit) (default "permit")
  -debug
        enable debug mode
  -disable_gpu
        if true, disable GPU
  -discord_webhook string
        Discord webhook URL for notifications
  -headless
        enable headless browser (default true)
  -interval int
        interval between checks (minutes) (default 3)
  -locations string
        comma-seperated list of locations to check (valid options: ahoskie,durham-south,hillsborough,raleigh-east,raleigh-west,cary,durham-east,fuquay-varina,garner,raleigh-north) (default "cary,durham-east,durham-south")
  -stop_on_failure
        if true, stop completely on a failure instead of just logging
  -timeout int
        timeout (seconds) (default 60)
```

### Examples

Run in headless mode:

```
go run . -locations=cary,durham-east,durham-south -discord_webhook=[WEBHOOK]
```

Show the browser with a timeout of 2 minutes each check (across all locations) and an interval of 10 minutes:

```
go run . -headless=false -locations=cary,durham-east,durham-south -discord_webhook=[WEBHOOK] -timeout=120 -interval=10
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
