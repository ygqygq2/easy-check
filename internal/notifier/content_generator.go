package notifier

type ContentGenerator interface {
    GenerateContent(alerts []*AlertItem) (string, error)
}
