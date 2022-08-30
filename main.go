package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hidez8891/shm"
	"github.com/ncruces/zenity"
)

const IS_ACTIVE_TAG = "ACTIVE_MONITOR_ADJ"
const BRIGHTNESS_VAL = "BRIGHTNESS_MONITOR_ADJ"
var empty_buf = make([]byte,256)
func setProgress(val int) error{
    r, err := shm.Open(IS_ACTIVE_TAG, 256)
    defer r.Close()
    w, err := shm.Open(BRIGHTNESS_VAL, 256)
    defer w.Close()

    if err != nil {
        log.Fatal(err)
    }
    rbuf := make ([]byte, 256)
    r.ReadAt(rbuf, 0)
    isactive := strings.Trim(string(rbuf),"\x00")

    log.Println("isactive is","or not", isactive)
    log.Println("len", len(isactive))

    if strings.Compare(isactive, "active") == 0{
        log.Println("IT'S ACTIVE")
        if err != nil {
            log.Fatal(err)
        }
        w.WriteAt(empty_buf,0)
        w.WriteAt([]byte(fmt.Sprintf("%d:%d:",time.Now().UnixMilli(), val)),0)
        log.Println("value written:",fmt.Sprintf("%d", val))
        //time.Sleep(100*time.Millisecond)
        //
    }else {
        log.Println("NOT ACTIVE")
        r.WriteAt([]byte("active"),0)

        cb, err := getCurrentBrightness()
        if err != nil{
            return err
        }
        cb = setBrightness(cb, val)

        dlg, err := zenity.Progress( zenity.Title("Monitor Settings"),  zenity.NoCancel() )
        if err != nil {
            return err
        }
        defer dlg.Close()
        log.Println("value is ", val)

        dlg.Text(fmt.Sprintf("Brightness: %d%%", cb))
        dlg.Value(cb)

        oldval:=""


        for i:=0;i<100;i++ {
            wbuf:= make([]byte, 256)
            w.ReadAt(wbuf, 0)
            bval := strings.Trim(string(wbuf), "\x00")
            parts := strings.Split(bval, ":")

            if len(parts) == 3 && oldval != parts[0]{

                if v, err := strconv.Atoi(parts[1]); err == nil{
                    i = 0
                    log.Printf("v is %d\n", v)
                    cb = setBrightness(cb, v)
                    dlg.Text(fmt.Sprintf("Brightness: %d%%", cb))
                    dlg.Value(cb)
                }
            }
            time.Sleep(10 * time.Millisecond)
            oldval = parts[0]
        }


        dlg.Complete()
        r.WriteAt(empty_buf,0)
    }

    return nil


}

func getCurrentBrightness() (int, error){
    m:=regexp.MustCompile("VCP 10 (\\w*)")

    cmd_instance :=exec.Command(exePath, "detect")
    cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

    monitors, _ := cmd_instance.Output()
    fmt.Println(string(monitors))

    //if strings.Index(string(monitors),"Dell") == -1 {
        //return 0, errors.New("monitor not found")
    //}

    cmd_instance = exec.Command(exePath, "getvcp", "0", "10")
    cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
    out, _ := cmd_instance.Output()



    matches := m.FindStringSubmatch(string(out))
    if len(matches) > 1 {
        fmt.Println(matches[1])
        curBrightness, _ := strconv.ParseInt(matches[1], 16, 64)
        return int(curBrightness), nil

    }
    return 0, errors.New("Cannot get brightness value")
}

func setBrightness(cb int, adj int) int{
        

    curBrightness := cb + adj


    if curBrightness > 100 {
        curBrightness = 100
    }else if curBrightness < 0 {
        curBrightness = 0
    }



    cmd_instance := exec.Command(exePath, "setvcp", "0", "10", fmt.Sprintf("%x", curBrightness))
    cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
    cmd_instance.Run()

    return curBrightness

}   

var exePath string
func main(){

    ex, err := os.Executable()
    if err!= nil {
        panic(err)
    }
    exePath = filepath.Join(filepath.Dir(ex), "winddcutil.exe")
    log.Println(exePath)

    intVar, err := strconv.Atoi(os.Args[1])
    if err != nil {
        return
    }
    setProgress(intVar)
}
