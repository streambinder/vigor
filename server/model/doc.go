// Package model contains database models with GORM annotations.
//
// Code Generation:
// This package uses go:generate to automatically generate Dart models for the Flutter app.
// Run `go generate` from this directory to regenerate Dart models in app/lib/models/

//go:generate sh -c "cd ../../codegen && go run . -models ../server/model -output ../app/lib/models"

package model
