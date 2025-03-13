package checker

type Pinger interface {
    Ping(host string, count int, timeout int) (string, error)
}
