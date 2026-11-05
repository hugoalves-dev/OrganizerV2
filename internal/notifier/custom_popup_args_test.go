package notifier

import (
    "encoding/json"
    "flag"
    "path/filepath"
    "testing"

    "github.com/vitorhugo-java/organizerv2/internal/config"
)

// Regression for #19: passing "--drag-enabled", "true" stops the Go flag parser
// before "--choices", leaving the popup with an empty destination list.
func TestPopupCommandArgsParsedByHelper(t *testing.T) {
    root:=t.TempDir()
    file:=filepath.Join(root,"Executables","Maestri-Setup (2) 日本語.exe")
    destination:=filepath.Join(root,"Desktop")
    event:=notificationEvent{ID:"test-id",CurrentPath:file,Category:"Executables"}
    for _,tc:=range []struct{
        name string
        drag,openFile,openLocation,confirm,moveTo bool
        wantChoices int
    }{
        {"all actions enabled",true,true,true,true,true,2},
        {"disabled actions",false,false,false,false,false,1},
        {"mixed actions",false,true,false,true,true,2},
    } {
        t.Run(tc.name,func(t *testing.T){
            cfg:=config.NotificationConfig{
                Custom:config.CustomNotificationConfig{
                    Position:"bottom_right",
                    DurationSeconds:15,
                    DragAndDrop:config.CustomDragAndDropConfig{Enabled:tc.drag},
                },
                Actions:config.NotificationActions{
                    OpenFile:tc.openFile,
                    OpenLocation:tc.openLocation,
                    Confirm:tc.confirm,
                    MoveTo:tc.moveTo,
                },
            }
            shortcuts:=[]resolvedShortcut{{ID:"desktop-id",Name:"Desktop",Path:destination}}
            args,err:=popupCommandArgs(event,cfg,shortcuts,"127.0.0.1:12345","secret")
            if err!=nil{t.Fatal(err)}

            // Match organizer-popup's actual flag names, types and defaults.
            fs:=flag.NewFlagSet("organizer-popup",flag.ContinueOnError)
            parsedFile:=fs.String("file","","")
            category:=fs.String("category","","")
            duration:=fs.Int("duration",15,"")
            eventID:=fs.String("event","","")
            callback:=fs.String("callback","","")
            token:=fs.String("token","","")
            choicesRaw:=fs.String("choices","[]","")
            position:=fs.String("position","bottom_right","")
            drag:=fs.Bool("drag-enabled",true,"")
            openFile:=fs.Bool("open-file",true,"")
            openLocation:=fs.Bool("open-location",true,"")
            confirm:=fs.Bool("confirm",true,"")
            if err:=fs.Parse(args);err!=nil{t.Fatalf("failed to parse helper arguments: %v",err)}
            if fs.NArg()!=0{t.Fatalf("unexpected positional arguments: %q",fs.Args())}
            if *parsedFile!=file || *category!="Executables" || *duration!=15 ||
                *eventID!="test-id" || *callback!="127.0.0.1:12345" || *token!="secret" ||
                *position!="bottom_right" {
                t.Fatalf("lost required CLI flags: file=%q category=%q duration=%d event=%q callback=%q position=%q",
                    *parsedFile,*category,*duration,*eventID,*callback,*position)
            }
            if *drag!=tc.drag || *openFile!=tc.openFile ||
                *openLocation!=tc.openLocation || *confirm!=tc.confirm {
                t.Fatalf("booleans parsed incorrectly: drag=%t file=%t location=%t confirm=%t",
                    *drag,*openFile,*openLocation,*confirm)
            }
            var choices []struct{ ID string `json:"id"`; Name string `json:"name"` }
            if err:=json.Unmarshal([]byte(*choicesRaw),&choices);err!=nil{t.Fatalf("invalid choices: %v",err)}
            if len(choices)!=tc.wantChoices{t.Fatalf("got %d choices, want %d: %+v",len(choices),tc.wantChoices,choices)}
            if choices[0].ID!=currentDestinationSelectionID{t.Fatalf("current destination missing: %+v",choices)}
            if tc.wantChoices==2 && choices[1].ID!="desktop-id"{
                t.Fatalf("configured destination missing: %+v",choices)
            }
        })
    }
}
