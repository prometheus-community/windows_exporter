package rotate

// Rotating things
import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FilePerm defines the permissions that Writer will use for all
// the files it creates (0644 = owner can read/write, others can read)
const (
	FilePerm   = os.FileMode(0644)
	DateFormat = "2006-01-02" // Go uses this specific date format as template
)

// FileWriter implements the io.Writer interface and writes to the
// filename specified.
// Will rotate at the specified interval and/or when the current file size exceeds maxSizeInBytes
// At rotation time, current file is renamed and a new file is created.
// If the number of archives exceeds maxArchives, older files are deleted.
type FileWriter struct {
	filename                 string        // The current log file path (e.g., "/var/log/app.log")
	filenameRotationTemplate string        // Template for rotated files (e.g., "/var/log/app-%s-%s.log")
	current                  *os.File      // The currently open file handle
	interval                 time.Duration // How often to rotate (e.g., 24h for daily)
	maxSizeInBytes           int64         // Maximum file size before rotation (e.g., 10MB)
	maxArchives              int           // How many old log files to keep (-1 = keep all)
	expireTime               time.Time     // When the next time-based rotation should happen
	bytesWritten             int64         // Counter of bytes written to current file
	sync.Mutex                             // Protects against concurrent writes
}

// NewFileWriter creates a new rotating file writer.
// Parameters:
//   - filename: path to the log file (e.g., "app.log")
//   - interval: rotation interval (e.g., 24*time.Hour for daily rotation, 0 to disable)
//   - maxSizeInBytes: max file size before rotation (e.g., 10*1024*1024 for 10MB, 0 to disable)
//   - maxArchives: number of old files to keep (-1 = keep all, 0 = keep none)
func NewFileWriter(filename string, interval time.Duration, maxSizeInBytes int64, maxArchives int) (*FileWriter, error) {
	// Create the template for rotated filenames
	// If filename is "app.log", template becomes "app-%s-%s.log"
	// First %s will be date (2006-01-02), second %s will be unix timestamp
	ext := filepath.Ext(filename)                       // e.g., ".log"
	prefix := strings.TrimSuffix(filename, ext)         // e.g., "app"
	filenameRotationTemplate := prefix + "-%s-%s" + ext // e.g., "app-%s-%s.log"

	w := &FileWriter{
		filename:                 filename,
		filenameRotationTemplate: filenameRotationTemplate,
		interval:                 interval,
		maxSizeInBytes:           maxSizeInBytes,
		maxArchives:              maxArchives,
	}

	// Calculate when the first rotation should happen (if time-based rotation is enabled)
	if interval > 0 {
		w.expireTime = time.Now().Add(interval)
	}

	// Open the log file for the first time
	if err := w.openCurrent(); err != nil {
		return nil, err
	}

	return w, nil
}

// Write implements io.Writer interface. This is called whenever something writes to the log.
// The io.Writer interface requires a method: Write(p []byte) (n int, err error)
func (w *FileWriter) Write(p []byte) (n int, err error) {
	// Lock the mutex to prevent concurrent writes from corrupting the file
	w.Lock()
	defer w.Unlock() // defer means "run this when the function exits"

	// Check if we need to rotate before writing
	if err := w.rotateIfNeeded(); err != nil {
		return 0, err
	}

	// Write the actual data to the file
	n, err = w.current.Write(p)

	// Update our byte counter
	w.bytesWritten += int64(n)

	return n, err
}

// openCurrent opens (or creates) the current log file and resets counters.
func (w *FileWriter) openCurrent() error {
	// Check if the file already exists to get its current size
	info, err := os.Stat(w.filename)
	if err != nil && !os.IsNotExist(err) {
		// An error occurred that's not "file doesn't exist"
		return fmt.Errorf("failed to stat log file: %w", err)
	}

	// Open or create the file
	// os.O_WRONLY = write-only
	// os.O_CREATE = create if doesn't exist
	// os.O_APPEND = append to end of file (don't overwrite)
	file, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, FilePerm)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	w.current = file

	// If the file existed, set bytesWritten to its current size
	// Otherwise, bytesWritten stays at 0
	if info != nil {
		w.bytesWritten = info.Size()
	} else {
		w.bytesWritten = 0
	}

	return nil
}

// rotateIfNeeded checks if rotation is needed and performs it.
func (w *FileWriter) rotateIfNeeded() error {
	// Check if we need to rotate based on:
	// 1. Time: interval is set AND current time is after expireTime
	// 2. Size: maxSizeInBytes is set AND we've written more than the limit
	if (w.interval > 0 && time.Now().After(w.expireTime)) ||
		(w.maxSizeInBytes > 0 && w.bytesWritten >= w.maxSizeInBytes) {

		if err := w.rotate(); err != nil {
			// Ignore rotation errors and keep the log open
			// This prevents losing logs if rotation fails
			fmt.Printf("unable to rotate the file %q, %s", w.filename, err.Error())
		}

		// Open a fresh log file
		return w.openCurrent()
	}
	return nil
}

// rotate performs the actual rotation: closes current file, renames it, and cleans up old files.
func (w *FileWriter) rotate() (err error) {
	// Close the current log file
	if err := w.current.Close(); err != nil {
		return err
	}

	// Create a unique filename for the rotated file
	// Use year-month-date for readability, unix time to make the file name unique with second precision
	now := time.Now()
	rotatedFilename := fmt.Sprintf(w.filenameRotationTemplate, now.Format(DateFormat), strconv.FormatInt(now.Unix(), 10))
	// Example: "app.log" becomes "app-2026-01-08-1704672000.log"

	if err := os.Rename(w.filename, rotatedFilename); err != nil {
		return err
	}

	// Update the next rotation time if time-based rotation is enabled
	if w.interval > 0 {
		w.expireTime = time.Now().Add(w.interval)
	}

	// Clean up old rotated files if we have too many
	return w.purgeArchivesIfNeeded()
}

// purgeArchivesIfNeeded deletes old log files if we exceed maxArchives.
func (w *FileWriter) purgeArchivesIfNeeded() (err error) {
	if w.maxArchives == -1 {
		// Skip archiving - keep all files
		return nil
	}

	// Find all rotated log files matching our pattern
	// filepath.Glob finds files matching a pattern (* = wildcard)
	var matches []string
	if matches, err = filepath.Glob(fmt.Sprintf(w.filenameRotationTemplate, "*", "*")); err != nil {
		return err
	}

	// If there are more archives than the configured maximum, delete the oldest ones
	if len(matches) > w.maxArchives {
		// Sort files alphanumerically - older dates/timestamps come first
		sort.Strings(matches)

		// Delete the oldest files
		// matches[:len(matches)-w.maxArchives] gets the files we want to delete
		// For example: if we have 10 files and maxArchives=5, delete files 0-4 (keep 5-9)
		for _, filename := range matches[:len(matches)-w.maxArchives] {
			if err := os.Remove(filename); err != nil {
				return err
			}
		}
	}
	return nil
}

// Close closes the current log file. Should be called when shutting down.
func (w *FileWriter) Close() error {
	w.Lock()
	defer w.Unlock()

	if w.current != nil {
		return w.current.Close()
	}
	return nil
}
