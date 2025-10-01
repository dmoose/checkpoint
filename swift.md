# macOS Checkpoint Summary Application - SwiftUI Specification

## 1. Application Overview

**Checkpoint Summary** is a native macOS SwiftUI application that aggregates and displays development progress across multiple LLM-assisted projects. The app scans user-configured root directories, identifies projects tracked by the Checkpoint tool, processes their changelogs, generates daily summaries, and provides a unified dashboard for viewing project progress.

**Target:** macOS 15+ (Sequoia), Swift 6, SwiftUI
**Integration:** Apple Intelligence for summarization (macOS Sequoia 15.1+, with graceful fallback)

---

## 2. Core Architecture

### 2.1 Application Lifecycle

```
App Launch
  ↓
Load Configuration (root directories, output location)
  ↓
User Initiates "Process Projects" or scheduled trigger
  ↓
Discovery Phase:
  - Recursively scan all configured roots
  - Identify git repositories with .checkpoint-status files
  - Build project registry
  ↓
Processing Phase (per project):
  - Read existing daily summaries to find last processed date
  - Read .changelog file
  - Find first unprocessed entry
  - Collect all entries since last processed date
  - Bucket entries by local date
  - Write daily summaries for project
  - Track all dates with project data
  ↓
Cross-Project Phase:
  - For each date with project data
  - Check if cross-project summary exists
  - Generate if missing
  ↓
UI Update & Display
```

### 2.2 Processing State Management

The app tracks processing state to enable incremental updates:

- **Last processed timestamp per project** (stored locally)
- **Processed changelog entries** (identified by commit hash)
- **Generated summary dates** (tracked as set across all projects)
- **Cross-project summary dates** (separate tracking)

---

## 3. Output Directory Structure

```
{OutputRoot}/
├── projects/
│   ├── {ProjectName1}/
│   │   └── daily_summaries/
│   │       ├── {ProjectName1}_2025-10-22.md
│   │       ├── {ProjectName1}_2025-10-21.md
│   │       └── ...
│   ├── {ProjectName2}/
│   │   └── daily_summaries/
│   │       ├── {ProjectName2}_2025-10-22.md
│   │       └── ...
│   └── ...
├── overview/
│   ├── SUMMARY_2025-10-22.md
│   ├── SUMMARY_2025-10-21.md
│   └── ...
└── app_state.json  # Persisted tracking data
```

**File Naming Convention:**
- Project summaries: `{ProjectName}_{YYYY-MM-DD}.md`
- Cross-project summaries: `SUMMARY_{YYYY-MM-DD}.md`
- Files are indexed by date for easy chronological access

---

## 4. Data Models

### 4.1 Core Structures

```swift
// MARK: - Configuration
struct AppConfiguration: Codable {
    var rootDirectories: [String]          // User-configured scan roots
    var outputLocation: String              // Where summaries are written
    var lastProcessingTimestamp: Date?      // When app last ran
}

// MARK: - Project Discovery
struct ProjectRegistry {
    var projects: [CheckpointProject]
    var lastScannedDate: Date
}

struct CheckpointProject: Identifiable, Hashable {
    let id: UUID
    let projectPath: String                 // Full path to git root
    let projectName: String                 // Derived from directory name
    let statusFilePath: String              // Path to .checkpoint-status
    let changelogPath: String               // Path to .changelog

    // Tracking state
    var lastProcessedCommitHash: String?    // Last processed entry
    var lastProcessedDate: Date?            // Date of last processing
    var dailySummariesPath: String          // Output path for summaries
    var summaryFilesByDate: [String: String] = [:] // Date string -> file path
}

// MARK: - Changelog Entry (parsed from YAML)
struct ChangelogEntry: Identifiable {
    let id: String                          // Derived from commit hash
    let schemaVersion: String
    let timestamp: Date                     // ISO8601 parsed to Date
    let summary: String
    let details: String?
    let changeType: String                  // feature, fix, refactor, docs, perf, other
    let scope: String?
    let commitHash: String?

    // Derived
    var localDate: Date                     // Timestamp in user's timezone, normalized to midnight
}

// MARK: - Processing Results
struct DayBucket: Identifiable {
    let id: String                          // Date string (YYYY-MM-DD)
    let date: Date
    let entries: [ChangelogEntry]
    let project: CheckpointProject
}

struct ProcessedDay {
    let date: Date
    let projects: [String]                  // Project names with entries that day
    let hasCrossProjectSummary: Bool
}

// MARK: - Summary File Content
struct SummaryMetadata {
    let date: Date
    let project: String?                    // nil for cross-project summaries
    let generatedDate: Date
    let entriesCount: Int
}

// MARK: - App State
struct AppProcessingState: Codable {
    var projects: [ProjectProcessingState] = []
    var processedDates: [String] = []       // YYYY-MM-DD format, sorted reverse
    var crossProjectDates: [String] = []    // YYYY-MM-DD format, sorted reverse
}

struct ProjectProcessingState: Codable {
    let projectPath: String
    let projectName: String
    var lastProcessedCommitHash: String?
    var lastProcessedDate: String?          // ISO8601
    var summaryDates: [String] = []         // YYYY-MM-DD format
}
```

---

## 5. File System Operations

### 5.1 Configuration Management

**Configuration File:** `~/Library/Application Support/CheckpointSummary/config.json`

```swift
protocol ConfigurationManager {
    func loadConfiguration() -> AppConfiguration
    func saveConfiguration(_ config: AppConfiguration)
    func updateRootDirectories(_ paths: [String])
    func setOutputLocation(_ path: String)
}
```

**User-Facing Configuration:**
- Dialog to add/remove root directories
- Dialog to select output directory
- Persist in standard macOS app preferences location

