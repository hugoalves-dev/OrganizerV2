package notifier

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestOptionalPopupLicenseNoticesAreComplete(t *testing.T) {
    root:=filepath.Join("..","..")
    checks:=map[string][]string{
        "LICENSE":{"MIT License"},
        "LICENSES/MIT.txt":{"MIT License"},
        "LICENSES/AGPL-3.0.txt":{"GNU AFFERO GENERAL PUBLIC LICENSE"},
        "LICENSES/AGPL-NOTICE.txt":{"Copyright (c) 2026 Eaxum","LICENSE.drag-rs"},
        "LICENSES/LICENSE.drag-rs":{"MIT License","CrabNebula Ltd."},
        "THIRD_PARTY_NOTICES.md":{"wails-dragout","LICENSES/LICENSE.drag-rs"},
    }
    for path,needles:=range checks{
        t.Run(path,func(t *testing.T){
            data,err:=os.ReadFile(filepath.Join(root,filepath.FromSlash(path)))
            if err!=nil{t.Fatal(err)}
            for _,needle:=range needles{
                if !strings.Contains(string(data),needle){t.Fatalf("%s missing required notice %q",path,needle)}
            }
        })
    }
}
