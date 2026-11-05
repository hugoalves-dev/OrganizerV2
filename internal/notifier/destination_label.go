package notifier

import "path/filepath"

func currentDestinationLabel(directory string) string {
    name := filepath.Base(filepath.Clean(directory))
    if name == "" || name == "." {
        name = filepath.Clean(directory)
    }
    return name + " (current)"
}