### 5.2 Project Discovery

```swift
protocol ProjectDiscovery {
    func scanForProjects(in roots: [String]) async -> [CheckpointProject]
    func isValidProject(at path: String) -> Bool
    func findCheckpointStatusFile(in path: String) -> String?
    func extractProjectName(from path: String) -> String
}
```

**Algorithm:**
- Walk directory tree recursively
- At each level, check for `.git` directory
- If git repo found, look for `.checkpoint-status`
- If status file found, add to registry
- Stop recursing into non-git directories

### 5.3 Changelog Parsing

```swift
protocol ChangelogParser {
    func parseChangelog(at path: String) -> [ChangelogEntry]
    func extractUnprocessedEntries(
        all: [ChangelogEntry],
        lastProcessed: String?
    ) -> [ChangelogEntry]
}
```

**Format:** YAML, entries separated by `---`

**Parsing Strategy:**
- Read file from end backwards (append-only)
- Parse each YAML document
- Build array of `ChangelogEntry` structs
- Handle mixed schema versions gracefully (forgiving of missing fields)

### 5.4 Summary Generation

```swift
protocol SummaryGenerator {
    func generateProjectDaySummary(
        entries: [ChangelogEntry],
        project: String,
        date: Date,
        useAppleIntelligence: Bool
    ) async -> String

    func generateCrossProjectSummary(
        dayBuckets: [DayBucket],
        date: Date,
        useAppleIntelligence: Bool
    ) async -> String
}
```

**Output Format:** Markdown

**Project Summary Example:**

```markdown
# {ProjectName} - {Date: YYYY-MM-DD}

## Summary
{AI-generated narrative if available, or structured list}

## Changes

### feature: {scope} - {summary}
- Details: {details}
- Commit: {hash}

### fix: Database connection pooling
- Details: ...
- Commit: ...

[All entries for the day, grouped by change_type]
```

**Cross-Project Summary Example:**

```markdown
# Development Summary - {Date: YYYY-MM-DD}

## Overview
{AI-generated overview of all projects}

## Projects Updated

### ProjectName1
- {summary}
- {summary}

### ProjectName2
- {summary}

[All projects with entries, organized hierarchically]
```

---

## 6. Processing Logic

### 6.1 Incremental Processing Algorithm

```
For each project in registry:
  1. Create output directory if missing
  2. Scan existing daily_summaries/ to find most recent file
  3. Extract date from filename (YYYY-MM-DD)
  4. Load app state to find lastProcessedCommitHash for this project
  5. Read changelog file
  6. Parse all YAML entries

  7. Find entries newer than last processed:
     - If lastProcessedCommitHash known: start from entry after it
     - If unknown: start from first entry newer than last processed date
     - If no tracking: process all

  8. Bucket unprocessed entries by local date
  9. For each date bucket:
     - Generate project summary
     - Write to {ProjectName}_{YYYY-MM-DD}.md
     - Track date in processedDates set

  10. Update app state:
      - lastProcessedCommitHash = most recent entry processed
      - lastProcessedDate = most recent date processed
      - summaryDates = all dates written

Collect all unique dates from all projects into processedDates set

For each date in processedDates:
  1. Check if overview/SUMMARY_{YYYY-MM-DD}.md exists
  2. If exists: skip
  3. If missing:
     - Collect all project summaries for that date
     - Read each summary file
     - Generate cross-project summary
     - Write to overview/SUMMARY_{YYYY-MM-DD}.md
     - Add to crossProjectDates set

Save app state to disk
```

### 6.2 Resume Capability

- App state persisted to `app_state.json` after each run
- On app launch or reprocessing, read app state
- Use stored `lastProcessedCommitHash` to resume
- Only process new entries since last run
- Regenerate if user deletes summary files

---

## 7. Apple Intelligence Integration

### 7.1 Conditional Availability

```swift
protocol AppleIntelligenceProvider {
    var isAvailable: Bool { get }  // macOS 15.1+
    func summarizeText(_ text: String) async throws -> String
}

// Fallback implementation if not available
class NoOpIntelligenceProvider: AppleIntelligenceProvider {
    var isAvailable: Bool { false }
    func summarizeText(_ text: String) async throws -> String {
        return text  // Return structured data as-is
    }
}
```

### 7.2 Summarization Strategy

**If Apple Intelligence available:**
- Convert changelog entries to natural language
- Pass to Apple Intelligence summarization API
- Include in summary as "AI Summary" section
- Fall back to structured format if API fails

**If not available:**
- Present structured data in formatted markdown
- Group entries by change_type
- Use clear headers and formatting

```swift
func createSummaryText(from entries: [ChangelogEntry]) -> String {
    var text = ""
    for changeType in ["feature", "fix", "refactor", "docs", "perf", "other"] {
        let filtered = entries.filter { $0.changeType == changeType }
        if !filtered.isEmpty {
            text += "\n## \(changeType.capitalized)\n"
            for entry in filtered {
                text += "- \(entry.summary)\n"
                if let details = entry.details {
                    text += "  \(details)\n"
                }
            }
        }
    }
    return text
}
```

---

## 8. User Interface Specification

### 8.1 Main Window Structure

```
┌─────────────────────────────────────────────┐
│ Checkpoint Summary              ⚙️ 🔄 ➕  │
├──────────┬──────────────────────────────────┤
│          │                                  │
│ Sidebar  │      Main Content Area           │
│          │                                  │
│ Overview │  Overview Screen                 │
│ ──────── │  (showing cross-project summaries)
│ Projects │  OR Project Detail Screen        │
│ ──────── │  (showing project-specific data)
│   • Proj1│  OR Summary Text View            │
│   • Proj2│  (markdown display)              │
│   • Proj3│                                  │
│          │                                  │
└──────────┴──────────────────────────────────┘
```

