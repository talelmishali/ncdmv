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
$ go run . -locations=cary,durham-east,durham-south,fuquay-varina,garner,hillsborough,raleigh-east,raleigh-north,raleigh-west -discord_webhook=[WEBHOOK]
```

Open the browser and set a timeout of 2 minutes for a single check (across all locations):

```
$ go run . -headless=false -timeout=120 -locations=cary,durham-east,durham-south,fuquay-varina,garner,hillsborough,raleigh-east,raleigh-north,raleigh-west -discord_webhook=[WEBHOOK]
```

Run using Docker:

```
$ docker run --rm ghcr.io/aksiksi/ncdmv:latest ncdmv [...]
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
