# Task Tracker CLI 

A simple command-line task tracker built in Go that helps you manage your tasks with priority levels and visualize your progress with charts.

## Features

- ✨ Add tasks with priority levels (1-5)
- ✅ Mark tasks as complete
- 📋 List all tasks with color-coded status and priority
- 📊 Generate visual statistics and charts
- 💾 Persistent storage using SQLite
- 🎨 Color-coded terminal output

## Installation

### Prerequisites

- Go 1.23.3 or higher
- SQLite


### Building from source

```bash
# Clone the repository
git clone https://github.com/Phillip-D-Shields/tracker-thingie
cd tracker-thingie

# Install dependencies
go mod tidy

# Build the binary
go build -o tasks
```

## Usage

### Add a new task

```bash
# Add a task with priority (1-5)
./tasks add "Complete project documentation" -p 3
./tasks add "Urgent meeting prep" -p 5
```

### List tasks

```bash
./tasks list
```

### Complete a task

```bash
# Replace <id> with the task ID from the list command
./tasks complete <id>
```

### View statistics

```bash
./tasks stats
```
This will generate two HTML files:
- `task_priority.html`: Bar chart showing tasks by priority
- `task_completion.html`: Pie chart showing completion status

### Help

```bash
./tasks --help
```

## Task Priority Levels

- Priority 1 (🟢): Low priority
- Priority 2 (🔵): Normal priority
- Priority 3 (🟡): Medium priority
- Priority 4 (🔴): High priority
- Priority 5 (🔴): Critical priority

## Database

The app uses SQLite for storage. The database file (`tasks.db`) is created automatically in the current directory when you add your first task.

## Charts

After running the `stats` command, open the generated HTML files in your web browser to view:
- Task distribution by priority
- Completion status
- Visual statistics

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details