### 8.2 Overview Screen

**Purpose:** Display all days with cross-project summaries in reverse chronological order

**UI Elements:**
- List of dates (newest first) with `SUMMARY_{YYYY-MM-DD}.md` files
- Each row shows:
  - Date (YYYY-MM-DD)
  - Number of projects updated that day
  - Short preview of summary (first 100 characters)
- Click to open summary in text view
- "Process Now" button (top right) to trigger processing

**Layout:**
```swift
VStack {
    HStack {
        Text("Daily Overviews")
            .font(.title)
        Spacer()
        Button(action: processNow) {
            Label("Process Now", systemImage: "arrow.clockwise")
        }
        Button(action: showSettings) {
            Image(systemName: "gear")
        }
    }
    .padding()

    List(summaryDates, id: \.self) { dateString in
        NavigationLink(destination: SummaryTextView(summaryPath: path(for: dateString))) {
            VStack(alignment: .leading, spacing: 4) {
                Text(dateString)
                    .font(.headline)
                Text("\(projectsForDate[dateString]?.count ?? 0) projects updated")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
        }
    }
}
```

### 8.3 Project List & Selection

**Sidebar View:**
- Expandable/selectable list of all projects with daily summaries
- Show project name + count of daily summaries
- Highlight selected project

**Behavior:**
- Clicking project name navigates to Project Detail Screen
- Show visual indicator for selected project

```swift
List(projects, id: \.id, selection: $selectedProject) { project in
    NavigationLink(destination: ProjectDetailScreen(project: project)) {
        VStack(alignment: .leading) {
            Text(project.projectName)
                .font(.headline)
            Text("\(project.summaryFilesByDate.count) summaries")
                .font(.caption)
                .foregroundColor(.secondary)
        }
    }
}
```

### 8.4 Project Detail Screen

**Purpose:** Display project-specific information and daily summaries

**Sections:**
1. **Project Header**
   - Project name
   - Path to git repository
   - Last processed date
   - Count of daily summaries

2. **Daily Summaries List**
   - All summaries for this project in reverse chronological order
   - Date + count of entries that day
   - Click to open in text view

**Layout:**
```swift
VStack {
    VStack(alignment: .leading, spacing: 8) {
        Text(project.projectName)
            .font(.title)
        Text(project.projectPath)
            .font(.caption)
            .foregroundColor(.secondary)
        HStack {
            Text("Summaries: \(project.summaryFilesByDate.count)")
            Spacer()
            Text("Last: \(project.lastProcessedDate?.formatted() ?? "—")")
        }
        .font(.caption)
    }
    .padding()

    List(sortedSummaries, id: \.key) { dateString, filePath in
        NavigationLink(destination: SummaryTextView(summaryPath: filePath)) {
            Text(dateString)
        }
    }
}
.navigationTitle("Project Details")
```

### 8.5 Summary Text View

**Purpose:** Display markdown summary file content

**UI Elements:**
- Read-only text display (or attributed text for basic markdown formatting)
- Scrollable
- Copy button
- Back/close navigation

**Note on MVP:** Uses plain text or basic text styling. Future iteration will add third-party markdown viewer (e.g., **MarkdownUI** or **Splash**).

```swift
VStack {
    ScrollView {
        Text(summaryContent)
            .font(.system(.body, design: .monospaced))
            .padding()
            .frame(maxWidth: .infinity, alignment: .leading)
    }

    HStack {
        Spacer()
        Button(action: copySummary) {
            Label("Copy", systemImage: "doc.on.doc")
        }
    }
    .padding()
}
.navigationTitle("Summary")
```

### 8.6 Settings/Configuration Screen

**Modal or Sheet:**
- Add/remove root directories
- Set output location
- Show current configuration
- Manual cache clear option
- Processing status/logs

```swift
Form {
    Section("Scan Directories") {
        ForEach(configuration.rootDirectories, id: \.self) { dir in
            HStack {
                Text(dir)
                Spacer()
                Button(role: .destructive) {
                    configuration.rootDirectories.removeAll { $0 == dir }
                } label: {
                    Image(systemName: "xmark")
                }
            }
        }
        Button(action: addDirectory) {
            Label("Add Directory", systemImage: "plus")
        }
    }

    Section("Output Location") {
        HStack {
            Text(configuration.outputLocation)
            Spacer()
            Button(action: chooseOutputLocation) {
                Text("Change")
            }
        }
    }

    Section("Actions") {
        Button(role: .destructive) {
            clearCache()
        } label: {
            Text("Clear Cache")
        }
    }
}
```

---

## 9. Navigation & State Management

### 9.1 Navigation Structure

```
Root
├── Overview Screen (default)
│   └── Summary Text View (selected summary)
├── Projects Sidebar
│   └── Project Detail Screen
│       └── Summary Text View (selected project summary)
└── Settings (modal/sheet)
```

### 9.2 State Container

```swift
@main
struct CheckpointSummaryApp: App {
    @StateObject private var coordinator = AppCoordinator()
    @StateObject private var configManager = ConfigurationManager()

    var body: some Scene {
        WindowGroup {
            MainView()
                .environmentObject(coordinator)
                .environmentObject(configManager)
        }
    }
}

@MainActor
class AppCoordinator: ObservableObject {
    @Published var selectedProject: CheckpointProject?
    @Published var selectedSummaryPath: String?
    @Published var projects: [CheckpointProject] = []
    @Published var summaryDates: [String] = []
    @Published var isProcessing = false
    @Published var lastProcessingError: String?

    func processProjects() async { /* ... */ }
    func loadState() async { /* ... */ }
}
```

