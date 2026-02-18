#!/bin/bash

# Integration test script for UI Automation backend
set -e

BASE_URL="http://localhost:8080"
COOKIES="cookies.txt"

echo "=== UI Automation Backend Integration Tests ==="
echo

# Cleanup
rm -f $COOKIES
rm -f test_image.png

# Create a test image file
echo "Creating test image..."
echo "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==" | base64 -d > test_image.png

echo "=== Scenario 1: Basic Project Flow ==="
echo

# 1. Register user
echo "1. Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -c $COOKIES \
  -d '{"email":"test@example.com","username":"testuser","password":"password123"}')
echo "Register response: $REGISTER_RESPONSE"
echo

# 2. Login
echo "2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c $COOKIES \
  -d '{"email":"test@example.com","password":"password123"}')
echo "Login response: $LOGIN_RESPONSE"
echo

# 2b. Validate session with /auth/me
echo "2b. Validating session with /auth/me..."
ME_RESPONSE=$(curl -s -X GET $BASE_URL/api/v1/auth/me -b $COOKIES)
echo "Me response: $ME_RESPONSE"
echo

# 3. Create project
echo "3. Creating project..."
PROJECT_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/projects \
  -H "Content-Type: application/json" \
  -b $COOKIES \
  -d '{"name":"Test Project","description":"Integration test project"}')
PROJECT_ID=$(echo $PROJECT_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "Project created with ID: $PROJECT_ID"
echo "Response: $PROJECT_RESPONSE"
echo

# 4. List projects
echo "4. Listing projects..."
LIST_PROJECTS=$(curl -s -X GET "$BASE_URL/api/v1/projects?limit=10" -b $COOKIES)
echo "Projects: $LIST_PROJECTS"
echo

# 5. Get project by ID
echo "5. Getting project by ID..."
GET_PROJECT=$(curl -s -X GET "$BASE_URL/api/v1/projects/$PROJECT_ID" -b $COOKIES)
echo "Project details: $GET_PROJECT"
echo

# 6. Update project
echo "6. Updating project..."
UPDATE_PROJECT=$(curl -s -X PUT "$BASE_URL/api/v1/projects/$PROJECT_ID" \
  -H "Content-Type: application/json" \
  -b $COOKIES \
  -d '{"name":"Updated Project Name"}')
echo "Update response: $UPDATE_PROJECT"
echo

echo "=== Scenario 2: Test Procedure with Versioning ==="
echo

# 1. Create test procedure
echo "1. Creating test procedure..."
PROCEDURE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/projects/$PROJECT_ID/procedures" \
  -H "Content-Type: application/json" \
  -b $COOKIES \
  -d '{
    "name":"Login Test Procedure",
    "description":"Test login functionality",
    "steps":[
      {"action":"navigate","url":"https://example.com/login"},
      {"action":"type","selector":"#username","value":"testuser"},
      {"action":"type","selector":"#password","value":"password"},
      {"action":"click","selector":"#login-button"}
    ]
  }')
