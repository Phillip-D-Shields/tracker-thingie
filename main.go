package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

type Task struct {
	ID          int64
	Title       string
	Priority    int
	DueDate     *time.Time
	Completed   bool
	CreatedAt   time.Time
	CompletedAt *time.Time
}

var (
	green   = color.New(color.FgGreen).SprintFunc()
	red     = color.New(color.FgRed).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	blue    = color.New(color.FgCyan).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
)

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var rootCmd = &cobra.Command{
		Use:   "tasks",
		Short: "A simple task tracker",
	}

	var addCmd = &cobra.Command{
		Use:   "add [task title]",
		Short: "Add a new task",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			priority, _ := cmd.Flags().GetInt("priority")
			title := args[0]
			if err := addTask(db, title, priority); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s: %s (Priority: %s)\n",
				green("Added task"),
				blue(title),
				yellow(fmt.Sprintf("%d", priority)))
		},
	}
	addCmd.Flags().IntP("priority", "p", 1, "Task priority (1-5)")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Run: func(cmd *cobra.Command, args []string) {
			tasks, err := listTasks(db)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(magenta("\nTask List:"))
			fmt.Println(strings.Repeat("=", 50))

			for _, task := range tasks {
				status := red("[ ]")
				if task.Completed {
					status = green("[✓]")
				}
				priorityColor := getPriorityColor(task.Priority)
				fmt.Printf("%s %d. %s (Priority: %s)\n",
					status,
					task.ID,
					blue(task.Title),
					priorityColor(fmt.Sprintf("%d", task.Priority)))
			}
			fmt.Println(strings.Repeat("=", 50))
		},
	}

	var statsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Show task statistics and charts",
		Run: func(cmd *cobra.Command, args []string) {
			generateStats(db)
		},
	}

	var completeCmd = &cobra.Command{
		Use:   "complete [task ID]",
		Short: "Mark a task as completed",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			if err := completeTask(db, id); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s task %s\n",
				green("Completed"),
				blue(id))
		},
	}

	rootCmd.AddCommand(addCmd, listCmd, completeCmd, statsCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func getPriorityColor(priority int) func(a ...interface{}) string {
	switch priority {
	case 1:
		return green
	case 2:
		return blue
	case 3:
		return yellow
	case 4, 5:
		return red
	default:
		return blue
	}
}

func generateStats(db *sql.DB) error {
	// Get statistics from database
	var totalTasks, completedTasks int
	var priorityStats []struct {
		Priority int
		Count    int
	}

	err := db.QueryRow(`
        SELECT COUNT(*), SUM(CASE WHEN completed THEN 1 ELSE 0 END)
        FROM tasks
    `).Scan(&totalTasks, &completedTasks)
	if err != nil {
		return err
	}

	rows, err := db.Query(`
        SELECT priority, COUNT(*) 
        FROM tasks 
        GROUP BY priority 
        ORDER BY priority
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var stat struct {
			Priority int
			Count    int
		}
		if err := rows.Scan(&stat.Priority, &stat.Count); err != nil {
			return err
		}
		priorityStats = append(priorityStats, stat)
	}

	// Create charts
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Tasks by Priority",
		}),
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
		}),
	)

	// Prepare data for charts
	priorities := make([]string, 0)
	counts := make([]opts.BarData, 0)
	for _, stat := range priorityStats {
		priorities = append(priorities, fmt.Sprintf("Priority %d", stat.Priority))
		counts = append(counts, opts.BarData{Value: stat.Count})
	}

	bar.SetXAxis(priorities).AddSeries("Tasks", counts)

	// Create pie chart for completion status
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Task Completion Status",
		}),
	)

	completionData := []opts.PieData{
		{Name: "Completed", Value: completedTasks},
		{Name: "Pending", Value: totalTasks - completedTasks},
	}
	pie.AddSeries("Completion", completionData)

	// Save charts to HTML files
	f1, _ := os.Create("task_priority.html")
	bar.Render(f1)
	f2, _ := os.Create("task_completion.html")
	pie.Render(f2)

	// Print summary statistics
	fmt.Println(magenta("\nTask Statistics:"))
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total Tasks: %s\n", blue(fmt.Sprintf("%d", totalTasks)))
	fmt.Printf("Completed Tasks: %s\n", green(fmt.Sprintf("%d", completedTasks)))
	fmt.Printf("Completion Rate: %s\n",
		yellow(fmt.Sprintf("%.1f%%", float64(completedTasks)/float64(totalTasks)*100)))
	fmt.Println("\nCharts have been generated:")
	fmt.Println(green("- task_priority.html"))
	fmt.Println(green("- task_completion.html"))

	return nil
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            priority INTEGER DEFAULT 1,
            due_date DATETIME,
            completed BOOLEAN DEFAULT FALSE,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            completed_at DATETIME
        )
    `)
	return db, err
}

func addTask(db *sql.DB, title string, priority int) error {
	_, err := db.Exec(`
        INSERT INTO tasks (title, priority) 
        VALUES (?, ?)
    `, title, priority)
	return err
}

func listTasks(db *sql.DB) ([]Task, error) {
	rows, err := db.Query(`
        SELECT id, title, priority, completed 
        FROM tasks 
        ORDER BY priority DESC, created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Priority, &t.Completed); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func completeTask(db *sql.DB, id string) error {
	_, err := db.Exec(`
        UPDATE tasks 
        SET completed = TRUE, completed_at = CURRENT_TIMESTAMP 
        WHERE id = ?
    `, id)
	return err
}