---

## 10. Processing & Background Operations

### 10.1 Async Processing

```swift
func processProjects() async {
    defer { isProcessing = false }
    isProcessing = true

    do {
        // Phase 1: Discovery
        let projects = try await projectDiscovery.scanForProjects(
            in: configManager.configuration.rootDirectories
        )

        // Phase 2: Process each project
        var allProcessedDates = Set<String>()
        for project in projects {
            let dates = try await processProject(project)
            allProcessedDates.formUnion(dates)
        }

        // Phase 3: Generate cross-project summaries
        try await generateCrossProjectSummaries(for: allProcessedDates)

        // Phase 4: Update UI
        await MainActor.run {
            self.projects = projects
            self.summaryDates = Array(allProcessedDates).sorted(by: >)
        }
    } catch {
        lastProcessingError = error.localizedDescription
    }
}
```

### 10.2 Error Handling

- Handle missing files gracefully
- Skip corrupted changelog entries
- Log warnings but continue processing
- Display non-fatal errors in UI
- Persist state even if processing partially fails

---

## 11. Data Persistence

### 11.1 App State File

**Location:** `OutputRoot/app_state.json`

**Contents:**
```json
{
  "projects": [
    {
      "projectPath": "/Users/dev/project1",
      "projectName": "project1",
      "lastProcessedCommitHash": "abc123def456",
      "lastProcessedDate": "2025-10-22T00:00:00Z",
      "summaryDates": ["2025-10-22", "2025-10-21"]
    }
  ],
  "processedDates": ["2025-10-22", "2025-10-21"],
  "crossProjectDates": ["2025-10-22", "2025-10-21"]
}
```

### 11.2 Caching Strategy

- Cache project list in memory (updated on each "Process Now")
- Cache parsed changelogs temporarily during processing
- Disk cache for app state (persisted after each processing run)
- Allow manual cache clear in settings

---

## 12. Error Handling & Robustness

### 12.1 Common Failure Scenarios

| Scenario | Behavior |
|----------|----------|
| `.checkpoint-status` missing | Skip project, log warning, continue |
| `.changelog` file corrupted | Log error, skip that project |
| YAML parsing fails | Log entry details, skip entry, continue |
| Output directory not writable | Show error, ask user to fix |
| Apple Intelligence unavailable | Use fallback (structured markdown) |
| No changes to process | Show message "No new summaries" |
| Permissions denied on root dir | Log and skip that root, continue with others |

### 12.2 Logging

- Write processing logs to `OutputRoot/app.log`
- Include timestamps, severity (INFO, WARN, ERROR)
- Show last 20 log entries in settings view (future iteration)
- Rotate log file when > 1MB

---

## 13. MVP Scope vs. Future Enhancements

### MVP (In Scope for Initial Release)

- ✅ Configuration of root directories and output location
- ✅ Recursive project discovery with `.checkpoint-status` validation
- ✅ Changelog parsing and incremental processing
- ✅ Daily summary generation (structured markdown)
- ✅ Cross-project summary generation
- ✅ Apple Intelligence integration for summarization (graceful fallback)
- ✅ Overview screen showing cross-project summaries by date
- ✅ Project list and detail screens
- ✅ Summary text view for reading markdown
- ✅ "Process Now" manual trigger
- ✅ Basic settings/configuration UI
- ✅ App state persistence and resume capability

### Future Enhancements (Out of Scope for MVP)

- 🔮 Markdown viewer library integration for rich formatting
- 🔮 Project statistics (commits/day, change type distribution)
- 🔮 PLAN.md creation/editing interface in app
- 🔮 Scheduled automatic processing (background task)
- 🔮 Lock file detection (.checkpoint-lock from Go tool)
- 🔮 iCloud sync of summaries
- 🔮 Export summaries to PDF or HTML
- 🔮 Search across all summaries
- 🔮 Tags and filtering by project or change type
- 🔮 Custom summary templates
- 🔮 Webhook notifications when summaries generated
- 🔮 Time range filtering (date picker for week/month views)
- 🔮 Customizable theme and appearance

---

## 14. Technical Considerations

### 14.1 Swift 6 & Concurrency

- Use `async/await` for all file I/O and processing
- Use `MainActor` for UI updates
- Use structured concurrency with `Task` groups for parallel processing
- Avoid `@escaping` closures; prefer `async/await`

### 14.2 File System Access

- Request user permission for directory access
- Use `NSOpenPanel` for directory selection
- Store permitted paths in app preferences
- Handle symlinks and permissions gracefully

### 14.3 macOS 15+ Compatibility

- Use SwiftUI for all UI
- Leverage native macOS controls (NSToolbar, sidebar)
- Use `@Environment` for system properties (color scheme, etc.)
- Test on Sequoia (15.x) and Tahoe (26.x)

### 14.4 YAML Parsing

- Use Codable with custom decoders or lightweight YAML library
- Recommend: `Yams` library (pure Swift, no C dependencies)
- Fallback: `yaml.v3` equivalent or manual TOML parsing if YAML unavailable

### 14.5 Dependencies (Minimal)

**Required:**
- `Yams` (YAML parsing)

**Optional (for future):**
- `MarkdownUI` (markdown rendering, v0.4+)
- `Splash` (syntax highlighting)

---

## 15. AppKit Integration Points

### 15.1 Native macOS Features to Leverage

- `NSOpenPanel` for directory/file selection
- `NSSavePanel` for output location
- Notifications for processing completion
- Draggable path indicators
- Right-click context menus on projects/summaries
- Keyboard shortcuts (Cmd+R for refresh, Cmd+, for settings)

