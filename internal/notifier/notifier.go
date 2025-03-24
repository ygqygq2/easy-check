package notifier

type Notifier interface {
    SendNotification(host, description string) error
}