PROCEDURE_ID=$(echo $PROCEDURE_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*' | head -1)
echo "Test procedure created with ID: $PROCEDURE_ID"
echo "Response: $PROCEDURE_RESPONSE"
echo

# 2. Create test run v1
echo "2. Creating test run for v1..."
RUN1_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/procedures/$PROCEDURE_ID/runs" \
  -H "Content-Type: application/json" \
  -b $COOKIES \
  -d '{"notes":"First test run"}')
RUN1_ID=$(echo $RUN1_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "Test run created with ID: $RUN1_ID"
echo

# 3. Start test run
echo "3. Starting test run..."
START_RUN=$(curl -s -X POST "$BASE_URL/api/v1/runs/$RUN1_ID/start" \
  -H "Content-Type: application/json" \
  -b $COOKIES)
echo "Start response: $START_RUN"
echo

# 4. Complete test run
echo "4. Completing test run..."
COMPLETE_RUN=$(curl -s -X POST "$BASE_URL/api/v1/runs/$RUN1_ID/complete" \
  -H "Content-Type: application/json" \
  -b $COOKIES \
  -d '{"status":"passed","notes":"All tests passed"}')
echo "Complete response: $COMPLETE_RUN"
echo

# 5. Update test procedure (in-place)
echo "5. Updating test procedure (in-place)..."
UPDATE_PROCEDURE=$(curl -s -X PUT "$BASE_URL/api/v1/projects/$PROJECT_ID/procedures/$PROCEDURE_ID" \
  -H "Content-Type: application/json" \
  -b $COOKIES \
  -d '{"description":"Updated description for login test"}')
echo "Update response: $UPDATE_PROCEDURE"
echo

# 6. Create new version
echo "6. Creating new version of test procedure..."
VERSION_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/projects/$PROJECT_ID/procedures/$PROCEDURE_ID/versions" \
  -H "Content-Type: application/json" \
  -b $COOKIES)
VERSION2_ID=$(echo $VERSION_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*' | head -1)
echo "New version created with ID: $VERSION2_ID"
echo "Response: $VERSION_RESPONSE"
echo

# 7. Get version history
echo "7. Getting version history..."
VERSION_HISTORY=$(curl -s -X GET "$BASE_URL/api/v1/projects/$PROJECT_ID/procedures/$PROCEDURE_ID/versions" -b $COOKIES)
echo "Version history: $VERSION_HISTORY"
echo

# 8. Create test run with new version
echo "8. Creating test run for v2..."
RUN2_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/procedures/$VERSION2_ID/runs" \
  -H "Content-Type: application/json" \
  -b $COOKIES \
  -d '{"notes":"Test run with v2"}')
RUN2_ID=$(echo $RUN2_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "Test run v2 created with ID: $RUN2_ID"
echo

echo "=== Scenario 3: Asset Management ==="
echo

# 1. Upload image asset
echo "1. Uploading image asset..."
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/runs/$RUN2_ID/assets" \
  -b $COOKIES \
  -F "file=@test_image.png" \
  -F "asset_type=image" \
  -F "description=Test screenshot")
ASSET_ID=$(echo $UPLOAD_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "Asset uploaded with ID: $ASSET_ID"
echo "Response: $UPLOAD_RESPONSE"
echo

# 2. List assets
echo "2. Listing assets for test run..."
LIST_ASSETS=$(curl -s -X GET "$BASE_URL/api/v1/runs/$RUN2_ID/assets" -b $COOKIES)
echo "Assets: $LIST_ASSETS"
echo

# 3. Download asset
echo "3. Downloading asset..."
curl -s -X GET "$BASE_URL/api/v1/runs/$RUN2_ID/assets/$ASSET_ID" \
  -b $COOKIES \
  -o downloaded_asset.png
if [ -f downloaded_asset.png ]; then
  echo "Asset downloaded successfully"
  ls -lh downloaded_asset.png
  rm downloaded_asset.png
fi
echo

# 4. Delete asset
echo "4. Deleting asset..."
DELETE_ASSET=$(curl -s -X DELETE "$BASE_URL/api/v1/runs/$RUN2_ID/assets/$ASSET_ID" -b $COOKIES)
echo "Delete response: $DELETE_ASSET"
echo

echo "=== Scenario 4: List All Data ==="
echo

# List test procedures
echo "Listing all test procedures..."
LIST_PROCEDURES=$(curl -s -X GET "$BASE_URL/api/v1/projects/$PROJECT_ID/procedures" -b $COOKIES)
echo "Test procedures: $LIST_PROCEDURES"
echo

# List test runs
echo "Listing all test runs..."
LIST_RUNS=$(curl -s -X GET "$BASE_URL/api/v1/procedures/$PROCEDURE_ID/runs" -b $COOKIES)
echo "Test runs: $LIST_RUNS"
echo

# Get test run details
echo "Getting test run details..."
GET_RUN=$(curl -s -X GET "$BASE_URL/api/v1/runs/$RUN1_ID" -b $COOKIES)
echo "Run details: $GET_RUN"
echo

echo "=== Cleanup ==="
echo

# Delete project (cascade will delete procedures and runs)
echo "Deleting project..."
DELETE_PROJECT=$(curl -s -X DELETE "$BASE_URL/api/v1/projects/$PROJECT_ID" -b $COOKIES)
echo "Delete response: $DELETE_PROJECT"
echo

# Logout
echo "Logging out..."
LOGOUT=$(curl -s -X POST $BASE_URL/api/v1/auth/logout -b $COOKIES -c $COOKIES)
echo "Logout response: $LOGOUT"
echo

# Cleanup files
rm -f $COOKIES test_image.png

echo "=== All integration tests completed successfully! ==="
