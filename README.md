# Offline Species Explorer

A cross-platform desktop app for browsing taxonomic hierarchies of species. Works fully offline with a local SQLite database. Supports syncing with the Nimbus API to fetch or update species data.

## Features

- **Offline-first**: All data stored in embedded SQLite (pure Go, no CGO)
- **Search**: Case-insensitive partial match by common or scientific name
- **Taxonomic tree**: Indented hierarchy (Kingdom → Phylum → Class → Order → Family → Species)
- **Sync**: Update species data from the live Nimbus API
- **Fetch**: Add new species from the API and save them locally
- **100+ pre-loaded species** with realistic taxonomy (animals, plants, fungi)

## Prerequisites

- Go 1.21+
- [Wails v2 CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Platform-specific dependencies (see [Wails docs](https://wails.io/docs/gettingstarted/installation#platform-specific-dependencies))

## Build

```bash
# Install dependencies
go mod tidy

# Development (live reload)
wails dev

# Production build (optimized, -ldflags "-s -w" to reduce binary size)
wails build -ldflags "-s -w"

# The binary will be in build/bin/
```

## Project Structure

```
offline-species-explorer/
├── frontend/
│   ├── index.html      # Main HTML layout
│   ├── main.js         # Vanilla JS frontend logic
│   └── styles.css      # Light/dark CSS with responsive layout
├── app.go              # Wails-bound App struct: Search, Sync, Fetch
├── db.go               # SQLite schema, CRUD operations, connection
├── species_seed.go     # Hardcoded seed data (100+ species)
├── main.go             # Wails entry point, asset embedding
├── go.mod / go.sum     # Go module
├── wails.json          # Wails project config
└── README.md
```

## Usage

1. **Search**: Type a common or scientific name in the search bar and press Enter.
2. **Select**: Click a result to view its full taxonomic hierarchy.
3. **Sync**: Click "Sync from Nimbus API" to fetch updated data from the API.
4. **Fetch new**: If no local results are found, enter a name and click "Fetch from API".

## API

The Nimbus API endpoint used:

```
GET https://nimbus.space/api/v1/bio/species?name=<urlencoded_name>
```

Returns JSON:
```json
{
  "common_name": "Blue Whale",
  "scientific_name": "Balaenoptera musculus",
  "kingdom": "Animalia",
  "phylum": "Chordata",
  "class": "Mammalia",
  "order": "Artiodactyla",
  "family": "Balaenopteridae"
}
```

## Scaling to 5000+ Species

Replace the hardcoded seed list in `species_seed.go` with CSV loading:

```go
func SeedDatabase(db *sql.DB) error {
    f, _ := os.Open("species.csv")
    defer f.Close()
    r := csv.NewReader(f)
    // ... INSERT each row in a transaction
}
```

## License

MIT
