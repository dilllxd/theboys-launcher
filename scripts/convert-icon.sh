#!/bin/bash

# Convert Windows ICO to macOS ICNS format
# Usage: ./convert-icon.sh

set -e

ICON_DIR="resources/darwin"
ICONS_DIR="$ICON_DIR/TheBoysLauncher.iconset"
ICNS_FILE="$ICON_DIR/TheBoysLauncher.icns"
SOURCE_ICON="icon.ico"

echo "Converting Windows ICO to macOS ICNS format..."

# Create directories
mkdir -p "$ICONS_DIR"
mkdir -p "$ICON_DIR"

# Check if source icon exists
if [ ! -f "$SOURCE_ICON" ]; then
    echo "Error: Source icon not found: $SOURCE_ICON"
    exit 1
fi

# Check if sips is available (macOS built-in tool)
if ! command -v sips &> /dev/null; then
    echo "Warning: sips not found. Trying alternative methods..."

    # Try to use iconutil if available (macOS)
    if command -v iconutil &> /dev/null; then
        echo "Using iconutil for icon conversion..."

        # Create iconset directory structure
        for size in 16 32 64 128 256 512 1024; do
            mkdir -p "$ICONS_DIR/icon_${size}x${size}"
            mkdir -p "$ICONS_DIR/icon_${size}x${size}@2x"
        done

        # Try to convert using ImageMagick if available
        if command -v convert &> /dev/null; then
            echo "Using ImageMagick to generate icon sizes..."

            # Generate required icon sizes
            for size in 16 32 64 128 256 512 1024; do
                echo "Generating ${size}x${size} icon..."
                convert "$SOURCE_ICON" -resize ${size}x${size} "$ICONS_DIR/icon_${size}x${size}/icon.png"

                # Generate @2x versions
                double_size=$((size * 2))
                if [ $double_size -le 1024 ]; then
                    echo "Generating ${double_size}x${double_size} (@2x) icon..."
                    convert "$SOURCE_ICON" -resize ${double_size}x${double_size} "$ICONS_DIR/icon_${size}x${size}@2x/icon_@2x.png"
                fi
            done
        else
            echo "Error: Neither sips nor ImageMagick available for icon conversion"
            echo "Please install ImageMagick: brew install imagemagick"
            exit 1
        fi

        # Create ICNS from iconset
        echo "Creating ICNS file from iconset..."
        iconutil -c icns "$ICONS_DIR" -o "$ICNS_FILE"

    else
        echo "Error: Neither sips nor iconutil available for icon conversion"
        echo "Please install Xcode Command Line Tools"
        exit 1
    fi
else
    echo "Using sips for icon conversion..."

    # Create iconset directory structure
    for size in 16 32 64 128 256 512 1024; do
        mkdir -p "$ICONS_DIR/icon_${size}x${size}"
        mkdir -p "$ICONS_DIR/icon_${size}x${size}@2x"
    done

    # Generate required icon sizes using sips
    for size in 16 32 64 128 256 512 1024; do
        echo "Generating ${size}x${size} icon..."
        sips -z $size $size "$SOURCE_ICON" --out "$ICONS_DIR/icon_${size}x${size}/icon.png"

        # Generate @2x versions
        double_size=$((size * 2))
        if [ $double_size -le 1024 ]; then
            echo "Generating ${double_size}x${double_size} (@2x) icon..."
            sips -z $double_size $double_size "$SOURCE_ICON" --out "$ICONS_DIR/icon_${size}x${size}@2x/icon_@2x.png"
        fi
    done

    # Create ICNS from iconset
    echo "Creating ICNS file from iconset..."
    iconutil -c icns "$ICONS_DIR" -o "$ICNS_FILE"
fi

# Clean up iconset directory
rm -rf "$ICONS_DIR"

if [ -f "$ICNS_FILE" ]; then
    echo "✓ ICNS file created successfully!"
    echo "Location: $ICNS_FILE"
    echo "ICNS size: $(du -sh "$ICNS_FILE" | cut -f1)"
else
    echo "❌ Failed to create ICNS file"
    exit 1
fi

echo "Icon conversion completed!"