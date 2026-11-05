package main

import (
  "bytes"
  "encoding/json"
  "errors"
  "fmt"
  "net/http"
  "time"
)

func (s *PopupService) Ready() error {s.place();return s.request("/ready",nil)}
func (s *PopupService) Action(action, selection string) error {
  if action!="open_location" && action!="open_file" && action!="confirm" {
    return errors.New("unsupported action")
  }
  valid:=false
  for _,c:=range s.options.Choices {if c.ID==selection {valid=true;break}}
  if !valid{return errors.New("unsupported selection")}
  return s.request("/action",map[string]string{"action":action,"selection":selection})
}
func (s *PopupService) request(endpoint string,data any) error {
  encoded,err:=json.Marshal(data)
  if err!=nil{return err}
  req,err:=http.NewRequest(http.MethodPost,"http://"+s.callback+endpoint,bytes.NewReader(encoded))
  if err!=nil{return err}
  req.Header.Set("X-Organizer-Key",s.authKey)
  client:=&http.Client{Timeout:3*time.Second}
  response,err:=client.Do(req)
  if err!=nil{return err}
  defer response.Body.Close()
  if response.StatusCode!=http.StatusNoContent{return fmt.Errorf("notifier rejected request: %d",response.StatusCode)}
  return nil
}