### 15.2 Menu Bar Items

- Menu > Checkpoint Summary
  - Process Now (Cmd+R)
  - Settings (Cmd+,)
  - About
  - Quit

---

## 16. Data Flow Diagram

```
┌─────────────────────────┐
│   User Configuration    │
│  (root dirs, output)    │
└────────────┬────────────┘
             ↓
     ┌───────────────┐
     │ App Launch    │
     │ Load State    │
     └───────┬───────┘
             ↓
    ┌────────────────────┐
    │ User Clicks        │
    │ "Process Now"      │
    └────────┬───────────┘
             ↓
   ┌─────────────────────────┐
   │ Project Discovery       │
   │ (scan roots for .git +  │
   │  .checkpoint-status)    │
   └────────┬────────────────┘
            ↓
   ┌────────────────────────────┐
   │ For Each Project:          │
   │ - Read changelog           │
   │ - Find new entries         │
   │ - Bucket by date           │
   │ - Generate summaries       │
   │ - Track processed dates    │
   └────────┬───────────────────┘
            ↓
   ┌────────────────────────────┐
   │ Generate Cross-Project     │
   │ Summaries for all          │
   │ processed dates            │
   └────────┬───────────────────┘
            ↓
   ┌────────────────────────────┐
   │ Save App State             │
   │ Update UI                  │
   │ Show Results               │
   └────────────────────────────┘
```

---

## 17. File Structure & Key Classes

```
CheckpointSummary/
├── App/
│   ├── CheckpointSummaryApp.swift       # @main entry point
│   └── AppCoordinator.swift             # State management
├── Views/
│   ├── MainView.swift                   # Root container
│   ├── OverviewScreen.swift             # Cross-project summaries
│   ├── ProjectDetailScreen.swift        # Project details + summaries
│   ├── SummaryTextView.swift            # Markdown display
│   ├── Sidebar.swift                    # Project list
│   └── SettingsView.swift               # Configuration
├── Models/
│   ├── ProjectModels.swift              # CheckpointProject, DayBucket
│   ├── ChangelogModels.swift            # ChangelogEntry
│   ├── AppStateModels.swift             # AppProcessingState
│   └── ConfigurationModels.swift        # AppConfiguration
├── Services/
│   ├── ProjectDiscovery.swift           # Scanning logic
│   ├── ChangelogParser.swift            # YAML parsing
│   ├── SummaryGenerator.swift           # Summary creation
│   ├── AppleIntelligenceProvider.swift  # AI integration
│   ├── ConfigurationManager.swift       # Settings persistence
│   └── FileSystemManager.swift          # I/O operations
├── Utilities/
│   ├── Logger.swift                     # Logging
│   ├── DateFormatter.swift              # Date handling
│   └── PathNormalizer.swift             # Path utilities
└── Resources/
    └── Localizable.strings              # i18n (future)
```

---

## 18. Implementation Roadmap

### Phase 1: Foundation
- App structure and navigation
- Configuration UI
- File system access setup

### Phase 2: Core Logic
- Project discovery
- Changelog parsing
- Summary generation (basic markdown)

### Phase 3: UI & Display
- Overview and project screens
- Text view for summaries
- Sidebar navigation

### Phase 4: Apple Intelligence
- Integration with summarization API
- Fallback handling
- Error management

### Phase 5: Polish & Testing
- Error handling and edge cases
- Performance optimization
- User testing and refinement

---

This specification provides a complete blueprint for implementing the macOS Checkpoint Summary application. Ready to proceed with SwiftUI code implementation?

## The Even Simpler Approach (What You're Getting At)

**Don't store app state for processing markers at all.** Let the filesystem tell you what's been processed:

```
On each run:

For each project:
  1. Find all existing summary files for this project
     (e.g., scan project_summaries/ dir)

  2. Extract the most recent date from filenames
     (e.g., "project1_2025-10-22.md" → 2025-10-22)

  3. Read changelog entries

  4. Filter: keep only entries timestamped after that date

  5. Process & write new summaries for new dates
```

**This is safer and simpler because:**

- ✅ **Filesystem is source of truth** – If summary file exists, that date was processed
- ✅ **Natural regeneration** – Delete a summary file → app regenerates it next run
- ✅ **No state file needed** – Zero consistency issues
- ✅ **Handles manual edits** – User adds new entry to old date in changelog → finds existing summary for that date → notices it's newer and regenerates
- ✅ **Atomic** – Each summary file is a checkpoint; either it exists or it doesn't
- ✅ **Observable** – User can inspect output dir and see exactly what's been processed (it's all there)

---

## Trade-offs Compared to Hash Tracking

