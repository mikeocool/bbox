package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
	"unsafe"

	"github.com/parquet-go/parquet-go"
)

// Record represents our parquet schema
type Record struct {
	ID     string `parquet:"id"`
	Name   string `parquet:"name"`
	City   string `parquet:"city"`
	Region string `parquet:"region"`
	Geom   []byte `parquet:"geom"`
}

func main() {
	var (
		outputFile = flag.String("output", "test_data.parquet", "Output parquet file path")
		count      = flag.Int("count", 1000, "Number of records to generate")
		seed       = flag.Int64("seed", 0, "Random seed (0 for current time)")
	)
	flag.Parse()

	// Set random seed
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rand.Seed(*seed)

	fmt.Printf("Generating %d records to %s (seed: %d)\n", *count, *outputFile, *seed)

	// Create output file
	file, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	// Create parquet writer
	writer := parquet.NewGenericWriter[Record](file)
	defer writer.Close()

	// Generate and write records
	records := make([]Record, *count)
	for i := range *count {
		records[i] = generateRecord(i)
	}

	// Write all records in batches
	batchSize := 1000
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		_, err := writer.Write(records[i:end])
		if err != nil {
			log.Fatalf("Failed to write records: %v", err)
		}
	}

	fmt.Printf("Successfully wrote %d records to %s\n", *count, *outputFile)
}

func generateRecord(index int) Record {
	// Generate random coordinates (roughly within continental US bounds)
	lon := -125.0 + rand.Float64()*50.0 // -125 to -75 longitude
	lat := 25.0 + rand.Float64()*25.0   // 25 to 50 latitude

	return Record{
		ID:     fmt.Sprintf("ID_%06d", index),
		Name:   generateRandomName(),
		City:   generateRandomCity(),
		Region: generateRandomRegion(),
		Geom:   encodePointAsWKB(lon, lat),
	}
}

func generateRandomName() string {
	firstNames := []string{"Alice", "Bob", "Charlie", "Diana", "Edward", "Fiona", "George", "Helen", "Ivan", "Julia"}
	lastNames := []string{"Anderson", "Brown", "Clark", "Davis", "Evans", "Fisher", "Garcia", "Harris", "Johnson", "King"}

	return firstNames[rand.Intn(len(firstNames))] + " " + lastNames[rand.Intn(len(lastNames))]
}

func generateRandomCity() string {
	cities := []string{
		"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia",
		"San Antonio", "San Diego", "Dallas", "San Jose", "Austin", "Jacksonville",
		"Fort Worth", "Columbus", "Charlotte", "San Francisco", "Indianapolis", "Seattle",
		"Denver", "Washington", "Boston", "El Paso", "Nashville", "Detroit", "Oklahoma City",
	}
	return cities[rand.Intn(len(cities))]
}

func generateRandomRegion() string {
	regions := []string{
		"Northeast", "Southeast", "Midwest", "Southwest", "West", "Pacific",
		"Mountain", "Central", "Atlantic", "Great Lakes", "Gulf Coast", "Plains",
	}
	return regions[rand.Intn(len(regions))]
}

// encodePointAsWKB creates a WKB (Well-Known Binary) representation of a point
func encodePointAsWKB(x, y float64) []byte {
	// WKB Point structure:
	// - Byte order (1 byte): 1 = little endian
	// - Geometry type (4 bytes): 1 = Point
	// - X coordinate (8 bytes, double)
	// - Y coordinate (8 bytes, double)

	wkb := make([]byte, 21) // 1 + 4 + 8 + 8 = 21 bytes

	// Byte order (little endian)
	wkb[0] = 1

	// Geometry type (Point = 1)
	binary.LittleEndian.PutUint32(wkb[1:5], 1)

	// X coordinate
	binary.LittleEndian.PutUint64(wkb[5:13], uint64(binary.LittleEndian.Uint64((*[8]byte)(unsafe.Pointer(&x))[:8])))

	// Y coordinate
	binary.LittleEndian.PutUint64(wkb[13:21], uint64(binary.LittleEndian.Uint64((*[8]byte)(unsafe.Pointer(&y))[:8])))

	return wkb
}
