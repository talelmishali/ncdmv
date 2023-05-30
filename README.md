## ncdmv

### Examples

Run in headless mode:

```
$ go run . -locations=cary,durham-east,durham-south,fuquay-varina,garner,hillsborough,raleigh-east,raleigh-north,raleigh-west -discord_webhook=[WEBHOOK]
```

Open the browser and set a timeout of 2 minutes for a single check (across all locations):

```
$ go run . -headless=false -timeout=120 -locations=cary,durham-east,durham-south,fuquay-varina,garner,hillsborough,raleigh-east,raleigh-north,raleigh-west -discord_webhook=[WEBHOOK]
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