| Scenario | Hash Tracking | Filesystem-Based |
|----------|---------------|------------------|
| New entry added to old date in changelog | Would process only new entry | Would regenerate entire date (safe, expected) |
| Summary file deleted | Hash remains; would skip that date | Date would be missing; regenerated next run (correct behavior) |
| Fast resume | Fast (don't re-scan that date) | Slightly slower (check file dates, but negligible) |
| Code complexity | Store/load state, match hashes | Just read filenames, compare dates |

---

## My Revised Recommendation

**No app state file at all.** Just:

1. Recursively find projects (fresh each run)
2. For each project, scan `project_summaries/` dir to find newest date
3. Read changelog entries newer than that date
4. Process & write
5. Repeat for cross-project

**That's it.** No state file. Filesystem is your state machine.

You're right to push back on this. The hash was cargo-culting from typical "last processed ID" patterns. Your system doesn't need it.

> @ 2025-10-22 15:49:13 UTC
>
> what are line length issues in yaml, toml, your changelog schema is pretty 1 dimensional, I'm thinking change entries  are more of an array than a list, how is that handled in yaml/toml

<details><summary>Thinking</summary>

The user is asking about:
1. Line length issues in YAML and TOML
2. Whether their changelog schema is too 1-dimensional
3. How arrays/lists are handled in YAML/TOML for multiple entries

They're thinking about whether entries should be an array within a single document rather than multiple separate documents separated by `---`.

This is a good architectural question. Let me think through this:

Current approach (from the spec):
```yaml
---
schema_version: "1"
timestamp: "..."
summary: "..."
# ... more fields
---
schema_version: "1"
timestamp: "..."
summary: "..."
```

Multiple documents in one file, separated by `---`.

Proposed approach (what they're considering):
```yaml
version: "1"
entries:
  - timestamp: "..."
    summary: "..."
    # ... more fields
  - timestamp: "..."
    summary: "..."
    # ... more fields
```

Or in TOML:
```toml
version = "1"

[[entries]]
timestamp = "..."
summary = "..."

[[entries]]
timestamp = "..."
summary = "..."
```

Line length issues:
- YAML: No hard limit, but readability degrades after ~100-120 chars. Long strings can be wrapped with `|` or `>` for multiline.
- TOML: Similar, no hard limit. Uses `"""` for multiline strings.

Array representation:
- YAML: Uses `-` for array items, very clean
- TOML: Uses `[[table]]` for array of tables, or inline arrays

The user is right that their current approach treats each entry as its own document. An array approach would be more structured.

Let me discuss the pros/cons of both approaches.

</details>

Good architectural question. You've identified a real structural issue. Let me break this down:

---

## Current Approach (Multi-Document)

```yaml
---
schema_version: "1"
timestamp: "2025-10-22T16:00:00Z"
summary: "Added authentication"
details: "Implemented JWT tokens with refresh rotation"
change_type: "feature"
scope: "auth"
commit_hash: "abc123"
---
schema_version: "1"
timestamp: "2025-10-22T15:00:00Z"
summary: "Fixed connection leak"
change_type: "fix"
scope: "database"
commit_hash: "def456"
```

**Issues:**
- Schema version repeated for every entry (redundant)
- Parsing logic: split on `---`, parse each as separate YAML document
- Feels like a workaround, not a natural data structure
- Line length is fine (individual entries are small), but the repetition is wasteful

---

## Array Approach (Single Document with Entry Array)

### YAML Version

```yaml
schema_version: "1"
entries:
  - timestamp: "2025-10-22T16:00:00Z"
    summary: "Added authentication"
    details: "Implemented JWT tokens with refresh rotation"
    change_type: "feature"
    scope: "auth"
    commit_hash: "abc123"
  - timestamp: "2025-10-22T15:00:00Z"
    summary: "Fixed connection leak"
    change_type: "fix"
    scope: "database"
    commit_hash: "def456"
```

**Advantages:**
- ✅ Schema version declared once at top
- ✅ Entries are naturally an array (cleaner semantics)
- ✅ Single parse: one YAML document
- ✅ Easier to add metadata (last_updated, processing_notes, etc.)
- ✅ No delimiter confusion or edge cases

**Line length:** Still fine. Longest lines are ~50-60 chars. Multiline details use `|` or `>`:

```yaml
entries:
  - timestamp: "2025-10-22T16:00:00Z"
    summary: "Added authentication"
    details: |
      Implemented JWT tokens with refresh rotation.
      Added token expiry and refresh endpoints.
      Updated middleware chain for auth checks.
    change_type: "feature"
```

### TOML Version

```toml
schema_version = "1"

[[entries]]
timestamp = "2025-10-22T16:00:00Z"
summary = "Added authentication"
details = "Implemented JWT tokens with refresh rotation"
change_type = "feature"
scope = "auth"
commit_hash = "abc123"

[[entries]]
timestamp = "2025-10-22T15:00:00Z"
summary = "Fixed connection leak"
change_type = "fix"
scope = "database"
commit_hash = "def456"
```

**Line length:** TOML has same constraints as YAML. Multiline strings use `"""`:

```toml
[[entries]]
timestamp = "2025-10-22T16:00:00Z"
summary = "Added authentication"
details = """
Implemented JWT tokens with refresh rotation.
Added token expiry and refresh endpoints.
Updated middleware chain for auth checks.
"""
```

---

## Line Length Comparison

| Format | Typical Max | Hard Limit | Readability |
|--------|-------------|-----------|------------|
| YAML | 80-100 chars | None | Degrades after ~120 |
| TOML | 80-100 chars | None | Degrades after ~120 |
| Both support multiline strings with folding (`>`, `\|`, `"""`) | — | — | — |

**Your use case:** Short summaries (1 line), longer details (3-5 lines). **No line length issues** with either format.

---

## Parsing Implications

### Multi-Document (Current)

```go
// Ugly parsing
docs := strings.Split(content, "\n---\n")
for _, doc := range docs {
    entry := &ChangelogEntry{}
    yaml.Unmarshal([]byte(doc), entry)
    entries = append(entries, entry)
}
```

### Array Structure (Proposed)

```go
// Clean parsing
type ChangelogFile struct {
    SchemaVersion string
    Entries       []ChangelogEntry
}

var changelog ChangelogFile
yaml.Unmarshal(content, &changelog)
entries := changelog.Entries
```

**Much cleaner.** Standard `Codable` works directly.

---

## My Recommendation

**Switch to array structure.** Here's why:

| Aspect | Multi-Doc | Array |
|--------|-----------|-------|
| Parsing | Requires string split + loop | Direct unmarshaling |
| Schema versioning | Per-entry (wasteful) | Once at top (correct) |
| Data semantics | "List of documents" | "Collection of entries" (correct) |
| Append-only appending | Append `---\n{yaml}` | Read file → unmarshal → append to array → marshal → write |
| Line length | Fine | Fine |
| Swift parsing | More complex | Direct Codable |
| Go parsing | String manipulation | Direct unmarshaling |

**The append-only requirement changes slightly:**

Instead of:
```go
file.WriteString("\n---\n" + newEntryYAML)
```

You'd do:
```go
// Read entire file
changelog := ReadAndUnmarshal(filePath)

// Append new entry to array
changelog.Entries = append(changelog.Entries, newEntry)

// Write back entire file
MarshalAndWrite(filePath, changelog)
```

This is **still append-only conceptually** (you're adding to the entries list), but you're rewriting the whole file. At your scale (dozens-hundreds of projects, maybe 50-100 entries per project max), this is fine. File size is <100KB even for large projects.

---

## Trade-off: Append-Only Semantics

The multi-document approach *felt* append-only (literally append bytes to file). The array approach requires full re-write.

**But the multi-document approach has a hidden cost:**
- If parsing fails halfway through file, you lose some entries
- Corruption mid-file breaks the rest
- Harder to validate entire file

Array approach is actually **safer** because you parse the whole thing or nothing.

---

## My Vote

**Use array structure with single schema_version at top.** It's cleaner, more maintainable, and the re-write cost is negligible at your scale. Plus, it's easier for the macOS app to parse in Swift.

Should I update the schema in the specification, or do you want to prototype both approaches first?

> @ 2025-10-22 15:51:51 UTC
>
> you missed the point, per document append only is safer, I meant for the actually schema per document, your schema assumes a very precise checkpoint with few actual changes, i'm thinking a checkpoint may have a bug fix, a feature, a documentation update, etc. you schema makes capturing this challenging

<details><summary>Thinking</summary>

Ah, I see now. The user is saying:

1. Per-document append-only IS safer (I was wrong to second-guess that earlier)
2. The schema needs to support MULTIPLE changes within a single checkpoint, not one change per entry
3. My current schema treats each entry as a single change (one summary, one change_type, one scope)
4. But a real checkpoint might have: a bug fix in auth, a feature in dashboard, docs updates, perf improvements to database query layer, etc.

So the schema should be:
- One document per checkpoint (per commit)
- Multiple "changes" or "items" within that checkpoint
- Each change has its own type, scope, summary

This makes much more sense. A user checkpoints their work, which might touch multiple areas. The Go tool creates ONE checkpoint entry, but that entry contains an ARRAY of changes.

Current schema (wrong):
```yaml
---
schema_version: "1"
timestamp: "2025-10-22T16:00:00Z"
summary: "Added authentication"
details: "..."
change_type: "feature"
scope: "auth"
commit_hash: "abc123"
```

Revised schema (correct):
```yaml
---
schema_version: "1"
timestamp: "2025-10-22T16:00:00Z"
commit_hash: "abc123"
changes:
  - summary: "Added JWT authentication"
    details: "Implemented JWT tokens..."
    change_type: "feature"
    scope: "auth"
  - summary: "Fixed redirect bug in login flow"
    change_type: "fix"
    scope: "auth"
  - summary: "Updated auth documentation"
    change_type: "docs"
    scope: "auth"
  - summary: "Optimized token verification queries"
    change_type: "perf"
    scope: "database"
```

This makes way more sense. Each checkpoint (commit) can have multiple changes listed, each with its own type and scope. The timestamp and commit_hash are at the checkpoint level.

The multi-document approach offers robust protection and flexibility. Breaking changes into individual, atomic documents ensures data integrity. If one document gets corrupted, the rest remain intact. The append-only model allows seamless additions without complex restructuring. Parsing from the end backwards provides a natural chronological review, while enabling incremental processing for applications like the Mac app.

</details>

Ah, now I understand. You're right on both counts:

1. **Per-document append-only IS safer** – I was wrong to second-guess that
2. **Multiple changes per checkpoint** – My schema is too rigid

---

## The Real Problem with Current Schema

You're capturing ONE change per entry:

```yaml
---
schema_version: "1"
timestamp: "2025-10-22T16:00:00Z"
summary: "Added authentication"
change_type: "feature"
scope: "auth"
commit_hash: "abc123"
```

But a real checkpoint (commit) might contain:
- Auth feature (JWT implementation)
- Database fix (connection pooling)
- Documentation update (API docs)
- Performance improvement (query optimization)

Your schema forces you to choose **one** and ignore the rest, or create multiple commits for a single logical checkpoint. Neither is right.

---

## Revised Schema: Changes Array Per Checkpoint

```yaml
---
schema_version: "1"
timestamp: "2025-10-22T16:00:00Z"
commit_hash: "abc123def456"
changes:
  - summary: "Implemented JWT authentication"
    details: "Added token generation and validation middleware"
    change_type: "feature"
    scope: "auth"
  - summary: "Fixed connection pool exhaustion"
    details: "Resolved issue where stale connections weren't cleaned up"
    change_type: "fix"
    scope: "database"
  - summary: "Updated authentication guide"
    change_type: "docs"
    scope: "auth"
  - summary: "Optimized token verification queries"
    details: "Added caching layer for token lookups"
    change_type: "perf"
    scope: "auth"
---
schema_version: "1"
timestamp: "2025-10-22T15:00:00Z"
commit_hash: "xyz789uvw012"
changes:
  - summary: "Refactored query builder"
    change_type: "refactor"
    scope: "database"
```

**This is much better because:**

- ✅ Single checkpoint = single commit = single document (atomic)
- ✅ Captures all work done in that commit
- ✅ Multiple scopes and types naturally expressed
- ✅ Per-document append-only is **genuinely safe**
- ✅ If corruption happens at end of file, you lose only that checkpoint, not everything before it

---

## TOML Version

```toml
schema_version = "1"
timestamp = "2025-10-22T16:00:00Z"
commit_hash = "abc123def456"

[[changes]]
summary = "Implemented JWT authentication"
details = "Added token generation and validation middleware"
change_type = "feature"
scope = "auth"

[[changes]]
summary = "Fixed connection pool exhaustion"
details = "Resolved issue where stale connections weren't cleaned up"
change_type = "fix"
scope = "database"

[[changes]]
summary = "Updated authentication guide"
change_type = "docs"
scope = "auth"

[[changes]]
summary = "Optimized token verification queries"
details = "Added caching layer for token lookups"
change_type = "perf"
scope = "auth"

---

schema_version = "1"
timestamp = "2025-10-22T15:00:00Z"
commit_hash = "xyz789uvw012"

[[changes]]
summary = "Refactored query builder"
change_type = "refactor"
scope = "database"
```

---

## Parsing & Processing

### Go Tool Input Template (What LLM Sees)

```yaml
schema_version: "1"
timestamp: "2025-10-22T16:00:00Z"
commit_hash: ""  # tool fills this

git_status: |
  M  src/auth.go
  M  src/database.go
  M  docs/auth.md

diff_file: ".checkpoint-diff"

# List all changes made in this checkpoint
# Each can have different type and scope
changes:
  - summary: "[FILL IN: what changed]"
    details: "[OPTIONAL: longer description]"
    change_type: "[FILL IN: feature|fix|refactor|docs|perf|other]"
    scope: "[FILL IN: affected component]"
  - summary: "[FILL IN: another change]"
    change_type: "[FILL IN]"
    scope: "[FILL IN]"
  # Add more changes as needed
```

### Go Tool Processing

```go
type CheckpointEntry struct {
    SchemaVersion string `yaml:"schema_version"`
    Timestamp     string `yaml:"timestamp"`
    CommitHash    string `yaml:"commit_hash"`
    Changes       []Change `yaml:"changes"`
}

type Change struct {
    Summary    string `yaml:"summary"`
    Details    string `yaml:"details,omitempty"`
    ChangeType string `yaml:"change_type"`
    Scope      string `yaml:"scope,omitempty"`
}
```

---

## Mac App Processing

**Parsing per-document (unchanged from your intent):**

```swift
// Read file from end backwards
// For each YAML document (split on ---):
// Parse into CheckpointEntry with array of Changes
// Process all changes from that checkpoint
```

**Summarization per checkpoint:**

For a checkpoint with multiple changes:

```markdown
# {Project} - {Date}

## Checkpoint {CommitHash}

### feature: auth - Implemented JWT authentication
- Added token generation and validation middleware

### fix: database - Fixed connection pool exhaustion
- Resolved issue where stale connections weren't cleaned up

### docs: auth - Updated authentication guide

### perf: auth - Optimized token verification queries
- Added caching layer for token lookups

---
```

Or with Apple Intelligence:

```markdown
# {Project} - {Date}

## Summary
In this update, we implemented JWT-based authentication, fixed a database connection pooling issue, and optimized related queries. Documentation was also updated.

## Changes

### feature: auth - Implemented JWT authentication
...

### fix: database - Fixed connection pool exhaustion
...

[etc]
```

---

## Why This Solves Your Problem

| Aspect | Old Schema | New Schema |
|--------|-----------|-----------|
| Multi-change checkpoint | Forced to split or ignore | Natural array |
| Line length | Fine | Still fine |
| Per-document safety | Works but feels wrong | **Correct**: one commit = one document = atomic |
| Append-only | Awkward with multi-change | **Natural**: append whole checkpoint |
| Granularity | Too fine (one change per doc) | **Right** (one checkpoint per doc) |
| Parsing complexity | String split + loop | String split + loop (same) |
| Data semantics | "List of entries" | "List of checkpoints, each with multiple changes" |

---
**Recommendation:** Build for Sequoia 15.1 as minimum, with graceful fallback for older macOS. By the time users upgrade past Sequoia, any new Tahoe features will be automatically available through the system APIs.


_References_:
[^1]: [Use Apple Intelligence on your Mac - Apple Support (QA)](https://support.apple.com/en-qa/guide/mac-help/mchl46361784/mac)
[^2]: [How to get Apple Intelligence - Apple Support (HK)](https://support.apple.com/en-hk/121115)
[^3]: [Summarize notifications and reduce interruptions with Apple ...](https://support.apple.com/guide/mac-help/summarize-notifications-reduce-interruptions-mchldf5e4cb6/mac)
[^4]: [macOS Tahoe vs Sequoia: What’s Actually Different?](https://www.macobserver.com/wwdc/macos-tahoe-vs-sequoia-whats-different/)
[^5]: [How to get Apple Intelligence - Apple Support](https://support.apple.com/en-us/121115)
[^6]: [Use Apple Intelligence in Messages on Mac](https://support.apple.com/guide/mac-help/use-apple-intelligence-in-messages-mchl00c5b2c3/mac)
[^7]: [Use Apple Intelligence on your Mac](https://support.apple.com/guide/mac-help/intro-to-apple-intelligence-mchl46361784/mac)
[^8]: [Get web page summaries with Apple Intelligence on Mac](https://support.apple.com/gu-in/guide/mac-help/mchl62d5873e/15.0/mac/15.0)
