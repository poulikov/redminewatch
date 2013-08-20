package main

import (
    "fmt"
    "log"
    "time"
    "bytes"
    "os"
    "os/signal"
    "redminewatch/redmine"
    dbus "launchpad.net/~jamesh/go-dbus/trunk"
)

func main() {
    in := make(chan int, 1)
    out := make(chan int, 1)

    Notify("Redmine Watcher Start")

    go loopCheck(in, out)

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, os.Kill)

    <-c

    in <- 1
    // do something you need to do before the program will has finished
    <-out
}

func CheckNewTasks() {
    tasks, err := redmine.LoadTasks()
    if err != nil {
        Notify(err.Error())
        return
    }
    log.Println("Loaded new tasks")

    now := time.Now()
    yesterday := now.Add(-10 * time.Minute)
    older := tasks.OlderThan(&yesterday)

    var message bytes.Buffer
    for indx := range older {
        message.WriteString(fmt.Sprintf(
            "%s %s\n",
            older[indx].Subject,
            older[indx].LastUpdate().Format("2006-01-02 15:04")));

    }

    if len(older) > 0 {
        Notify(message.String())
    }
}

func loopCheck(in, out chan int) {
    CheckNewTasks()
    stop := false
    for !stop {
        select {
        case <- in:
            stop = true
        case <-time.After(10 * time.Second):
            CheckNewTasks()
        }
    }

    out <- 1
}

func Notify(msg string) {
    conn, err := dbus.Connect(dbus.SessionBus)
    if err != nil {
        log.Println("Can't connect to DBus")
        return
    }

    obj := conn.Object("org.freedesktop.Notifications", dbus.ObjectPath("/org/freedesktop/Notifications"))

    var id_num_to_repl uint32
    var actions_list []string
    var hint map[string]dbus.Variant
    var tmlimit int32 = -1
    reply, err := obj.Call(
        "org.freedesktop.Notifications",
        "Notify",
        "RedMine Client",
        id_num_to_repl,
        "",
        "RedMine Task Update",
        msg,
        actions_list,
        hint,
        tmlimit,
    )

    if err != nil {
        panic(err)
    }

    var notification_id uint32
    if err := reply.Args(&notification_id); err != nil {
        panic(err)
    }
}
