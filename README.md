# sodoff

Dead simple CLI tool to tell you which bloody trains are delayed/cancelled, innit?

![img.png](img.png)

## CLI Usage

A token for the use of the National Rail SOAP API is required. You can get one [here.](https://realtime.nationalrail.co.uk/OpenLDBWSRegistration/).

```sh
# Check onced
train-cli check --from "StationA" --to "StationB"

# Continuously check every 2 seconds
train-cli check --from "StationA" --to "StationB" --continuous
```

> **Note**: The `--from` and `--to` flags are __NOT__ required. Recommended to use the in-built input search.
