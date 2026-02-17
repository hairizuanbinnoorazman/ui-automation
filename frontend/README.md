# UI Automation Frontend

An Elm frontend for the UI Automation testing platform, built with Material Components Web for a modern, responsive interface.

## Features

- **Authentication**: Login and registration with session management
- **Projects Management**: Create, view, update, and delete test projects
- **Test Procedures**: Manage test procedures with versioning support
- **Test Runs**: Execute tests, track status, and manage assets
- **Asset Management**: Upload and view test artifacts (images, videos, documents)
- **Material Design**: Clean, modern UI using Material Components Web

## Project Structure

```
frontend/
├── src/
│   ├── App.elm                    # Main application with routing
│   ├── Types.elm                  # Shared type definitions and JSON codecs
│   ├── API.elm                    # HTTP API client
│   └── Pages/
│       ├── Login.elm              # Login and registration
│       ├── Projects.elm           # Projects CRUD operations
│       ├── TestProcedures.elm     # Test procedures with versioning
│       └── TestRuns.elm           # Test runs with asset management
├── elm.json                       # Elm dependencies
├── index.html                     # HTML entry point
└── README.md                      # This file
```

## Prerequisites

- [Elm 0.19.1](https://guide.elm-lang.org/install/elm.html)
- Backend server running on http://localhost:8080

## Installation

1. Install Elm if you haven't already:
```bash
npm install -g elm
```

2. Navigate to the frontend directory:
```bash
cd frontend
```

3. Install Elm dependencies:
```bash
elm make src/App.elm
```

## Development

### Build the application

```bash
elm make src/App.elm --output=elm.js
```

### Development with live reload

For a better development experience, use `elm-live`:

```bash
npm install -g elm-live
elm-live src/App.elm --open -- --output=elm.js
```

This will:
- Compile your Elm code
- Start a local development server
- Automatically reload when you make changes
- Open your browser to http://localhost:8000

### Production build

For a production build with optimizations:

```bash
elm make src/App.elm --output=elm.js --optimize
```

Then minify the output (optional):

```bash
npm install -g uglify-js
uglifyjs elm.js --compress 'pure_funcs=[F2,F3,F4,F5,F6,F7,F8,F9,A2,A3,A4,A5,A6,A7,A8,A9],pure_getters,keep_fargs=false,unsafe_comps,unsafe' | uglifyjs --mangle --output elm.min.js
```

## Running the Application

1. Make sure the backend server is running:
```bash
cd ..
make run
```

2. Open `index.html` in your browser, or serve it with a simple HTTP server:
```bash
# Using Python 3
python -m http.server 8000

# Using Node.js http-server
npx http-server -p 8000
```

3. Navigate to http://localhost:8000

## Usage

### First Time Setup

1. **Register**: Click "Need an account? Register" on the login page
2. **Login**: Enter your credentials to access the application

### Managing Projects

1. Navigate to the Projects page
2. Click "Create Project" to add a new project
3. Edit or delete projects using the action buttons in the table

### Managing Test Procedures

1. Select a project from the Projects page
2. Navigate to Test Procedures
3. Create procedures with test steps (JSON format)
4. Create new versions when you want to preserve history
5. View version history to see all versions of a procedure

### Managing Test Runs

1. Select a test procedure
2. Navigate to Test Runs
3. Create a new test run
4. Start the run to change status to "Running"
5. Upload assets using the API (see backend README)
6. Complete the run with a final status (Passed/Failed/Skipped)

## API Configuration

The frontend expects the backend API to be available at:
```
http://localhost:8080/api/v1
```

To change this, edit `src/API.elm`:

```elm
baseUrl : String
baseUrl =
    "http://your-backend-url/api/v1"
```

## Architecture

### Module Overview

- **App.elm**: Main application orchestrating routing, navigation, and page composition
- **Types.elm**: All domain types (User, Project, TestProcedure, TestRun, etc.) with JSON encoders/decoders
- **API.elm**: HTTP client for backend communication with type-safe API calls
- **Pages/*.elm**: Individual page modules with their own Model, Msg, update, and view functions

### State Management

Each page manages its own state:
- Login page: Authentication state
- Projects page: List of projects, dialog states for create/edit/delete
- Test Procedures page: Procedures list, versioning, dialogs
- Test Runs page: Runs list, selected run details, assets

### Routing

The application uses URL-based routing:
- `/` - Login page
- `/projects` - Projects list
- `/projects/{projectId}/procedures` - Test procedures for a project
- `/procedures/{procedureId}/runs` - Test runs for a procedure

## Dependencies

### Elm Packages

- `elm/browser` - Browser application framework
- `elm/core` - Core Elm functionality
- `elm/html` - HTML rendering
- `elm/http` - HTTP requests
- `elm/json` - JSON encoding/decoding
- `elm/time` - Time handling
- `elm/url` - URL parsing and routing
- `elm/file` - File handling
- `aforemny/material-components-web-elm` - Material Design components

### External Resources

- Material Components Web CSS/JS (loaded from CDN)
- Material Icons font
- Roboto font

## Development Tips

### Hot Reloading

Use `elm-live` for the best development experience with automatic reloading.

### Debugging

Elm provides excellent compile-time error messages. Read them carefully - they usually point you to the exact problem.

For runtime debugging, use the browser's developer console.

### Type Safety

Elm's type system prevents many runtime errors. If your code compiles, it's likely to work correctly. Pay attention to:
- JSON decoders matching the backend response format
- Route parsing matching your URL structure
- Message passing between parent and child modules

### Code Organization

Each page module follows the Elm Architecture:
1. **Model**: The state of the page
2. **Msg**: All possible messages/events
3. **update**: State transitions based on messages
4. **view**: Render the UI based on current state

## Common Issues

### CORS Errors

If you see CORS errors, ensure your backend allows requests from the frontend origin. Add CORS headers in your backend configuration.

### Module Not Found

If Elm can't find a module, check:
- The module name matches the filename
- The file is in the correct directory
- `elm.json` has the correct source directories

### JSON Decoding Errors

If API responses fail to decode:
- Check that the decoder in `Types.elm` matches the backend response
- Verify field names match exactly (case-sensitive)
- Check that types match (String vs Int, etc.)

## Future Enhancements

- File upload UI for test run assets
- Inline image/video preview
- Advanced filtering and search
- Export test results
- Dark mode support
- Internationalization

## Contributing

When adding new features:
1. Add types to `Types.elm`
2. Add API functions to `API.elm`
3. Create or update page modules
4. Update routing in `App.elm` if needed
5. Test with the backend API

## License

Same as the main UI Automation project.
