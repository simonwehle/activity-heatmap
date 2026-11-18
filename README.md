# Activity Heatmap

Display activity routes on a map

## Prerequitsites

### Activities

`activity-heatmap` currently supports `.gpx` and `.fit` files to for display on the map.
Add your GPX files in the `/activities` folder.

The easiest way to get activities from Garmin is to spin up a [Endurain](https://endurain.com/) container and point `activity-heatmap` to the Endurain activities folder.

Alternatively you can manually connect your watch and copy `.fit` files from `/GARMIN/Activity` to your `/activities` folder. For Mac you can use OpenMTP

```sh
brew install --cask openmtp
```

Afterwards you need to convert the `.fit` files to `.gpx`

```bash
brew install gpsbabel
```

```bash
for f in *.fit; do gpsbabel -i garmin_fit -f "$f" -o gpx -F "${f%.fit}.gpx"; done
```

## Usage

### Locally

```
MAPLIBRE_STYLE="https://tiles.openfreemap.org/styles/liberty" go run .
```

You can view the map via http://localhost:3465

### Docker

Rename the `/docker/.env.example` to `.env` and set your API token.
Start the container with

```
docker compose up -d
```
