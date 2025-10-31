#!/bin/bash

echo "Upload Notifications Test"
echo "==============================="
echo "This test program verifies the fixed upload notification implementation."
echo "It verifies that:"
echo "1. Progress dialogs appear correctly in the main thread"
echo "2. Success/error dialogs appear after upload completes"
echo "3. Threading synchronization works without freezing"
echo "4. Diagnostic logging is working"
echo ""
echo "Starting test application..."
echo ""

# Check if executable exists
if [ ! -f "upload_test" ]; then
    echo "Building test application..."
    go build -o upload_test main.go
    if [ $? -ne 0 ]; then
        echo "Failed to build test application"
        exit 1
    fi
fi

# Run the test
echo "Running upload notifications test..."
./upload_test

echo ""
echo "Test completed. Press any key to exit..."
read -n 1