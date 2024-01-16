#!/bin/bash
echo "Installing gothon..."
cd cmd/gothon || (echo "Must run script from project root!" && exit)
go clean
go mod tidy
go install .
cd ../..
echo "...gothon installed."
