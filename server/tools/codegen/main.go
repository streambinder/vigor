package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var (
		modelDir   = flag.String("models", "", "Path to Go models directory (required)")
		outputDir  = flag.String("output", "", "Path to Dart models output directory (required)")
		modulePath = flag.String("module", "", "Go module path for imports (e.g., github.com/dpucci/vigor/server)")
	)
	flag.Parse()

	if *modelDir == "" || *outputDir == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -models <go-models-dir> -output <dart-output-dir> [-module <go-module-path>]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Resolve absolute paths
	absModelDir, err := filepath.Abs(*modelDir)
	if err != nil {
		log.Fatalf("Failed to resolve model directory: %v", err)
	}

	absOutputDir, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("Failed to resolve output directory: %v", err)
	}

	log.Printf("Generating Dart models from Go structs...")
	log.Printf("  Source: %s", absModelDir)
	log.Printf("  Output: %s", absOutputDir)

	// Parse Go structs
	parser := NewParser(absModelDir, *modulePath)
	structs, err := parser.ParsePackage()
	if err != nil {
		log.Fatalf("Failed to parse Go structs: %v", err)
	}

	log.Printf("Found %d structs to generate", len(structs))

	// Generate Dart files
	generator := NewGenerator(absOutputDir)
	for _, s := range structs {
		if err := generator.Generate(s); err != nil {
			log.Fatalf("Failed to generate Dart for %s: %v", s.Name, err)
		}
		log.Printf("  ✓ Generated %s.dart", toSnakeCase(s.Name))
	}

	log.Printf("Done! Run 'dart run build_runner build' in your Flutter project.")
}
