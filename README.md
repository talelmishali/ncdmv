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
